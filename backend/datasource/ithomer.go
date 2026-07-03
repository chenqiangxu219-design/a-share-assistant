package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"a-share-assistant/backend/model"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Ithomer (同花顺) provides hot stocks and concept data
type Ithomer struct {
	client *http.Client
}

func NewIthomer() *Ithomer {
	return &Ithomer{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (i *Ithomer) Name() string { return "ithomer" }

// HotStock represents a trending stock from 同花顺
type HotStock struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ChangePct float64 `json:"change_pct"`
	Volume    float64 `json:"volume"`
	Concept   string  `json:"concept"` // related concept/theme
}

// ConceptTopic represents a concept/theme with attribution
type ConceptTopic struct {
	Name      string   `json:"name"`
	ChangePct float64  `json:"change_pct"`
	Stocks    []string `json:"stocks"` // leading stocks
	Reason    string   `json:"reason"` // attribution
}

func (i *Ithomer) GetHotStocks() ([]HotStock, error) {
	// 同花顺热点: use EastMoney's hot stock ranking as fallback
	// (ithomer's direct API requires JS challenge, use a more accessible endpoint)
	u := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:67,m:0+t:68,m:0+t:69,m:0+t:70,m:0+t:71,m:1+t:21,m:1+t:23,m:1+t:24,m:1+t:25,m:1+t:168,m:1+t:171,m:1+t:172,m:1+t:173,m:1+t:174,m:1+t:29,m:1+t:200,m:1+t:201,m:1+t:202,m:1+t:203,m:1+t:204"

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ithomer] get hot stocks: %w", err)
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
				F2  string  `json:"f2"`  // name
				F3  float64 `json:"f3"`  // change_pct
				F4  float64 `json:"f4"`  // price
				F12 string  `json:"f12"` // code
				F14 string  `json:"f14"` // market
				F15 float64 `json:"f15"` // high
				F16 float64 `json:"f16"` // low
				F17 float64 `json:"f17"` // volume
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[ithomer] unmarshal: %w", err)
	}

	var stocks []HotStock
	for _, item := range em.Data.Items {
		code := strings.TrimPrefix(item.F12, "sh")
		code = strings.TrimPrefix(code, "sz")
		stocks = append(stocks, HotStock{
			Code:      code,
			Name:      item.F2,
			ChangePct: item.F3,
			Volume:    item.F17,
		})
	}
	return stocks, nil
}

func (i *Ithomer) GetConcepts() ([]ConceptTopic, error) {
	// Get concept board data from EastMoney (more stable than ithomer direct)
	u := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=50&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:90+t:13,m:90+t:14,m:90+t:15,m:90+t:16"

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ithomer] get concepts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ithomer] read concepts: %w", err)
	}

	var em struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				F2  string  `json:"f2"`  // name
				F3  float64 `json:"f3"`  // change_pct
				F12 string  `json:"f12"` // board code
				F14 string  `json:"f14"` // lead stock name
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[ithomer] unmarshal concepts: %w", err)
	}

	var concepts []ConceptTopic
	for _, item := range em.Data.Items {
		ct := ConceptTopic{
			Name:      item.F2,
			ChangePct: item.F3,
			Reason:    fmt.Sprintf("%s 领涨", item.F14),
		}
		concepts = append(concepts, ct)
	}
	return concepts, nil
}

// GetQuote stub (ithomer is primarily for hot stocks, not real-time quotes)
func (i *Ithomer) GetQuote(code string) (*model.Quote, error) {
	return nil, fmt.Errorf("[ithomer] quote not supported")
}

// GetKLines stub
func (i *Ithomer) GetKLines(code, period string, count int) ([]model.KLine, error) {
	return nil, fmt.Errorf("[ithomer] klines not supported")
}

// Keep imports used
var _ = strconv.Atoi
var _ = simplifiedchinese.GBK
var _ = transform.NewReader
var _ = strings.TrimSpace
