package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"a-share-assistant/backend/model"
	"a-share-assistant/backend/strategy"

	"github.com/gin-gonic/gin"
)

// AlertHandler manages alert-related API endpoints
type AlertHandler struct {
	DS     *Handler
	Engine *strategy.AlertEngine
	mu     sync.RWMutex
	alerts []model.Alert
}

// NewAlertHandler creates a new AlertHandler
func NewAlertHandler(handler *Handler) *AlertHandler {
	return &AlertHandler{
		DS:     handler,
		Engine: strategy.NewAlertEngine(),
		alerts: make([]model.Alert, 0, 100),
	}
}

// GetAlerts returns alerts for given stock codes
// GET /api/alerts?codes=sh600519,sz000001
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	codesParam := c.Query("codes")
	if codesParam == "" {
		c.JSON(http.StatusOK, h.GetRecentAlerts(50))
		return
	}

	codes := strings.Split(codesParam, ",")
	var alerts []model.Alert

	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}

		quote, err := h.DS.DS.GetQuote(code)
		if err != nil {
			continue
		}

		klines, err := h.DS.DS.GetKLines(code, "d", 100)
		if err != nil {
			continue
		}

		ind := h.DS.CalcIndicators(klines)
		newAlerts := h.Engine.RunAlerts(code, quote.Name, *quote, ind)
		alerts = append(alerts, newAlerts...)
	}

	h.mu.Lock()
	h.alerts = append(h.alerts, alerts...)
	if len(h.alerts) > 200 {
		h.alerts = h.alerts[len(h.alerts)-200:]
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, alerts)
}

// GetRecentAlerts returns the most recent alerts
func (h *AlertHandler) GetRecentAlerts(n int) []model.Alert {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.alerts) == 0 {
		return []model.Alert{}
	}

	start := 0
	if n < len(h.alerts) {
		start = len(h.alerts) - n
	}
	return h.alerts[start:]
}

// PostMonitorConfig updates monitoring configuration
// POST /api/alerts/monitor
func (h *AlertHandler) PostMonitorConfig(c *gin.Context) {
	var config model.AlertMonitorConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if config.IntervalMs <= 0 {
		config.IntervalMs = 3000
	}

	h.mu.Lock()
	h.alerts = h.alerts[:0]
	h.mu.Unlock()

	// Start background monitoring goroutine
	go h.runMonitor(config)

	c.JSON(http.StatusOK, gin.H{
		"message": "监控已启动",
		"config":  config,
	})
}

// runMonitor periodically checks alerts for configured codes
func (h *AlertHandler) runMonitor(config model.AlertMonitorConfig) {
	interval := time.Duration(config.IntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		codes := config.Codes
		h.mu.RUnlock()

		if len(codes) == 0 {
			continue
		}

		var alerts []model.Alert
		for _, code := range codes {
			code = strings.ToUpper(strings.TrimSpace(code))
			if code == "" {
				continue
			}

			quote, err := h.DS.DS.GetQuote(code)
			if err != nil {
				continue
			}

			klines, err := h.DS.DS.GetKLines(code, "d", 100)
			if err != nil {
				continue
			}

			ind := h.DS.CalcIndicators(klines)
			newAlerts := h.Engine.RunAlerts(code, quote.Name, *quote, ind)
			alerts = append(alerts, newAlerts...)
		}

		h.mu.Lock()
		h.alerts = append(h.alerts, alerts...)
		if len(h.alerts) > 200 {
			h.alerts = h.alerts[len(h.alerts)-200:]
		}
		h.mu.Unlock()

		// Broadcast via WebSocket
		h.DS.Hub.BroadcastAlerts(alerts)
	}
}

// RegisterAlertRoutes registers alert monitoring routes
func (h *AlertHandler) RegisterAlertRoutes(g *gin.RouterGroup) {
	g.GET("/alerts", h.GetAlerts)
	g.POST("/alerts/monitor", h.PostMonitorConfig)
}
