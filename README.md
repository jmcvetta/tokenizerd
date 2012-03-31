# Gokenizer

Gokenizer presents a websocket API for tokenizing and detokenizing arbitrary data, represented as JSON key/value pairs.


## Installation

	$ go install gokenizer.go

## Usage

### Connect

Connect to Gokenizer with a websocket client.  You can use http://websocket.org/echo to experiment.

Tokenizer: ws://localhost:3000/v1/tokenize

Detokenizer: ws://localhost:3000/v1/detokenize

### Tokenize Request

Issue a JSON request:

	{
		"ReqId": "an arbitrary string identifying this request",
		"Data": {
			"fieldname1": "fieldvalue1",
			"field name 2": "field  value 2",
		}
	}

### Tokenize Response

[to be completed]
