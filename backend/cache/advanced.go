package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"a-share-assistant/backend/model"
)

// GetQuoteByCode returns quote for a specific code.
func (s *Store) GetQuoteByCode(code string) (*model.Quote, bool) {
	return s.GetQuote(code)
}

// SetQuoteWithTimestamp sets a quote with a specific timestamp.
func (s *Store) SetQuoteWithTimestamp(code string, quote *model.Quote, ts time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(quote)
	if err != nil {
		return
	}

	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO quotes (code, data, updated_at) VALUES (?, ?, ?)",
		code, data, ts.Unix(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache setQuoteWithTimestamp: %v\n", err)
	}
}

// GetCachedKLinesKeys returns all cached K-line keys.
func (s *Store) GetCachedKLinesKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT key FROM klines")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err == nil {
			keys = append(keys, key)
		}
	}
	return keys
}

// GetCachedKLinesCodes returns unique stock codes from K-lines cache.
func (s *Store) GetCachedKLinesCodes() []string {
	keys := s.GetCachedKLinesKeys()
	seen := make(map[string]struct{})
	var codes []string
	for _, key := range keys {
		code := key[:strings.IndexByte(key, ':')]
		if _, ok := seen[code]; !ok {
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
	}
	return codes
}

// GetAllCodes returns all unique stock codes from both quotes and K-lines.
func (s *Store) GetAllCodes() []string {
	quoteCodes := s.GetCachedCodes()
	klineCodes := s.GetCachedKLinesCodes()

	seen := make(map[string]struct{})
	var codes []string
	for _, code := range quoteCodes {
		if _, ok := seen[code]; !ok {
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
	}
	for _, code := range klineCodes {
		if _, ok := seen[code]; !ok {
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
	}
	return codes
}

// LoadFromLegacyMap migrates data from the old in-memory cache format.
func (s *Store) LoadFromLegacyMap(quotes map[string]*model.Quote, klines map[string][]model.KLine) {
	// Import quotes
	if len(quotes) > 0 {
		s.ImportQuoteBatch(quotes)
	}

	// Import K-lines
	if len(klines) > 0 {
		s.mu.Lock()
		defer s.mu.Unlock()

		now := time.Now().Unix()
		tx, err := s.db.Begin()
		if err != nil {
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare("INSERT OR REPLACE INTO klines (key, data, updated_at) VALUES (?, ?, ?)")
		if err != nil {
			return
		}
		defer stmt.Close()

		for key, kl := range klines {
			data, _ := json.Marshal(kl)
			_, _ = stmt.Exec(key, data, now)
		}

		_ = tx.Commit()
	}
}

// ExportForLegacy returns data in the old in-memory cache format.
func (s *Store) ExportForLegacy() (map[string]*model.Quote, map[string][]model.KLine) {
	quotes := s.GetAllQuotes()

	klineKeys := s.GetCachedKLinesKeys()
	klines := make(map[string][]model.KLine)
	for _, key := range klineKeys {
		if kl, ok := s.GetKLinesRaw(key); ok {
			klines[key] = kl
		}
	}

	return quotes, klines
}

// BatchGetKLines retrieves multiple K-line sets at once.
func (s *Store) BatchGetKLines(keys []string) map[string][]model.KLine {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]model.KLine)
	if len(keys) == 0 {
		return result
	}

	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := fmt.Sprintf("SELECT key, data FROM klines WHERE key IN (%s)",
		joinPlaceholders(placeholders))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var data string
		if err := rows.Scan(&key, &data); err != nil {
			continue
		}
		var klines []model.KLine
		if err := json.Unmarshal([]byte(data), &klines); err != nil {
			continue
		}
		result[key] = klines
	}
	return result
}

// BatchSetKLines inserts multiple K-line sets in a single transaction.
func (s *Store) BatchSetKLines(klines map[string][]model.KLine) {
	if len(klines) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO klines (key, data, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for key, kl := range klines {
		data, _ := json.Marshal(kl)
		_, _ = stmt.Exec(key, data, now)
	}

	_ = tx.Commit()
}

// MigrateFromLegacy migrates all data from the old in-memory cache to SQLite.
func (s *Store) MigrateFromLegacy(
	quotes map[string]*model.Quote,
	quoteTts map[string]time.Time,
	klines map[string][]model.KLine,
	klineTts map[string]time.Time,
) {
	// Import quotes with timestamps
	if len(quotes) > 0 {
		s.mu.Lock()
		defer s.mu.Unlock()

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
			ts := quoteTts[code]
			if ts.IsZero() {
				ts = time.Now()
			}
			_, _ = stmt.Exec(code, data, ts.Unix())
		}

		_ = tx.Commit()
	}

	// Import K-lines with timestamps
	if len(klines) > 0 {
		s.mu.Lock()
		defer s.mu.Unlock()

		tx, err := s.db.Begin()
		if err != nil {
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare("INSERT OR REPLACE INTO klines (key, data, updated_at) VALUES (?, ?, ?)")
		if err != nil {
			return
		}
		defer stmt.Close()

		for key, kl := range klines {
			data, _ := json.Marshal(kl)
			ts := klineTts[key]
			if ts.IsZero() {
				ts = time.Now()
			}
			_, _ = stmt.Exec(key, data, ts.Unix())
		}

		_ = tx.Commit()
	}
}

// BatchSetQuotes inserts multiple quotes in a single transaction.
func (s *Store) BatchSetQuotes(quotes map[string]*model.Quote) {
	s.ImportQuoteBatch(quotes)
}

// RefreshIfNeeded refreshes data if it's older than TTL.
func (s *Store) RefreshIfNeeded(code string, fetcher func() (*model.Quote, error)) (*model.Quote, error) {
	if quote, ok := s.GetQuote(code); ok {
		return quote, nil
	}

	quote, err := fetcher()
	if err != nil {
		return nil, err
	}

	s.SetQuote(code, quote)
	return quote, nil
}

// BatchRefreshIfNeeded refreshes multiple quotes.
func (s *Store) BatchRefreshIfNeeded(codes []string, fetcher func(code string) (*model.Quote, error)) map[string]*model.Quote {
	cached := s.BatchGetQuotes(codes)
	needRefresh := make([]string, 0)

	for _, code := range codes {
		if _, ok := cached[code]; !ok {
			needRefresh = append(needRefresh, code)
		}
	}

	if len(needRefresh) == 0 {
		return cached
	}

	fresh := make(map[string]*model.Quote)
	for _, code := range needRefresh {
		quote, err := fetcher(code)
		if err == nil {
			fresh[code] = quote
		}
	}

	s.ImportQuoteBatch(fresh)

	// Merge cached and fresh
	result := make(map[string]*model.Quote)
	for k, v := range cached {
		result[k] = v
	}
	for k, v := range fresh {
		result[k] = v
	}
	return result
}

// ImportFromJSON reads quotes from a JSON array string.
func (s *Store) ImportFromJSON(data string) error {
	var quotes []*model.Quote
	if err := json.Unmarshal([]byte(data), &quotes); err != nil {
		return err
	}

	quoteMap := make(map[string]*model.Quote)
	for _, q := range quotes {
		quoteMap[q.Code] = q
	}
	s.ImportQuoteBatch(quoteMap)
	return nil
}

// ExportToJSON returns all quotes as a JSON array string.
func (s *Store) ExportToJSON() (string, error) {
	quotes := s.ExportAllQuotes()
	data, err := json.Marshal(quotes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BatchInvalidate removes multiple quotes from cache.
func (s *Store) BatchInvalidate(codes []string) {
	if len(codes) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	placeholders := make([]string, len(codes))
	args := make([]interface{}, len(codes))
	for i, code := range codes {
		placeholders[i] = "?"
		args[i] = code
	}

	query := fmt.Sprintf("DELETE FROM quotes WHERE code IN (%s)",
		joinPlaceholders(placeholders))
	_, _ = s.db.Exec(query, args...)
}

// BatchInvalidateKLines removes multiple K-line sets from cache.
func (s *Store) BatchInvalidateKLines(keys []string) {
	if len(keys) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := fmt.Sprintf("DELETE FROM klines WHERE key IN (%s)",
		joinPlaceholders(placeholders))
	_, _ = s.db.Exec(query, args...)
}

// TouchQuote updates the timestamp of a cached quote without changing data.
func (s *Store) TouchQuote(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(
		"UPDATE quotes SET updated_at = ? WHERE code = ?",
		time.Now().Unix(), code,
	)
}

// TouchKLines updates the timestamp of cached K-lines without changing data.
func (s *Store) TouchKLines(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(
		"UPDATE klines SET updated_at = ? WHERE key = ?",
		time.Now().Unix(), key,
	)
}

// Lock acquires the write lock.
func (s *Store) Lock() {
	s.mu.Lock()
}

// Unlock releases the write lock.
func (s *Store) Unlock() {
	s.mu.Unlock()
}

// RLock acquires the read lock.
func (s *Store) RLock() {
	s.mu.RLock()
}

// RUnlock releases the read lock.
func (s *Store) RUnlock() {
	s.mu.RUnlock()
}

// DB returns the underlying *sql.DB for advanced operations.
func (s *Store) DB() *sql.DB {
	return s.db
}

// RefreshStale refreshes all quotes older than TTL.
func (s *Store) RefreshStale(fetcher func(code string) (*model.Quote, error)) int {
	codes := s.GetCachedCodes()
	refreshed := 0

	for _, code := range codes {
		age := s.QuoteAge(code)
		if age > s.quoteTTL {
			if quote, err := fetcher(code); err == nil {
				s.SetQuote(code, quote)
				refreshed++
			}
		}
	}
	return refreshed
}

// Preload preloads quotes for a list of codes.
func (s *Store) Preload(codes []string, fetcher func(code string) (*model.Quote, error)) int {
	loaded := 0
	for _, code := range codes {
		if !s.HasQuote(code) {
			if quote, err := fetcher(code); err == nil {
				s.SetQuote(code, quote)
				loaded++
			}
		}
	}
	return loaded
}

// PreloadKLines preloads K-lines for a list of codes.
func (s *Store) PreloadKLines(codes []string, period string, count int, fetcher func(code, period string, count int) ([]model.KLine, error)) int {
	loaded := 0
	for _, code := range codes {
		key := CacheKey(code, period)
		if !s.HasKLines(key) {
			if klines, err := fetcher(code, period, count); err == nil {
				s.SetKLines(key, klines)
				loaded++
			}
		}
	}
	return loaded
}

// GetCodeForName finds a stock code by name.
func (s *Store) GetCodeForName(name string) []string {
	codes := s.GetCachedCodes()
	var results []string

	for _, code := range codes {
		if quote, ok := s.GetQuoteRaw(code); ok {
			if strings.Contains(quote.Name, name) {
				results = append(results, code)
			}
		}
	}
	return results
}

// SearchCodes searches for stocks by code or name.
func (s *Store) SearchCodes(query string) []string {
	query = strings.ToUpper(query)

	codes := s.GetCachedCodes()
	var results []string

	for _, code := range codes {
		if strings.Contains(code, query) {
			results = append(results, code)
			continue
		}
		if quote, ok := s.GetQuoteRaw(code); ok {
			if strings.Contains(quote.Name, query) {
				results = append(results, code)
			}
		}
	}
	return results
}

// GetQuoteByCodeOrName finds a quote by code or name.
func (s *Store) GetQuoteByCodeOrName(query string) (*model.Quote, bool) {
	// Try as code first
	code := strings.ToUpper(query)
	if quote, ok := s.GetQuote(code); ok {
		return quote, true
	}

	// Try as name
	matches := s.GetCodeForName(query)
	if len(matches) > 0 {
		return s.GetQuote(matches[0])
	}
	return nil, false
}

// GetCacheEntry returns the cache entry for a code.
func (s *Store) GetCacheEntry(code string) map[string]interface{} {
	quote, age, ok := s.GetQuoteWithAge(code)
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"code":     code,
		"quote":    quote,
		"age":      age.String(),
		"is_fresh": age < s.quoteTTL,
	}
}

// GetCacheEntryForKLines returns the cache entry for K-lines.
func (s *Store) GetCacheEntryForKLines(code, period string) map[string]interface{} {
	key := CacheKey(code, period)
	klines, age, ok := s.GetKLinesWithAge(key)
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"code":     code,
		"period":   period,
		"klines":   klines,
		"age":      age.String(),
		"is_fresh": age < s.klineTTL,
	}
}

// TouchCode updates timestamps for all cache entries of a code.
func (s *Store) TouchCode(code string) {
	s.TouchQuote(code)

	keys := s.GetCachedKLinesKeys()
	for _, key := range keys {
		if strings.HasPrefix(key, code+":") {
			s.TouchKLines(key)
		}
	}
}
