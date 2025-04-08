package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/AbhishekMashetty/url-shortner/utils"
	"github.com/gorilla/mux"
)

type Storage interface {
	Save(shortCode, originalURL, tenant string) error
	Get(shortCode, tenant string) (string, bool)
}

type URLHandler struct {
	store Storage
}

func NewURLHandler(s Storage) *URLHandler {
	return &URLHandler{store: s}
}

func (h *URLHandler) Home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	// Extract tenant from subdomain
	host := r.Host
	parts := strings.Split(host, ".")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Missing or invalid tenant", http.StatusUnauthorized)
		return
	}
	tenantID := parts[0]

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Generate and save short code
	shortCode := utils.GenerateShortCode()
	if err := h.store.Save(shortCode, originalURL, tenantID); err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortCode)

	// API response
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"short_url": shortURL,
			"original":  originalURL,
		})
		return
	}

	// HTML template response
	tmpl := template.Must(template.ParseFiles("templates/result.html"))
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, map[string]string{
		"ShortURL": shortURL,
	})
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Extract tenant from subdomain
	host := r.Host
	parts := strings.Split(host, ".")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Missing or invalid tenant", http.StatusUnauthorized)
		return
	}
	tenantID := parts[0]

	// Get short code from path
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	// Lookup in store
	originalURL, ok := h.store.Get(shortCode, tenantID)
	if !ok {
		// Return JSON if API client
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Short URL not found",
			})
			return
		}
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
