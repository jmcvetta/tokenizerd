# Gokenizer

Gokenizer presents a websocket API for tokenizing and detokenizing arbitrary
data, represented as JSON key/value pairs.


## Installation

	$ go install gokenizer.go

## Usage

### Start MongoDB

Gokenizer uses MongoDB as a datastore.  Installation instructions for MongoDB
can be [found here](http://www.mongodb.org/display/DOCS/Quickstart).

Currently Gokenizer connects to MongoDB on the default port with no security.
This will be improved in a future version.

### Connect

Connect to Gokenizer with a websocket client.  You can use [Echo
Test](http://websocket.org/echo.html) to experiment.



### Tokenize

Connect to the websocket:

	ws://localhost:3000/v1/tokenize

Issue a JSON request:

	{
		"ReqId": "an arbitrary string identifying this request",
		"Data": {
			"fieldname1": "fieldvalue1",
			"field name 2": "field  value 2"
		}
	}

Response:

	{
		"ReqId": "an arbitrary string identifying this request",
		"Status": "Success",
		"Error": "",
		"Data": {
			"field name 2": "OTMyMzgzMDAw",
			"fieldname1": "OTMwMjkxMDAw"
		}
	}

### Detokenize

Connect to the websocket:

	ws://localhost:3000/v1/detokenize

Issue a JSON request:

	{
		"ReqId": "foobar",
		"Data": {
			"field name 2": "OTMyMzgzMDAw",
			"fieldname1": "OTMwMjkxMDAw",
			"fieldname 3": "non-existent token string"
		}
	}

Response:

	{
		"ReqId": "foobar",
		"Status": "Success", 
		"Error": "",
		"Data": {
			"field name 2": {
				"Token": "OTMyMzgzMDAw",
				"Found":true,
				"Text": "field value 2"
			},
			"fieldname1": {
				"Token": "OTMwMjkxMDAw",
				"Found":true,
				"Text":"fieldvalue1"
			},
			"fieldname 3": {
				"Token":"non-existent token string",
				"Found": false,
				"Text":""
			}
		}
	}
