package api

import (
	"net/http"
	"strings"

	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/strategy"
	"github.com/gin-gonic/gin"
)

// GetDepth handles GET /api/depth/:code
func (h *Handler) GetDepth(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "python service not available"})
		return
	}
	depth, err := py.GetDepth(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, depth)
}

// GetNews handles GET /api/news/:code
func (h *Handler) GetNews(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "python service not available"})
		return
	}
	items, err := py.GetNews(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"news": items})
}

// GetSentiment handles GET /api/news/sentiment/:code
func (h *Handler) GetSentiment(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))

	// Get news articles
	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "python service not available"})
		return
	}

	items, err := py.GetNews(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Convert to NewsArticle
	articles := make([]datasource.NewsArticle, len(items))
	for i, item := range items {
		articles[i] = datasource.NewsArticle{
			Title:   item.Title,
			Content: item.Content,
			Source:  "akshare",
			Time:    item.Time,
			Code:    code,
		}
	}

	// Analyze sentiment
	results, err := strategy.AnalyzeSentiment(articles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sentiment": results})
}

// AnalyzeNews handles POST /api/news/analyze
func (h *Handler) AnalyzeNews(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	code := strings.ToUpper(req.Code)
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	// Get news articles
	articles, err := datasource.GetNewsArticles(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get quote for stock name
	quote, _ := h.DS.GetQuote(code)
	name := code
	if quote != nil {
		name = quote.Name
	}

	// Analyze sentiment
	summary, err := strategy.AnalyzeSentimentWithSummary(articles, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetResearch handles GET /api/research/:code
func (h *Handler) GetResearch(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "python service not available"})
		return
	}
	items, err := py.GetResearch(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"research": items})
}

// GetHeatmap handles GET /api/heatmap
func (h *Handler) GetHeatmap(c *gin.Context) {
	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "python service not available"})
		return
	}
	items, err := py.GetHeatmap()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sectors": items})
}

// GetHotStocks handles GET /api/hot-stocks
// Tries Python akshare first, falls back to Ithomer
func (h *Handler) GetHotStocks(c *gin.Context) {
	py := h.DS.PYC
	if py != nil && py.IsReady() {
		items, err := py.GetHotStocksAkshare()
		if err == nil && len(items) > 0 {
			c.JSON(http.StatusOK, gin.H{"hot_stocks": items})
			return
		}
	}
	// Fallback to Ithomer
	ithomer := h.DS.Ithomer
	if ithomer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}
	stocks, err := ithomer.GetHotStocks()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hot_stocks": stocks})
}

// GetConcepts handles GET /api/concepts
// Tries Python akshare sector boards first, falls back to Ithomer
func (h *Handler) GetConcepts(c *gin.Context) {
	py := h.DS.PYC
	if py != nil && py.IsReady() {
		boards, err := py.GetSectorBoards()
		if err == nil && len(boards) > 0 {
			concepts := make([]datasource.ConceptTopic, len(boards))
			for i, b := range boards {
				concepts[i] = datasource.ConceptTopic{
					Name:      b.Name,
					ChangePct: b.ChangePct,
				}
			}
			c.JSON(http.StatusOK, gin.H{"concepts": concepts})
			return
		}
	}
	// Fallback to Ithomer
	ithomer := h.DS.Ithomer
	if ithomer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}
	concepts, err := ithomer.GetConcepts()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": concepts})
}

// SearchHandler handles POST /api/search
func (h *Handler) SearchHandler(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}
	iwencai := h.DS.Iwencai
	if iwencai == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "iwencai not configured"})
		return
	}
	results, err := iwencai.Search(req.Query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// RegisterDataRoutes registers data source routes (depth, news, research, heatmap, etc.)
func (h *Handler) RegisterDataRoutes(g *gin.RouterGroup) {
	g.GET("/depth/:code", h.GetDepth)
	g.GET("/news/:code", h.GetNews)
	g.GET("/research/:code", h.GetResearch)
	g.GET("/heatmap", h.GetHeatmap)
	g.GET("/hot-stocks", h.GetHotStocks)
	g.GET("/concepts", h.GetConcepts)
	g.POST("/search", h.SearchHandler)
	g.GET("/news/sentiment/:code", h.GetSentiment)
	g.POST("/news/analyze", h.AnalyzeNews)
}
