package indicator

import "a-share-assistant/backend/model"

// CalcMACD calculates MACD indicator
func CalcMACD(klines []model.KLine) model.MACDIndicator {
	n := len(klines)
	if n == 0 {
		return model.MACDIndicator{}
	}

	closes := make([]float64, n)
	for i, k := range klines {
		closes[i] = k.Close
	}

	// EMA calculation
	ema12 := calcEMA(closes, 12)
	ema26 := calcEMA(closes, 26)

	dif := make([]float64, n)
	for i := 0; i < n; i++ {
		dif[i] = ema12[i] - ema26[i]
	}

	dea := calcEMA(dif, 9)
	hist := make([]float64, n)
	for i := 0; i < n; i++ {
		hist[i] = (dif[i] - dea[i]) * 2
	}

	return model.MACDIndicator{
		DIF:  dif,
		DEA:  dea,
		Hist: hist,
	}
}

func calcEMA(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)

	// First value is SMA
	k := 2.0 / float64(period+1)

	// Calculate SMA for first EMA value
	sum := 0.0
	p := period
	if p > n {
		p = n
	}
	for i := 0; i < p; i++ {
		sum += data[i]
	}
	result[p-1] = sum / float64(p)

	// Calculate EMA for remaining values
	for i := p; i < n; i++ {
		result[i] = data[i]*k + result[i-1]*(1-k)
	}

	return result
}
