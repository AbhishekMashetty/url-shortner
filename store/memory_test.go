package store

import "testing"

func TestURLStore_SaveAndGet(t *testing.T) {
	s := NewURLStore()
	shortCode := "abc123"
	originalURL := "https://google.com"

	s.Save(shortCode, originalURL)
	got, ok := s.Get(shortCode)

	if !ok {
		t.Fatal("expected URL to exist")
	}

	if got != originalURL {
		t.Errorf("expected %s, got %s", originalURL, got)
	}
}
