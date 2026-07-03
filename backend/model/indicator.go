package model

// IndicatorResult holds all technical indicators for a stock
type IndicatorResult struct {
	Code  string        `json:"code"`
	MA    MAIndicators  `json:"ma"`
	MACD  MACDIndicator `json:"macd"`
	RSI   RSIIndicator  `json:"rsi"`
	KDJ   KDJIndicator  `json:"kdj"`
	BOLL  BOLLIndicator `json:"boll"`
	Vol   VolIndicator  `json:"volume"`
}

type MAIndicators struct {
	MA5  []float64 `json:"ma5"`
	MA10 []float64 `json:"ma10"`
	MA20 []float64 `json:"ma20"`
	MA60 []float64 `json:"ma60"`
}

type MACDIndicator struct {
	DIF  []float64 `json:"dif"`
	DEA  []float64 `json:"dea"`
	Hist []float64 `json:"hist"`
}

type RSIIndicator struct {
	RSI6  []float64 `json:"rsi6"`
	RSI12 []float64 `json:"rsi12"`
	RSI24 []float64 `json:"rsi24"`
}

type KDJIndicator struct {
	K []float64 `json:"k"`
	D []float64 `json:"d"`
	J []float64 `json:"j"`
}

type BOLLIndicator struct {
	Mid  []float64 `json:"mid"`
	Up   []float64 `json:"upper"`
	Down []float64 `json:"lower"`
}

type VolIndicator struct {
	Ratio    []float64 `json:"ratio"`    // volume ratio
	TurnOver []float64 `json:"turnover"` // turnover rate
}
