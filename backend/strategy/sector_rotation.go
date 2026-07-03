package strategy

import (
	"fmt"
	"sort"
	"a-share-assistant/backend/datasource"
)

// SectorRotationEngine analyzes capital flow + sector performance
// to detect rotation patterns across the market
type SectorRotationEngine struct{}

func NewSectorRotationEngine() *SectorRotationEngine {
	return &SectorRotationEngine{}
}

// RotationSummary represents the overall market rotation analysis
type RotationSummary struct {
	Stage        string                  `json:"stage"`         // overall market rotation stage
	Description  string                  `json:"description"`   // human-readable description
	TopSectors   []datasource.StrongSector `json:"top_sectors"`  // top 5 performing sectors
	BottomSectors []datasource.StrongSector `json:"bottom_sectors"` // bottom 3 sectors
	AvgChange    float64                 `json:"avg_change"`    // average sector change
	BullishPct   float64                 `json:"bullish_pct"`   // percentage of bullish sectors
}

// AnalyzeMarketRotation analyzes the overall market rotation pattern
func (e *SectorRotationEngine) AnalyzeMarketRotation(sectors []datasource.SectorData) RotationSummary {
	// Calculate statistics for each stage
	stageCount := map[string]int{
		"startup":   0,
		"accelerate": 0,
		"peak":      0,
		"decline":   0,
		"neutral":   0,
	}

	var totalChange float64
	var bullishCount int
	var strongSectors []datasource.StrongSector

	for _, s := range sectors {
		stage := s.DetectRotation()
		stageCount[string(stage)]++
		totalChange += s.ChangePct
		if s.ChangePct > 0 {
			bullishCount++
		}
		strongSectors = append(strongSectors, datasource.StrongSector{
			SectorData: s,
			Stage:      stage,
			Flow:       s.FlowDirection(),
			Strength:   s.StrengthScore(),
		})
	}

	// Sort by strength score
	sort.Slice(strongSectors, func(i, j int) bool {
		return strongSectors[i].Strength > strongSectors[j].Strength
	})

	// Determine overall market stage
	avgChange := totalChange / float64(len(sectors))
	bullishPct := float64(bullishCount) / float64(len(sectors)) * 100

	stage := e.determineRotationStage(stageCount, avgChange, bullishPct)
	description := e.buildDescription(stage, stageCount, avgChange, strongSectors)

	return RotationSummary{
		Stage:         stage,
		Description:   description,
		TopSectors:    strongSectors[:min(5, len(strongSectors))],
		BottomSectors: strongSectors[max(0, len(strongSectors)-3):],
		AvgChange:     avgChange,
		BullishPct:    bullishPct,
	}
}

// AnalyzeSectorRotation is a convenience method that returns the rotation summary as a string
func (e *SectorRotationEngine) AnalyzeSectorRotation(sectors []datasource.SectorData) string {
	summary := e.AnalyzeMarketRotation(sectors)
	return summary.Description
}

// detectRotationPattern identifies the specific rotation pattern
func (e *SectorRotationEngine) detectRotationPattern(sectors []datasource.SectorData) string {
	// Sort by change_pct
	sorted := make([]datasource.SectorData, len(sectors))
	copy(sorted, sectors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ChangePct > sorted[j].ChangePct
	})

	if len(sorted) < 2 {
		return "单一板块主导"
	}

	// Check if top performers are leading
	top1 := sorted[0]
	top5Avg := 0.0
	for i := 0; i < min(5, len(sorted)); i++ {
		top5Avg += sorted[i].ChangePct
	}
	top5Avg /= float64(min(5, len(sorted)))

	// Check for sector concentration
	top1Pct := top1.ChangePct
	if top1Pct > top5Avg*1.5 {
		return "龙头板块领涨"
	}
	if top5Avg > 2.0 {
		return "多板块共振"
	}
	return "轮动活跃"
}

// determineRotationStage classifies the overall market rotation stage
func (e *SectorRotationEngine) determineRotationStage(stageCount map[string]int, avgChange float64, bullishPct float64) string {
	// If most sectors are accelerating and avg change is positive
	if stageCount["accelerate"] > stageCount["peak"] && avgChange > 1.0 {
		return "加速上行"
	}
	if stageCount["peak"] > stageCount["accelerate"] && avgChange > 2.0 {
		return "见顶分化"
	}
	if stageCount["decline"] > stageCount["accelerate"] && avgChange < 0.5 {
		return "资金流出"
	}
	if stageCount["startup"] > stageCount["decline"] && bullishPct > 50 {
		return "启动轮动"
	}
	return "震荡整理"
}

// buildDescription creates a human-readable description of the rotation
func (e *SectorRotationEngine) buildDescription(stage string, stageCount map[string]int, avgChange float64, strongSectors []datasource.StrongSector) string {
	pattern := e.detectRotationPattern(nil)

	desc := fmt.Sprintf("当前市场处于「%s」阶段，%s。", stage, pattern)

	if len(strongSectors) > 0 {
		top := strongSectors[0]
		desc += fmt.Sprintf("领涨板块为「%s」(+%0.2f%%)，", top.Name, top.ChangePct)
	}

	desc += fmt.Sprintf("平均涨幅 %0.2f%%。", avgChange)

	return desc
}
