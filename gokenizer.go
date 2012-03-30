/*
                                   Gokenizer
                             A Document Tokenizer


@author: Jason McVetta <jason.mcvetta@gmail.com>
@copyright: (c) 2012 Jason McVetta
@license: GPL v3 - http://www.gnu.org/copyleft/gpl.html

*/

package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// Maybe these should be more similar to HTTP response codes.
const (
	invalidRequest = "Invalid Request"
	success        = "Success"
)

var TokenNotFound = errors.New("Foobar")

type TokenizeRequest struct {
	ReqId string            // Request ID string will be returned unchanged with the response to this request
	Data  map[string]string // Maps fieldnames to text
}

type TokenizeReponse struct {
	ReqId  string            // Request ID string from orginating TokenizeRequest
	Status string            // Status code
	Error  string            // Error message if any
	Data   map[string]string // Maps fieldnames to token strings
}

type DetokenizeRequest struct {
	ReqId string // Request ID string will be returned unchanged with the response to this request
	Data  map[string]string
}

type foundToken struct {
	// Is it really pointful to return the token?
	Token string // The token we looked up
	Found bool   // Was the token found in the database?
	Text  string // The text it represents, if found
}

type DetokenizeReponse struct {
	ReqId  string                // Request ID string from orginating TokenizeRequest
	Status string                // Status code
	Error  string                // Error message if any
	Data   map[string]foundToken // Maps fieldnames to foundToken instances
}

type newTokenizeRequest struct {
	fieldname string      // Name of the field from which this text came
	text      string      // The original text
	replyto   chan string // Channel on which to return tokenized text
}

type tokenizedText struct {
	Fieldname string // Name of the field from which this text came
	Text      string // The original text
	Token     string // A token representing, but not programmatically derived from, the original text
}

type Tokenizer struct {
	session *mgo.Session
	reqs    chan newTokenizeRequest
}

func (t Tokenizer) run() {
	for {
		select {
		case req := <-t.reqs:
			t.newToken(req)
		}
	}
}

func (t *Tokenizer) tokenCollection() *mgo.Collection {
	// lightweight operation, involves no network communication
	col := t.session.DB("gokenizer").C("tokens")
	return col
}

func (t *Tokenizer) proposeToken() string {
	// Create a proposed token value, based on the current timestamp plus a
	// random integer.  This should *usually* produce unique tokens - however
	// there is no guarantee of this, so it is necessary to check that the 
	// token does not already exist.
	// Proposed token
	token_int := time.Now().Second()
	bigrand, _ := rand.Int(rand.Reader, big.NewInt(10000000))
	token_int += int(bigrand.Int64())
	token := strconv.Itoa(token_int)
	token = base64.StdEncoding.EncodeToString([]byte(token))
	return token
}

func (t *Tokenizer) newToken(req newTokenizeRequest) {
	var token string
	var count int
	var err error
	col := t.tokenCollection()
	for {
		token = t.proposeToken()
		// Make sure this token does not already exist
		count, err = col.Find(bson.M{
			"fieldname": req.fieldname,
			"token":     token,
		}).Count()
		if err != nil {
			panic(err)
		}
		if count > 0 {
			// token already exists, try again
			continue
		}
		break
	}
	// No one else is using this token, so let's save it to db
	err = col.Insert(&tokenizedText{req.fieldname, req.text, token})
	if err != nil {
		panic(err)
	}
	log.Println("New token: " + token)
	req.replyto <- token
	return
}

func (t *Tokenizer) GetToken(fieldname string, text string) string {
	log.Println("Get Token")
	log.Println("  Fieldname: " + fieldname)
	log.Println("  Text:      " + text)
	var token string
	col := t.tokenCollection()
	result := tokenizedText{}
	switch err := col.Find(bson.M{"fieldname": fieldname, "text": text}).One(&result); true {
	default:
		log.Panic(err)
	case nil == err:
		token = result.Token
		log.Println("Found existing token: " + token)
	case err == mgo.NotFound:
		log.Println("No existing token found.  Requesting a new token.")
		replychan := make(chan string)
		req := newTokenizeRequest{
			fieldname: fieldname,
			text:      text,
			replyto:   replychan,
		}
		t.reqs <- req
		token = <-req.replyto
	}
	return token
}

func (t *Tokenizer) GetText(fieldname string, token string) (string, error) {
	log.Println("Get Text")
	log.Println("  Fieldname: " + fieldname)
	log.Println("  Token:      " + token)
	var text string
	var err error
	col := t.tokenCollection()
	result := tokenizedText{}
	query := col.Find(bson.M{"fieldname": fieldname, "token": token})
	switch db_err := query.One(&result); true {
	case db_err == mgo.NotFound:
		log.Println("Token not found in DB")
		err = TokenNotFound
		return text, err
	case db_err != nil:
		log.Panic(err)
	}
	text = result.Text
	log.Println("Found text for token: " + text)
	return text, err
}

type wsHandler func(ws *websocket.Conn)

func (t *Tokenizer) EchoHandler() wsHandler {
	return func(ws *websocket.Conn) {
		log.Println("Starting websocket handler.")
		log.Println("    Local Address: " + ws.LocalAddr().String())
		log.Println("    Remote Address: " + ws.RemoteAddr().String())
		for {
			var message string
			var err error
			err = websocket.Message.Receive(ws, &message)
			switch err != nil {
			case err == io.EOF:
				log.Println("Websocket disconnecting")
				return
			default:
				log.Panic(err)
			}
			fieldname := ""
			token := t.GetToken(fieldname, message)
			err = websocket.Message.Send(ws, token)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}

func (t *Tokenizer) JsonTokenizer() wsHandler {
	return func(ws *websocket.Conn) {
		dec := json.NewDecoder(ws)
		enc := json.NewEncoder(ws)
		for {
			var request TokenizeRequest
			// Read one request from the socket and attempt to decode
			switch err := dec.Decode(&request); true {
			case err == io.EOF:
				log.Println("Websocket disconnecting")
				return
			case err != nil:
				// Request could not be decoded - return error
				response := TokenizeReponse{Status: invalidRequest, Error: err.Error()}
				enc.Encode(&response)
			}
			content := make(map[string]string)
			for fieldname, text := range request.Data {
				content[fieldname] = t.GetToken(fieldname, text)
			}
			response := TokenizeReponse{
				ReqId:  request.ReqId,
				Status: success,
				Data:   content,
			}
			enc.Encode(response)
		}
	}
}

func (t *Tokenizer) JsonDetokenizer() wsHandler {
	return func(ws *websocket.Conn) {
		dec := json.NewDecoder(ws)
		enc := json.NewEncoder(ws)
		for {
			var request DetokenizeRequest
			// Read one request from the socket and attempt to decode
			switch err := dec.Decode(&request); true {
			case err == io.EOF:
				log.Println("Websocket disconnecting")
				return
			case err != nil:
				// Request could not be decoded - return error
				response := DetokenizeReponse{Status: invalidRequest, Error: err.Error()}
				enc.Encode(&response)
				return
			}
			data := make(map[string]foundToken)
			for fieldname, token := range request.Data {
				ft := foundToken{
					Token: token,
				}
				text, err := t.GetText(fieldname, token)
				switch err != nil {
				case err == TokenNotFound:
					ft.Found = false
				default:
					log.Panic(err)
				}
				ft.Text = text
				ft.Found = true
				data[fieldname] = ft
			}
			response := DetokenizeReponse{
				ReqId:  request.ReqId,
				Status: success,
				Data:   data,
			}
			enc.Encode(response)
		}
	}
}
func NewTokenizer() Tokenizer {
	//
	// Setup database connection
	//
	log.Println("Connecting to MongoDB")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	//
	// Ensure DB uses indexes.  If indexes already exist, this is a noop.
	//
	col := session.DB("gokenizer").C("tokens")
	col.EnsureIndex(mgo.Index{
		Key:      []string{"fieldname", "text"},
		Unique:   true,
		DropDups: false,
		Sparse:   true,
	})
	col.EnsureIndex(mgo.Index{
		Key:      []string{"fieldname", "token"},
		Unique:   true,
		DropDups: false,
		Sparse:   true,
	})
	//
	// Initialize tokenizer
	//
	t := Tokenizer{
		session: session,
		reqs:    make(chan newTokenizeRequest),
	}
	go t.run()
	return t
}

func tokenEcho(t *Tokenizer, message string) string {
	fieldname := ""
	token := t.GetToken(fieldname, message)
	return token
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	t := NewTokenizer()
	//
	// Initialize websockets
	//
	echoHandler := t.EchoHandler()
	jtok := t.JsonTokenizer()
	jdetok := t.JsonDetokenizer()
	log.Println("Starting websocket listener.\n")
	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/v1/tokenize", websocket.Handler(jtok))
	http.Handle("/v1/detokenize", websocket.Handler(jdetok))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
