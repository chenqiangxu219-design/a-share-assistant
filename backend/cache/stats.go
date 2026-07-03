package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"a-share-assistant/backend/model"
)

// DumpAll returns all cached data for debugging.
func (s *Store) DumpAll() string {
	quotes := s.GetAllQuotes()
	var result string
	result += fmt.Sprintf("Cached stocks: %d\n", len(quotes))
	for code, q := range quotes {
		result += fmt.Sprintf("  %s: %s %.2f (%.2f%%)\n", code, q.Name, q.Price, q.ChangePct)
	}
	return result
}

// ExportAllQuotes returns all quotes as JSON-serializable format.
func (s *Store) ExportAllQuotes() []*model.Quote {
	quotes := s.GetAllQuotes()
	result := make([]*model.Quote, 0, len(quotes))
	for _, q := range quotes {
		result = append(result, q)
	}
	return result
}

// Stats returns cache statistics.
func (s *Store) Stats() map[string]interface{} {
	qc, kc := s.Count()
	return map[string]interface{}{
		"quotes":    qc,
		"klines":    kc,
		"quote_ttl": s.quoteTTL.String(),
		"kline_ttl": s.klineTTL.String(),
	}
}

// DumpStats returns all cached data as a map for the /api/cache/stats endpoint.
func (s *Store) DumpStats() map[string]interface{} {
	return s.Stats()
}

// LoadFromFile reads quotes from a JSON file and imports them.
func (s *Store) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var quotes []*model.Quote
	if err := json.Unmarshal(data, &quotes); err != nil {
		return err
	}

	quoteMap := make(map[string]*model.Quote)
	for _, q := range quotes {
		quoteMap[q.Code] = q
	}
	s.ImportQuoteBatch(quoteMap)
	return nil
}

// SaveToFile exports all quotes to a JSON file.
func (s *Store) SaveToFile(path string) error {
	quotes := s.GetAllQuotes()
	if len(quotes) == 0 {
		return fmt.Errorf("no quotes to save")
	}

	data, err := json.MarshalIndent(quotes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetStats returns cache statistics.
func (s *Store) GetStats() map[string]interface{} {
	return s.Stats()
}

// ResetStats resets all statistics (for testing).
func (s *Store) ResetStats() {
	// No-op for now, reserved for future use
}

// IsEmpty checks if the cache has no data.
func (s *Store) IsEmpty() bool {
	qc, kc := s.Count()
	return qc == 0 && kc == 0
}

// TotalRecords returns the total number of cached records.
func (s *Store) TotalRecords() int {
	qc, kc := s.Count()
	return qc + kc
}

// UniqueStocks returns the number of unique stocks cached.
func (s *Store) UniqueStocks() int {
	codes := s.GetAllCodes()
	return len(codes)
}

// HasQuote checks if a quote exists (not expired).
func (s *Store) HasQuote(code string) bool {
	_, ok := s.GetQuote(code)
	return ok
}

// HasKLines checks if K-lines exist (not expired).
func (s *Store) HasKLines(key string) bool {
	_, ok := s.GetKLines(key)
	return ok
}

// QuoteAge returns how old a cached quote is.
func (s *Store) QuoteAge(code string) time.Duration {
	ts := s.GetLastUpdated(code)
	if ts.IsZero() {
		return 0
	}
	return time.Since(ts)
}

// KLineAge returns how old cached K-lines are.
func (s *Store) KLineAge(key string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ts int64
	err := s.db.QueryRow(
		"SELECT updated_at FROM klines WHERE key = ?",
		key,
	).Scan(&ts)
	if err != nil || ts == 0 {
		return 0
	}
	return time.Since(time.Unix(ts, 0))
}

// GetLastUpdated returns the last update time for a stock.
func (s *Store) GetLastUpdated(code string) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ts int64
	err := s.db.QueryRow(
		"SELECT MAX(updated_at) FROM quotes WHERE code = ?",
		code,
	).Scan(&ts)
	if err != nil || ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// GetCacheSize returns the current cache file size.
func (s *Store) GetCacheSize() (int64, error) {
	path := s.dsn
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetCacheSizeMB returns the cache file size in MB.
func (s *Store) GetCacheSizeMB() float64 {
	size, err := s.GetCacheSize()
	if err != nil {
		return 0
	}
	return float64(size) / (1024 * 1024)
}

// GetCacheAge returns how old the cache file is.
func (s *Store) GetCacheAge() (time.Duration, error) {
	path := s.dsn
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return time.Since(info.ModTime()), nil
}

// GetCacheModifiedTime returns the last modification time of the cache file.
func (s *Store) GetCacheModifiedTime() (time.Time, error) {
	path := s.dsn
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// GetCacheCreatedTime returns the creation time of the cache file.
func (s *Store) GetCacheCreatedTime() (time.Time, error) {
	path := s.dsn
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// GetCacheStats returns comprehensive cache statistics.
func (s *Store) GetCacheStats() map[string]interface{} {
	sizeMB := s.GetCacheSizeMB()
	age, _ := s.GetCacheAge()
	qc, kc := s.Count()
	uniqueStocks := s.UniqueStocks()

	return map[string]interface{}{
		"size_mb":       sizeMB,
		"age":           age.String(),
		"quotes_count":  qc,
		"klines_count":  kc,
		"unique_stocks": uniqueStocks,
		"total_records": qc + kc,
		"db_path":       s.dsn,
	}
}

// CacheStats provides detailed cache statistics.
func (s *Store) CacheStats() map[string]interface{} {
	qc, kc := s.Count()
	uniqueStocks := s.UniqueStocks()

	quoteTTL, klineTTL := s.GetTTL()

	return map[string]interface{}{
		"quotes_count":      qc,
		"klines_count":      kc,
		"unique_stocks":     uniqueStocks,
		"quote_ttl":         quoteTTL.String(),
		"kline_ttl":         klineTTL.String(),
		"db_path":           s.dsn,
		"total_records":     qc + kc,
		"quote_ttl_seconds": int64(quoteTTL.Seconds()),
		"kline_ttl_seconds": int64(klineTTL.Seconds()),
	}
}

// GetQuoteWithAge returns a quote and its age.
func (s *Store) GetQuoteWithAge(code string) (*model.Quote, time.Duration, bool) {
	quote, ok := s.GetQuote(code)
	if !ok {
		return nil, 0, false
	}
	age := s.QuoteAge(code)
	return quote, age, true
}

// GetKLinesWithAge returns K-lines and their age.
func (s *Store) GetKLinesWithAge(key string) ([]model.KLine, time.Duration, bool) {
	klines, ok := s.GetKLines(key)
	if !ok {
		return nil, 0, false
	}
	age := s.KLineAge(key)
	return klines, age, true
}

// GetQuoteStats returns statistics for a specific stock.
func (s *Store) GetQuoteStats(code string) map[string]interface{} {
	age := s.QuoteAge(code)
	hasQuote := s.HasQuote(code)
	hasKLines := false

	// Check if any K-lines exist for this code
	keys := s.GetCachedKLinesKeys()
	for _, key := range keys {
		if strings.HasPrefix(key, code+":") {
			hasKLines = true
			break
		}
	}

	return map[string]interface{}{
		"code":              code,
		"has_quote":         hasQuote,
		"has_klines":        hasKLines,
		"quote_age":         age.String(),
		"quote_age_seconds": int64(age.Seconds()),
	}
}

// GetPopularCodes returns the most recently accessed codes.
func (s *Store) GetPopularCodes(limit int) []string {
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT code FROM quotes ORDER BY updated_at DESC LIMIT ?",
		limit,
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

// SetMaxSize sets a soft limit on database file size.
func (s *Store) SetMaxSize(maxSizeMB int) {
	if maxSizeMB <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check current size
	info, err := os.Stat(s.dsn)
	if err != nil {
		return
	}

	maxBytes := int64(maxSizeMB * 1024 * 1024)
	if info.Size() > maxBytes {
		// Trim oldest records
		s.TrimToMax(50)
		s.Vacuum()
	}
}
