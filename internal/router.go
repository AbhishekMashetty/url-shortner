package internal

import (
	"github.com/AbhishekMashetty/url-shortner/handlers"
	"github.com/AbhishekMashetty/url-shortner/store"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	tenantStore := store.NewTenantStore()
	handler := handlers.NewURLHandler(tenantStore)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.Home).Methods("GET")
	r.HandleFunc("/shorten", handler.Shorten).Methods("POST")
	r.HandleFunc("/{shortCode}", handler.Redirect).Methods("GET")

	return r
}
