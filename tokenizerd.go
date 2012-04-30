// Copyright 2012 Jason McVetta.  This is Free Software, released under the 
// terms of the GNU Public License version 3.

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/jmcvetta/tokenizer"
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
	//
	// Initialize tokenizer
	//
	t := tokenizer.NewMongoTokenizer(db)
	tok := WsTokenize(t)
	detok := WsDetokenize(t)
	//
	// Start websocket listener
	//
	log.Println("Starting websocket listener on ", *listenUrl)
	http.Handle("/v1/tokenize", websocket.Handler(tok))
	http.Handle("/v1/detokenize", websocket.Handler(detok))
	// listenUrl := "heliotropi.cc:3000"
	err = http.ListenAndServe(*listenUrl, nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
