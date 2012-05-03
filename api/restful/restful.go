package restful

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"github.com/jmcvetta/tokenizer"
	"log"
	"net/http"
)

func tokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["s"]
		log.Println("RESTful Tokenize:", s)
		token, err := t.Tokenize(s)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprint(w, token)
	}
}

func detokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["s"]
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

func Router(t tokenizer.Tokenizer) *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/v1/rest/").Subrouter()
	s.HandleFunc("/tokenize/{s}", tokenizeHandler(t))
	s.HandleFunc("/detokenize/{s}", detokenizeHandler(t))
	return r
}
