package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Iwencai provides natural language stock search via 问财 (iwencai) API
type Iwencai struct {
	apiKey string
	client *http.Client
}

func NewIwencai() *Iwencai {
	apiKey := os.Getenv("IWENCAI_API_KEY")
	return &Iwencai{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (iw *Iwencai) Name() string { return "iwencai" }

// SearchResult represents a stock found by NL search
type SearchResult struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Match string  `json:"match"` // why it matches the query
}

// Search performs natural language stock search
func (iw *Iwencai) Search(query string) ([]SearchResult, error) {
	if iw.apiKey == "" {
		return nil, fmt.Errorf("[iwencai] API key not configured (set IWENCAI_API_KEY)")
	}

	// Parse NL query if it looks like natural language
	nl := ParseNL(query)
	if nl.Query != "" && len(nl.Filters) > 0 {
		// Rebuild query with filter syntax for iwencai
		query = rebuildQuery(nl)
	}

	// iwencai REST API
	u := "https://api.iwencai.com/api/stock/search"
	body := map[string]interface{}{
		"query":    query,
		"key":      iw.apiKey,
		"page":     1,
		"per_page": 20,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", u, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := iw.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[iwencai] search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("[iwencai] search: status %d: %s", resp.StatusCode, string(body))
	}

	// Parse iwencai response
	var result struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				Code  string  `json:"code"`
				Name  string  `json:"name"`
				Price float64 `json:"price"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[iwencai] decode: %w", err)
	}

	var results []SearchResult
	for _, item := range result.Data.Items {
		results = append(results, SearchResult{
			Code:  item.Code,
			Name:  item.Name,
			Price: item.Price,
			Match: query,
		})
	}
	return results, nil
}

// rebuildQuery converts a NLQuery back to a filter-friendly string for iwencai.
// e.g. NLQuery{Query: "白酒", Filters: [{pe_ttm, "<", 20}]} → "白酒 市盈率小于20"
func rebuildQuery(q NLQuery) string {
	parts := []string{q.Query}
	for _, f := range q.Filters {
		parts = append(parts, fieldToChinese(f.Field)+" "+opToChinese(f.Op)+" "+itoa(f.Value))
	}
	return strings.Join(parts, " ")
}

func fieldToChinese(field string) string {
	switch field {
	case "pe_ttm":
		return "市盈率"
	case "pb":
		return "市净率"
	case "volume_ratio":
		return "成交量"
	case "turnover_rate":
		return "换手率"
	case "total_market_cap":
		return "总市值"
	case "change_pct":
		return "涨跌幅"
	default:
		return field
	}
}

func opToChinese(op string) string {
	switch op {
	case ">":
		return "大于"
	case "<":
		return "小于"
	case ">=":
		return "大于等于"
	case "<=":
		return "小于等于"
	case "=":
		return "等于"
	default:
		return op
	}
}
