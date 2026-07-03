package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"a-share-assistant/backend/model"
)

type EastMoney struct {
	client *http.Client
}

func NewEastMoney() *EastMoney {
	return &EastMoney{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *EastMoney) Name() string { return "eastmoney" }

func (e *EastMoney) GetQuote(code string) (*model.Quote, error) {
	secID := codeToSecID(code)
	fields := "f43,f44,f45,f46,f47,f48,f50,f57,f58,f60,f116,f117,f168,f169,f170,f171,f23,f24,f25,f26,f27"
	u := fmt.Sprintf(
		"https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=%s&ut=fa5fd1943c7b386f172d6893dbfba10b",
		secID, fields,
	)

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[eastmoney] get quote: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] read body: %w", err)
	}

	var em struct {
		Code int         `json:"code"`
		Data EMQuoteData `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[eastmoney] unmarshal: %w", err)
	}

	if em.Data.Name == "" {
		return nil, fmt.Errorf("[eastmoney] stock %s not found", code)
	}

	q := &model.Quote{
		Code:         code,
		Name:         em.Data.Name,
		Price:        em.Data.Price,
		Change:       em.Data.Price - em.Data.PrePrice,
		ChangePct:    em.Data.ChgPct,
		Open:         em.Data.Open,
		High:         em.Data.High,
		Low:          em.Data.Low,
		Yesterday:    em.Data.PrePrice,
		Volume:       em.Data.Volume,
		Turnover:     em.Data.Amount,
		TurnoverRate: em.Data.Turnover,
		Amount:       em.Data.Amount,
		TimeStamp:    time.Now(),
	}
	return q, nil
}

func (e *EastMoney) GetKLines(code, period string, count int) ([]model.KLine, error) {
	secID := codeToSecID(code)
	klt := map[string]string{
		"d": "101", "w": "102", "m": "103",
		"1m": "1", "5m": "5", "15m": "15", "30m": "30", "60m": "60",
	}[period]
	if klt == "" {
		klt = "101"
	}

	params := url.Values{}
	params.Set("secid", secID)
	params.Set("klt", klt)
	params.Set("fqt", "1") // forward adjusted
	params.Set("fields", "f1,f2,f3,f4,f5,f6")
	params.Set("lmt", strconv.Itoa(count))
	params.Set("end", "20500101")
	params.Set("ut", "fa5fd1943c7b386f172d6893dbfba10b")

	u := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?%s", params.Encode())

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Referer", "https://quote.eastmoney.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] get klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[eastmoney] get klines: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] read klines: %w", err)
	}

	var em struct {
		Code int `json:"code"`
		Data struct {
			Items []string `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[eastmoney] unmarshal klines: %w", err)
	}

	// Parse K-line data
	var klines []model.KLine
	for _, item := range em.Data.Items {
		parts := strings.Split(item, ",")
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

// CapitalFlow represents daily capital flow data
type CapitalFlow struct {
	Date       string  `json:"date"`
	MainNet    float64 `json:"main_net"`     // 主力净流入 (估算)
	MainNetPct float64 `json:"main_net_pct"` // 净流入占比
	LargeNet   float64 `json:"large_net"`    // 大单净流入
	MediumNet  float64 `json:"medium_net"`   // 中单净流入
	SmallNet   float64 `json:"small_net"`    // 小单净流入
}

// GetCapitalFlow estimates capital flow from K-line data
// Uses price change * volume as proxy for money flow
func (e *EastMoney) GetCapitalFlow(code string, days int) ([]CapitalFlow, error) {
	klines, err := e.GetKLines(code, "d", days)
	if err != nil {
		return nil, err
	}

	var flows []CapitalFlow
	for _, k := range klines {
		// Estimate: main force flow = price_change * volume * 0.6
		priceChange := k.Close - k.Open
		amount := priceChange * k.Volume // 元
		mainNet := amount * 0.6
		largeNet := amount * 0.4
		mediumNet := -amount * 0.3
		smallNet := -amount * 0.1

		// Percentage: main_net / total_amount
		totalAmount := k.Volume * k.Close
		mainNetPct := 0.0
		if totalAmount > 0 {
			mainNetPct = mainNet / totalAmount * 100
		}

		flows = append(flows, CapitalFlow{
			Date:       k.Date,
			MainNet:    mainNet,
			MainNetPct: mainNetPct,
			LargeNet:   largeNet,
			MediumNet:  mediumNet,
			SmallNet:   smallNet,
		})
	}

	return flows, nil
}

func (e *EastMoney) GetFinance(code string) (*model.FinanceData, error) {
	secID := codeToSecID(code)
	fields := "f50,f23,f24,f25,f26,f27"
	u := fmt.Sprintf(
		"https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=%s&ut=fa5fd1943c7b386f172d6893dbfba10b",
		secID, fields,
	)

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] get finance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[eastmoney] get finance: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[eastmoney] read body: %w", err)
	}

	var em struct {
		Code int         `json:"code"`
		Data EMQuoteData `json:"data"`
	}
	if err := json.Unmarshal(body, &em); err != nil {
		return nil, fmt.Errorf("[eastmoney] unmarshal: %w", err)
	}

	if em.Data.Name == "" {
		return nil, fmt.Errorf("[eastmoney] stock %s not found", code)
	}

	return &model.FinanceData{
		Code:           code,
		Name:           em.Data.Name,
		PE_TTM:         em.Data.PE_TTM,
		PB:             em.Data.PB,
		ROE:            em.Data.ROE,
		Total_Market_Cap: em.Data.TotalCap,
		Circulating_MCap: em.Data.CircCap,
	}, nil
}

type EMQuoteData struct {
	ChangeAmt float64 `json:"f49"`
	PrePrice  float64 `json:"f46"`
	Open      float64 `json:"f47"`
	Price     float64 `json:"f43"`
	High      float64 `json:"f44"`
	Low       float64 `json:"f45"`
	ChgPct    float64 `json:"f48"`
	Name      string  `json:"f50"`
	Code      string  `json:"f51"`
	Amount    float64 `json:"f57"`
	Volume    float64 `json:"f116"`
	Turnover  float64 `json:"f117"`
	PE_TTM    float64 `json:"f23"`
	PB        float64 `json:"f24"`
	ROE       float64 `json:"f25"`
	TotalCap  float64 `json:"f26"`
	CircCap   float64 `json:"f27"`
}

func codeToSecID(code string) string {
	code = strings.TrimPrefix(code, "sh")
	code = strings.TrimPrefix(code, "sz")
	if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") || strings.HasPrefix(code, "11") {
		return "1." + code
	}
	return "0." + code
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
