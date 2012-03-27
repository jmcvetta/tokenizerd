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
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

type newTokenRequest struct {
	fieldname string      // Name of the field from which this text came
	text      string      // The original text
	replyto   chan string // Channel on which to return tokenized text
}

type tokenizedText struct {
	Text  string // The original text
	Token string // A token representing, but not programmatically derived from, the original text
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

func (t *Tokenizer) tokenCollection(fieldname string) *mgo.Collection {
	// lightweight operation, involves no network communication
	col := t.session.DB("tokens").C(fieldname)
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
	col := t.tokenCollection(req.fieldname)
	for {
		token = t.proposeToken()
		// Make sure this token does not already exist
		count, err = col.Find(bson.M{"token": token}).Count()
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
	err = col.Insert(&tokenizedText{req.text, token})
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
	col := t.tokenCollection(fieldname)
	result := tokenizedText{}
	err := col.Find(bson.M{"text": text}).One(&result)
	var token string
	switch {
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

func (t *Tokenizer) NewHandler() func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Starting websocket handler.")
		log.Println("    Local Address: " + ws.LocalAddr().String())
		log.Println("    Remote Address: " + ws.RemoteAddr().String())
		for {
			var text string
			recv_err := websocket.Message.Receive(ws, &text)
			switch {
			case recv_err != nil && recv_err.Error() == "EOF":
				log.Println("Websocket disconnecting")
				return
			case recv_err != nil:
				log.Fatalln("Error receiving message from websocket: " + recv_err.Error())
			}
			fieldname := "foobar" // TODO: fieldname support not yet implemented
			token := t.GetToken(fieldname, text)
			send_err := websocket.Message.Send(ws, token)
			if send_err != nil {
				log.Fatalln("Error sending message to websocket: " + send_err.Error())
			}
		}
	}
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	//
	// Setup database connection
	//
	log.Println("Connecting to MongoDB")
	session, err := mgo.Dial("localhost")
	session
	if err != nil {
		log.Panic(err)
	}
	//
	// Initialize tokenizer
	//
	t := Tokenizer{
		session: session,
		reqs:    make(chan newTokenRequest),
	}
	go t.run()
	//
	// Initialize websockets
	//
	handler := t.NewHandler()
	log.Println("Starting websocket listener.\n")
	http.Handle("/", websocket.Handler(handler))
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
