// Copyright 2012 Jason McVetta.  This is Free Software, released under the 
// terms of the GNU Public License version 3.

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/bmizerany/pat"
	"github.com/jmcvetta/tokenizer"
	"github.com/jmcvetta/tokenizerd/api/rest"
	"github.com/jmcvetta/tokenizerd/api/ws"
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
	// Setup MongoDB connection
	//
	log.Println("Connecting to MongoDB on", *mongoUrl)
	session, err := mgo.Dial(*mongoUrl)
	if err != nil {
		log.Fatalln(err)
	}
	db := session.DB("tokenizer")
	//
	// Initialize Tokenizer
	//
	t := tokenizer.NewMongoTokenizer(db)
	//
	// Register URLs
	//
	mux := pat.New()
	mux.Get("/v1/rest/tokenize/:string", rest.TokenizeHandler(t))
	mux.Get("/v1/rest/detokenize/:token", rest.DetokenizeHandler(t))
	mux.Get("/v1/ws/tokenize", websocket.Handler(ws.Tokenize(t)))
	mux.Get("/v1/ws/detokenize", websocket.Handler(ws.Detokenize(t)))
	http.Handle("/", mux)
	//
	// Start HTTP server
	//
	log.Println("Starting HTTP server on", *listenUrl)
	err = http.ListenAndServe(*listenUrl, nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}
