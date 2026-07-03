package api

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"a-share-assistant/backend/chat"
	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/indicator"
	"a-share-assistant/backend/model"
	"a-share-assistant/backend/strategy"
	"a-share-assistant/backend/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Handler holds all dependencies for API handlers
type Handler struct {
	DS        *datasource.Manager
	Cache     *QuoteCache
	Hub       *ws.Hub
	Engine    *strategy.CompositeStrategy
	Portfolio *PortfolioManager
	Alerts    *AlertHandler
	Memory    *chat.ConversationStore
	PS        *datasource.PyClient // Python microservice client
}

func NewHandler(ds *datasource.Manager, cache *QuoteCache, hub *ws.Hub) *Handler {
	h := &Handler{
		DS:     ds,
		Cache:  cache,
		Hub:    hub,
		Engine: strategy.NewCompositeStrategy(),
		Memory: chat.NewConversationStore(),
	}
	h.Alerts = NewAlertHandler(h)
	return h
}

// SetPyClient sets the Python microservice client
func (h *Handler) SetPyClient(pc *datasource.PyClient) {
	h.PS = pc
}

// CalcIndicators wraps indicator calculation
func (h *Handler) CalcIndicators(klines []model.KLine) model.IndicatorResult {
	return indicator.CalcAllIndicators(klines, "")
}

// GetQuote handles GET /api/quote/:code
func (h *Handler) GetQuote(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))

	// Check cache first (for recent quotes)
	if cached, ok := h.Cache.GetQuote(code); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	quote, err := h.DS.GetQuote(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.Cache.SetQuote(code, quote)
	h.Hub.BroadcastQuote(quote)

	c.JSON(http.StatusOK, quote)
}

// GetMultiQuote handles GET /api/quote?codes=sh600519,sz000001
func (h *Handler) GetMultiQuote(c *gin.Context) {
	codesParam := c.Query("codes")
	if codesParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codes parameter required"})
		return
	}

	codes := strings.Split(codesParam, ",")
	var quotes []*model.Quote
	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		q, err := h.DS.GetQuote(code)
		if err == nil {
			quotes = append(quotes, q)
			h.Cache.SetQuote(code, q)
		}
	}

	c.JSON(http.StatusOK, quotes)
}

// GetKLines handles GET /api/kline/:code?period=d&count=100
func (h *Handler) GetKLines(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	period := c.DefaultQuery("period", "d")
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	cacheKey := CacheKey(code, period)

	// Check cache
	if cached, ok := h.Cache.GetKLines(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	klines, err := h.DS.GetKLines(code, period, count)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.Cache.SetKLines(cacheKey, klines)

	c.JSON(http.StatusOK, klines)
}

// GetIndicators handles GET /api/indicators/:code?period=d&count=100
func (h *Handler) GetIndicators(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	period := c.DefaultQuery("period", "d")
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	klines, err := h.DS.GetKLines(code, period, count)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ind := indicator.CalcAllIndicators(klines, code)

	c.JSON(http.StatusOK, ind)
}

// GetSignals handles GET /api/signals/:code
func (h *Handler) GetSignals(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	period := c.DefaultQuery("period", "d")
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	// Get quote for price and name
	quote, err := h.DS.GetQuote(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get K-lines for indicator calculation
	klines, err := h.DS.GetKLines(code, period, count)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ind := indicator.CalcAllIndicators(klines, code)

	// Run composite strategy
	composite := h.Engine.Analyze(code, quote.Name, quote.Price, ind)

	// Also get individual signals
	engine := strategy.NewEngine()
	signals := engine.Analyze(ind, quote.Price)

	result := gin.H{
		"composite":  composite,
		"signals":    signals,
		"indicators": ind,
	}

	c.JSON(http.StatusOK, result)
}

// GetFinance handles GET /api/finance/:code
func (h *Handler) GetFinance(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))

	finance, err := h.DS.GetFinance(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, finance)
}

// WSHandler handles WebSocket upgrade
func (h *Handler) WSHandler(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Same-origin or non-browser
			}
			// Allow localhost/dev origins
			allowed := []string{
				"http://localhost:5173",
				"http://127.0.0.1:5173",
				"https://localhost:5173",
			}
			for _, a := range allowed {
				if origin == a {
					return true
				}
			}
			// Allow file:// (Electron)
			if strings.HasPrefix(origin, "file://") {
				return true
			}
			return false
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.Hub.ServeWS(conn)
}

// GetStockList handles GET /api/stocks — returns cached stock list for frontend search
func (h *Handler) GetStockList(c *gin.Context) {
	quotes := h.Cache.GetAllQuotes()
	type StockInfo struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	var stocks []StockInfo

	if len(quotes) > 0 {
		// Use cached quotes if available
		for code, q := range quotes {
			stocks = append(stocks, StockInfo{Code: code, Name: q.Name})
		}
	} else {
		// Return default stock list from model when cache is empty
		defaultStocks := []StockInfo{
			{Code: "600519", Name: "贵州茅台"},
			{Code: "000858", Name: "五粮液"},
			{Code: "601318", Name: "中国平安"},
			{Code: "600036", Name: "招商银行"},
			{Code: "300750", Name: "宁德时代"},
			{Code: "688981", Name: "中芯国际"},
			{Code: "000001", Name: "平安银行"},
		}
		stocks = defaultStocks
	}

	sort.Slice(stocks, func(i, j int) bool { return stocks[i].Code < stocks[j].Code })
	c.JSON(http.StatusOK, gin.H{"stocks": stocks})
}

// GetAllStocks handles GET /api/stocks/all — returns full A-share market from akshare
func (h *Handler) GetAllStocks(c *gin.Context) {
	// Check if Python service is available
	if h.PS != nil && h.PS.IsReady() {
		stocks, err := h.PS.GetAllStocks()
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"stocks": stocks, "total": len(stocks)})
			return
		}
		// Log error but continue to fallback
		log.Printf("WARNING: Failed to get all stocks from Python service: %v", err)
	}

	// Fallback to cached + hardcoded stocks
	h.GetStockList(c)
}

// RegisterQuoteRoutes registers quote/indicator/signal routes
func (h *Handler) RegisterQuoteRoutes(g *gin.RouterGroup) {
	g.GET("/quote/:code", h.GetQuote)
	g.GET("/quote", h.GetMultiQuote)
	g.GET("/stocks", h.GetStockList)
	g.GET("/stocks/all", h.GetAllStocks)
	g.GET("/finance/:code", h.GetFinance)
	g.GET("/kline/:code", h.GetKLines)
	g.GET("/indicators/:code", h.GetIndicators)
	g.GET("/signals/:code", h.GetSignals)
}
