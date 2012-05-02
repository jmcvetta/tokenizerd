package restful

import (
	"code.google.com/p/gorilla/pat"
)

func Router() *pat.Router {
	r := pat.New()
	r.Get("/tokenize/{text}", TokenizeHandler)
	r.Get("/detokenize/{text}", DetokenizeHandler)
	return r
}

func TokenizeHandler(w ResponseWriter, r *Request) {
}

func DetokenizeHandler(w ResponseWriter, r *Request) {
}
