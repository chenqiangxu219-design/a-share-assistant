package api

import (
	"net/http"
	"sort"
	"strings"

	"a-share-assistant/backend/backtest"
	"a-share-assistant/backend/indicator"
	"a-share-assistant/backend/model"

	"github.com/gin-gonic/gin"
)

// BacktestHandler handles POST /api/backtest
func (h *Handler) BacktestHandler(c *gin.Context) {
	var req struct {
		Code       string   `json:"code" binding:"required"`
		Name       string   `json:"name"`
		Strategy   string   `json:"strategy"`
		Strategies []string `json:"strategies"`
		Start      string   `json:"start"`
		End        string   `json:"end"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.Code = strings.ToUpper(req.Code)

	// Get K-lines
	klines, err := h.DS.GetKLines(req.Code, "d", 500)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get name if not provided
	if req.Name == "" {
		if q, err := h.DS.GetQuote(req.Code); err == nil {
			req.Name = q.Name
		}
	}

	// Filter by date range if specified
	if req.Start != "" {
		klines = filterKLines(klines, req.Start, req.End)
	}

	// Multi-strategy mode
	if len(req.Strategies) > 0 {
		var results []model.BacktestResult
		for _, strat := range req.Strategies {
			btReq := backtest.BacktestRequest{
				Code:     req.Code,
				Name:     req.Name,
				Strategy: strat,
				Start:    req.Start,
				End:      req.End,
			}
			engine := backtest.NewEngine()
			result := engine.Run(klines, btReq)
			results = append(results, result)
		}
		c.JSON(http.StatusOK, gin.H{"results": results})
		return
	}

	// Single strategy mode (backward compatible)
	btReq := backtest.BacktestRequest{
		Code:     req.Code,
		Name:     req.Name,
		Strategy: req.Strategy,
		Start:    req.Start,
		End:      req.End,
	}
	engine := backtest.NewEngine()
	result := engine.Run(klines, btReq)

	c.JSON(http.StatusOK, result)
}

func filterKLines(klines []model.KLine, start, end string) []model.KLine {
	var filtered []model.KLine
	for _, k := range klines {
		if k.Date >= start && (end == "" || k.Date <= end) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

// ScreenerHandler handles POST /api/screener
func (h *Handler) ScreenerHandler(c *gin.Context) {
	var req struct {
		Codes          []string `json:"codes"`
		MinScore       int      `json:"min_score"`
		MaxScore       int      `json:"max_score"`
		MaxRSI6        float64  `json:"max_rsi6"`
		MinVolumeRatio float64  `json:"min_volume_ratio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	type ScreenerResult struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Score    int     `json:"score"`
		Signal   string  `json:"signal"`
		RSI6     float64 `json:"rsi6"`
		VolRatio float64 `json:"volume_ratio"`
	}

	var results []ScreenerResult

	for _, code := range req.Codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}

		// Get K-lines
		klines, err := h.DS.GetKLines(code, "d", 100)
		if err != nil {
			continue
		}

		// Get quote
		quote, err := h.DS.GetQuote(code)
		if err != nil {
			continue
		}

		ind := indicator.CalcAllIndicators(klines, code)
		sig := h.Engine.Analyze(code, quote.Name, quote.Price, ind)

		// Apply filters
		rsi6 := 50.0
		if len(ind.RSI.RSI6) > 0 {
			rsi6 = ind.RSI.RSI6[len(ind.RSI.RSI6)-1]
		}
		volRatio := 1.0
		if len(ind.Vol.Ratio) > 0 {
			volRatio = ind.Vol.Ratio[len(ind.Vol.Ratio)-1]
		}

		if req.MinScore != 0 && sig.Score < req.MinScore {
			continue
		}
		if req.MaxScore != 0 && sig.Score > req.MaxScore {
			continue
		}
		if req.MaxRSI6 > 0 && rsi6 > req.MaxRSI6 {
			continue
		}
		if req.MinVolumeRatio > 0 && volRatio < req.MinVolumeRatio {
			continue
		}

		results = append(results, ScreenerResult{
			Code:     code,
			Name:     quote.Name,
			Price:    quote.Price,
			Score:    sig.Score,
			Signal:   sig.Direction,
			RSI6:     rsi6,
			VolRatio: volRatio,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	c.JSON(http.StatusOK, results)
}

// RegisterBacktestRoutes registers backtest and screener routes
func (h *Handler) RegisterBacktestRoutes(g *gin.RouterGroup) {
	g.POST("/backtest", h.BacktestHandler)
	g.POST("/screener", h.ScreenerHandler)
}
