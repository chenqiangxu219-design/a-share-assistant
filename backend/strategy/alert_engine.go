package strategy

import (
	"fmt"
	"time"
	"a-share-assistant/backend/model"
)

// AlertEngine detects abnormal market conditions
type AlertEngine struct {
	VolumeSurgeRatio float64
	PriceSpikePct    float64
	LimitUpPct       float64
	LimitDownPct     float64
	MACDDivergenceN  int
}

// NewAlertEngine creates AlertEngine with default thresholds
func NewAlertEngine() *AlertEngine {
	return &AlertEngine{
		VolumeSurgeRatio:  2.0,
		PriceSpikePct:     5.0,
		LimitUpPct:        9.8,
		LimitDownPct:      -9.8,
		MACDDivergenceN:   5,
	}
}

// RunAlerts checks all alert rules and returns triggered alerts
func (e *AlertEngine) RunAlerts(code, name string, quote model.Quote, ind model.IndicatorResult) []model.Alert {
	alerts := make([]model.Alert, 0)

	alerts = append(alerts, e.checkVolumeSurge(code, name, quote, ind)...)
	alerts = append(alerts, e.checkPriceSpike(code, name, quote)...)
	alerts = append(alerts, e.checkLimitUp(code, name, quote)...)
	alerts = append(alerts, e.checkLimitDown(code, name, quote)...)
	alerts = append(alerts, e.checkMACDDivergence(code, name, quote, ind)...)

	return alerts
}

// checkVolumeSurge: volume > 5-day average * 2
func (e *AlertEngine) checkVolumeSurge(code, name string, quote model.Quote, ind model.IndicatorResult) []model.Alert {
	var alerts []model.Alert
	n := len(ind.Vol.Ratio)
	if n == 0 {
		return alerts
	}

	ratio := ind.Vol.Ratio[n-1]
	if ratio > e.VolumeSurgeRatio {
		severity := model.AlertWarning
		if ratio > 3.0 {
			severity = model.AlertCritical
		}
		alerts = append(alerts, model.Alert{
			Code:     code,
			Name:     name,
			Type:     model.AlertVolumeSurge,
			Message:  fmt.Sprintf("放量: 量比=%.1f, 成交量=%.0f万手", ratio, quote.Volume/10000),
			Severity: severity,
			Time:     time.Now(),
		})
	}
	return alerts
}

// checkPriceSpike: change_pct > 5% (not limit up)
func (e *AlertEngine) checkPriceSpike(code, name string, quote model.Quote) []model.Alert {
	if quote.ChangePct > e.PriceSpikePct && quote.ChangePct < e.LimitUpPct {
		return []model.Alert{{
			Code:     code,
			Name:     name,
			Type:     model.AlertPriceSpike,
			Message:  fmt.Sprintf("急涨: +%.2f%%, 价格=%.2f", quote.ChangePct, quote.Price),
			Severity: model.AlertWarning,
			Time:     time.Now(),
		}}
	}
	if quote.ChangePct < -e.PriceSpikePct && quote.ChangePct > e.LimitDownPct {
		return []model.Alert{{
			Code:     code,
			Name:     name,
			Type:     model.AlertPriceSpike,
			Message:  fmt.Sprintf("急跌: %.2f%%, 价格=%.2f", quote.ChangePct, quote.Price),
			Severity: model.AlertWarning,
			Time:     time.Now(),
		}}
	}
	return nil
}

// checkLimitUp: price reaches limit up
func (e *AlertEngine) checkLimitUp(code, name string, quote model.Quote) []model.Alert {
	if quote.ChangePct >= e.LimitUpPct {
		return []model.Alert{{
			Code:     code,
			Name:     name,
			Type:     model.AlertLimitUp,
			Message:  fmt.Sprintf("涨停: +%.2f%%, 价格=%.2f", quote.ChangePct, quote.Price),
			Severity: model.AlertCritical,
			Time:     time.Now(),
		}}
	}
	return nil
}

// checkLimitDown: price reaches limit down
func (e *AlertEngine) checkLimitDown(code, name string, quote model.Quote) []model.Alert {
	if quote.ChangePct <= e.LimitDownPct {
		return []model.Alert{{
			Code:     code,
			Name:     name,
			Type:     model.AlertLimitDown,
			Message:  fmt.Sprintf("跌停: %.2f%%, 价格=%.2f", quote.ChangePct, quote.Price),
			Severity: model.AlertCritical,
			Time:     time.Now(),
		}}
	}
	return nil
}

// checkMACDDivergence: MACD top/bottom divergence
func (e *AlertEngine) checkMACDDivergence(code, name string, quote model.Quote, ind model.IndicatorResult) []model.Alert {
	n := len(ind.MACD.Hist)
	if n < e.MACDDivergenceN+1 {
		return nil
	}

	// Check for top divergence: price makes new high but MACD histogram is lower
	var topDivergence bool
	var bottomDivergence bool

	for i := n - e.MACDDivergenceN; i < n; i++ {
		if i == 0 {
			continue
		}
		// Top divergence: current price higher than recent but histogram lower
		if quote.Price > ind.MACD.DIF[i-1]*100 && ind.MACD.Hist[i] < ind.MACD.Hist[i-1] {
			topDivergence = true
		}
		// Bottom divergence: current price lower than recent but histogram higher
		if quote.Price < ind.MACD.DIF[i-1]*100 && ind.MACD.Hist[i] > ind.MACD.Hist[i-1] {
			bottomDivergence = true
		}
	}

	var alerts []model.Alert
	if topDivergence {
		alerts = append(alerts, model.Alert{
			Code:     code,
			Name:     name,
			Type:     model.AlertMACDDivergence,
			Message:  "MACD顶背离: 价格新高但MACD动能减弱",
			Severity: model.AlertWarning,
			Time:     time.Now(),
		})
	}
	if bottomDivergence {
		alerts = append(alerts, model.Alert{
			Code:     code,
			Name:     name,
			Type:     model.AlertMACDDivergence,
			Message:  "MACD底背离: 价格新低但MACD动能增强",
			Severity: model.AlertWarning,
			Time:     time.Now(),
		})
	}
	return alerts
}
