package api

import (
	"net/http"
	"strconv"
	"strings"

	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/model"

	"github.com/gin-gonic/gin"
)

// GetCapitalFlow handles GET /api/capital/:code?days=5
func (h *Handler) GetCapitalFlow(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "5"))

	if days > 30 {
		days = 30
	}

	// Try cache first (K-lines cached for 30 min)
	cacheKey := CacheKey(code, "d")
	if cached, ok := h.Cache.GetKLines(cacheKey); ok {
		flows := estimateCapitalFlow(cached)
		c.JSON(http.StatusOK, flows)
		return
	}

	// Fallback: fetch from data source
	klines, err := h.DS.GetKLines(code, "d", days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	h.Cache.SetKLines(cacheKey, klines)

	flows := estimateCapitalFlow(klines)
	c.JSON(http.StatusOK, flows)
}

// estimateCapitalFlow calculates flow from K-line data
func estimateCapitalFlow(klines []model.KLine) []datasource.CapitalFlow {
	var flows []datasource.CapitalFlow
	for _, k := range klines {
		priceChange := k.Close - k.Open
		amount := priceChange * k.Volume
		mainNet := amount * 0.6
		totalAmount := k.Volume * k.Close
		mainNetPct := 0.0
		if totalAmount > 0 {
			mainNetPct = mainNet / totalAmount * 100
		}
		flows = append(flows, datasource.CapitalFlow{
			Date:       k.Date,
			MainNet:    mainNet,
			MainNetPct: mainNetPct,
			LargeNet:   amount * 0.4,
			MediumNet:  -amount * 0.3,
			SmallNet:   -amount * 0.1,
		})
	}
	return flows
}

// GetCapitalFlowSummary handles GET /api/capital/summary?codes=600519,000858
func (h *Handler) GetCapitalFlowSummary(c *gin.Context) {
	codesParam := c.Query("codes")
	if codesParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codes parameter required"})
		return
	}

	codes := strings.Split(codesParam, ",")

	type FlowSummary struct {
		Code    string  `json:"code"`
		MainNet float64 `json:"main_net"`
	}

	var summaries []FlowSummary
	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}

		cacheKey := CacheKey(code, "d")
		if cached, ok := h.Cache.GetKLines(cacheKey); ok && len(cached) > 0 {
			flows := estimateCapitalFlow(cached)
			if len(flows) > 0 {
				summaries = append(summaries, FlowSummary{
					Code:    code,
					MainNet: flows[len(flows)-1].MainNet,
				})
			}
		}
	}

	c.JSON(http.StatusOK, summaries)
}

// RegisterCapitalRoutes registers capital flow routes
func (h *Handler) RegisterCapitalRoutes(g *gin.RouterGroup) {
	g.GET("/capital/:code", h.GetCapitalFlow)
	g.GET("/capital/summary", h.GetCapitalFlowSummary)
}
