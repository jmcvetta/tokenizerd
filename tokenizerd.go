/*
                                  tokenizerd
                               A Data Tokenizer


@author: Jason McVetta <jason.mcvetta@gmail.com>
@copyright: (c) 2012 Jason McVetta
@license: GPL v3 - http://www.gnu.org/copyleft/gpl.html

********************************************************************************
Tokenizerd is free software: you can redistribute it and/or modify it under the
terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

Tokenizerd is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
Tokenizerd.  If not, see <http://www.gnu.org/licenses/>.
********************************************************************************

*/

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"launchpad.net/mgo"
	"log"
	"net/http"
	"github.com/jmcvetta/tokenizer"
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
	db := session.DB("gokenizer")
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
