package backtest

import (
	"math"

	"a-share-assistant/backend/indicator"
	"a-share-assistant/backend/model"
	"a-share-assistant/backend/strategy"
)

// Engine runs backtests
type Engine struct {
	Strategy *strategy.CompositeStrategy
}

func NewEngine() *Engine {
	return &Engine{
		Strategy: strategy.NewCompositeStrategy(),
	}
}

// BacktestRequest defines a backtest request
type BacktestRequest struct {
	Code     string  `json:"code" binding:"required"`
	Name     string  `json:"name"`
	Strategy string  `json:"strategy"`
	Start    string  `json:"start"`
	End      string  `json:"end"`
}

// Run executes a backtest
func (e *Engine) Run(klines []model.KLine, req BacktestRequest) model.BacktestResult {
	if len(klines) < 30 {
		return model.BacktestResult{
			Code:     req.Code,
			Name:     req.Name,
			Strategy: req.Strategy,
			Start:    req.Start,
			End:      req.End,
		}
	}

	initialCash := 1000000.0
	cash := initialCash
	shares := 0
	var netValues []model.NetValue
	wins := 0
	losses := 0
	peakNV := initialCash
	maxDD := 0.0
	entryPrice := 0.0

	for i := 30; i < len(klines); i++ {
		// Calculate indicators for data up to current point
		subKlines := klines[:i+1]
		ind := indicator.CalcAllIndicators(subKlines, req.Code)

		// Get signal
		sig := e.Strategy.Analyze(req.Code, req.Name, subKlines[i].Close, ind)

		// Execute trades
		if sig.Direction == "buy" && sig.Score >= 3 && shares == 0 {
			// Buy: use 90% of cash
			amount := cash * 0.9
			shares = int(amount / subKlines[i].Close)
			if shares < 100 {
				shares = 0
			}
			cost := float64(shares) * subKlines[i].Close * 1.001
			cash -= cost
			entryPrice = subKlines[i].Close
		}

		if sig.Direction == "sell" && sig.Score <= -3 && shares > 0 {
			// Sell all
			proceeds := float64(shares) * subKlines[i].Close * 0.999
			cash += proceeds
			if subKlines[i].Close > entryPrice {
				wins++
			} else {
				losses++
			}
			shares = 0
		}

		// Calculate net value
		netValue := cash + float64(shares)*subKlines[i].Close
		netValues = append(netValues, model.NetValue{
			Date:     subKlines[i].Date,
			NetValue: math.Round(netValue*100) / 100,
		})

		if netValue > peakNV {
			peakNV = netValue
		}
		dd := (peakNV - netValue) / peakNV * 100
		if dd > maxDD {
			maxDD = dd
		}
	}

	// Sell remaining at end
	if shares > 0 {
		cash += float64(shares) * klines[len(klines)-1].Close * 0.999
	}

	finalNV := cash
	totalReturn := (finalNV - initialCash) / initialCash * 100
	// Simple Sharpe ratio approximation
	sharpe := 0.0
	if len(netValues) > 1 {
		var returns []float64
		for i := 1; i < len(netValues); i++ {
			if netValues[i-1].NetValue > 0 {
				r := (netValues[i].NetValue - netValues[i-1].NetValue) / netValues[i-1].NetValue
				returns = append(returns, r)
			}
		}
		if len(returns) > 1 {
			var sum, sumSq float64
			for _, r := range returns {
				sum += r
				sumSq += r * r
			}
			mean := sum / float64(len(returns))
			variance := sumSq/float64(len(returns)) - mean*mean
			if variance > 0 {
				sharpe = mean / math.Sqrt(variance) * math.Sqrt(252)
			}
		}
	}

	totalTrades := wins + losses
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(wins) / float64(totalTrades) * 100
	}

	return model.BacktestResult{
		Code:          req.Code,
		Name:          req.Name,
		Strategy:      req.Strategy,
		Start:         req.Start,
		End:           req.End,
		TotalReturn:   math.Round(totalReturn*100) / 100,
		MaxDrawdown:   math.Round(maxDD*100) / 100,
		SharpeRatio:   math.Round(sharpe*100) / 100,
		WinRate:       math.Round(winRate*100) / 100,
		TradeCount:    totalTrades,
		NetValueCurve: netValues,
	}
}
