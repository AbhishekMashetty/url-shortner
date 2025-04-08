package internal

import (
	"github.com/AbhishekMashetty/url-shortner/handlers"

	"github.com/gorilla/mux"
)

func SetupRouter(handler *handlers.URLHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handler.Home).Methods("GET")
	r.HandleFunc("/shorten", handler.Shorten).Methods("POST")
	r.HandleFunc("/{shortCode}", handler.Redirect).Methods("GET")
	return r
}
