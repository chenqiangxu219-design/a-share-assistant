package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DefaultCacheKey generates a default cache key for K-lines.
func DefaultCacheKey(code, period string) string {
	return CacheKey(code, period)
}

// ParseCacheKey extracts code and period from a cache key.
func ParseCacheKey(key string) (string, string) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, "d"
}

// WithQuoteTTL returns a new store with the specified quote TTL.
func (s *Store) WithQuoteTTL(ttl time.Duration) *Store {
	return &Store{
		db:       s.db,
		quoteTTL: ttl,
		klineTTL: s.klineTTL,
	}
}

// WithKLineTTL returns a new store with the specified K-line TTL.
func (s *Store) WithKLineTTL(ttl time.Duration) *Store {
	return &Store{
		db:       s.db,
		quoteTTL: s.quoteTTL,
		klineTTL: ttl,
	}
}

// WithTTL returns a new store with the specified TTLs.
func (s *Store) WithTTL(quoteTTL, klineTTL time.Duration) *Store {
	return &Store{
		db:       s.db,
		quoteTTL: quoteTTL,
		klineTTL: klineTTL,
	}
}

// WithDB returns a new store with the specified database.
func (s *Store) WithDB(db *sql.DB) *Store {
	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA busy_timeout=5000")

	return &Store{
		db:       db,
		quoteTTL: s.quoteTTL,
		klineTTL: s.klineTTL,
	}
}

// WithDSN returns a new store with the specified DSN.
func (s *Store) WithDSN(dsn string) (*Store, error) {
	return NewStoreWithTTL(dsn, s.quoteTTL, s.klineTTL)
}

// WithPath returns a new store with the specified path.
func (s *Store) WithPath(path string) (*Store, error) {
	return s.WithDSN(path)
}

// WithDir returns a new store in the specified directory.
func (s *Store) WithDir(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "cache.db")
	return s.WithPath(path)
}

// WithTempDir creates a store in a temporary directory.
func (s *Store) WithTempDir() (*Store, error) {
	dir, err := os.MkdirTemp("", "cache-*")
	if err != nil {
		return nil, err
	}
	return s.WithDir(dir)
}

// WithMemory creates an in-memory store.
func (s *Store) WithMemory() (*Store, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	_, _ = db.Exec("PRAGMA journal_mode=WAL")

	schema, _ := schemaFS.ReadFile("schema.sql")
	_, err = db.Exec(string(schema))
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{
		db:       db,
		quoteTTL: s.quoteTTL,
		klineTTL: s.klineTTL,
	}, nil
}
