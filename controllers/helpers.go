package controllers

import (
	"log"
	"net/http"

	"github.com/gorilla/schema"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true) // ignore the CSRF token key
	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}

	return nil
}

func handleUnknownError(err error, w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
	log.Println(err)
}
