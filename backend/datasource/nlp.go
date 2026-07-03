package datasource

import (
	"regexp"
	"strconv"
	"strings"
)

// NLQuery represents a parsed natural language stock query
type NLQuery struct {
	Query   string // sector/industry name (e.g. "白酒", "银行")
	Filters []Filter
}

// Filter is a single constraint on stock attributes
type Filter struct {
	Field string // e.g. "pe_ttm", "volume_ratio", "pb"
	Op    string // "<", ">", "<=", ">=", "="
	Value float64
}

// String returns the filter as "field op value"
func (f Filter) String() string {
	return f.Field + f.Op + itoa(f.Value)
}

// ParseNL parses a natural language query into structured form.
// Examples:
//   "低估值白酒股" → query="白酒", filters=[pe_ttm<20]
//   "市盈率小于20的银行股" → query="银行", filters=[pe_ttm<20]
//   "成交量放大的股票" → query="", filters=[volume_ratio>1.5]
func ParseNL(s string) NLQuery {
	s = strings.TrimSpace(s)
	if s == "" {
		return NLQuery{}
	}

	// Step 1: Extract filters from known patterns
	query := s
	filters := extractFilters(s)

	// Step 2: Remove filter mentions from query text
	query = removeFilterMentions(query)

	// Step 3: Extract sector/industry keywords
	if sector := extractSector(query); sector != "" {
		query = sector
	}

	return NLQuery{Query: query, Filters: filters}
}

// --- Filter extraction ---

var (
	// 市盈率/PE < 20, 大于 30, 小于 15
	peRe = regexp.MustCompile(`(?:市盈率|PE)[^0-9]*([<>]=?|大于|小于|高于|低于|超过|不到|不超过|大于等于|小于等于)\s*([0-9.]+)`)
	// 市净率/PB
	pbRe = regexp.MustCompile(`(?:市净率|PB)[^0-9]*([<>]=?|大于|小于|高于|低于|超过|不到|不超过|大于等于|小于等于)\s*([0-9.]+)`)
	// 成交量放大/缩小/放量/缩量
	volRe = regexp.MustCompile(`成交量(?:放大|放量|缩小|缩量|增加|减少)`)
	// 换手率
	tbRe = regexp.MustCompile(`(?:换手率)[^0-9]*([<>]=?|大于|小于|高于|低于|超过|不到|大于等于|小于等于)\s*([0-9.]+)`)
	// 市值
	mcapRe = regexp.MustCompile(`(?:市值|总市值|流通市值)[^0-9]*([<>]=?|大于|小于|高于|低于|超过|不到|大于等于|小于等于)\s*([0-9.]+)`)
	// 涨幅/跌幅
	pctRe = regexp.MustCompile(`(?:涨幅|涨幅|跌幅|涨跌幅)[^0-9]*([<>]=?|大于|小于|高于|低于|超过|不到|大于等于|小于等于)\s*([0-9.]+)`)
)

func extractFilters(s string) []Filter {
	var filters []Filter

	if m := peRe.FindStringSubmatch(s); m != nil {
		filters = append(filters, Filter{Field: "pe_ttm", Op: opFromStr(m[1]), Value: parseF(m[2])})
	}
	if m := pbRe.FindStringSubmatch(s); m != nil {
		filters = append(filters, Filter{Field: "pb", Op: opFromStr(m[1]), Value: parseF(m[2])})
	}
	if m := tbRe.FindStringSubmatch(s); m != nil {
		filters = append(filters, Filter{Field: "turnover_rate", Op: opFromStr(m[1]), Value: parseF(m[2])})
	}
	if m := mcapRe.FindStringSubmatch(s); m != nil {
		filters = append(filters, Filter{Field: "total_market_cap", Op: opFromStr(m[1]), Value: parseF(m[2])})
	}
	if m := pctRe.FindStringSubmatch(s); m != nil {
		filters = append(filters, Filter{Field: "change_pct", Op: opFromStr(m[1]), Value: parseF(m[2])})
	}

	// Volume keywords → volume_ratio threshold
	if volRe.MatchString(s) {
		filters = append(filters, Filter{Field: "volume_ratio", Op: ">", Value: 1.5})
	}

	return filters
}

func opFromStr(s string) string {
	switch s {
	case ">", "<", ">=", "<=", "=":
		return s
	case "大于", "高于", "超过", "大于等于":
		return ">"
	case "小于", "低于", "不到", "小于等于":
		return "<"
	default:
		return ">"
	}
}

func parseF(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// --- Sector extraction ---

var sectorKeywords = []string{
	"白酒", "银行", "医药", "医疗", "消费", "科技", "新能源", "半导体",
	"芯片", "人工智能", "AI", "机器人", "国防", "军工", "汽车", "地产",
	"保险", "券商", "证券", "煤炭", "钢铁", "电力", "通信", "传媒",
	"农业", "旅游", "教育", "金融", "医药生物", "计算机", "电子",
	"机械", "化工", "建材", "有色", "石油", "运输", "纺织", "食品",
	"家电", "轻工", "商贸", "公用", "建筑", "交运", "石化",
}

func extractSector(s string) string {
	for _, kw := range sectorKeywords {
		if strings.Contains(s, kw) {
			return kw
		}
	}
	return ""
}

// --- Filter removal from query text ---

func removeFilterMentions(s string) string {
	// Remove "XX的" patterns
	s = regexp.MustCompile(`[^\s，,。]+的`).ReplaceAllString(s, "")
	// Remove filter mentions
	for _, re := range []*regexp.Regexp{peRe, pbRe, volRe, tbRe, mcapRe, pctRe} {
		s = re.ReplaceAllString(s, "")
	}
	s = strings.TrimSpace(s)
	return s
}

func itoa(v float64) string {
	if v == float64(int(v)) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
