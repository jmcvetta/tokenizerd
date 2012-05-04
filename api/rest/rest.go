package rest

import (
	"fmt"
	"github.com/jmcvetta/tokenizer"
	"log"
	"net/http"
)

func TokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		s := req.URL.Query().Get(":string")
		log.Println("RESTful Tokenize:", s)
		token, err := t.Tokenize(s)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprint(w, token)
	}
}

func DetokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		s := req.URL.Query().Get(":token")
		log.Println("RESTful Detokenize:", s)
		orig, err := t.Detokenize(s)
		switch {
		case err == tokenizer.TokenNotFound:
			http.Error(w, "Token not found", 404)
		case err != nil:
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprint(w, orig)
	}
}
