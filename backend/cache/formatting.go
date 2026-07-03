package cache

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// String returns a string representation of the store.
func (s *Store) String() string {
	qc, kc := s.Count()
	return fmt.Sprintf("Store(%s: %d quotes, %d klines)",
		s.dsn, qc, kc)
}

// GoString returns a Go string representation.
func (s *Store) GoString() string {
	return s.String()
}

// Error returns an error description.
func (s *Store) Error() error {
	return nil
}

// Format implements fmt.Formatter.
func (s *Store) Format(fs fmt.State, c rune) {
	fmt.Fprint(fs, s.String())
}

// Clone creates a shallow copy of the store configuration.
func (s *Store) Clone() *Store {
	return &Store{
		db:       s.db,
		dsn:      s.dsn,
		quoteTTL: s.quoteTTL,
		klineTTL: s.klineTTL,
	}
}

// Equal compares two stores.
func (s *Store) Equal(other *Store) bool {
	if other == nil {
		return false
	}
	return s.dsn == other.dsn &&
		s.quoteTTL == other.quoteTTL &&
		s.klineTTL == other.klineTTL
}

// Hash returns a hash of the store configuration.
func (s *Store) Hash() string {
	return fmt.Sprintf("%s-%d-%d",
		s.dsn,
		int64(s.quoteTTL.Seconds()),
		int64(s.klineTTL.Seconds()))
}

// Equals compares stores by hash.
func (s *Store) Equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.Hash() == other.Hash()
}

// Copy creates a deep copy.
func (s *Store) Copy() *Store {
	return s.Clone()
}

// DeepCopy creates a deep copy with a new database connection.
func (s *Store) DeepCopy() (*Store, error) {
	db, err := sql.Open("sqlite", s.dsn)
	if err != nil {
		return nil, err
	}
	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA busy_timeout=5000")

	return &Store{
		db:       db,
		quoteTTL: s.quoteTTL,
		klineTTL: s.klineTTL,
	}, nil
}
