package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/AbhishekMashetty/url-shortner/store"
	"github.com/gorilla/mux"
)

const testConnStr = os.Getenv("TEST_CONN_STR")

// "postgres://abhishekmasetty@localhost:5432/url_shortner_test?sslmode=disable"

func setupTestHandler(t *testing.T) *URLHandler {
	db, err := store.NewPostgresStore(testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	_, err = db.DB.Exec(`DELETE FROM urls`)
	if err != nil {
		t.Fatalf("Failed to clean up test DB: %v", err)
	}
	return NewURLHandler(db)
}

func setupRouter(handler *URLHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/shorten", handler.Shorten).Methods("POST")
	r.HandleFunc("/{shortCode}", handler.Redirect).Methods("GET")
	return r
}

func TestPostgresHandler_ShortenAndRedirect(t *testing.T) {
	h := setupTestHandler(t)
	r := setupRouter(h)

	// Shorten request
	form := url.Values{}
	form.Add("url", "https://example.com")

	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Host = "a.localhost"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var res struct {
		ShortURL string `json:"short_url"`
		Original string `json:"original"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	parts := strings.Split(res.ShortURL, "/")
	shortCode := parts[len(parts)-1]

	// Redirect request
	req2 := httptest.NewRequest("GET", "/"+shortCode, nil)
	req2.Host = "a.localhost"
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusFound {
		t.Fatalf("Expected 302 Found, got %d", rr2.Code)
	}

	loc := rr2.Header().Get("Location")
	if loc != "https://example.com" {
		t.Errorf("Expected redirect to https://example.com, got %s", loc)
	}
}

func TestPostgresHandler_UnknownShortCode(t *testing.T) {
	h := setupTestHandler(t)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	req.Host = "a.localhost"
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rr.Code)
	}
}
