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
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
	"io"
)

// Maybe these should be more similar to HTTP response codes.
const (
	invalidRequest = "Invalid Request"
	success        = "Success"
)

type ApiRequest struct {
	ReqId   string // Request ID string will be returned unchanged with the response to this request
	Content map[string]string
}

type ApiResponse struct {
	ReqId   string // Request ID string from orginating ApiRequest
	Status  string // Status code
	Error   string // Error message if any
	Content map[string]string
}

type newTokenRequest struct {
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
	reqs    chan newTokenRequest
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

func (t *Tokenizer) newToken(req newTokenRequest) {
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
		req := newTokenRequest{
			fieldname: fieldname,
			text:      text,
			replyto:   replychan,
		}
		t.reqs <- req
		token = <-req.replyto
	}
	return token
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

// A fieldHandler tokenizes or detokenizes a fieldname/text pair
type apiHandler func(request *ApiRequest, response *ApiResponse)

func (t *Tokenizer) jsonHandler(ws *websocket.Conn, h apiHandler) {
	dec := json.NewDecoder(ws)
	enc := json.NewEncoder(ws)
	for {
		var request ApiRequest
		var response ApiResponse
		// Read one request from the socket and attempt to decode
		err := dec.Decode(&request)
		switch {
		case err == io.EOF:
			log.Panic(err)
			log.Println("Websocket disconnecting")
        	return
        case err != nil:
        	// Request could not be decoded - return error
			response = ApiResponse{Status: invalidRequest, Error: err.Error()}
			enc.Encode(&response)
        }
        // Call API Handler
        h(&request, &response)
        /*
		content := make(map[string]string)
		for fieldname, text := range request.Content {
			content[fieldname] = fh(fieldname, text)
		}
		response = ApiResponse{
			ReqId: request.ReqId,
			Status: success,
			Content: content,
		}
		*/
		enc.Encode(response)
	}
}

func (t *Tokenizer) tokenizeReq(request *ApiRequest, response *ApiResponse) {
	content := make(map[string]string)
	for fieldname, text := range request.Content {
		content[fieldname] = t.GetToken(fieldname, text)
	}
	response.ReqId = request.ReqId
	response.Status = success
	response.Content = content
}


func (t *Tokenizer) JsonTokenizer() wsHandler {
	// Is this really the right way to do this?
	// See:  http://groups.google.com/group/golang-nuts/browse_thread/thread/2d3c573a05f72d69?pli=1
	return func(ws *websocket.Conn) {
		t.jsonHandler(ws, func(rq *ApiRequest, rp *ApiResponse) { t.tokenizeReq(rq, rp) } )
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
		reqs:    make(chan newTokenRequest),
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
	handler := t.JsonTokenizer()
	log.Println("Starting websocket listener.\n")
	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/", websocket.Handler(handler))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
