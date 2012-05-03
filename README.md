# tokenizerd - A data tokenization server

Tokenizerd presents REST and JSON over websocket APIs for tokenizing and
detokenizing arbitrary strings.  A token uniquely represents, but is not
programmatically derived from, the original string.

See Wikipedia's page on
[Tokenization](http://en.wikipedia.org/wiki/Tokenization_(data_security\)) for
background on why one might want to tokenize some data.

# Security Note

Tokenizerd does not currently implement any sort of access control or transport
encryption. Use it at your own risk!


# Usage

## Run tokenizerd

By default tokenizerd connects to MongoDB on localhost over the default port,
and listens for websocket connections on ws://localhost:3000.  These can be
changed with command line flags:

	$ ./tokenizerd -help
	Usage of ./tokenizerd:
	  -mongo="localhost": URL of MongoDB server
	  -url="localhost:3000": Host/port on which to run websocket listener


## REST API

### Tokenize

	http://localhost:3000/v1/rest/tokenize/{string}

Returns status code 200 and a token string, or status code 500 and an error
message.

### Detokenize

	http://localhost:3000/v1/rest/detokenize/{string}

Returns status code 200 and the original string; status code 404, indicating no
such token exists in the database; or status code 500 and an error message.


## JSON over Websocket API

Connect to tokenizerd with a websocket client.  You can use [Echo
Test](http://websocket.org/echo.html) to experiment.

### Tokenize

Connect to the websocket:

	ws://localhost:3000/v1/ws/tokenize

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

	ws://localhost:3000/v1/ws/detokenize

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


# License

Tokenizerd is Free Software, released under the terms of the
[GPL](http://www.gnu.org/copyleft/gpl.html)
