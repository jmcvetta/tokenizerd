// Copyright 2012 Jason McVetta.  This is Free Software, released under the 
// terms of the GNU Public License version 3.

package main

import (
	// "code.google.com/p/go.net/websocket"
	"flag"
	"github.com/jmcvetta/tokenizer"
	"github.com/jmcvetta/tokenizerd/api/restful"
	// "github.com/jmcvetta/tokenizerd/api/ws"
	"launchpad.net/mgo"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	//
	// Parse command line
	//
	listenUrl := flag.String("url", "localhost:3000", "Host/port on which to run websocket listener")
	mongoUrl := flag.String("mongo", "localhost", "URL of MongoDB server")
	flag.Parse()
	//
	// Setup database connection
	//
	log.Println("Connecting to MongoDB on", *mongoUrl)
	session, err := mgo.Dial(*mongoUrl)
	if err != nil {
		log.Fatalln(err)
	}
	db := session.DB("tokenizer")
	// Get a tokenizer
	t := tokenizer.NewMongoTokenizer(db)
	//
	// Register websocket handlers
	//
	// tok := ws.Tokenize(t)
	// detok := ws.Detokenize(t)
	// http.Handle("/v1/ws/tokenize", websocket.Handler(tok))
	// http.Handle("/v1/ws/detokenize", websocket.Handler(detok))
	//
	// RESTful Handler
	//
	http.Handle("/", restful.Router(t))
	//
	// Start listener
	//
	log.Println("Listening on ", *listenUrl)
	err = http.ListenAndServe(*listenUrl, nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
