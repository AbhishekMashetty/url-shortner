package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AbhishekMashetty/url-shortner/store"
	"github.com/AbhishekMashetty/url-shortner/utils"
	"github.com/gorilla/mux"
)

type URLHandler struct {
	Tenants *store.TenantStore
}

func NewURLHandler(ts *store.TenantStore) *URLHandler {
	return &URLHandler{Tenants: ts}
}

func (h *URLHandler) Home(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
		<head><title>URL Shortener</title></head>
		<body>
			<h1>URL Shortener</h1>
			<form method="POST" action="/shorten">
				<input type="text" name="url" placeholder="Enter URL" style="width: 300px;" required />
				<input type="submit" value="Shorten">
			</form>
		</body>
		</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	parts := strings.Split(host, ".")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Missing or invalid tenant", http.StatusUnauthorized)
		return
	}
	tenantID := parts[0]
	store := h.Tenants.GetStore(tenantID)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortCode := utils.GenerateShortCode()
	store.Save(shortCode, originalURL)

	shortURL := fmt.Sprintf("http://%s/%s", r.Host, shortCode)

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"short_url": shortURL,
			"original":  originalURL,
		})
		return
	}

	html := fmt.Sprintf(`<a href="%s">%s</a>`, shortURL, shortURL)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Enforce tenant from subdomain
	host := r.Host
	parts := strings.Split(host, ".")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Missing or invalid tenant", http.StatusUnauthorized)
		return
	}
	tenantID := parts[0]
	store := h.Tenants.GetStore(tenantID)

	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	originalURL, ok := store.Get(shortCode)
	if !ok {
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
