// Copyright 2012 Jason McVetta.  This is Free Software, released under the 
// terms of the GNU Public License version 3.

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/bmizerany/pat"
	"github.com/jmcvetta/mgourl"
	"github.com/jmcvetta/tokenizer"
	"github.com/jmcvetta/tokenizerd/api/rest"
	"github.com/jmcvetta/tokenizerd/api/ws"
	"github.com/russross/blackfriday"
	"launchpad.net/mgo"
	"log"
	"net/http"
)

const (
	homeMarkdown = `# Tokenizerd
	
A data tokenization server

## REST API

### Tokenize

	/v1/rest/tokenize/{string}

Returns status code 200 and a token string, or status code 500 and an error
message.

### Detokenize

	/v1/rest/detokenize/{token}

Returns status code 200 and the original string; status code 404, indicating no
such token exists in the database; or status code 500 and an error message.
`
)

func HomePageHandler(w http.ResponseWriter, req *http.Request) {
	output := blackfriday.MarkdownCommon([]byte(homeMarkdown))
	w.Write(output)
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	//
	// Parse command line
	//
	listenUrl := flag.String("url", "localhost:3000", "Host/port on which to run websocket listener")
	mongoUrl := flag.String("mongo", "localhost", "URL of MongoDB server")
	flag.Parse()
	// Extract DB name from DB URL, if present
	dbName := "tokenizer" // If no DB name specified, use "tokenizer"
	switch _, auth, _, err := mgourl.ParseURL(*mongoUrl); true {
		case err != nil:
			log.Fatal("Could not parse MongoDB URL:", err)
		case auth.Db != "":
			dbName = auth.Db
	}
	//
	// Setup MongoDB connection
	//
	log.Println("Connecting to MongoDB on", *mongoUrl)
	session, err := mgo.Dial(*mongoUrl)
	if err != nil {
		log.Fatalln(err)
	}
	db := session.DB(dbName)
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
	mux.Get("/", http.HandlerFunc(HomePageHandler))
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
