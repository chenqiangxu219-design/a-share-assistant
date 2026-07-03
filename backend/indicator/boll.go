package indicator

import (
	"math"
	"a-share-assistant/backend/model"
)

// CalcBOLL calculates Bollinger Bands
func CalcBOLL(klines []model.KLine) model.BOLLIndicator {
	n := len(klines)
	if n == 0 {
		return model.BOLLIndicator{}
	}

	period := 20
	width := 2.0

	mid := make([]float64, n)
	upper := make([]float64, n)
	lower := make([]float64, n)

	for i := 0; i < n; i++ {
		if i < period-1 {
			continue
		}

		// Calculate MA
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += klines[j].Close
		}
		mid[i] = sum / float64(period)

		// Calculate standard deviation
		var sqSum float64
		for j := i - period + 1; j <= i; j++ {
			diff := klines[j].Close - mid[i]
			sqSum += diff * diff
		}
		std := math.Sqrt(sqSum / float64(period))

		upper[i] = mid[i] + width*std
		lower[i] = mid[i] - width*std
	}

	return model.BOLLIndicator{Mid: mid, Up: upper, Down: lower}
}
