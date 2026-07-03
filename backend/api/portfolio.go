package api

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/model"

	"github.com/gin-gonic/gin"
)

// PortfolioManager manages simulated trading
type PortfolioManager struct {
	mu          sync.Mutex
	cash        float64
	positions   map[string]*model.Position
	trades      []model.TradeRecord
	tradeID     int
	initialCash float64
}

func NewPortfolioManager() *PortfolioManager {
	initial := 1000000.0 // 100万
	return &PortfolioManager{
		cash:        initial,
		positions:   make(map[string]*model.Position),
		initialCash: initial,
	}
}

// GetPortfolio returns the current portfolio state
func (p *PortfolioManager) GetPortfolio(quotes map[string]*model.Quote) model.Portfolio {
	p.mu.Lock()
	defer p.mu.Unlock()

	var posList []model.Position
	for code, pos := range p.positions {
		curPrice := pos.CostPrice
		if q, ok := quotes[code]; ok {
			curPrice = q.Price
		}
		pnl := (curPrice - pos.CostPrice) * float64(pos.Shares)
		pnlPct := 0.0
		if pos.CostPrice > 0 {
			pnlPct = (curPrice - pos.CostPrice) / pos.CostPrice * 100
		}

		posList = append(posList, model.Position{
			Code:      code,
			Name:      pos.Name,
			Shares:    pos.Shares,
			CostPrice: pos.CostPrice,
			CurPrice:  math.Round(curPrice*100) / 100,
			PnL:       math.Round(pnl*100) / 100,
			PnLPct:    math.Round(pnlPct*100) / 100,
		})
	}

	sort.Slice(posList, func(i, j int) bool {
		return posList[i].Code < posList[j].Code
	})

	totalValue := p.cash
	for _, pos := range posList {
		totalValue += pos.CurPrice * float64(pos.Shares)
	}

	pnl := totalValue - p.initialCash
	pnlPct := pnl / p.initialCash * 100

	return model.Portfolio{
		Cash:       math.Round(p.cash*100) / 100,
		Positions:  posList,
		Trades:     p.trades,
		TotalValue: math.Round(totalValue*100) / 100,
		PnL:        math.Round(pnl*100) / 100,
		PnLPct:     math.Round(pnlPct*100) / 100,
	}
}

// Buy executes a simulated buy order
func (p *PortfolioManager) Buy(code, name string, price float64, shares int) (*model.TradeRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	amount := price * float64(shares)
	// Add 0.1% transaction cost
	fee := amount * 0.001
	total := amount + fee

	if total > p.cash {
		return nil, fmt.Errorf("资金不足")
	}

	p.cash -= total

	if pos, ok := p.positions[code]; ok {
		// Average up
		totalCost := pos.CostPrice*float64(pos.Shares) + amount
		totalShares := pos.Shares + shares
		pos.CostPrice = totalCost / float64(totalShares)
		pos.Shares = totalShares
	} else {
		p.positions[code] = &model.Position{
			Code:      code,
			Name:      name,
			Shares:    shares,
			CostPrice: price,
		}
	}

	p.tradeID++
	trade := model.TradeRecord{
		ID:        p.tradeID,
		Code:      code,
		Name:      name,
		Direction: "buy",
		Price:     price,
		Shares:    shares,
		Amount:    total,
		Time:      time.Now(),
	}
	p.trades = append(p.trades, trade)

	return &trade, nil
}

// Sell executes a simulated sell order
func (p *PortfolioManager) Sell(code string, price float64, shares int) (*model.TradeRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pos, ok := p.positions[code]
	if !ok || pos.Shares < shares {
		return nil, fmt.Errorf("持仓不足")
	}

	amount := price * float64(shares)
	fee := amount * 0.001
	net := amount - fee

	p.cash += net

	if pos.Shares == shares {
		delete(p.positions, code)
	} else {
		pos.Shares -= shares
	}

	p.tradeID++
	trade := model.TradeRecord{
		ID:        p.tradeID,
		Code:      code,
		Name:      pos.Name,
		Direction: "sell",
		Price:     price,
		Shares:    shares,
		Amount:    net,
		Time:      time.Now(),
	}
	p.trades = append(p.trades, trade)

	return &trade, nil
}

// getQuotesForPositions fetches current quotes for all active positions
func (p *PortfolioManager) getQuotesForPositions(ds *datasource.Manager) map[string]*model.Quote {
	p.mu.Lock()
	codes := make([]string, 0, len(p.positions))
	for code := range p.positions {
		codes = append(codes, code)
	}
	p.mu.Unlock()

	quotes := make(map[string]*model.Quote)
	for _, code := range codes {
		q, err := ds.GetQuote(code)
		if err == nil {
			quotes[code] = q
		}
	}
	return quotes
}

// PortfolioHandler handles GET /api/portfolio
func (h *Handler) PortfolioHandler(c *gin.Context) {
	// Fetch current quotes for all position codes so P&L is correct
	quotes := h.Portfolio.getQuotesForPositions(h.DS)
	portfolio := h.Portfolio.GetPortfolio(quotes)
	if portfolio.Positions == nil {
		portfolio.Positions = []model.Position{}
	}
	if portfolio.Trades == nil {
		portfolio.Trades = []model.TradeRecord{}
	}
	c.JSON(http.StatusOK, portfolio)
}

// BuyHandler handles POST /api/portfolio/buy
func (h *Handler) BuyHandler(c *gin.Context) {
	var req struct {
		Code   string  `json:"code" binding:"required"`
		Price  float64 `json:"price" binding:"required"`
		Shares int     `json:"shares" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	code := strings.ToUpper(req.Code)
	name := ""
	// Try to get name from quote
	if q, err := h.DS.GetQuote(code); err == nil {
		name = q.Name
	}

	trade, err := h.Portfolio.Buy(code, name, req.Price, req.Shares)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trade)
}

// SellHandler handles POST /api/portfolio/sell
func (h *Handler) SellHandler(c *gin.Context) {
	var req struct {
		Code   string  `json:"code" binding:"required"`
		Price  float64 `json:"price" binding:"required"`
		Shares int     `json:"shares" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	code := strings.ToUpper(req.Code)
	trade, err := h.Portfolio.Sell(code, req.Price, req.Shares)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trade)
}

// RegisterPortfolioRoutes registers portfolio simulation routes
func (h *Handler) RegisterPortfolioRoutes(g *gin.RouterGroup) {
	g.GET("/portfolio", h.PortfolioHandler)
	g.POST("/portfolio/buy", h.BuyHandler)
	g.POST("/portfolio/sell", h.SellHandler)
}
