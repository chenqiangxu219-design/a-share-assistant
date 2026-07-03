package strategy

import (
	"a-share-assistant/backend/model"
)

// Strategy defines a trading strategy
type Strategy interface {
	Name() string
	Evaluate(ind model.IndicatorResult, price float64) []model.Signal
}

// Engine manages multiple strategies
type Engine struct {
	strategies []Strategy
}

func NewEngine() *Engine {
	return &Engine{
		strategies: []Strategy{
			NewMACrossStrategy(),
			NewMACDDCrossStrategy(),
			NewBOLLBreakStrategy(),
		},
	}
}

// Analyze runs all strategies and returns composite signals
func (e *Engine) Analyze(ind model.IndicatorResult, price float64) []model.Signal {
	allSignals := make([]model.Signal, 0)
	for _, s := range e.strategies {
		signals := s.Evaluate(ind, price)
		allSignals = append(allSignals, signals...)
	}
	return allSignals
}

// CompositeScore calculates a composite score from all signals
// Positive = bullish, Negative = bearish
func CompositeScore(signals []model.Signal) int {
	score := 0
	for _, s := range signals {
		if s.Direction == "buy" {
			score += s.Strength
		} else {
			score -= s.Strength
		}
	}
	return score
}
