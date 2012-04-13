/*
                                   Gokenizer
                               A Data Tokenizer
API


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
	"io"
	"log"
	"tokenizer"
)

// Maybe these should be more similar to HTTP response codes.
const (
	invalidRequest = "Invalid Request"
	success        = "Success"
)

type JsonTokenizeRequest struct {
	ReqId string            // Request ID string will be returned unchanged with the response to this request
	Data  map[string]string // Maps fieldnames to text
}

type TokenizeReponse struct {
	ReqId  string            // Request ID string from orginating JsonTokenizeRequest
	Status string            // Status code
	Error  string            // Error message if any
	Data   map[string]string // Maps fieldnames to token strings
}

type DetokenizeRequest struct {
	ReqId string // Request ID string will be returned unchanged with the response to this request
	Data  map[string]string
}

type foundToken struct {
	// Is it really pointful to return the token?
	Token string // The token we looked up
	Found bool   // Was the token found in the database?
	Text  string // The text it represents, if found
}

type DetokenizeReponse struct {
	ReqId  string                // Request ID string from orginating JsonTokenizeRequest
	Status string                // Status code
	Error  string                // Error message if any
	Data   map[string]foundToken // Maps fieldnames to foundToken instances
}

type wsHandler func(ws *websocket.Conn)

func WsTokenize(t tokenizer.Tokenizer) wsHandler {
	return func(ws *websocket.Conn) {
		log.Println("New websocket connection")
		log.Println("    Location:  ", ws.Config().Location)
		log.Println("    Origin:    ", ws.Config().Origin)
		log.Println("    Protocol:  ", ws.Config().Protocol)
		dec := json.NewDecoder(ws)
		enc := json.NewEncoder(ws)
		for {
			var request JsonTokenizeRequest
			// Read one request from the socket and attempt to decode
			switch err := dec.Decode(&request); true {
			case err == io.EOF:
				log.Println("Websocket disconnecting")
				return
			case err != nil:
				// Request could not be decoded - return error
				response := TokenizeReponse{Status: invalidRequest, Error: err.Error()}
				enc.Encode(&response)
				log.Println("Invalid request - websocket disconnecting")
				return
			}
			data := make(map[string]string)
			for fieldname, text := range request.Data {
				data[fieldname] = t.Tokenize(text)
			}
			response := TokenizeReponse{
				ReqId:  request.ReqId,
				Status: success,
				Data:   data,
			}
			enc.Encode(response)
		}
	}
}

// A websocket handler for detokenization
func WsDetokenize(t tokenizer.Tokenizer) wsHandler {
	return func(ws *websocket.Conn) {
		dec := json.NewDecoder(ws)
		enc := json.NewEncoder(ws)
		for {
			var request DetokenizeRequest
			// Read one request from the socket and attempt to decode
			switch err := dec.Decode(&request); true {
			case err == io.EOF:
				log.Println("Websocket disconnecting")
				return
			case err != nil:
				// Request could not be decoded - return error
				response := DetokenizeReponse{Status: invalidRequest, Error: err.Error()}
				enc.Encode(&response)
				return
			}
			data := make(map[string]foundToken)
			for fieldname, token := range request.Data {
				ft := foundToken{
					Token: token,
				}
				text, err := t.Detokenize(token)
				switch {
				case nil == err:
					ft.Text = text
					ft.Found = true
				case err == tokenizer.TokenNotFound:
					ft.Found = false
				case err != nil:
					log.Panic(err)
				}
				data[fieldname] = ft
			}
			response := DetokenizeReponse{
				ReqId:  request.ReqId,
				Status: success,
				Data:   data,
			}
			enc.Encode(response)
		}
	}
}
