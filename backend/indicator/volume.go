package indicator

import "a-share-assistant/backend/model"

// CalcVolIndicators calculates volume ratio and turnover rate
func CalcVolIndicators(klines []model.KLine) model.VolIndicator {
	n := len(klines)
	if n == 0 {
		return model.VolIndicator{}
	}

	ratio := make([]float64, n)
	turnOver := make([]float64, n)

	ma5Vol := CalcMA(volumeArray(klines), 5)

	for i := 0; i < n; i++ {
		if ma5Vol[i] > 0 {
			ratio[i] = klines[i].Volume / ma5Vol[i]
		}
		// Turnover rate is typically calculated as volume / total_shares * 100
		// For simplicity, we'll use volume / 5-day average as a proxy
		if i > 0 && klines[i-1].Volume > 0 {
			turnOver[i] = klines[i].Volume / klines[i-1].Volume * 100 - 100
		}
	}

	return model.VolIndicator{
		Ratio:    ratio,
		TurnOver: turnOver,
	}
}

func volumeArray(klines []model.KLine) []float64 {
	vols := make([]float64, len(klines))
	for i, k := range klines {
		vols[i] = k.Volume
	}
	return vols
}

// CalcAllIndicators calculates all indicators for a set of K-lines
func CalcAllIndicators(klines []model.KLine, code string) model.IndicatorResult {
	return model.IndicatorResult{
		Code:   code,
		MA:     CalcMAIndicators(klines),
		MACD:   CalcMACD(klines),
		RSI:    CalcRSIIndicators(klines),
		KDJ:    CalcKDJ(klines),
		BOLL:   CalcBOLL(klines),
		Vol:    CalcVolIndicators(klines),
	}
}
