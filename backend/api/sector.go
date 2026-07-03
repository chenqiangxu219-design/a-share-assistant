package api

import (
	"a-share-assistant/backend/datasource"
	"a-share-assistant/backend/strategy"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSectorRotation handles GET /api/sectors/rotation
// Returns rotation analysis with stage labels and capital flow
func (h *Handler) GetSectorRotation(c *gin.Context) {
	// Try Python service first
	py := h.DS.PYC
	if py != nil && py.IsReady() {
		boards, err := py.GetSectorBoards()
		if err == nil && len(boards) > 0 {
			sectors := make([]datasource.SectorData, len(boards))
			for i, b := range boards {
				sectors[i] = datasource.SectorData{
					Name:      b.Name,
					ChangePct: b.ChangePct,
				}
			}

			engine := strategy.NewSectorRotationEngine()
			summary := engine.AnalyzeMarketRotation(sectors)

			c.JSON(http.StatusOK, gin.H{
				"rotation": summary,
				"sectors":  sectors,
			})
			return
		}
	}

	// Fallback to Ithomer
	ithomer := h.DS.Ithomer
	if ithomer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}

	concepts, err := ithomer.GetConcepts()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	sectors := make([]datasource.SectorData, len(concepts))
	for i, c := range concepts {
		sectors[i] = datasource.SectorData{
			Name:        c.Name,
			ChangePct:   c.ChangePct,
			LeaderStock: c.Reason,
		}
	}

	engine := strategy.NewSectorRotationEngine()
	summary := engine.AnalyzeMarketRotation(sectors)

	c.JSON(http.StatusOK, gin.H{
		"rotation": summary,
		"sectors":  sectors,
	})
}

// GetSectorHeatmap handles GET /api/sectors/heatmap
// Returns heatmap data with rotation metadata
func (h *Handler) GetSectorHeatmap(c *gin.Context) {
	py := h.DS.PYC
	if py != nil && py.IsReady() {
		items, err := py.GetHeatmap()
		if err == nil && len(items) > 0 {
			sectors := make([]datasource.SectorData, len(items))
			for i, it := range items {
				sectors[i] = datasource.SectorData{
					Name:        it.Name,
					ChangePct:   it.ChangePct,
					LeaderStock: it.LeadStock,
				}
			}

			engine := strategy.NewSectorRotationEngine()
			summary := engine.AnalyzeMarketRotation(sectors)

			c.JSON(http.StatusOK, gin.H{
				"sectors":  sectors,
				"rotation": summary,
			})
			return
		}
	}

	// Fallback to Ithomer
	ithomer := h.DS.Ithomer
	if ithomer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}

	concepts, err := ithomer.GetConcepts()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	sectors := make([]datasource.SectorData, len(concepts))
	for i, c := range concepts {
		sectors[i] = datasource.SectorData{
			Name:        c.Name,
			ChangePct:   c.ChangePct,
			LeaderStock: c.Reason,
		}
	}

	engine := strategy.NewSectorRotationEngine()
	summary := engine.AnalyzeMarketRotation(sectors)

	c.JSON(http.StatusOK, gin.H{
		"sectors":  sectors,
		"rotation": summary,
	})
}

// RegisterSectorRoutes registers sector analysis routes
func (h *Handler) RegisterSectorRoutes(g *gin.RouterGroup) {
	g.GET("/sectors/rotation", h.GetSectorRotation)
	g.GET("/sectors/heatmap", h.GetSectorHeatmap)
}
