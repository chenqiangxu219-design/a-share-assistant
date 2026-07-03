package indicator

import "a-share-assistant/backend/model"

// CalcMA calculates moving average
func CalcMA(closes []float64, period int) []float64 {
	result := make([]float64, len(closes))
	for i := 0; i < len(closes); i++ {
		if i < period-1 {
			result[i] = 0
			continue
		}
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += closes[j]
		}
		result[i] = sum / float64(period)
	}
	return result
}

// CalcMAIndicators calculates MA5, MA10, MA20, MA60
func CalcMAIndicators(klines []model.KLine) model.MAIndicators {
	if len(klines) == 0 {
		return model.MAIndicators{}
	}
	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}
	return model.MAIndicators{
		MA5:  CalcMA(closes, 5),
		MA10: CalcMA(closes, 10),
		MA20: CalcMA(closes, 20),
		MA60: CalcMA(closes, 60),
	}
}
