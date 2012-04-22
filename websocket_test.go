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
	"encoding/json"
	"fmt"
	"github.com/jmcvetta/goutil"
	"log"
	"net/http"
	"testing"
)


func getWebsocket(t *testing.T) *websocket.Conn {
	origin := "http://localhost/"
	url := "ws://localhost:3500/v1/tokenize"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		msg := "Could not conect to websocket.  Is Gokenizer running?"
		log.Println(msg)
		t.Fatal(err)
	}
	return ws
}

func runServer() {
	//
	// Use a fake tokenizer since we are only interested in testing the API.
	//
	fake := FakeTokenizer{}
	tok := WsTokenize(fake)
	detok := WsDetokenize(fake)
	//
	// Start websocket listener
	//
	http.Handle("/v1/tokenize", websocket.Handler(tok))
	http.Handle("/v1/detokenize", websocket.Handler(detok))
	err := http.ListenAndServe(":3500", nil)
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

// Tests tokenization 
func TestWsTokenize(t *testing.T) {
	go runServer()
	var err error
	//
	// Prepare some random data
	//  
	reqid := goutil.RandAlphanumeric(8, 8)
	origData := make(map[string]string)
	for i := 0; i < 10; i++ {
		fieldname := goutil.RandAlphanumeric(8, 8)
		field := goutil.RandAlphanumeric(8, 8)
		origData[fieldname] = field
	}
	//
	// Setup API connection
	//
	ws := getWebsocket(t)
	dec := json.NewDecoder(ws)
	enc := json.NewEncoder(ws)
	//
	// Tokenize
	//
	req := JsonTokenizeRequest{
		ReqId: reqid,
		Data:  origData,
	}
	if err = enc.Encode(req); err != nil {
		t.Fatal(err)
	}
	var resp TokenizeReponse
	if err = dec.Decode(&resp); err != nil {
		t.Fatal(err)
	}
	// Since the FakeTokenizer returns the original string as the token string,
	// we can easily check whether the API is properly handling our request.
	for field, orig := range origData {
		token := resp.Data[field]
		if orig != token {
			msg := fmt.Sprintf("Tokenization failure: %s != %s", orig, token)
			t.Error(msg)
		}
	}
	//
	// Detokenize
	//
}
