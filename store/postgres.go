package store

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL")
	return &PostgresStore{DB: db}, nil
}

func (s *PostgresStore) Save(shortCode, originalURL, tenant string) error {
	_, err := s.DB.Exec(`
		INSERT INTO urls (shortcode, original_url, tenant)
		VALUES ($1, $2, $3)
		ON CONFLICT (shortcode) DO NOTHING
	`, shortCode, originalURL, tenant)
	return err
}

func (s *PostgresStore) Get(shortCode, tenant string) (string, bool) {
	var url string
	err := s.DB.QueryRow(`
		SELECT original_url FROM urls
		WHERE shortcode = $1 AND tenant = $2
	`, shortCode, tenant).Scan(&url)

	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		log.Println("❌ DB error:", err)
		return "", false
	}
	return url, true
}
