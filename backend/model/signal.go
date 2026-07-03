package model

import "time"

// Signal represents a trading signal
type Signal struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Direction string    `json:"direction"` // "buy" or "sell"
	Strength  int       `json:"strength"`  // 1-5
	Score     int       `json:"score"`     // composite score
	Indicators []string `json:"indicators"` // triggering indicators
	Price     float64   `json:"price"`
	Message   string    `json:"message"`
	Time      time.Time `json:"time"`
}

// BacktestResult represents backtest output
type BacktestResult struct {
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Strategy       string     `json:"strategy"`
	Start          string     `json:"start"`
	End            string     `json:"end"`
	TotalReturn    float64    `json:"total_return"`
	MaxDrawdown    float64    `json:"max_drawdown"`
	SharpeRatio    float64    `json:"sharpe_ratio"`
	WinRate        float64    `json:"win_rate"`
	TradeCount     int        `json:"trade_count"`
	NetValueCurve  []NetValue `json:"net_value_curve"`
}

type NetValue struct {
	Date     string  `json:"date"`
	NetValue float64 `json:"net_value"`
}

// Portfolio represents a simulated portfolio
type Portfolio struct {
	Cash      float64        `json:"cash"`
	Positions []Position     `json:"positions"`
	Trades    []TradeRecord  `json:"trades"`
	TotalValue float64       `json:"total_value"`
	PnL        float64       `json:"pnL"`
	PnLPct     float64       `json:"pnL_pct"`
}

type Position struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Shares     int     `json:"shares"`
	CostPrice  float64 `json:"cost_price"`
	CurPrice   float64 `json:"current_price"`
	PnL        float64 `json:"pnL"`
	PnLPct     float64 `json:"pnL_pct"`
}

type TradeRecord struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Direction string    `json:"direction"` // "buy" or "sell"
	Price     float64   `json:"price"`
	Shares    int       `json:"shares"`
	Amount    float64   `json:"amount"`
	Time      time.Time `json:"time"`
}
