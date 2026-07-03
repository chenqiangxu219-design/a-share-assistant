package indicator

import "a-share-assistant/backend/model"

// CalcRSI calculates RSI indicator
func CalcRSI(klines []model.KLine, period int) []float64 {
	n := len(klines)
	if n < period+1 {
		return make([]float64, n)
	}

	gains := make([]float64, n)
	losses := make([]float64, n)

	for i := 1; i < n; i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			gains[i] = change
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -change
		}
	}

	rsi := make([]float64, n)
	rsi[0] = 50

	// Initial average
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - 100/(1+rs)
	}

	// Smoothed average
	for i := period + 1; i < n; i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - 100/(1+rs)
		}
	}

	return rsi
}

func CalcRSIIndicators(klines []model.KLine) model.RSIIndicator {
	return model.RSIIndicator{
		RSI6:  CalcRSI(klines, 6),
		RSI12: CalcRSI(klines, 12),
		RSI24: CalcRSI(klines, 24),
	}
}
