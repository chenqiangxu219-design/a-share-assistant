package indicator

import "a-share-assistant/backend/model"

// CalcKDJ calculates KDJ indicator
func CalcKDJ(klines []model.KLine) model.KDJIndicator {
	n := len(klines)
	if n == 0 {
		return model.KDJIndicator{}
	}

	period := 9

	k := make([]float64, n)
	d := make([]float64, n)
	j := make([]float64, n)

	// Initialize all at 50 to prevent zero-fill triggering false oversold signals
	for i := 0; i < n; i++ {
		k[i] = 50
		d[i] = 50
		j[i] = 50
	}

	for i := 0; i < n; i++ {
		if i < period-1 {
			continue
		}

		// Find lowest low and highest high in period
		lowest := klines[i].Low
		highest := klines[i].High
		for ii := i - period + 1; ii < i; ii++ {
			if klines[ii].Low < lowest {
				lowest = klines[ii].Low
			}
			if klines[ii].High > highest {
				highest = klines[ii].High
			}
		}

		// Calculate RSV
		var rsv float64
		if highest == lowest {
			rsv = 50
		} else {
			rsv = (klines[i].Close - lowest) / (highest - lowest) * 100
		}

		if i == period-1 {
			k[i] = 50
			d[i] = 50
		} else {
			k[i] = k[i-1]*2.0/3.0 + rsv/3.0
			d[i] = d[i-1]*2.0/3.0 + k[i]/3.0
		}
		j[i] = 3*k[i] - 2*d[i]
	}

	return model.KDJIndicator{K: k, D: d, J: j}
}
