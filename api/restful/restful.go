package restful

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"github.com/jmcvetta/tokenizer"
	"html"
	"log"
	"net/http"
)

func TokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["s"]
		log.Println("RESTful Tokenize:", s)
		s = html.UnescapeString(s)
		token, err := t.Tokenize(s)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprint(w, html.EscapeString(token))
	}
}

func DetokenizeHandler(t tokenizer.Tokenizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["s"]
		log.Println("RESTful Detokenize:", s)
		s = html.UnescapeString(s)
		orig, err := t.Detokenize(s)
		switch {
		case err == tokenizer.TokenNotFound:
			http.Error(w, "Token not found", 404)
		case err != nil:
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprint(w, html.EscapeString(orig))
	}
}

func Router(t tokenizer.Tokenizer) *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/v1/rest/").Subrouter()
	s.HandleFunc("/tokenize/{s}", TokenizeHandler(t))
	s.HandleFunc("/detokenize/{s}", DetokenizeHandler(t))
	return r
}
