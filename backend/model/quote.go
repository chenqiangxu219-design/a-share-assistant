package model

import "time"

// Quote represents real-time stock quote
type Quote struct {
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Price        float64   `json:"price"`
	Change       float64   `json:"change"`
	ChangePct    float64   `json:"change_pct"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Yesterday    float64   `json:"yesterday_close"`
	Volume       float64   `json:"volume"`
	Turnover     float64   `json:"turnover"`
	TurnoverRate float64   `json:"turnover_rate"`
	Amount       float64   `json:"amount"`
	TimeStamp    time.Time `json:"time"`
}

// FinanceData holds financial indicators
type FinanceData struct {
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	PE_TTM            float64 `json:"pe_ttm"`
	PB                float64 `json:"pb"`
	ROE               float64 `json:"roe"`
	Total_Market_Cap  float64 `json:"total_market_cap"`  // 总市值 (元)
	Circulating_MCap  float64 `json:"circulating_market_cap"` // 流通市值 (元)
}

// KLine represents a single K-line candle
type KLine struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
	Amount float64 `json:"amount"`
}
