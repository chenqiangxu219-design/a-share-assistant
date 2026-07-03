package model

import "time"

// AlertType represents the type of alert
type AlertType string

const (
	AlertVolumeSurge    AlertType = "volume_surge"
	AlertPriceSpike     AlertType = "price_spike"
	AlertLimitUp        AlertType = "limit_up"
	AlertLimitDown      AlertType = "limit_down"
	AlertMACDDivergence AlertType = "macd_divergence"
)

// AlertSeverity represents the severity level
type AlertSeverity string

const (
	AlertInfo     AlertSeverity = "info"
	AlertWarning  AlertSeverity = "warning"
	AlertCritical AlertSeverity = "critical"
)

// Alert represents an abnormal market alert
type Alert struct {
	Code     string         `json:"code"`
	Name     string         `json:"name"`
	Type     AlertType      `json:"type"`
	Message  string         `json:"message"`
	Severity AlertSeverity  `json:"severity"`
	Time     time.Time      `json:"time"`
}

// AlertMonitorConfig holds monitoring configuration
type AlertMonitorConfig struct {
	Codes      []string `json:"codes"`
	Enabled    bool     `json:"enabled"`
	IntervalMs int      `json:"interval_ms"`
}
