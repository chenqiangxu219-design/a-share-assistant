package strategy

import (
	"fmt"
	"time"
	"a-share-assistant/backend/model"
)

type MACDDCross struct{}

func NewMACDDCrossStrategy() *MACDDCross { return &MACDDCross{} }

func (m *MACDDCross) Name() string { return "MACD_Cross" }

func (m *MACDDCross) Evaluate(ind model.IndicatorResult, _ float64) []model.Signal {
	var signals []model.Signal
	n := len(ind.MACD.DIF)
	if n < 2 {
		return signals
	}

	// DIF crosses DEA (golden/death cross)
	difCurr := ind.MACD.DIF[n-1]
	difPrev := ind.MACD.DIF[n-2]
	deaCurr := ind.MACD.DEA[n-1]
	deaPrev := ind.MACD.DEA[n-2]

	if difPrev <= deaPrev && difCurr > deaCurr {
		// Golden cross above/below zero axis matters
		zone := "零轴下方"
		if deaCurr >= 0 {
			zone = "零轴上方"
		}
		strength := 3
		if deaCurr >= 0 {
			strength = 4 // stronger when above zero
		}
		signals = append(signals, model.Signal{
			Direction:  "buy",
			Strength:   strength,
			Indicators: []string{"MACD金叉", zone},
			Message:    fmt.Sprintf("MACD金叉%s: DIF(%0.4f) 上穿 DEA(%0.4f)", zone, difCurr, deaCurr),
			Time:       time.Now(),
		})
	}

	if difPrev >= deaPrev && difCurr < deaCurr {
		zone := "零轴下方"
		if deaCurr >= 0 {
			zone = "零轴上方"
		}
		strength := 3
		if deaCurr >= 0 {
			strength = 4
		}
		signals = append(signals, model.Signal{
			Direction:  "sell",
			Strength:   strength,
			Indicators: []string{"MACD死叉", zone},
			Message:    fmt.Sprintf("MACD死叉%s: DIF(%0.4f) 下穿 DEA(%0.4f)", zone, difCurr, deaCurr),
			Time:       time.Now(),
		})
	}

	// Histogram momentum change
	if n >= 3 {
		histCurr := ind.MACD.Hist[n-1]
		histPrev := ind.MACD.Hist[n-2]
		if histPrev < 0 && histCurr > 0 && histCurr > histPrev {
			signals = append(signals, model.Signal{
				Direction:  "buy",
				Strength:   2,
				Indicators: []string{"MACD绿柱缩短"},
				Message:    fmt.Sprintf("MACD绿柱缩短，空头动能减弱: HIST(%0.4f)", histCurr),
				Time:       time.Now(),
			})
		}
		if histPrev > 0 && histCurr < 0 && histCurr < histPrev {
			signals = append(signals, model.Signal{
				Direction:  "sell",
				Strength:   2,
				Indicators: []string{"MACD红柱缩短"},
				Message:    fmt.Sprintf("MACD红柱缩短，多头动能减弱: HIST(%0.4f)", histCurr),
				Time:       time.Now(),
			})
		}
	}

	return signals
}
