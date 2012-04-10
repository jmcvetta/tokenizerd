/*
                                   Gokenizer
Test Suite


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

package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/jmcvetta/goutil"
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
		t.Fatal(err)
	}
	return ws
}

// Tests tokenization 
func TestRoundTrip(t *testing.T) {
	var err error
	//
	// Prepare some random data
	//  
	reqid := goutil.RandString(8, 128)
	origData := make(map[string]string)
	for i := 0; i < 10; i++ {
		fieldname := goutil.RandString(8, 128)
		field := goutil.RandString(8, 128)
		origData[fieldname] = field
	}
	//
	// Tokenize
	//
	req := TokenizeRequest{
		ReqId: reqid,
		Data:  origData,
	}
	t.Log("Tokenize request:", req)
	ws := getWebsocket(t)
	dec := json.NewDecoder(ws)
	if _, err = ws.Write([]byte(tokenizeReq)); err != nil {
		t.Fatal(err)
	}
	var resp TokenizeReponse
	if err = dec.Decode(&resp); err != nil {
		t.Fatal(err)
	}
}
