package strategy

import (
	"fmt"
	"time"
	"a-share-assistant/backend/model"
)

type BOLLBreak struct{}

func NewBOLLBreakStrategy() *BOLLBreak { return &BOLLBreak{} }

func (b *BOLLBreak) Name() string { return "BOLL_Break" }

func (b *BOLLBreak) Evaluate(ind model.IndicatorResult, price float64) []model.Signal {
	var signals []model.Signal
	n := len(ind.BOLL.Mid)
	if n < 2 {
		return signals
	}

	closePrice := price

	if closePrice <= 0 {
		return signals
	}

	mid := ind.BOLL.Mid[n-1]
	upper := ind.BOLL.Up[n-1]
	lower := ind.BOLL.Down[n-1]

	if mid <= 0 || upper <= 0 || lower <= 0 {
		return signals
	}

	// Price breaks above upper band (overbought → sell)
	if closePrice > upper {
		signals = append(signals, model.Signal{
			Direction:  "sell",
			Strength:   3,
			Indicators: []string{"布林带突破上轨"},
			Message:    fmt.Sprintf("股价突破布林带上轨: 价格(%0.2f) > 上轨(%0.2f)", closePrice, upper),
			Time:       time.Now(),
		})
	}

	// Price breaks below lower band (oversold → buy)
	if closePrice < lower {
		signals = append(signals, model.Signal{
			Direction:  "buy",
			Strength:   3,
			Indicators: []string{"布林带下轨支撑"},
			Message:    fmt.Sprintf("股价触及布林带下轨: 价格(%0.2f) < 下轨(%0.2f)", closePrice, lower),
			Time:       time.Now(),
		})
	}

	// Band narrowing (squeeze → breakout coming)
	if n >= 20 {
		prevWidth := (ind.BOLL.Up[n-20] - ind.BOLL.Down[n-20]) / mid
		currWidth := (upper - lower) / mid
		if prevWidth > 0 && currWidth < prevWidth*0.5 {
			signals = append(signals, model.Signal{
				Direction:  "buy",
				Strength:   2,
				Indicators: []string{"布林带收窄"},
				Message:    fmt.Sprintf("布林带收窄，变盘信号: 带宽从(%0.2f)收窄至(%0.2f)", prevWidth, currWidth),
				Time:       time.Now(),
			})
		}
	}

	// RSI oversold/overbought confirmation
	if len(ind.RSI.RSI6) >= 2 {
		rsi6 := ind.RSI.RSI6[len(ind.RSI.RSI6)-1]
		if rsi6 < 30 {
			signals = append(signals, model.Signal{
				Direction:  "buy",
				Strength:   2,
				Indicators: []string{"RSI超卖"},
				Message:    fmt.Sprintf("RSI6 超卖: %0.2f", rsi6),
				Time:       time.Now(),
			})
		}
		if rsi6 > 70 {
			signals = append(signals, model.Signal{
				Direction:  "sell",
				Strength:   2,
				Indicators: []string{"RSI超买"},
				Message:    fmt.Sprintf("RSI6 超买: %0.2f", rsi6),
				Time:       time.Now(),
			})
		}
	}

	// KDJ extreme
	if len(ind.KDJ.K) >= 2 {
		kn := len(ind.KDJ.K)
		kv := ind.KDJ.K[kn-1]
		jv := ind.KDJ.J[kn-1]
		if kv > 0 && kv < 20 && jv < 10 {
			signals = append(signals, model.Signal{
				Direction:  "buy",
				Strength:   2,
				Indicators: []string{"KDJ低位"},
				Message:    fmt.Sprintf("KDJ低位: K=%0.2f, J=%0.2f", kv, jv),
				Time:       time.Now(),
			})
		}
		if kv > 80 && jv > 90 {
			signals = append(signals, model.Signal{
				Direction:  "sell",
				Strength:   2,
				Indicators: []string{"KDJ高位"},
				Message:    fmt.Sprintf("KDJ高位: K=%0.2f, J=%0.2f", kv, jv),
				Time:       time.Now(),
			})
		}
	}

	return signals
}
