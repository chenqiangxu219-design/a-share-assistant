package datasource

import (
	"math"
)

// SectorData represents a sector/industry board with performance metrics
type SectorData struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	ChangePct   float64 `json:"change_pct"`
	Volume      float64 `json:"volume"`
	LeaderStock string  `json:"leader_stock"`
	CapitalFlow float64 `json:"capital_flow"` // 主力净流入 (亿元)
}

// RotationStage represents the lifecycle stage of a sector
type RotationStage string

const (
	StageStartup   RotationStage = "startup"    // 启动 — 资金刚流入
	StageAccelerate RotationStage = "accelerate" // 加速 — 持续走强
	StagePeak      RotationStage = "peak"       // 见顶 — 涨幅过大
	StageDecline   RotationStage = "decline"    // 衰退 — 资金流出
	StageNeutral   RotationStage = "neutral"    // 中性 — 无明显信号
)

// DetectRotation determines the rotation stage for each sector
// based on price change, volume surge, and capital flow
func (s *SectorData) DetectRotation() RotationStage {
	// Capital flow direction
	if s.CapitalFlow < -0.5 {
		return StageDecline
	}

	switch {
	case s.ChangePct > 3.0 && s.Volume > 1.5:
		return StagePeak
	case s.ChangePct > 1.5 && s.Volume > 1.2:
		return StageAccelerate
	case s.ChangePct > 0.5:
		return StageStartup
	default:
		return StageNeutral
	}
}

// CapitalFlowDirection represents capital flow direction
type CapitalFlowDirection string

const (
	FlowIn  CapitalFlowDirection = "in"
	FlowOut CapitalFlowDirection = "out"
	FlowFlat CapitalFlowDirection = "flat"
)

// FlowDirection returns the capital flow direction indicator
func (s *SectorData) FlowDirection() CapitalFlowDirection {
	if s.CapitalFlow > 0.5 {
		return FlowIn
	}
	if s.CapitalFlow < -0.5 {
		return FlowOut
	}
	return FlowFlat
}

// StrongSector represents a sector with strong performance
type StrongSector struct {
	SectorData
	Stage     RotationStage   `json:"stage"`
	Flow      CapitalFlowDirection `json:"flow"`
	Strength  float64         `json:"strength"` // composite strength score
}

// StrengthScore calculates a composite strength score for ranking
func (s *SectorData) StrengthScore() float64 {
	// Weight: 60% price change, 20% volume, 20% capital flow
	priceScore := math.Abs(s.ChangePct) / 5.0
	volScore := math.Min(s.Volume/2.0, 1.0)
	flowScore := math.Min(math.Abs(s.CapitalFlow)/3.0, 1.0)
	return priceScore*0.6 + volScore*0.2 + flowScore*0.2
}

// SectorRotationResult bundles sector data with rotation metadata
type SectorRotationResult struct {
	Sectors       []SectorData     `json:"sectors"`
	TopSectors    []StrongSector   `json:"top_sectors"`
	BottomSectors []StrongSector   `json:"bottom_sectors"`
	AvgChange     float64          `json:"avg_change"`
	BullishPct    float64          `json:"bullish_pct"`
	Stage         RotationStage    `json:"stage"`
	Description   string           `json:"description"`
}

// RotateStageLabel returns a Chinese label for the rotation stage
func (s RotationStage) RotateStageLabel() string {
	switch s {
	case StageStartup:
		return "启动"
	case StageAccelerate:
		return "加速"
	case StagePeak:
		return "见顶"
	case StageDecline:
		return "衰退"
	default:
		return "中性"
	}
}
