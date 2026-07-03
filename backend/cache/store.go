package cache

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"a-share-assistant/backend/model"
)

//go:embed schema.sql
var schemaFS embed.FS

// Store provides SQLite-backed caching for quotes and K-lines.
type Store struct {
	mu       sync.RWMutex
	db       *sql.DB
	dsn      string
	quoteTTL time.Duration
	klineTTL time.Duration
}

// NewStore creates a new store with the given DSN.
func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA busy_timeout=5000")

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("exec schema: %w", err)
	}

	return &Store{
		db:       db,
		dsn:      dsn,
		quoteTTL: 5 * time.Minute,
		klineTTL: 30 * time.Minute,
	}, nil
}

// NewStoreWithTTL creates a store with custom TTLs.
func NewStoreWithTTL(dsn string, quoteTTL, klineTTL time.Duration) (*Store, error) {
	s, err := NewStore(dsn)
	if err != nil {
		return nil, err
	}
	s.quoteTTL = quoteTTL
	s.klineTTL = klineTTL
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetQuote returns cached quote if not expired.
func (s *Store) GetQuote(code string) (*model.Quote, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		"SELECT data FROM quotes WHERE code = ? AND updated_at > ?",
		code, time.Now().Add(-s.quoteTTL).Unix(),
	)
	var data string
	if err := row.Scan(&data); err != nil {
		return nil, false
	}
	var quote model.Quote
	if err := json.Unmarshal([]byte(data), &quote); err != nil {
		return nil, false
	}
	return &quote, true
}

// SetQuote caches a quote.
func (s *Store) SetQuote(code string, quote *model.Quote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(quote)
	if err != nil {
		return
	}

	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO quotes (code, data, updated_at) VALUES (?, ?, ?)",
		code, data, time.Now().Unix(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache setQuote: %v\n", err)
	}
}

// GetKLines returns cached K-lines if not expired.
func (s *Store) GetKLines(key string) ([]model.KLine, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(
		"SELECT data FROM klines WHERE key = ? AND updated_at > ?",
		key, time.Now().Add(-s.klineTTL).Unix(),
	)
	var data string
	if err := row.Scan(&data); err != nil {
		return nil, false
	}
	var klines []model.KLine
	if err := json.Unmarshal([]byte(data), &klines); err != nil {
		return nil, false
	}
	return klines, true
}

// SetKLines caches K-line data.
func (s *Store) SetKLines(key string, klines []model.KLine) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(klines)
	if err != nil {
		return
	}

	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO klines (key, data, updated_at) VALUES (?, ?, ?)",
		key, data, time.Now().Unix(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache setKLines: %v\n", err)
	}
}

// GetCachedCodes returns all cached stock codes within TTL.
func (s *Store) GetCachedCodes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT code FROM quotes WHERE updated_at > ?",
		time.Now().Add(-s.quoteTTL).Unix(),
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			codes = append(codes, code)
		}
	}
	return codes
}

// DefaultDSN returns the default database path.
func DefaultDSN() string {
	dir := filepath.Join(os.Getenv("HOME"), ".a-share-assistant")
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir cache dir: %v\n", err)
	}
	return filepath.Join(dir, "cache.db")
}

// ClearExpired removes all expired records.
func (s *Store) ClearExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	_, _ = s.db.Exec("DELETE FROM quotes WHERE updated_at < ?", now-int64(s.quoteTTL.Seconds()))
	_, _ = s.db.Exec("DELETE FROM klines WHERE updated_at < ?", now-int64(s.klineTTL.Seconds()))
}

// GetAllQuotes returns all cached quotes regardless of TTL (for initial load).
func (s *Store) GetAllQuotes() map[string]*model.Quote {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT code, data FROM quotes")
	if err != nil {
		return nil
	}
	defer rows.Close()

	quotes := make(map[string]*model.Quote)
	for rows.Next() {
		var code string
		var data string
		if err := rows.Scan(&code, &data); err != nil {
			continue
		}
		var quote model.Quote
		if err := json.Unmarshal([]byte(data), &quote); err != nil {
			continue
		}
		quotes[code] = &quote
	}
	return quotes
}

// ImportQuoteBatch bulk-inserts quotes.
func (s *Store) ImportQuoteBatch(quotes map[string]*model.Quote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO quotes (code, data, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for code, q := range quotes {
		data, _ := json.Marshal(q)
		_, _ = stmt.Exec(code, data, now)
	}

	_ = tx.Commit()
}

// Count returns number of records in each table.
func (s *Store) Count() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var qc, kc int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM quotes").Scan(&qc)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM klines").Scan(&kc)
	return qc, kc
}

// BatchGetQuotes gets multiple quotes at once.
func (s *Store) BatchGetQuotes(codes []string) map[string]*model.Quote {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*model.Quote)
	if len(codes) == 0 {
		return result
	}

	// Build query with placeholders
	placeholders := make([]string, len(codes))
	args := make([]interface{}, len(codes))
	for i, code := range codes {
		placeholders[i] = "?"
		args[i] = code
	}

	query := fmt.Sprintf("SELECT code, data FROM quotes WHERE code IN (%s)",
		joinPlaceholders(placeholders))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var data string
		if err := rows.Scan(&code, &data); err != nil {
			continue
		}
		var quote model.Quote
		if err := json.Unmarshal([]byte(data), &quote); err != nil {
			continue
		}
		result[code] = &quote
	}
	return result
}

// joinPlaceholders creates a comma-separated string of placeholders.
func joinPlaceholders(placeholders []string) string {
	return strings.Join(placeholders, ",")
}

// CacheKey generates a cache key for K-lines.
func CacheKey(code, period string) string {
	return fmt.Sprintf("%s:%s", code, period)
}

// GetKLinesByCodePeriod is a convenience method to get K-lines by code and period.
func (s *Store) GetKLinesByCodePeriod(code, period string) ([]model.KLine, bool) {
	key := CacheKey(code, period)
	return s.GetKLines(key)
}

// SetKLinesByCodePeriod is a convenience method to set K-lines by code and period.
func (s *Store) SetKLinesByCodePeriod(code, period string, klines []model.KLine) {
	key := CacheKey(code, period)
	s.SetKLines(key, klines)
}

// InvalidateQuote removes a quote from cache.
func (s *Store) InvalidateQuote(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("DELETE FROM quotes WHERE code = ?", code)
}

// InvalidateKLines removes K-lines from cache.
func (s *Store) InvalidateKLines(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("DELETE FROM klines WHERE key = ?", key)
}

// InvalidateAll removes all cached data.
func (s *Store) InvalidateAll() {
	s.Reset()
}

// Exists checks if a key exists in cache (regardless of TTL).
func (s *Store) Exists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM quotes WHERE code = ?", code).Scan(&count)
	return count > 0
}

// ExistsKLines checks if K-lines exist in cache (regardless of TTL).
func (s *Store) ExistsKLines(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM klines WHERE key = ?", key).Scan(&count)
	return count > 0
}

// GetQuoteRaw returns the raw quote (may be expired) for recovery.
func (s *Store) GetQuoteRaw(code string) (*model.Quote, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow("SELECT data FROM quotes WHERE code = ?", code)
	var data string
	if err := row.Scan(&data); err != nil {
		return nil, false
	}
	var quote model.Quote
	if err := json.Unmarshal([]byte(data), &quote); err != nil {
		return nil, false
	}
	return &quote, true
}

// GetKLinesRaw returns raw K-lines (may be expired) for recovery.
func (s *Store) GetKLinesRaw(key string) ([]model.KLine, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow("SELECT data FROM klines WHERE key = ?", key)
	var data string
	if err := row.Scan(&data); err != nil {
		return nil, false
	}
	var klines []model.KLine
	if err := json.Unmarshal([]byte(data), &klines); err != nil {
		return nil, false
	}
	return klines, true
}

// RefreshBatch updates quotes for multiple codes in a single transaction.
func (s *Store) RefreshBatch(quotes map[string]*model.Quote) {
	s.ImportQuoteBatch(quotes)
}

// InvalidateCode removes all cache entries for a stock code.
func (s *Store) InvalidateCode(code string) {
	s.InvalidateQuote(code)

	// Find and remove all K-lines for this code
	keys := s.GetCachedKLinesKeys()
	for _, key := range keys {
		if strings.HasPrefix(key, code+":") {
			s.InvalidateKLines(key)
		}
	}
}

// InvalidateCodeBatch removes all cache entries for multiple stock codes.
func (s *Store) InvalidateCodeBatch(codes []string) {
	for _, code := range codes {
		s.InvalidateCode(code)
	}
}

// GetQuoteOrRaw returns cached quote if valid, otherwise raw (may be expired).
func (s *Store) GetQuoteOrRaw(code string) (*model.Quote, bool, bool) {
	// First try valid cache
	if quote, ok := s.GetQuote(code); ok {
		return quote, true, false // quote, found, isFresh
	}
	// Fall back to raw (may be expired)
	if quote, ok := s.GetQuoteRaw(code); ok {
		return quote, true, false // quote, found, isStale
	}
	return nil, false, false
}

// GetKLinesOrRaw returns cached K-lines if valid, otherwise raw.
func (s *Store) GetKLinesOrRaw(key string) ([]model.KLine, bool, bool) {
	if klines, ok := s.GetKLines(key); ok {
		return klines, true, false
	}
	if klines, ok := s.GetKLinesRaw(key); ok {
		return klines, true, false
	}
	return nil, false, false
}

// BatchGetQuotesByCodesOrNames gets quotes for multiple codes/names.
func (s *Store) BatchGetQuotesByCodesOrNames(queries []string) map[string]*model.Quote {
	result := make(map[string]*model.Quote)

	for _, q := range queries {
		if quote, ok := s.GetQuoteByCodeOrName(q); ok {
			result[q] = quote
		}
	}
	return result
}

// GetQuotesForWatchlist gets quotes for a watchlist of codes.
func (s *Store) GetQuotesForWatchlist(codes []string) map[string]*model.Quote {
	return s.BatchGetQuotes(codes)
}

// UpdateWatchlist refreshes quotes for a watchlist.
func (s *Store) UpdateWatchlist(codes []string, fetcher func(code string) (*model.Quote, error)) map[string]*model.Quote {
	return s.BatchRefreshIfNeeded(codes, fetcher)
}

// GetCachedQuoteForCode returns the cached quote for a code.
func (s *Store) GetCachedQuoteForCode(code string) (*model.Quote, bool) {
	return s.GetQuote(code)
}

// CacheQuote caches a quote for a code.
func (s *Store) CacheQuote(code string, quote *model.Quote) {
	s.SetQuote(code, quote)
}

// GetCachedKLinesForCodePeriod returns cached K-lines for a code and period.
func (s *Store) GetCachedKLinesForCodePeriod(code, period string) ([]model.KLine, bool) {
	return s.GetKLinesByCodePeriod(code, period)
}

// CacheKLines caches K-lines for a code and period.
func (s *Store) CacheKLines(code, period string, klines []model.KLine) {
	s.SetKLinesByCodePeriod(code, period, klines)
}

// InvalidateCodeCache removes all cached data for a code.
func (s *Store) InvalidateCodeCache(code string) {
	s.InvalidateCode(code)
}
