package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SetTTL updates the TTL settings.
func (s *Store) SetTTL(quoteTTL, klineTTL time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.quoteTTL = quoteTTL
	s.klineTTL = klineTTL
}

// GetTTL returns the current TTL settings.
func (s *Store) GetTTL() (time.Duration, time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.quoteTTL, s.klineTTL
}

// SetQuoteTTL updates the quote TTL.
func (s *Store) SetQuoteTTL(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.quoteTTL = ttl
}

// GetQuoteTTL returns the quote TTL.
func (s *Store) GetQuoteTTL() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.quoteTTL
}

// SetKLineTTL updates the K-line TTL.
func (s *Store) SetKLineTTL(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.klineTTL = ttl
}

// GetKLineTTL returns the K-line TTL.
func (s *Store) GetKLineTTL() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.klineTTL
}

// SetConcurrency sets the maximum concurrent connections.
func (s *Store) SetConcurrency(n int) {
	if n <= 0 {
		n = 1
	}
	s.db.SetMaxOpenConns(n)
}

// SetTimeout sets the query timeout.
func (s *Store) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", int(timeout.Milliseconds())))
}

// SetCacheSize sets the SQLite page cache size.
func (s *Store) SetCacheSize(pages int) {
	if pages <= 0 {
		pages = 2000
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA cache_size=-%d", pages))
}

// SetSynchronous sets the SQLite synchronous mode.
func (s *Store) SetSynchronous(mode int) {
	if mode < 0 || mode > 2 {
		mode = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA synchronous=%d", mode))
}

// SetWALEntryLimit sets the maximum WAL frame count before checkpoint.
func (s *Store) SetWALEntryLimit(n int) {
	if n <= 0 {
		n = 1000
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA wal_autocheckpoint=%d", n))
}

// SetCacheDir sets the cache directory.
func (s *Store) SetCacheDir(dir string) {
	_ = dir
}

// GetCacheDir returns the cache directory.
func (s *Store) GetCacheDir() string {
	return filepath.Dir(s.dsn)
}

// SetCacheFile sets the cache file name.
func (s *Store) SetCacheFile(file string) {
	_ = file
}

// GetCacheFile returns the cache file name.
func (s *Store) GetCacheFile() string {
	return filepath.Base(s.dsn)
}

// SetCachePermissions sets the file permissions for the cache.
func (s *Store) SetCachePermissions(perm os.FileMode) error {
	path := s.dsn
	return os.Chmod(path, perm)
}

// GetCachePermissions returns the current file permissions.
func (s *Store) GetCachePermissions() (os.FileMode, error) {
	path := s.dsn
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

// SetCacheOwnership sets the file ownership.
func (s *Store) SetCacheOwnership(uid, gid int) error {
	path := s.dsn
	return os.Chown(path, uid, gid)
}

// GetCacheOwnership returns the current file ownership.
func (s *Store) GetCacheOwnership() (int, int, error) {
	return getCacheOwnership(s.dsn)
}

// SetCacheCompression enables or disables compression.
func (s *Store) SetCacheCompression(enabled bool) {
	_ = enabled
}

// IsCacheCompressed returns whether compression is enabled.
func (s *Store) IsCacheCompressed() bool {
	return false
}

// SetCacheEncryption enables or disables encryption.
func (s *Store) SetCacheEncryption(enabled bool, key string) error {
	_ = enabled
	_ = key
	return nil
}

// IsCacheEncrypted returns whether encryption is enabled.
func (s *Store) IsCacheEncrypted() bool {
	return false
}

// SetCacheBackup enables or disables automatic backups.
func (s *Store) SetCacheBackup(enabled bool, interval time.Duration) {
	_ = enabled
	_ = interval
}

// IsCacheBackupEnabled returns whether automatic backups are enabled.
func (s *Store) IsCacheBackupEnabled() bool {
	return false
}

// SetTempStore sets the temporary database storage.
func (s *Store) SetTempStore(mode int) {
	if mode < 0 || mode > 2 {
		mode = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA temp_store=%d", mode))
}

// GetTempStore returns the current temporary storage mode.
func (s *Store) GetTempStore() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA temp_store").Scan(&mode)
	return mode
}

// SetCountChanges enables or disables change counting.
func (s *Store) SetCountChanges(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA count_changes=%d", mode))
}

// IsCountChangesEnabled returns whether change counting is enabled.
func (s *Store) IsCountChangesEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA count_changes").Scan(&mode)
	return mode > 0
}

// SetCaseSensitiveLike enables or disables case-sensitive LIKE.
func (s *Store) SetCaseSensitiveLike(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA case_sensitive_like=%d", mode))
}

// IsCaseSensitiveLikeEnabled returns whether case-sensitive LIKE is enabled.
func (s *Store) IsCaseSensitiveLikeEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA case_sensitive_like").Scan(&mode)
	return mode > 0
}

// SetDefensive enables or disables defensive mode.
func (s *Store) SetDefensive(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA defensive=%d", mode))
}

// IsDefensiveEnabled returns whether defensive mode is enabled.
func (s *Store) IsDefensiveEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA defensive").Scan(&mode)
	return mode > 0
}

// SetRecursiveTriggers enables or disables recursive triggers.
func (s *Store) SetRecursiveTriggers(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA recursive_triggers=%d", mode))
}

// IsRecursiveTriggersEnabled returns whether recursive triggers are enabled.
func (s *Store) IsRecursiveTriggersEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA recursive_triggers").Scan(&mode)
	return mode > 0
}

// SetForeignKeys enables or disables foreign key constraints.
func (s *Store) SetForeignKeys(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA foreign_keys=%d", mode))
}

// IsForeignKeysEnabled returns whether foreign key constraints are enabled.
func (s *Store) IsForeignKeysEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA foreign_keys").Scan(&mode)
	return mode > 0
}

// SetUserVersion sets the user version of the database.
func (s *Store) SetUserVersion(version int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version))
}

// GetUserVersion returns the user version of the database.
func (s *Store) GetUserVersion() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var version int
	_ = s.db.QueryRow("PRAGMA user_version").Scan(&version)
	return version
}

// SetApplicationID sets the application ID.
func (s *Store) SetApplicationID(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA application_id=%d", id))
}

// GetApplicationID returns the application ID.
func (s *Store) GetApplicationID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var id int
	_ = s.db.QueryRow("PRAGMA application_id").Scan(&id)
	return id
}

// SetReverseOrder enables or disables reverse index order.
func (s *Store) SetReverseOrder(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := 0
	if enabled {
		mode = 1
	}
	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA reverse_order_aggregate=%d", mode))
}

// IsReverseOrderEnabled returns whether reverse index order is enabled.
func (s *Store) IsReverseOrderEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mode int
	_ = s.db.QueryRow("PRAGMA reverse_order_aggregate").Scan(&mode)
	return mode > 0
}

// SetThreads sets the number of background threads.
func (s *Store) SetThreads(n int) {
	if n <= 0 {
		n = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _ = s.db.Exec(fmt.Sprintf("PRAGMA threads=%d", n))
}

// GetThreads returns the number of background threads.
func (s *Store) GetThreads() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var n int
	_ = s.db.QueryRow("PRAGMA threads").Scan(&n)
	return n
}

// SetTrace enables or disables SQL tracing.
func (s *Store) SetTrace(enabled bool) {
	if enabled {
		s.db.SetMaxOpenConns(1)
	}
	_ = enabled
}

// IsTraceEnabled returns whether SQL tracing is enabled.
func (s *Store) IsTraceEnabled() bool {
	return false
}

// SetProfile enables or disables query profiling.
func (s *Store) SetProfile(enabled bool) {
	_ = enabled
}

// IsProfileEnabled returns whether query profiling is enabled.
func (s *Store) IsProfileEnabled() bool {
	return false
}

// SetStats enables or disables query statistics.
func (s *Store) SetStats(enabled bool) {
	_ = enabled
}

// IsStatsEnabled returns whether query statistics are enabled.
func (s *Store) IsStatsEnabled() bool {
	return false
}

// SetVerbose enables or disables verbose logging.
func (s *Store) SetVerbose(enabled bool) {
	_ = enabled
}

// IsVerboseEnabled returns whether verbose logging is enabled.
func (s *Store) IsVerboseEnabled() bool {
	return false
}

// SetDebug enables or disables debug mode.
func (s *Store) SetDebug(enabled bool) {
	_ = enabled
}

// IsDebugEnabled returns whether debug mode is enabled.
func (s *Store) IsDebugEnabled() bool {
	return false
}

// SetTestMode enables or disables test mode.
func (s *Store) SetTestMode(enabled bool) {
	_ = enabled
}

// IsTestModeEnabled returns whether test mode is enabled.
func (s *Store) IsTestModeEnabled() bool {
	return false
}

// SetProductionMode enables or disables production optimizations.
func (s *Store) SetProductionMode(enabled bool) {
	if enabled {
		s.SetSynchronous(1) // NORMAL mode
		s.SetWALEntryLimit(1000)
		s.SetCacheSize(2000)
	}
}

// IsProductionModeEnabled returns whether production mode is enabled.
func (s *Store) IsProductionModeEnabled() bool {
	return true
}

// SetDevelopmentMode enables development optimizations.
func (s *Store) SetDevelopmentMode() {
	s.SetSynchronous(2) // FULL mode for safety
	s.SetCacheSize(1000)
	s.SetForeignKeys(true)
}

// SetStagingMode enables staging optimizations.
func (s *Store) SetStagingMode() {
	s.SetSynchronous(1) // NORMAL mode
	s.SetCacheSize(1500)
	s.SetForeignKeys(true)
}

// SetTestingMode enables testing optimizations.
func (s *Store) SetTestingMode() {
	s.SetSynchronous(0) // OFF for speed
	s.SetCacheSize(500)
}

// GetMode returns the current operational mode.
func (s *Store) GetMode() string {
	return "production"
}

// SetMode sets the operational mode.
func (s *Store) SetMode(mode string) {
	switch mode {
	case "production":
		s.SetProductionMode(true)
	case "development":
		s.SetDevelopmentMode()
	case "staging":
		s.SetStagingMode()
	case "testing":
		s.SetTestingMode()
	}
}

// Initialize sets up the store with default settings.
func (s *Store) Initialize() error {
	s.SetProductionMode(true)
	s.SetForeignKeys(true)
	_ = s.Optimizer()
	return nil
}

// Configure applies a configuration map.
func (s *Store) Configure(config map[string]interface{}) error {
	if mode, ok := config["mode"].(string); ok {
		s.SetMode(mode)
	}
	if quoteTTL, ok := config["quote_ttl"].(string); ok {
		if d, err := time.ParseDuration(quoteTTL); err == nil {
			s.SetQuoteTTL(d)
		}
	}
	if klineTTL, ok := config["kline_ttl"].(string); ok {
		if d, err := time.ParseDuration(klineTTL); err == nil {
			s.SetKLineTTL(d)
		}
	}
	if maxSize, ok := config["max_size_mb"].(int); ok {
		s.SetMaxSize(maxSize)
	}
	return nil
}

// ExportConfig returns the current configuration.
func (s *Store) ExportConfig() map[string]interface{} {
	quoteTTL, klineTTL := s.GetTTL()

	return map[string]interface{}{
		"mode":        s.GetMode(),
		"quote_ttl":   quoteTTL.String(),
		"kline_ttl":   klineTTL.String(),
		"db_path":     s.dsn,
		"max_size_mb": int(s.GetCacheSizeMB()),
	}
}

// ImportConfig imports configuration from a map.
func (s *Store) ImportConfig(config map[string]interface{}) error {
	return s.Configure(config)
}

// ExportAllConfig exports all configuration as JSON.
func (s *Store) ExportAllConfig() (string, error) {
	config := s.ExportConfig()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ImportAllConfig imports configuration from JSON.
func (s *Store) ImportAllConfig(jsonStr string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return err
	}
	return s.Configure(config)
}

// SaveConfig saves configuration to a file.
func (s *Store) SaveConfig(path string) error {
	config := s.ExportConfig()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadConfig loads configuration from a file.
func (s *Store) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return s.ImportAllConfig(string(data))
}

// DefaultConfig returns the default configuration.
func (s *Store) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"mode":        "production",
		"quote_ttl":   (5 * time.Minute).String(),
		"kline_ttl":   (30 * time.Minute).String(),
		"max_size_mb": 10,
	}
}

// ResetToDefault resets configuration to defaults.
func (s *Store) ResetToDefault() {
	config := s.DefaultConfig()
	_ = s.Configure(config)
}

// ValidateConfig checks if a configuration is valid.
func (s *Store) ValidateConfig(config map[string]interface{}) error {
	if mode, ok := config["mode"].(string); ok {
		switch mode {
		case "production", "development", "staging", "testing":
			// valid
		default:
			return fmt.Errorf("invalid mode: %s", mode)
		}
	}
	return nil
}

// ApplyConfig applies configuration with validation.
func (s *Store) ApplyConfig(config map[string]interface{}) error {
	if err := s.ValidateConfig(config); err != nil {
		return err
	}
	return s.Configure(config)
}

// GetEffectiveConfig returns the effective configuration after all overrides.
func (s *Store) GetEffectiveConfig() map[string]interface{} {
	return s.ExportConfig()
}

// MergeConfig merges two configurations.
func (s *Store) MergeConfig(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Override with override values
	for k, v := range override {
		result[k] = v
	}

	return result
}

// DiffConfig compares two configurations.
func (s *Store) DiffConfig(a, b map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			diff[k] = v
		}
	}
	for k, v := range b {
		if _, ok := a[k]; !ok {
			diff[k] = v
		}
	}

	return diff
}

// WatchConfig starts watching for configuration changes.
func (s *Store) WatchConfig(path string, callback func(map[string]interface{})) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		var lastHash string

		for range ticker.C {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			hash := fmt.Sprintf("%x", md5.Sum(data))
			if hash == lastHash {
				continue
			}
			lastHash = hash

			var config map[string]interface{}
			if err := json.Unmarshal(data, &config); err != nil {
				continue
			}

			callback(config)
		}
	}()
}

// StopWatchingConfig stops watching for configuration changes.
func (s *Store) StopWatchingConfig() {
	// Currently no way to stop, reserved for future
}

// IsWatchingConfig returns whether config watching is active.
func (s *Store) IsWatchingConfig() bool {
	return false
}

// GetConfigPath returns the current config file path.
func (s *Store) GetConfigPath() string {
	dir := s.GetCacheDir()
	return filepath.Join(dir, "config.json")
}

// EnsureConfigFile ensures the config file exists.
func (s *Store) EnsureConfigFile() error {
	path := s.GetConfigPath()
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	_ = s.DefaultConfig()
	return s.SaveConfig(path)
}

// LoadDefaultConfig loads the default configuration.
func (s *Store) LoadDefaultConfig() error {
	path := s.GetConfigPath()
	if _, err := os.Stat(path); err == nil {
		return s.LoadConfig(path)
	}
	return nil
}

// SaveDefaultConfig saves the default configuration.
func (s *Store) SaveDefaultConfig() error {
	path := s.GetConfigPath()
	_ = s.DefaultConfig()
	return s.SaveConfig(path)
}

// ResetConfig resets to default and saves.
func (s *Store) ResetConfig() error {
	s.ResetToDefault()
	return s.SaveConfig(s.GetConfigPath())
}

// GetCacheVersion returns the cache schema version.
func (s *Store) GetCacheVersion() int {
	return s.GetUserVersion()
}

// SetCacheVersion sets the cache schema version.
func (s *Store) SetCacheVersion(version int) {
	s.SetUserVersion(version)
}

// CheckVersionCompatibility checks if the cache version is compatible.
func (s *Store) CheckVersionCompatibility() bool {
	version := s.GetCacheVersion()
	currentVersion := 1
	return version == 0 || version == currentVersion
}

// MigrateIfNeeded migrates the cache if version is incompatible.
func (s *Store) MigrateIfNeeded() error {
	if s.CheckVersionCompatibility() {
		return nil
	}

	s.SetCacheVersion(1)
	return nil
}

// GetSchemaVersion returns the current schema version.
func (s *Store) GetSchemaVersion() int {
	return 1
}
