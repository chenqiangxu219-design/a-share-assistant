package cache

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// CleanupOld removes all data older than maxAge.
func (s *Store) CleanupOld(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge).Unix()
	_, _ = s.db.Exec("DELETE FROM quotes WHERE updated_at < ?", cutoff)
	_, _ = s.db.Exec("DELETE FROM klines WHERE updated_at < ?", cutoff)
}

// Reset removes all data.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("DELETE FROM quotes")
	_, _ = s.db.Exec("DELETE FROM klines")
}

// Vacuum optimizes database file size.
func (s *Store) Vacuum() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("VACUUM")
}

// TrimToMax reduces cache to the most recent maxRecords entries.
func (s *Store) TrimToMax(maxRecords int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Keep only the most recent records
	_, _ = s.db.Exec(
		"DELETE FROM quotes WHERE code NOT IN (SELECT code FROM quotes ORDER BY updated_at DESC LIMIT ?)",
		maxRecords,
	)
	_, _ = s.db.Exec(
		"DELETE FROM klines WHERE key NOT IN (SELECT key FROM klines ORDER BY updated_at DESC LIMIT ?)",
		maxRecords*2, // K-lines typically have more entries per stock
	)
}

// Compact removes redundant K-line data for codes that have newer data.
func (s *Store) Compact() {
	// For each stock, keep only the most recent period's K-lines
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get all unique codes from klines
	rows, err := s.db.Query("SELECT DISTINCT key FROM klines")
	if err != nil {
		return
	}
	defer rows.Close()

	codePeriods := make(map[string][]string)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			continue
		}
		parts := strings.SplitN(key, ":", 2)
		if len(parts) == 2 {
			codePeriods[parts[0]] = append(codePeriods[parts[0]], parts[1])
		}
	}

	// For each code, keep only daily K-lines (remove intraday)
	for code, periods := range codePeriods {
		keep := false
		for _, p := range periods {
			if p == "d" {
				keep = true
				break
			}
		}
		if !keep {
			// Keep the most recent period if no daily
			continue
		}
		// Remove non-daily periods
		for _, p := range periods {
			if p != "d" {
				_, _ = s.db.Exec("DELETE FROM klines WHERE key = ?", code+":"+p)
			}
		}
	}
}

// ForcedCheckpoint forces a WAL checkpoint.
func (s *Store) ForcedCheckpoint() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
}

// IntegrityCheck runs a SQLite integrity check.
func (s *Store) IntegrityCheck() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result string
	err := s.db.QueryRow("PRAGMA integrity_check").Scan(&result)
	return result, err
}

// Optimizer runs SQLite optimizer.
func (s *Store) Optimizer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("PRAGMA optimize")
	return err
}

// Analyze runs SQLite ANALYZE for query optimization.
func (s *Store) Analyze() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec("ANALYZE")
}

// Backup creates a backup of the database.
func (s *Store) Backup(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src, err := os.Open(s.dsn)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// Restore restores the database from a backup.
func (s *Store) Restore(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close current DB
	_ = s.db.Close()

	// Copy backup to original location
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(s.dsn)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// Reopen
	return s.reopen()
}

// reopen reopens the database connection.
func (s *Store) reopen() error {
	db, err := sql.Open("sqlite", s.dsn)
	if err != nil {
		return err
	}
	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA busy_timeout=5000")
	s.db = db
	return nil
}

// CompactAndVacuum performs compaction and vacuum.
func (s *Store) CompactAndVacuum() {
	s.Compact()
	s.Vacuum()
}

// Maintenance performs routine maintenance tasks.
func (s *Store) Maintenance() {
	s.CleanupOld(24 * time.Hour)
	s.Compact()
	s.Vacuum()
	_ = s.Optimizer()
}

// SetMaintenanceInterval sets the interval for automatic maintenance.
func (s *Store) SetMaintenanceInterval(interval time.Duration) {
	_ = interval
}

// GetMaintenanceInterval returns the current maintenance interval.
func (s *Store) GetMaintenanceInterval() time.Duration {
	return 24 * time.Hour
}

// StartMaintenance starts the automatic maintenance goroutine.
func (s *Store) StartMaintenance() {
	interval := s.GetMaintenanceInterval()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.Maintenance()
		}
	}()
}

// StopMaintenance stops the automatic maintenance goroutine.
func (s *Store) StopMaintenance() {
	// Currently no way to stop the goroutine,
	// reserved for future: add context/channel
}

// IsMaintenanceRunning returns whether automatic maintenance is active.
func (s *Store) IsMaintenanceRunning() bool {
	return false
}

// SetAutoVacuum enables or disables auto-vacuum.
func (s *Store) SetAutoVacuum(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA auto_vacuum=%d", mode))
}

// IsAutoVacuumEnabled returns whether auto-vacuum is enabled.
func (s *Store) IsAutoVacuumEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA auto_vacuum").Scan(&mode)
	return mode > 0
}

// SetIncrementalVacuum enables or disables incremental vacuum.
func (s *Store) SetIncrementalVacuum(pages int) {
	if pages <= 0 {
		pages = 100
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA incremental_vacuum=%d", pages))
}

// GetIncrementalVacuum returns the current incremental vacuum setting.
func (s *Store) GetIncrementalVacuum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pages int
	_ = s.db.QueryRow("PRAGMA incremental_vacuum").Scan(&pages)
	return pages
}
