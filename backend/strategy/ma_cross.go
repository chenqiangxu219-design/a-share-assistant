package strategy

import (
	"fmt"
	"time"
	"a-share-assistant/backend/model"
)

type MACross struct{}

func NewMACrossStrategy() *MACross { return &MACross{} }

func (m *MACross) Name() string { return "MA_Cross" }

func (m *MACross) Evaluate(ind model.IndicatorResult, _ float64) []model.Signal {
	var signals []model.Signal

	// Check MA5/MA10 golden/death cross
	if len(ind.MA.MA5) >= 2 && len(ind.MA.MA10) >= 2 {
		n := len(ind.MA.MA5)
		ma5Curr := ind.MA.MA5[n-1]
		ma5Prev := ind.MA.MA5[n-2]
		ma10Curr := ind.MA.MA10[n-1]
		ma10Prev := ind.MA.MA10[n-2]

		if ma5Curr > 0 && ma10Curr > 0 {
			// Golden cross: MA5 crosses above MA10
			if ma5Prev <= ma10Prev && ma5Curr > ma10Curr {
				signals = append(signals, model.Signal{
					Direction:  "buy",
					Strength:   3,
					Indicators: []string{"MA5上穿MA10"},
					Message:    fmt.Sprintf("均线金叉: MA5(%0.2f) 上穿 MA10(%0.2f)", ma5Curr, ma10Curr),
					Time:       time.Now(),
				})
			}
			// Death cross: MA5 crosses below MA10
			if ma5Prev >= ma10Prev && ma5Curr < ma10Curr {
				signals = append(signals, model.Signal{
					Direction:  "sell",
					Strength:   3,
					Indicators: []string{"MA5下穿MA10"},
					Message:    fmt.Sprintf("均线死叉: MA5(%0.2f) 下穿 MA10(%0.2f)", ma5Curr, ma10Curr),
					Time:       time.Now(),
				})
			}
		}
	}

	// Check MA10/MA20 cross (stronger signal)
	if len(ind.MA.MA10) >= 2 && len(ind.MA.MA20) >= 2 {
		n := len(ind.MA.MA10)
		ma10Curr := ind.MA.MA10[n-1]
		ma10Prev := ind.MA.MA10[n-2]
		ma20Curr := ind.MA.MA20[n-1]
		ma20Prev := ind.MA.MA20[n-2]

		if ma10Curr > 0 && ma20Curr > 0 {
			if ma10Prev <= ma20Prev && ma10Curr > ma20Curr {
				signals = append(signals, model.Signal{
					Direction:  "buy",
					Strength:   4,
					Indicators: []string{"MA10上穿MA20"},
					Message:    fmt.Sprintf("中期均线金叉: MA10(%0.2f) 上穿 MA20(%0.2f)", ma10Curr, ma20Curr),
					Time:       time.Now(),
				})
			}
			if ma10Prev >= ma20Prev && ma10Curr < ma20Curr {
				signals = append(signals, model.Signal{
					Direction:  "sell",
					Strength:   4,
					Indicators: []string{"MA10下穿MA20"},
					Message:    fmt.Sprintf("中期均线死叉: MA10(%0.2f) 下穿 MA20(%0.2f)", ma10Curr, ma20Curr),
					Time:       time.Now(),
				})
			}
		}
	}

	return signals
}
