package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NewsArticle is a unified news article for sentiment analysis
type NewsArticle struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
	Time    string `json:"time"`
	Code    string `json:"code"`
}

// GetNewsArticles fetches news from Python microservice, with Ithomer fallback
func GetNewsArticles(code string) ([]NewsArticle, error) {
	// Try Python microservice first
	py := defaultManager.PYC
	if py != nil && py.IsReady() {
		items, err := py.GetNews(code)
		if err == nil && len(items) > 0 {
			articles := make([]NewsArticle, len(items))
			for i, item := range items {
				articles[i] = NewsArticle{
					Title:   item.Title,
					Content: item.Content,
					Source:  "akshare",
					Time:    item.Time,
					Code:    code,
				}
			}
			return articles, nil
		}
	}

	// Fallback to Ithomer
	ithomer := defaultManager.Ithomer
	if ithomer != nil {
		stocks, err := ithomer.GetHotStocks()
		if err == nil && len(stocks) > 0 {
			articles := make([]NewsArticle, len(stocks))
			for i, s := range stocks {
				articles[i] = NewsArticle{
					Title:   fmt.Sprintf("%s 异动", s.Name),
					Content: fmt.Sprintf("%s 涨跌幅: %+.2f%%", s.Name, s.ChangePct),
					Source:  "ithomer",
					Time:    time.Now().Format("2006-01-02 15:04"),
					Code:    s.Code,
				}
			}
			return articles, nil
		}
	}

	return nil, fmt.Errorf("no news data available")
}

// GetNewsFromIthomer fetches news directly from Ithomer API
func GetNewsFromIthomer(code string) ([]NewsArticle, error) {
	ithomer := defaultManager.Ithomer
	if ithomer == nil {
		return nil, fmt.Errorf("ithomer not configured")
	}

	// Use Ithomer's hot stocks as news source
	stocks, err := ithomer.GetHotStocks()
	if err != nil {
		return nil, fmt.Errorf("[ithomer] get news: %w", err)
	}

	articles := make([]NewsArticle, 0, len(stocks))
	for _, s := range stocks {
		articles = append(articles, NewsArticle{
			Title:   fmt.Sprintf("%s 异动", s.Name),
			Content: fmt.Sprintf("%s 涨跌幅: %+.2f%%", s.Name, s.ChangePct),
			Source:  "ithomer",
			Time:    time.Now().Format("2006-01-02 15:04"),
			Code:    s.Code,
		})
	}

	return articles, nil
}

// GetNewsFromPython fetches news from Python microservice
func GetNewsFromPython(code string) ([]NewsArticle, error) {
	py := defaultManager.PYC
	if py == nil || !py.IsReady() {
		return nil, fmt.Errorf("python service not ready")
	}

	items, err := py.GetNews(code)
	if err != nil {
		return nil, fmt.Errorf("[pyclient] get news: %w", err)
	}

	articles := make([]NewsArticle, len(items))
	for i, item := range items {
		articles[i] = NewsArticle{
			Title:   item.Title,
			Content: item.Content,
			Source:  "akshare",
			Time:    item.Time,
			Code:    code,
		}
	}

	return articles, nil
}

// SetDefaultManager sets the global manager reference
func SetDefaultManager(m *Manager) {
	defaultManager = m
}

var defaultManager *Manager

// HotStockNews fetches hot stock news from Ithomer via HTTP
type HotStockNews struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	ChangePct float64 `json:"change_pct"`
	Volume   float64 `json:"volume"`
	Concept  string  `json:"concept"`
}

func GetHotStockNews() ([]HotStockNews, error) {
	ithomer := defaultManager.Ithomer
	if ithomer == nil {
		return nil, fmt.Errorf("ithomer not configured")
	}

	u := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=30&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:67,m:0+t:68,m:0+t:69,m:0+t:70,m:0+t:71,m:1+t:21,m:1+t:23,m:1+t:24,m:1+t:25,m:1+t:168,m:1+t:171,m:1+t:172,m:1+t:173,m:1+t:174,m:1+t:29,m:1+t:200,m:1+t:201,m:1+t:202,m:1+t:203"

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ithomer] get hot stock news: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ithomer] read body: %w", err)
	}

	var em struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				F2  string  `json:"f2"`
				F3  float64 `json:"f3"`
				F4  float64 `json:"f4"`
				F12 string  `json:"f12"`
				F14 string  `json:"f14"`
				F15 float64 `json:"f15"`
				F16 float64 `json:"f16"`
				F17 float64 `json:"f17"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[ithomer] unmarshal: %w", err)
	}

	var stocks []HotStockNews
	for _, item := range em.Data.Items {
		code := strings.TrimPrefix(item.F12, "sh")
		code = strings.TrimPrefix(code, "sz")
		stocks = append(stocks, HotStockNews{
			Code:      code,
			Name:      item.F2,
			ChangePct: item.F3,
			Volume:    item.F17,
		})
	}

	return stocks, nil
}
