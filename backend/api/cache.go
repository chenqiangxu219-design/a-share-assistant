package api

import (
	"time"

	"a-share-assistant/backend/cache"
	"a-share-assistant/backend/model"
)

// QuoteCache is a facade over the SQLite-backed cache store.
// Method signatures are unchanged for compatibility with existing handlers.
type QuoteCache struct {
	store *cache.Store
}

// NewQuoteCache creates a cache with the given quote TTL.
// K-line TTL defaults to 30 minutes.
func NewQuoteCache(store *cache.Store) *QuoteCache {
	return &QuoteCache{store: store}
}

// CacheKey combines code and period into a cache key.
func CacheKey(code, period string) string {
	return code + ":" + period
}

func (c *QuoteCache) GetQuote(code string) (*model.Quote, bool) {
	return c.store.GetQuote(code)
}

func (c *QuoteCache) SetQuote(code string, q *model.Quote) {
	c.store.SetQuote(code, q)
}

func (c *QuoteCache) GetKLines(key string) ([]model.KLine, bool) {
	return c.store.GetKLines(key)
}

func (c *QuoteCache) SetKLines(key string, klines []model.KLine) {
	c.store.SetKLines(key, klines)
}

// GetCachedCodes returns all cached stock codes within TTL.
func (c *QuoteCache) GetCachedCodes() []string {
	return c.store.GetCachedCodes()
}

// GetAllQuotes returns all cached quotes (for initial frontend load).
func (c *QuoteCache) GetAllQuotes() map[string]*model.Quote {
	return c.store.GetAllQuotes()
}

// GetQuoteRaw returns the cached quote even if expired (for stale-while-revalidate).
func (c *QuoteCache) GetQuoteRaw(code string) (*model.Quote, bool) {
	return c.store.GetQuoteRaw(code)
}

// GetKLinesRaw returns cached K-lines even if expired.
func (c *QuoteCache) GetKLinesRaw(key string) ([]model.KLine, bool) {
	return c.store.GetKLinesRaw(key)
}

// Stats returns cache statistics.
func (c *QuoteCache) Stats() map[string]interface{} {
	return c.store.CacheStats()
}

// ClearExpired removes all expired records.
func (c *QuoteCache) ClearExpired() {
	c.store.ClearExpired()
}

// Maintenance performs cleanup of old records (24h+).
func (c *QuoteCache) Maintenance() {
	c.store.CleanupOld(24 * time.Hour)
}

// Count returns the number of cached quotes and K-line sets.
func (c *QuoteCache) Count() (int, int) {
	return c.store.Count()
}

// UniqueStocks returns the number of unique stocks cached.
func (c *QuoteCache) UniqueStocks() int {
	return c.store.UniqueStocks()
}

// InvalidateCode removes all cache entries for a stock code.
func (c *QuoteCache) InvalidateCode(code string) {
	c.store.InvalidateCode(code)
}

// RefreshBatch updates quotes for multiple codes atomically.
func (c *QuoteCache) RefreshBatch(quotes map[string]*model.Quote) {
	c.store.RefreshBatch(quotes)
}
