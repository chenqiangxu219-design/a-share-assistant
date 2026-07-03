package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"a-share-assistant/backend/model"
)

type Sina struct {
	client *http.Client
}

func NewSina() *Sina {
	return &Sina{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Sina) Name() string { return "sina" }

func (s *Sina) GetQuote(code string) (*model.Quote, error) {
	prefix := "sz"
	if strings.HasPrefix(code, "6") {
		prefix = "sh"
	}
	sym := prefix + code

	u := fmt.Sprintf("https://hq.sinajs.cn/list=%s", sym)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[sina] get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[sina] get quote: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[sina] read body: %w", err)
	}

	// Decode GBK
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded, err := io.ReadAll(transform.NewReader(strings.NewReader(string(body)), decoder))
	if err != nil {
		return nil, fmt.Errorf("[sina] decode gbk: %w", err)
	}

	line := strings.TrimSpace(string(decoded))
	if !strings.HasPrefix(line, "var hq_str_") {
		return nil, fmt.Errorf("[sina] invalid response for %s", code)
	}

	// Extract data between quotes
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("[sina] parse error for %s", code)
	}
	data := line[start+1 : end]
	parts := strings.Split(data, ",")
	if len(parts) < 33 {
		return nil, fmt.Errorf("[sina] insufficient data fields: %d", len(parts))
	}

	// Sina fields: 0=name, 1=open, 2=yesterday_close, 3=current_price,
	// 4=high, 5=low, 6=volume(手), 7=amount(元), 8=?, 9=?,
	// 30=date, 31=time
	q := &model.Quote{
		Code:      code,
		Name:      parts[0],
		Open:      parseFloat(parts[1]),
		Yesterday: parseFloat(parts[2]),
		Price:     parseFloat(parts[3]),
		High:      parseFloat(parts[4]),
		Low:       parseFloat(parts[5]),
		Volume:    parseFloat(parts[6]) * 10, // 手 to 股
		Amount:    parseFloat(parts[7]),
		Change:    parseFloat(parts[3]) - parseFloat(parts[2]),
		TimeStamp: time.Now(),
	}
	if q.Yesterday > 0 {
		q.ChangePct = q.Change / q.Yesterday * 100
	}

	return q, nil
}

func (s *Sina) GetKLines(code, period string, count int) ([]model.KLine, error) {
	prefix := "sz"
	if strings.HasPrefix(code, "6") {
		prefix = "sh"
	}
	sym := prefix + code

	// Sina daily K-line API
	scale := map[string]string{"d": "240", "w": "120", "m": "21"}[period]
	if scale == "" {
		scale = "240"
	}

	u := fmt.Sprintf(
		"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=%s&ma=5&enq=type:1&leng=%d",
		sym, scale, count,
	)

	resp, err := s.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[sina] get klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[sina] get klines: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[sina] read klines: %w", err)
	}

	// Strip callback wrapper (e.g., "var _data = [...]" or "callback([...])")
	start := bytes.Index(body, []byte("["))
	end := bytes.LastIndex(body, []byte("]"))
	if start != -1 && end != -1 && end > start {
		body = body[start : end+1]
	}

	var klines []model.KLine
	// Parse JSON array of objects (Sina returns numeric fields as strings)
	var items []struct {
		Date  string `json:"day"`
		Open  string `json:"open"`
		High  string `json:"high"`
		Low   string `json:"low"`
		Close string `json:"close"`
		Vol   string `json:"volume"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("[sina] parse klines: %w", err)
	}

	for _, item := range items {
		klines = append(klines, model.KLine{
			Date:   item.Date,
			Open:   parseFloat(item.Open),
			High:   parseFloat(item.High),
			Low:    parseFloat(item.Low),
			Close:  parseFloat(item.Close),
			Volume: parseFloat(item.Vol),
		})
	}

	return klines, nil
}

// GetFinance returns basic financial data (stub for Sina).
func (s *Sina) GetFinance(code string) (*model.FinanceData, error) {
	return nil, fmt.Errorf("[sina] finance data not supported")
}
