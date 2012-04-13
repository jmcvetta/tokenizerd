/*
                                   Gokenizer
                                  Test Suite

NOTE: Gokenizer application must be running in order to run tests.


@author: Jason McVetta <jason.mcvetta@gmail.com>
@copyright: (c) 2012 Jason McVetta
@license: GPL v3 - http://www.gnu.org/copyleft/gpl.html

********************************************************************************
This file is part of Gokenizer.

Gokenizer is free software: you can redistribute it and/or modify it under the
terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

Gokenizer is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
Gokenizer.  If not, see <http://www.gnu.org/licenses/>.
********************************************************************************

*/

package api

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/jmcvetta/goutil"
	"log"
	"testing"
)

var tokenizeReq = `
{
    "ReqId": "an arbitrary string identifying this request",
    "Data": {
        "fieldname1": "fieldvalue1",
        "field name 2": "field  value 2"
    }
}
`

func getWebsocket(t *testing.T) *websocket.Conn {
	origin := "http://localhost/"
	url := "ws://localhost:3000/v1/tokenize"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		msg := "Could not conect to websocket.  Is Gokenizer running?"
		log.Println(msg)
		t.Fatal(err)
	}
	return ws
}

func runServer() error {
	//
	// Use a fake tokenizer since we are only interested in testing the API.
	//
	fake := FakeTokenizer{}
	tok := HandlerTokenize(fake)
	detok := HandlerDetokenize(t)
	//
	// Start websocket listener
	//
	log.Println("Starting websocket listener on ", *listenUrl)
	http.Handle("/v1/tokenize", websocket.Handler(tok))
	http.Handle("/v1/detokenize", websocket.Handler(detok))
	// listenUrl := "heliotropi.cc:3000"
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: " + err.Error())
	}
}

// A simple-minded fake tokenizer for testing.  Original string and token are 
// always identical, so no storage or logic is required.
type FakeTokenizer struct{}

func (f FakeTokenizer) Tokenize(s string) string {
	return s
}

func (f FakeTokenizer) Detokenize(s string) (string, error) {
	return s, nil
}


func TestTokenizeApi(t *testing.T) {
}

func TestDetokenizeApi(t *testing.T) {
}

