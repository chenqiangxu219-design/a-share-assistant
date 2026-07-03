package cache

import (
	"encoding/json"
	"os"
	"time"
)

// NewInMemory creates a new in-memory store.
func NewInMemory() (*Store, error) {
	base := &Store{
		quoteTTL: 5 * time.Minute,
		klineTTL: 30 * time.Minute,
	}
	return base.WithMemory()
}

// NewInMemoryWithTTL creates an in-memory store with custom TTLs.
func NewInMemoryWithTTL(quoteTTL, klineTTL time.Duration) (*Store, error) {
	base := &Store{
		quoteTTL: quoteTTL,
		klineTTL: klineTTL,
	}
	return base.WithMemory()
}

// NewTemp creates a store in a temporary directory.
func NewTemp() (*Store, error) {
	base := &Store{
		quoteTTL: 5 * time.Minute,
		klineTTL: 30 * time.Minute,
	}
	return base.WithTempDir()
}

// NewTempWithTTL creates a temp store with custom TTLs.
func NewTempWithTTL(quoteTTL, klineTTL time.Duration) (*Store, error) {
	base := &Store{
		quoteTTL: quoteTTL,
		klineTTL: klineTTL,
	}
	return base.WithTempDir()
}

// NewDefault creates a store with default settings.
func NewDefault() (*Store, error) {
	return NewStore(DefaultDSN())
}

// New creates a store with the specified DSN.
func New(dsn string) (*Store, error) {
	return NewStore(dsn)
}

// NewWithConfig creates a store with configuration.
func NewWithConfig(dsn string, config map[string]interface{}) (*Store, error) {
	s, err := NewStore(dsn)
	if err != nil {
		return nil, err
	}
	_ = s.Configure(config)
	return s, nil
}

// NewFromFile creates a store from a config file.
func NewFromFile(configPath string) (*Store, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	dsn := DefaultDSN()
	if path, ok := config["db_path"].(string); ok && path != "" {
		dsn = path
	}

	return NewWithConfig(dsn, config)
}

// MustNew creates a store or panics.
func MustNew(dsn string) *Store {
	s, err := NewStore(dsn)
	if err != nil {
		panic(err)
	}
	return s
}

// MustNewDefault creates a default store or panics.
func MustNewDefault() *Store {
	return MustNew(DefaultDSN())
}

// MustNewInMemory creates an in-memory store or panics.
func MustNewInMemory() *Store {
	s, err := NewInMemory()
	if err != nil {
		panic(err)
	}
	return s
}

// MustNewTemp creates a temp store or panics.
func MustNewTemp() *Store {
	s, err := NewTemp()
	if err != nil {
		panic(err)
	}
	return s
}

// MustNewWithConfig creates a configured store or panics.
func MustNewWithConfig(dsn string, config map[string]interface{}) *Store {
	s, err := NewWithConfig(dsn, config)
	if err != nil {
		panic(err)
	}
	return s
}

// MustNewFromFile creates a store from file or panics.
func MustNewFromFile(configPath string) *Store {
	s, err := NewFromFile(configPath)
	if err != nil {
		panic(err)
	}
	return s
}

// NewOrPanic is an alias for MustNew.
func NewOrPanic(dsn string) *Store {
	return MustNew(dsn)
}
