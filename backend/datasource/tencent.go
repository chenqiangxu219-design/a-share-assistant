package datasource

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"a-share-assistant/backend/model"
)

type Tencent struct {
	client *http.Client
}

func NewTencent() *Tencent {
	return &Tencent{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Tencent) Name() string { return "tencent" }

func (t *Tencent) GetQuote(code string) (*model.Quote, error) {
	prefix := "sz"
	if strings.HasPrefix(code, "6") {
		prefix = "sh"
	}
	sym := prefix + code

	u := fmt.Sprintf("https://qt.gtimg.cn/q=%s", sym)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[tencent] get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[tencent] get quote: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[tencent] read body: %w", err)
	}

	// Decode GBK
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded, err := io.ReadAll(transform.NewReader(strings.NewReader(string(body)), decoder))
	if err != nil {
		return nil, fmt.Errorf("[tencent] decode gbk: %w", err)
	}

	line := strings.TrimSpace(string(decoded))
	if !strings.HasPrefix(line, "v") {
		return nil, fmt.Errorf("[tencent] invalid response for %s", code)
	}

	// Parse ~ delimited fields
	fields := strings.Split(line, "~")
	if len(fields) < 45 {
		return nil, fmt.Errorf("[tencent] insufficient fields: %d", len(fields))
	}

	// Tencent qt.gtimg.cn fields:
	// 1=market_code, 2=code, 3=name, 4=price, 5=??, 6=??, 7=??,
	// 12=yesterday_close, 13=open, 14=high, 15=low,
	// 17=??, 18=??, 19=??, 20=??,
	// 21=??, 22=??, 23=??, 24=??, 25=??,
	// 30=volume(手), 31=amount(万), 32=??, 33=??,
	// 34=??, 35=??, 36=??, 37=??, 38=??, 39=??, 40=??,
	// 41=??, 42=??, 43=??, 44=turnover_rate
	yesterday := parseFloat(fields[12])
	price := parseFloat(fields[4])
	open := parseFloat(fields[13])
	high := parseFloat(fields[14])
	low := parseFloat(fields[15])
	volume := parseFloat(fields[30]) * 100 // 手 to 股
	amount := parseFloat(fields[31]) * 10000 // 万 to 元

	q := &model.Quote{
		Code:         code,
		Name:         fields[3],
		Price:        price,
		Open:         open,
		High:         high,
		Low:          low,
		Yesterday:    yesterday,
		Volume:       volume,
		Amount:       amount,
		TurnoverRate: parseFloat(fields[44]),
		Change:       price - yesterday,
		TimeStamp:    time.Now(),
	}
	if yesterday > 0 {
		q.ChangePct = (price - yesterday) / yesterday * 100
	}

	return q, nil
}

func (t *Tencent) GetKLines(code, period string, count int) ([]model.KLine, error) {
	prefix := "sz"
	if strings.HasPrefix(code, "6") {
		prefix = "sh"
	}
	sym := prefix + code

	// Use Tencent daily K-line interface
	periodNum := map[string]string{"d": "1", "w": "2", "m": "3"}[period]
	if periodNum == "" {
		periodNum = "1"
	}

	u := fmt.Sprintf(
		"https://web.ifeng.com/cv6/stock/chart/ifengChart?code=%s&type=%s&avg=1&assist=MA&length=%d",
		sym, periodNum, count,
	)

	resp, err := t.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[tencent] get klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[tencent] get klines: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[tencent] read klines: %w", err)
	}

	// Decode GBK
	decoded, err := io.ReadAll(transform.NewReader(strings.NewReader(string(body)), simplifiedchinese.GBK.NewDecoder()))
	if err == nil {
		body = decoded
	}

	var klines []model.KLine
	// Parse the ifeng chart data format
	// The response contains date,open,high,low,close,volume data
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 10 {
			continue
		}
		// Split by comma and extract OHLCV
		parts := strings.Split(line, ",")
		if len(parts) >= 6 {
			klines = append(klines, model.KLine{
				Date:   parts[0],
				Open:   parseFloat(parts[1]),
				High:   parseFloat(parts[2]),
				Low:    parseFloat(parts[3]),
				Close:  parseFloat(parts[4]),
				Volume: parseFloat(parts[5]),
			})
		}
	}

	return klines, nil
}

// GetFinance returns basic financial data (stub for Tencent).
func (t *Tencent) GetFinance(code string) (*model.FinanceData, error) {
	return nil, fmt.Errorf("[tencent] finance data not supported")
}
