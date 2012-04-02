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
	"testing"
	"code.google.com/p/go.net/websocket"
	// "encoding/base64"
	// "encoding/json"
	// "errors"
	// "io"
	// "launchpad.net/mgo"
	// "launchpad.net/mgo/bson"
	// "log"
	// "net/http"
	// "strconv"
	// "time"
	// "flag"
)



func TestTokenize(t *testing.T) {
	origin := "http://localhost/"
	url := "ws://localhost:3000/v1/tokenize"
	client, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	println(client, err)
}