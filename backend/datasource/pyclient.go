package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"a-share-assistant/backend/model"
)

// PyClient is an HTTP client for the Python microservice
type PyClient struct {
	baseURL string
	client  *http.Client
	ready   bool
}

func NewPyClient(baseURL string) *PyClient {
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	pc := &PyClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
	// Check health
	if err := pc.HealthCheck(); err == nil {
		pc.ready = true
	}
	return pc
}

func (pc *PyClient) IsReady() bool {
	return pc.ready
}

func (pc *PyClient) HealthCheck() error {
	resp, err := pc.client.Get(pc.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// DepthData represents order book depth (bid/ask 5 levels)
type DepthData struct {
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Price          float64   `json:"price"`
	BidPrices      []float64 `json:"bid_prices"`
	BidVolumes     []float64 `json:"bid_volumes"`
	AskPrices      []float64 `json:"ask_prices"`
	AskVolumes     []float64 `json:"ask_volumes"`
	Open           float64   `json:"open"`
	High           float64   `json:"high"`
	Low            float64   `json:"low"`
	YesterdayClose float64   `json:"yesterday_close"`
	Volume         float64   `json:"volume"`
	Amount         float64   `json:"amount"`
}

func (pc *PyClient) GetDepth(code string) (*DepthData, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/mootdx/quote/%s", pc.baseURL, code)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get depth: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("[pyclient] get depth: status %d: %s", resp.StatusCode, string(body))
	}
	var depth DepthData
	if err := json.NewDecoder(resp.Body).Decode(&depth); err != nil {
		return nil, fmt.Errorf("[pyclient] decode depth: %w", err)
	}
	return &depth, nil
}

func (pc *PyClient) GetMootdxKLines(code, period string, count int) ([]model.KLine, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/mootdx/kline/%s?period=%s&count=%d", pc.baseURL, code, period, count)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get mootdx klines: %w", err)
	}
	defer resp.Body.Close()
	var klines []model.KLine
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return nil, fmt.Errorf("[pyclient] decode klines: %w", err)
	}
	return klines, nil
}

// NewsItem represents a stock news item
type NewsItem struct {
	Title   string `json:"title"`
	Digest  string `json:"digest"`
	Content string `json:"info_content"`
	URL     string `json:"info_url"`
	Time    string `json:"showtime"`
}

func (pc *PyClient) GetNews(code string) ([]NewsItem, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/akshare/news/%s", pc.baseURL, code)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get news: %w", err)
	}
	defer resp.Body.Close()
	var items []NewsItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[pyclient] decode news: %w", err)
	}
	return items, nil
}

// ResearchItem represents a research report
type ResearchItem struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Date   string `json:"date"`
	Digest string `json:"digest"`
	URL    string `json:"url"`
}

func (pc *PyClient) GetResearch(code string) ([]ResearchItem, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/akshare/research/%s", pc.baseURL, code)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get research: %w", err)
	}
	defer resp.Body.Close()
	var items []ResearchItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[pyclient] decode research: %w", err)
	}
	return items, nil
}

// HeatmapItem represents a sector in the heatmap
type HeatmapItem struct {
	Name         string  `json:"name"`
	ChangePct    float64 `json:"change_pct"`
	LeadStock    string  `json:"lead_stock"`
	LeadStockPct float64 `json:"lead_stock_pct"`
}

func (pc *PyClient) GetHeatmap() ([]HeatmapItem, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/akshare/heatmap", pc.baseURL)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get heatmap: %w", err)
	}
	defer resp.Body.Close()
	var items []HeatmapItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[pyclient] decode heatmap: %w", err)
	}
	return items, nil
}

// SectorBoard represents a concept board
type SectorBoard struct {
	Name      string  `json:"name"`
	ChangePct float64 `json:"change_pct"`
}

// GetHotStocksAkshare calls /akshare/hot-stocks
func (pc *PyClient) GetHotStocksAkshare() ([]HotStock, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/akshare/hot-stocks", pc.baseURL)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get hot stocks: %w", err)
	}
	defer resp.Body.Close()
	var items []HotStock
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[pyclient] decode hot stocks: %w", err)
	}
	return items, nil
}

func (pc *PyClient) GetSectorBoards() ([]SectorBoard, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/akshare/sector-boards", pc.baseURL)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get sector boards: %w", err)
	}
	defer resp.Body.Close()
	var items []SectorBoard
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("[pyclient] decode sector boards: %w", err)
	}
	return items, nil
}

// GetAllStocks returns all A-share stocks from akshare
func (pc *PyClient) GetAllStocks() ([]HotStock, error) {
	if !pc.ready {
		return nil, fmt.Errorf("python service not ready")
	}
	u := fmt.Sprintf("%s/api/stocks/all", pc.baseURL)
	resp, err := pc.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get all stocks: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Stocks []HotStock `json:"stocks"`
		Total  int        `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[pyclient] decode all stocks: %w", err)
	}
	return result.Stocks, nil
}

// pyclient.go doesn't need strings/url but let's keep imports clean
var _ = strings.Trim
var _ = url.Values{}
