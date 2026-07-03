package strategy

import (
	"fmt"
	"time"
	"a-share-assistant/backend/model"
)

// CompositeStrategy combines all signals into a single scored recommendation
type CompositeStrategy struct {
	engine *Engine
}

func NewCompositeStrategy() *CompositeStrategy {
	return &CompositeStrategy{
		engine: NewEngine(),
	}
}

// Analyze runs all strategies and returns a composite signal with score
func (c *CompositeStrategy) Analyze(code, name string, price float64, ind model.IndicatorResult) model.Signal {
	signals := c.engine.Analyze(ind, price)

	score := CompositeScore(signals)
	direction := "neutral"
	if score > 0 {
		direction = "buy"
	} else if score < 0 {
		direction = "sell"
	}

	// Normalize score to -10 to 10 range
	if score > 10 {
		score = 10
	}
	if score < -10 {
		score = -10
	}

	// Collect all indicator names
	indicators := make([]string, 0)
	for _, s := range signals {
		indicators = append(indicators, s.Indicators...)
	}

	// Build message
	var msg string
	if score > 0 {
		msg = fmt.Sprintf("综合评分 +%d — 买入信号，%d 个指标共振", score, countByDirection(signals, "buy"))
	} else if score < 0 {
		msg = fmt.Sprintf("综合评分 %d — 卖出信号，%d 个指标共振", score, countByDirection(signals, "sell"))
	} else {
		msg = "中性，无明显信号"
	}

	strength := abs(score) / 2
	if strength < 1 {
		strength = 1
	}
	if strength > 5 {
		strength = 5
	}

	return model.Signal{
		Code:       code,
		Name:       name,
		Direction:  direction,
		Strength:   strength,
		Score:      score,
		Indicators: indicators,
		Price:      price,
		Message:    msg,
		Time:       time.Now(),
	}
}

func countByDirection(signals []model.Signal, dir string) int {
	count := 0
	for _, s := range signals {
		if s.Direction == dir {
			count++
		}
	}
	return count
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
