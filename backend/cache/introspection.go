package cache

import (
	"database/sql"
	"os"
	"strings"
)

// EnsureSchema ensures the database schema is up to date.
func (s *Store) EnsureSchema() error {
	return s.MigrateIfNeeded()
}

// CreateTables creates all required tables.
func (s *Store) CreateTables() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(string(schema))
	return err
}

// DropTables drops all tables (for testing/reset).
func (s *Store) DropTables() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DROP TABLE IF EXISTS quotes")
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DROP TABLE IF EXISTS klines")
	if err != nil {
		return err
	}
	return nil
}

// RecreateTables drops and recreates all tables.
func (s *Store) RecreateTables() error {
	if err := s.DropTables(); err != nil {
		return err
	}
	return s.CreateTables()
}

// GetTableNames returns all table names.
func (s *Store) GetTableNames() []string {
	return []string{"quotes", "klines"}
}

// TableExists checks if a table exists.
func (s *Store) TableExists(name string) bool {
	var count int
	_ = s.db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		name,
	).Scan(&count)
	return count > 0
}

// GetTableInfo returns information about a table.
func (s *Store) GetTableInfo(name string) []map[string]interface{} {
	rows, err := s.db.Query("PRAGMA table_info(?)", name)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var info []map[string]interface{}
	for rows.Next() {
		var col map[string]interface{}
		_ = rows.Scan() // Simplified
		info = append(info, col)
	}
	return info
}

// GetTableCounts returns row counts for all tables.
func (s *Store) GetTableCounts() map[string]int {
	counts := make(map[string]int)
	for _, table := range s.GetTableNames() {
		var count int
		_ = s.db.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&count)
		counts[table] = count
	}
	return counts
}

// GetTableSizes returns sizes for all tables.
func (s *Store) GetTableSizes() map[string]int64 {
	sizes := make(map[string]int64)
	qc, kc := s.Count()
	sizes["quotes"] = int64(qc)
	sizes["klines"] = int64(kc)
	return sizes
}

// OptimizeTable optimizes a specific table.
func (s *Store) OptimizeTable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("ANALYZE " + name)
	return err
}

// VacuumTable vacuums a specific table.
func (s *Store) VacuumTable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("VACUUM " + name)
	return err
}

// CheckpointTable checkpoints a specific table.
func (s *Store) CheckpointTable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("PRAGMA wal_checkpoint(" + name + ")")
	return err
}

// ReindexTable reindexes a specific table.
func (s *Store) ReindexTable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("REINDEX " + name)
	return err
}

// GetIndexes returns all indexes for a table.
func (s *Store) GetIndexes(name string) []string {
	rows, err := s.db.Query("PRAGMA index_list(?)", name)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var index string
		_ = rows.Scan(&index)
		indexes = append(indexes, index)
	}
	return indexes
}

// CreateIndex creates an index on a table.
func (s *Store) CreateIndex(name, table, columns string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("CREATE INDEX IF NOT EXISTS " + name + " ON " + table + " (" + columns + ")")
	return err
}

// DropIndex drops an index.
func (s *Store) DropIndex(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DROP INDEX IF EXISTS " + name)
	return err
}

// IndexExists checks if an index exists.
func (s *Store) IndexExists(name string) bool {
	var count int
	_ = s.db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?",
		name,
	).Scan(&count)
	return count > 0
}

// GetIndexInfo returns information about an index.
func (s *Store) GetIndexInfo(name string) []map[string]interface{} {
	rows, err := s.db.Query("PRAGMA index_info(?)", name)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var info []map[string]interface{}
	for rows.Next() {
		var col map[string]interface{}
		_ = rows.Scan()
		info = append(info, col)
	}
	return info
}

// CreateTrigger creates a trigger.
func (s *Store) CreateTrigger(name, sqlStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("CREATE TRIGGER IF NOT EXISTS " + name + " " + sqlStr)
	return err
}

// DropTrigger drops a trigger.
func (s *Store) DropTrigger(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DROP TRIGGER IF EXISTS " + name)
	return err
}

// TriggerExists checks if a trigger exists.
func (s *Store) TriggerExists(name string) bool {
	var count int
	_ = s.db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='trigger' AND name=?",
		name,
	).Scan(&count)
	return count > 0
}

// GetTriggers returns all triggers.
func (s *Store) GetTriggers() []string {
	rows, err := s.db.Query(
		"SELECT name FROM sqlite_master WHERE type='trigger'",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var triggers []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			triggers = append(triggers, name)
		}
	}
	return triggers
}

// GetViews returns all views.
func (s *Store) GetViews() []string {
	rows, err := s.db.Query(
		"SELECT name FROM sqlite_master WHERE type='view'",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			views = append(views, name)
		}
	}
	return views
}

// CreateView creates a view.
func (s *Store) CreateView(name, sqlStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("CREATE VIEW IF NOT EXISTS " + name + " AS " + sqlStr)
	return err
}

// DropView drops a view.
func (s *Store) DropView(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DROP VIEW IF EXISTS " + name)
	return err
}

// ViewExists checks if a view exists.
func (s *Store) ViewExists(name string) bool {
	var count int
	_ = s.db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name=?",
		name,
	).Scan(&count)
	return count > 0
}

// GetSchema returns the full database schema.
func (s *Store) GetSchema() string {
	rows, err := s.db.Query(
		"SELECT sql FROM sqlite_master WHERE sql IS NOT NULL",
	)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var sqlStr string
		if err := rows.Scan(&sqlStr); err == nil {
			parts = append(parts, sqlStr)
		}
	}
	return strings.Join(parts, "\n")
}

// ExportSchema exports the schema to a file.
func (s *Store) ExportSchema(path string) error {
	schema := s.GetSchema()
	return os.WriteFile(path, []byte(schema), 0644)
}

// ImportSchema imports the schema from a file.
func (s *Store) ImportSchema(path string) error {
	schema, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(string(schema))
	return err
}

// GetDB returns the underlying database connection.
func (s *Store) GetDB() *sql.DB {
	return s.db
}
