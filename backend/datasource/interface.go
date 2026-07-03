package datasource

import (
	"fmt"

	"a-share-assistant/backend/model"
)

// Source is the interface all data sources must implement
type Source interface {
	Name() string
	GetQuote(code string) (*model.Quote, error)
	GetKLines(code, period string, count int) ([]model.KLine, error)
	GetFinance(code string) (*model.FinanceData, error)
}

// Manager provides unified access with fallback
type Manager struct {
	sources []Source
	PYC     *PyClient
	Ithomer *Ithomer
	Iwencai *Iwencai
}

// GetSources returns all registered sources
func (m *Manager) GetSources() []Source {
	return m.sources
}

func NewManager(sources ...Source) *Manager {
	return &Manager{sources: sources}
}

// SetPyClient sets the Python microservice client
func (m *Manager) SetPyClient(pc *PyClient) {
	m.PYC = pc
}

// SetIthomer sets the 同花顺热点 client
func (m *Manager) SetIthomer(it *Ithomer) {
	m.Ithomer = it
}

// SetIwencai sets the iwencai client
func (m *Manager) SetIwencai(iw *Iwencai) {
	m.Iwencai = iw
}

// GetQuote tries each source in order, returns first successful result
func (m *Manager) GetQuote(code string) (*model.Quote, error) {
	var lastErr error
	for _, src := range m.sources {
		q, err := src.GetQuote(code)
		if err == nil {
			return q, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// GetKLines tries each source in order
func (m *Manager) GetKLines(code, period string, count int) ([]model.KLine, error) {
	var lastErr error
	for _, src := range m.sources {
		klines, err := src.GetKLines(code, period, count)
		if err == nil && len(klines) > 0 {
			return klines, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// GetCapitalFlow estimates capital flow from K-lines
func (m *Manager) GetCapitalFlow(code string, days int) ([]CapitalFlow, error) {
	sources := m.GetSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("no data sources")
	}
	em, ok := sources[0].(*EastMoney)
	if !ok {
		return nil, fmt.Errorf("capital flow not supported")
	}
	return em.GetCapitalFlow(code, days)
}

// GetFinance returns financial data (PE/PB/ROE/market cap)
func (m *Manager) GetFinance(code string) (*model.FinanceData, error) {
	for _, src := range m.sources {
		if s, ok := src.(interface{ GetFinance(string) (*model.FinanceData, error) }); ok {
			return s.GetFinance(code)
		}
	}
	return nil, fmt.Errorf("no data source supports finance data")
}

// GetSectorRotation returns sector rotation data from available sources
func (m *Manager) GetSectorRotation() ([]SectorData, error) {
	// Try Python service first
	if m.PYC != nil && m.PYC.IsReady() {
		boards, err := m.PYC.GetSectorBoards()
		if err == nil && len(boards) > 0 {
			sectors := make([]SectorData, len(boards))
			for i, b := range boards {
				sectors[i] = SectorData{
					Name:      b.Name,
					ChangePct: b.ChangePct,
				}
			}
			return sectors, nil
		}
	}

	// Fallback to Ithomer
	if m.Ithomer != nil {
		concepts, err := m.Ithomer.GetConcepts()
		if err == nil {
			sectors := make([]SectorData, len(concepts))
			for i, c := range concepts {
				sectors[i] = SectorData{
					Name:      c.Name,
					ChangePct: c.ChangePct,
				}
			}
			return sectors, nil
		}
	}

	return nil, fmt.Errorf("no data source available for sector rotation")
}
