package mlog

import (
	"net/http"
	"strings"
)

func HandleOutput() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		write(w, Content(), http.StatusOK)
	}
}

func HandleGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		write(w, strings.Join(Groups(), "\n"), http.StatusOK)
	}
}

func HandleGroupOutput() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		write(w, ContentBy(r.PathValue("group")), http.StatusOK)
	}
}

func write(w http.ResponseWriter, content string, status int) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}
