package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AbhishekMashetty/url-shortner/store"
	"github.com/gorilla/mux"
)

func TestShortenAndRedirect(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods("POST")
	r.HandleFunc("/{shortCode}", h.Redirect).Methods("GET")

	// === Tenant A shortens a URL ===
	form := url.Values{}
	form.Add("url", "https://example.com")
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Host = "tenant-a.url-short.com"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	// Extract short code from tenant A’s store
	var shortCode string
	store := ts.GetStore("tenant-a")
	for k := range store.Store {
		shortCode = k
		break
	}
	if shortCode == "" {
		t.Fatal("expected a short code to be generated")
	}

	// === Simulate redirect ===
	req2 := httptest.NewRequest("GET", "/"+shortCode, nil)
	req2.Host = "tenant-a.url-short.com"
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rr2.Code)
	}

	location := rr2.Header().Get("Location")
	if location != "https://example.com" {
		t.Errorf("expected redirect to https://example.com, got %s", location)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/{shortCode}", h.Redirect).Methods("GET")

	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	req.Host = "tenant-a.url-short.com" // simulate subdomain-based tenant

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rr.Code)
	}

	body := rr.Body.String()
	expected := "Short URL not found"
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}

func TestRedirect_NotFound_JSON(t *testing.T) {
	// Setup tenant-aware store and handler
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	// Setup router
	r := mux.NewRouter()
	r.HandleFunc("/{shortCode}", h.Redirect).Methods("GET")

	// Use a fake tenant host: a.url-short.com
	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	req.Header.Set("Accept", "application/json")
	req.Host = "a.url-short.com" // 👈 simulate tenant "a"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	expected := "Short URL not found"
	if response["error"] != expected {
		t.Errorf("expected error message %q, got %q", expected, response["error"])
	}
}

func TestMultiTenantIsolation(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods("POST")
	r.HandleFunc("/{shortCode}", h.Redirect).Methods("GET")

	// Tenant A shortens a URL
	form := url.Values{}
	form.Add("url", "https://example.com")
	reqA := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	reqA.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqA.Header.Set("Accept", "application/json")
	reqA.Host = "tenant-a.url-short.com"

	rrA := httptest.NewRecorder()
	r.ServeHTTP(rrA, reqA)

	if rrA.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for tenant A, got %d", rrA.Code)
	}

	// Extract short code
	var shortCode string
	storeA := ts.GetStore("tenant-a")
	for k := range storeA.Store {
		shortCode = k
		break
	}
	if shortCode == "" {
		t.Fatal("short code not generated")
	}

	// Tenant B tries to access it
	reqB := httptest.NewRequest("GET", "/"+shortCode, nil)
	reqB.Header.Set("Accept", "application/json")
	reqB.Host = "tenant-b.url-short.com"

	rrB := httptest.NewRecorder()
	r.ServeHTTP(rrB, reqB)

	if rrB.Code != http.StatusNotFound {
		t.Errorf("expected 404 for tenant B, got %d", rrB.Code)
	}

	// Tenant A accesses it — should succeed
	reqAGet := httptest.NewRequest("GET", "/"+shortCode, nil)
	reqAGet.Host = "tenant-a.url-short.com"

	rrAGet := httptest.NewRecorder()
	r.ServeHTTP(rrAGet, reqAGet)

	if rrAGet.Code != http.StatusFound {
		t.Errorf("expected 302 redirect for tenant A, got %d", rrAGet.Code)
	}

	location := rrAGet.Header().Get("Location")
	if location != "https://example.com" {
		t.Errorf("expected redirect to https://example.com, got %s", location)
	}
}

func TestMissingTenant(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods("POST")

	form := url.Values{}
	form.Add("url", "https://example.com")
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Host = "" // 👈 simulate no subdomain

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("expected non-200 when tenant is missing, got %d", rr.Code)
	}
}

func TestAPI_Shorten_JSONResponse(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods("POST")

	form := url.Values{}
	form.Add("url", "https://golang.org")

	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Host = "tenant-x.url-short.com"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if response["short_url"] == "" {
		t.Errorf("expected short_url in response, got: %v", response)
	}
	if response["original"] != "https://golang.org" {
		t.Errorf("expected original url to match input")
	}
}

func TestAPI_Shorten_MissingURL(t *testing.T) {
	ts := store.NewTenantStore()
	h := NewURLHandler(ts)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", h.Shorten).Methods("POST")

	req := httptest.NewRequest("POST", "/shorten", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Host = "tenant-x.url-short.com"

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}
