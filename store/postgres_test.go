package store

import (
	"testing"
)

var testConnStr = "postgres://abhishekmasetty@localhost:5432/url_shortner_test?sslmode=disable"

func setupTestDB(t *testing.T) *PostgresStore {
	db, err := NewPostgresStore(testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Clean up old data
	_, err = db.DB.Exec(`DELETE FROM urls`)
	if err != nil {
		t.Fatalf("Failed to clean up test DB: %v", err)
	}
	return db
}

func TestPostgres_ShortenAndRedirect(t *testing.T) {
	db := setupTestDB(t)

	tenant := "a"
	shortCode := "test123"
	originalURL := "https://example.com"

	// Save to DB
	err := db.Save(shortCode, originalURL, tenant)
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	// Retrieve from DB
	got, ok := db.Get(shortCode, tenant)
	if !ok {
		t.Fatal("Expected to find shortCode, but got false")
	}
	if got != originalURL {
		t.Errorf("Expected URL %q, got %q", originalURL, got)
	}
}
