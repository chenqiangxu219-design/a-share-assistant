package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"a-share-assistant/backend/chat"
	"a-share-assistant/backend/indicator"
	"a-share-assistant/backend/model"

	"github.com/gin-gonic/gin"

	"a-share-assistant/backend/strategy"
)

// Max iterations for the agent loop
const maxAgentIterations = 5

// Tool call parser: [TOOL tool_name] args [/TOOL]
var toolCallRe = regexp.MustCompile(`\[TOOL\s+(\w+)\]\s*(.*?)\s*\[/TOOL\]`)

// toolHandler executes a tool and returns formatted text result.
type toolHandler func(h *Handler, args string) string

// toolHandlers maps tool names to execution functions.
var toolHandlers = map[string]toolHandler{
	"get_quote":       getQuoteTool,
	"get_kline":       getKLineTool,
	"get_signals":     getSignalsTool,
	"get_news":        getNewsTool,
	"search":          searchTool,
	"get_hot":         getHotTool,
	"compare":         compareTool,        // NEW: multi-stock comparison
	"sector_analysis": sectorAnalysisTool, // NEW: sector rotation analysis
	"portfolio_review": portfolioReviewTool, // NEW: portfolio P&L review
}

// --- Signal Context ---

// SignalContext holds detailed signal information for AI interpretation
type SignalContext struct {
	Quote        *model.Quote           `json:"quote"`
	Signal       model.Signal           `json:"signal"`
	Indicators   model.IndicatorResult  `json:"indicators"`
	StrategySigs []model.Signal         `json:"strategy_signals"`
	Trend        string                 `json:"trend"` // "upward" / "downward" / "consolidating"
}

// buildSignalContext generates detailed signal context for a stock
func (h *Handler) buildSignalContext(code string) SignalContext {
	ctx := SignalContext{
		StrategySigs: make([]model.Signal, 0),
	}

	// Get quote
	quote, err := h.DS.GetQuote(code)
	if err == nil {
		ctx.Quote = quote
	}

	// Get K-lines and indicators
	klines, err := h.DS.GetKLines(code, "d", 60)
	if err == nil && len(klines) > 0 {
		ind := indicator.CalcAllIndicators(klines, code)
		sig := h.Engine.Analyze(code, quote.Name, quote.Price, ind)
		engine := strategy.NewEngine()
		allSigs := engine.Analyze(ind, quote.Price)

		ctx.Signal = sig
		ctx.Indicators = ind
		ctx.StrategySigs = allSigs

		// Determine trend from MA alignment
		if len(ind.MA.MA5) > 0 && len(ind.MA.MA20) > 0 {
			ma5 := getLastNonZero(ind.MA.MA5)
			ma20 := getLastNonZero(ind.MA.MA20)
			ma60 := getLastNonZero(ind.MA.MA60)
			if ma5 > ma20 && ma20 > ma60 {
				ctx.Trend = "多头排列（上涨趋势）"
			} else if ma5 < ma20 && ma20 < ma60 {
				ctx.Trend = "空头排列（下跌趋势）"
			} else {
				ctx.Trend = "震荡整理"
			}
		}
	}

	return ctx
}

// formatSignalContext formats the signal context as a readable string for the AI
func formatSignalContext(ctx SignalContext) string {
	var parts []string

	// Quote info
	if ctx.Quote != nil {
		parts = append(parts, fmt.Sprintf("## 实时报价\n- 名称: %s (%s)\n- 价格: %.2f\n- 涨跌: %+.2f%%\n- 高: %.2f 低: %.2f 开: %.2f",
			ctx.Quote.Name, ctx.Quote.Code, ctx.Quote.Price, ctx.Quote.ChangePct,
			ctx.Quote.High, ctx.Quote.Low, ctx.Quote.Open))
	}

	// Signal summary
	sig := ctx.Signal
	parts = append(parts, fmt.Sprintf("\n## 综合信号\n- 方向: %s\n- 强度: %d/5\n- 综合评分: %d\n- 说明: %s",
		sig.Direction, sig.Strength, sig.Score, sig.Message))

	// Individual strategy signals
	if len(ctx.StrategySigs) > 0 {
		strategyParts := make([]string, 0)
		for _, s := range ctx.StrategySigs {
			strategyParts = append(strategyParts, fmt.Sprintf("- [%s] %s (强度:%d)", s.Direction, s.Indicators[0], s.Strength))
		}
		parts = append(parts, fmt.Sprintf("\n## 各指标信号\n%s", strings.Join(strategyParts, "\n")))
	}

	// Indicator values
	ind := ctx.Indicators
	if len(ind.MA.MA5) > 0 {
		ma5 := getLastNonZero(ind.MA.MA5)
		ma10 := getLastNonZero(ind.MA.MA10)
		ma20 := getLastNonZero(ind.MA.MA20)
		ma60 := getLastNonZero(ind.MA.MA60)
		parts = append(parts, fmt.Sprintf("\n## 均线系统\n- MA5: %.2f  MA10: %.2f  MA20: %.2f  MA60: %.2f",
			ma5, ma10, ma20, ma60))

		// MA comparison
		if ma5 > 0 && ma20 > 0 {
			if ma5 > ma20 {
				parts = append(parts, "- MA5 > MA20: 短期均线在中期均线上方，偏多")
			} else {
				parts = append(parts, "- MA5 < MA20: 短期均线在中期均线下方，偏空")
			}
		}
	}

	if len(ind.MACD.DIF) > 0 {
		dif := getLastNonZero(ind.MACD.DIF)
		dea := getLastNonZero(ind.MACD.DEA)
		hist := getLastNonZero(ind.MACD.Hist)
		parts = append(parts, fmt.Sprintf("- MACD: DIF %.4f, DEA %.4f, HIST %.4f", dif, dea, hist))
		if dif > dea {
			parts = append(parts, "- DIF > DEA: 多头动能")
		} else {
			parts = append(parts, "- DIF < DEA: 空头动能")
		}
	}

	if len(ind.RSI.RSI6) > 0 {
		rsi6 := getLastNonZero(ind.RSI.RSI6)
		rsi12 := getLastNonZero(ind.RSI.RSI12)
		parts = append(parts, fmt.Sprintf("- RSI: RSI6 %.2f, RSI12 %.2f", rsi6, rsi12))
		if rsi6 > 70 {
			parts = append(parts, "- RSI6 > 70: 超买区域")
		} else if rsi6 < 30 {
			parts = append(parts, "- RSI6 < 30: 超卖区域")
		}
	}

	if len(ind.KDJ.K) > 0 {
		k := getLastNonZero(ind.KDJ.K)
		d := getLastNonZero(ind.KDJ.D)
		j := getLastNonZero(ind.KDJ.J)
		parts = append(parts, fmt.Sprintf("- KDJ: K %.2f, D %.2f, J %.2f", k, d, j))
		if j > 100 {
			parts = append(parts, "- J > 100: 超买")
		} else if j < 0 {
			parts = append(parts, "- J < 0: 超卖")
		}
	}

	if len(ind.BOLL.Mid) > 0 {
		mid := getLastNonZero(ind.BOLL.Mid)
		upper := getLastNonZero(ind.BOLL.Up)
		lower := getLastNonZero(ind.BOLL.Down)
		parts = append(parts, fmt.Sprintf("- BOLL: 上轨 %.2f, 中轨 %.2f, 下轨 %.2f", upper, mid, lower))
		if ctx.Quote != nil {
			if ctx.Quote.Price > upper {
				parts = append(parts, "- 股价突破上轨: 可能超买")
			} else if ctx.Quote.Price < lower {
				parts = append(parts, "- 股价触及下轨: 可能超卖")
			} else {
				parts = append(parts, "- 股价在中轨附近震荡")
			}
		}
	}

	if len(ctx.Trend) > 0 {
		parts = append(parts, fmt.Sprintf("\n## 趋势判断\n%s", ctx.Trend))
	}

	return strings.Join(parts, "\n")
}

// --- Tool Implementations ---

func getQuoteTool(h *Handler, args string) string {
	code := strings.ToUpper(strings.TrimSpace(args))
	if code == "" {
		return "错误: 请提供股票代码"
	}

	q, err := h.DS.GetQuote(code)
	if err != nil {
		return fmt.Sprintf("获取 %s 报价失败: %v", code, err)
	}

	// Cache the result
	h.Cache.SetQuote(code, q)

	return fmt.Sprintf(
		"%s (%s): 价格 %.2f, 涨跌 %+.2f%%, 涨跌幅 %+.2f%%, 最高 %.2f, 最低 %.2f, 成交量 %.0f, 成交额 %.0f",
		q.Name, code, q.Price, q.Change, q.ChangePct, q.High, q.Low, q.Volume, q.Amount,
	)
}

func getKLineTool(h *Handler, args string) string {
	parts := strings.Split(args, ",")
	if len(parts) < 1 {
		return "错误: 请提供股票代码。格式: code,period,count"
	}

	code := strings.ToUpper(strings.TrimSpace(parts[0]))
	period := "d"
	count := 30

	if len(parts) >= 2 {
		period = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		if c, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil && c > 0 {
			count = c
		}
	}

	if !strings.Contains("d,w,m", period) {
		return fmt.Sprintf("错误: 不支持的周期 %s，请使用 d(日)/w(周)/m(月)", period)
	}
	if count > 500 {
		count = 500
	}

	klines, err := h.DS.GetKLines(code, period, count)
	if err != nil {
		return fmt.Sprintf("获取 %s K 线失败: %v", code, err)
	}

	// Cache the result
	key := CacheKey(code, period)
	h.Cache.SetKLines(key, klines)

	if len(klines) == 0 {
		return fmt.Sprintf("%s 无 K 线数据", code)
	}

	// Format last N candles
	n := len(klines)
	if n > 10 {
		n = 10
	}

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		k := klines[len(klines)-n+i]
		lines[i] = fmt.Sprintf("%s | O:%.2f H:%.2f L:%.2f C:%.2f V:%.0f",
			k.Date, k.Open, k.High, k.Low, k.Close, k.Volume)
	}

	total := len(klines)
	return fmt.Sprintf("%s K 线(%s, 共%d根):\n%s", code, period, total,
		strings.Join(lines, "\n"))
}

func getSignalsTool(h *Handler, args string) string {
	code := strings.ToUpper(strings.TrimSpace(args))
	if code == "" {
		return "错误: 请提供股票代码"
	}

	// Get quote
	quote, err := h.DS.GetQuote(code)
	if err != nil {
		quote = &model.Quote{Code: code, Name: code} // fallback
	} else {
		h.Cache.SetQuote(code, quote)
	}

	// Get K-lines
	klines, err := h.DS.GetKLines(code, "d", 60)
	if err != nil || len(klines) == 0 {
		return fmt.Sprintf("%s 无足够 K 线数据计算技术指标", code)
	}

	key := CacheKey(code, "d")
	h.Cache.SetKLines(key, klines)

	// Calculate indicators
	ind := indicator.CalcAllIndicators(klines, code)
	sig := h.Engine.Analyze(code, quote.Name, quote.Price, ind)

	// Format signal
	indicators := strings.Join(sig.Indicators, ", ")
	if indicators == "" {
		indicators = "无"
	}

	return fmt.Sprintf(
		"%s (%s) 技术分析:\n方向: %s | 强度: %d/5 | 综合评分: %d\n触发指标: %s\n说明: %s",
		code, quote.Name, sig.Direction, sig.Strength, sig.Score, indicators, sig.Message,
	)
}

func getNewsTool(h *Handler, args string) string {
	code := strings.ToUpper(strings.TrimSpace(args))
	if code == "" {
		return "错误: 请提供股票代码"
	}

	py := h.DS.PYC
	if py == nil || !py.IsReady() {
		return "新闻服务暂不可用（Python 微服务未连接）"
	}

	items, err := py.GetNews(code)
	if err != nil {
		return fmt.Sprintf("获取 %s 新闻失败: %v", code, err)
	}

	if len(items) == 0 {
		return fmt.Sprintf("%s 暂无相关新闻", code)
	}

	// Format last 5
	n := len(items)
	if n > 5 {
		n = 5
	}

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("- %s (%s)", items[i].Title, items[i].Time)
	}

	return fmt.Sprintf("%s 相关新闻:\n%s", code, strings.Join(lines, "\n"))
}

func searchTool(h *Handler, args string) string {
	query := strings.TrimSpace(args)
	if query == "" {
		return "错误: 请提供搜索关键词"
	}

	iw := h.DS.Iwencai
	if iw == nil {
		return "搜索服务暂不可用"
	}

	results, err := iw.Search(query)
	if err != nil {
		return fmt.Sprintf("搜索 \"%s\" 失败: %v", query, err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("未找到与 \"%s\" 相关的股票", query)
	}

	n := len(results)
	if n > 10 {
		n = 10
	}

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		r := results[i]
		lines[i] = fmt.Sprintf("- %s (%s): %s", r.Name, r.Code, r.Match)
	}

	return fmt.Sprintf("搜索 \"%s\" 结果:\n%s", query, strings.Join(lines, "\n"))
}

func getHotTool(h *Handler, args string) string {
	// Try Python akshare first
	py := h.DS.PYC
	if py != nil && py.IsReady() {
		items, err := py.GetHotStocksAkshare()
		if err == nil && len(items) > 0 {
			n := len(items)
			if n > 10 {
				n = 10
			}
			lines := make([]string, n)
			for i := 0; i < n; i++ {
				lines[i] = fmt.Sprintf("- %s: %+.2f%%", items[i].Name, items[i].ChangePct)
			}
			return fmt.Sprintf("热门股票:\n%s", strings.Join(lines, "\n"))
		}
	}

	// Fallback to Ithomer
	ithomer := h.DS.Ithomer
	if ithomer == nil {
		return "热点数据暂不可用"
	}

	stocks, err := ithomer.GetHotStocks()
	if err != nil {
		return fmt.Sprintf("获取热点失败: %v", err)
	}

	if len(stocks) == 0 {
		return "暂无热门股票数据"
	}

	n := len(stocks)
	if n > 10 {
		n = 10
	}

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("- %s (%s): %+.2f%%", stocks[i].Name, stocks[i].Code, stocks[i].ChangePct)
	}

	return fmt.Sprintf("热门股票:\n%s", strings.Join(lines, "\n"))
}

// --- System Prompt ---

func buildAgentSystemPrompt(stockContext string) string {
	sb := &strings.Builder{}

	sb.WriteString("你是 A 股分析助手。你可以调用工具获取实时数据，然后给出分析结论。\n\n")

	sb.WriteString("## 可用工具\n")
	sb.WriteString("使用格式: [TOOL tool_name] 参数 [/TOOL]\n\n")

	// Tool definitions
	sb.WriteString("### get_quote — 获取股票实时报价\n")
	sb.WriteString("参数: 股票代码\n")
	sb.WriteString("示例: [TOOL get_quote] 600519 [/TOOL]\n\n")

	sb.WriteString("### get_kline — 获取 K 线数据\n")
	sb.WriteString("参数: code,period,count（period: d=日, w=周, m=月）\n")
	sb.WriteString("示例: [TOOL get_kline] 600519,d,30 [/TOOL]\n\n")

	sb.WriteString("### get_signals — 获取技术分析信号（含 MA/MACD/RSI/KDJ 等）\n")
	sb.WriteString("参数: 股票代码\n")
	sb.WriteString("示例: [TOOL get_signals] 600519 [/TOOL]\n\n")

	sb.WriteString("### get_news — 获取股票相关新闻\n")
	sb.WriteString("参数: 股票代码\n")
	sb.WriteString("示例: [TOOL get_news] 600519 [/TOOL]\n\n")

	sb.WriteString("### search — 自然语言搜索股票\n")
	sb.WriteString("参数: 搜索词\n")
	sb.WriteString("示例: [TOOL search] 低估值白酒股 [/TOOL]\n\n")

	sb.WriteString("### get_hot — 获取热门股票排行\n")
	sb.WriteString("参数: 无\n")
	sb.WriteString("示例: [TOOL get_hot] [/TOOL]\n\n")

	// Stock context (pre-fetched quote if code provided)
	if stockContext != "" {
		sb.WriteString("## 当前分析标的\n")
		sb.WriteString(stockContext)
		sb.WriteString("\n")
	}

	sb.WriteString("## 使用规则\n")
	sb.WriteString("1. 每次只调用一个工具\n")
	sb.WriteString("2. 收到工具返回结果后，继续分析或调用下一个工具\n")
	sb.WriteString("3. 收集足够数据后，直接给出分析结论（不再用 [TOOL]）\n")
	sb.WriteString("4. 回答格式: 先结论 → 数据支撑 → 风险提示\n")
	sb.WriteString("5. 末尾加 ⚠️仅供参考，不构成投资建议\n")

	return sb.String()
}

// --- Tool Call Parsing ---

// parseToolCall extracts tool name and args from LLM response.
// Returns (toolName, args, found). If found is false, it's a final answer.
func parseToolCall(response string) (string, string, bool) {
	matches := toolCallRe.FindStringSubmatch(response)
	if matches == nil {
		return "", "", false
	}
	return matches[1], strings.TrimSpace(matches[2]), true
}

// buildFinalSystemPrompt creates a lighter system prompt for the final analysis.
// It tells the LLM to synthesize collected data rather than call more tools.
func buildFinalSystemPrompt(_ string) string {
	return "你是 A 股分析助手。你已经收集了足够的数据，现在需要给出完整的分析结论。\n\n" +
		"回答格式:\n" +
		"1. **核心结论** — 一句话概括\n" +
		"2. **数据支撑** — 关键数据点\n" +
		"3. **风险提示** — 潜在风险\n" +
		"\n末尾加 ⚠️仅供参考，不构成投资建议\n" +
		"\n注意: 这是最终回答，不要再调用工具。"
}

// buildSignalInterpretPrompt creates a prompt for signal interpretation.
// It tells the LLM how to analyze the provided signal data.
func buildSignalInterpretPrompt(signalCtx SignalContext) string {
	sb := &strings.Builder{}

	sb.WriteString("你是 A 股技术分析专家。请根据以下信号数据，给出专业的信号解读。\n\n")

	sb.WriteString("## 信号解读规则\n")
	sb.WriteString("1. **综合判断**: 结合所有指标给出整体方向（买入/卖出/观望）\n")
	sb.WriteString("2. **指标共振**: 多个指标同向时信号更强，反向时需谨慎\n")
	sb.WriteString("3. **强弱判断**: 评分 > 3 为强信号，1-3 为中等，≤ 1 为弱信号\n")
	sb.WriteString("4. **趋势确认**: 趋势方向与信号方向一致时更可靠\n")
	sb.WriteString("5. **风险提示**: 指出潜在风险因素（超买/超卖/背离等）\n\n")

	sb.WriteString("## 回答格式\n")
	sb.WriteString("1. **信号结论** — 一句话总结当前信号\n")
	sb.WriteString("2. **信号分析** — 逐项分析关键指标\n")
	sb.WriteString("3. **操作建议** — 基于信号给出操作建议\n")
	sb.WriteString("4. **风险提示** — 指出需要注意的风险\n\n")

	// Format signal context
	sb.WriteString("## 当前信号数据\n")
	sb.WriteString(formatSignalContext(signalCtx))
	sb.WriteString("\n\n")

	// User question
	sb.WriteString("---\n")
	sb.WriteString("请基于以上数据给出信号解读。\n")

	return sb.String()
}

// --- Agent Loop ---

// runAgentLoop runs the agent loop with up to maxAgentIterations iterations.
// For non-streaming mode, it returns the final text.
func (h *Handler) runAgentLoop(userMessage string, systemPrompt string) string {
	messages := []chat.Message{
		{Role: "user", Content: userMessage},
	}

	for i := 0; i < maxAgentIterations; i++ {
		response, err := chat.ChatOnce(messages, systemPrompt)
		if err != nil {
			return fmt.Sprintf("[ERROR] LLM 调用失败: %v", err)
		}

		toolName, args, found := parseToolCall(response)
		if found {
			handler, ok := toolHandlers[toolName]
			if !ok {
				result := fmt.Sprintf("未知工具: %s", toolName)
				messages = append(messages, chat.Message{Role: "user", Content: "工具返回: " + result})
				continue
			}
			result := handler(h, args)
			messages = append(messages, chat.Message{Role: "user", Content: "工具返回: " + result})
			continue
		}

		// Final answer
		return response
	}

	return "[超时] 分析迭代次数已达上限，请尝试更具体的问题"
}

// runAgentLoopWithMemory is like runAgentLoop but with conversation memory.
func (h *Handler) runAgentLoopWithMemory(userMessage string, systemPrompt string, sessionID string) string {
	// Get recent conversation history (last 5 messages for context)
	history := h.Memory.GetHistory(sessionID, 5)

	// Build messages with history
	messages := make([]chat.Message, len(history))
	for i, msg := range history {
		messages[i] = msg
	}

	// Add current user message
	messages = append(messages, chat.Message{Role: "user", Content: userMessage})

	for i := 0; i < maxAgentIterations; i++ {
		response, err := chat.ChatOnce(messages, systemPrompt)
		if err != nil {
			return fmt.Sprintf("[ERROR] LLM 调用失败：%v", err)
		}

		toolName, args, found := parseToolCall(response)
		if found {
			handler, ok := toolHandlers[toolName]
			if !ok {
				result := fmt.Sprintf("未知工具：%s", toolName)
				messages = append(messages, chat.Message{Role: "user", Content: "工具返回：" + result})
				continue
			}
			result := handler(h, args)
			messages = append(messages, chat.Message{Role: "user", Content: "工具返回：" + result})
			continue
		}

		// Final answer - save to memory before returning
		h.Memory.AddMessage(sessionID, chat.Message{Role: "assistant", Content: response})
		return response
	}

	return "[超时] 分析迭代次数已达上限，请尝试更具体的问题"
}

// streamAgentLoop runs the agent loop with SSE streaming for tool status and final answer.
func (h *Handler) streamAgentLoop(userMessage, systemPrompt string, w http.ResponseWriter, flush func(), eventChan chan string) {
	messages := []chat.Message{
		{Role: "user", Content: userMessage},
	}

	for i := 0; i < maxAgentIterations; i++ {
		response, err := chat.ChatOnce(messages, systemPrompt)
		if err != nil {
			eventChan <- fmt.Sprintf("[ERROR] LLM 调用失败：%v", err)
			return
		}

		toolName, args, found := parseToolCall(response)
		if found {
			handler, ok := toolHandlers[toolName]
			if !ok {
				result := fmt.Sprintf("未知工具：%s", toolName)
				messages = append(messages, chat.Message{Role: "user", Content: "工具返回：" + result})
				eventChan <- fmt.Sprintf("[未知工具：%s]", toolName)
				continue
			}

			// Stream tool start event
			eventChan <- fmt.Sprintf("tool_start:%s:%s", toolName, args)

			result := handler(h, args)

			// Stream tool done event
			eventChan <- fmt.Sprintf("tool_done:%s", toolName)
			messages = append(messages, chat.Message{Role: "user", Content: "工具返回：" + result})
			continue
		}

		// Final answer — make a fresh streaming call with all context
		eventChan <- "final_start"

		// Add a final-answer instruction to get a comprehensive response
		finalMessages := append([]chat.Message{}, messages...)
		finalMessages = append(finalMessages, chat.Message{
			Role:    "assistant",
			Content: response,
		})
		finalMessages = append(finalMessages, chat.Message{
			Role: "user", Content: "基于以上收集的数据，给出完整的分析结论。先结论后分析，末尾加 ⚠️ 仅供参考。",
		})

		// Build system prompt for final answer (less emphasis on tools)
		finalSystem := systemPrompt + "\n\n请给出完整的分析结论，包括：当前趋势、关键指标解读、操作建议。末尾加 ⚠️ 仅供参考。"

		finalResponse, err := chat.ChatOnce(finalMessages, finalSystem)
		if err != nil {
			eventChan <- fmt.Sprintf("[ERROR] 最终分析失败：%v", err)
			return
		}

		eventChan <- finalResponse
	}
}

// streamAgentLoopWithMemory is like streamAgentLoop but with conversation memory.
func (h *Handler) streamAgentLoopWithMemory(userMessage, systemPrompt string, w http.ResponseWriter, flush func(), eventChan chan string, sessionID string) {
	// Get recent conversation history (last 5 messages for context)
	history := h.Memory.GetHistory(sessionID, 5)

	// Build messages with history
	messages := make([]chat.Message, len(history))
	for i, msg := range history {
		messages[i] = msg
	}

	// Add current user message
	messages = append(messages, chat.Message{Role: "user", Content: userMessage})

	for i := 0; i < maxAgentIterations; i++ {
		// Get full response first (to check for tool calls)
		response, err := chat.ChatOnce(messages, systemPrompt)
		if err != nil {
			eventChan <- fmt.Sprintf("[ERROR] LLM 调用失败: %v", err)
			return
		}

		toolName, args, found := parseToolCall(response)
		if found {
			handler, ok := toolHandlers[toolName]
			if !ok {
				result := fmt.Sprintf("未知工具: %s", toolName)
				messages = append(messages, chat.Message{Role: "user", Content: "工具返回: " + result})
				eventChan <- fmt.Sprintf("[未知工具: %s]", toolName)
				continue
			}

			// Stream tool start event
			eventChan <- fmt.Sprintf("tool_start:%s:%s", toolName, args)

			result := handler(h, args)

			// Stream tool done event
			eventChan <- fmt.Sprintf("tool_done:%s", toolName)
			messages = append(messages, chat.Message{Role: "user", Content: "工具返回: " + result})
			continue
		}

		// Final answer — make a fresh streaming call with all context
		eventChan <- "final_start"

		// Add a final-answer instruction to get a comprehensive response
		finalMessages := append([]chat.Message{}, messages...)
		finalMessages = append(finalMessages, chat.Message{
			Role:    "assistant",
			Content: response,
		})
		finalMessages = append(finalMessages, chat.Message{
			Role: "user", Content: "基于以上收集的数据，给出完整的分析结论。先结论后分析，末尾加 ⚠️仅供参考。",
		})

		// Build system prompt for final answer (less emphasis on tools)
		finalPrompt := buildFinalSystemPrompt("")

		chat.ChatStream(finalMessages, finalPrompt, w, flush, false)
		return
	}

	eventChan <- "[超时] 分析迭代次数已达上限"
}

// --- ChatHandler (Entry Point) ---

// ChatHandler handles POST /api/chat with agent loop and conversation memory
func (h *Handler) ChatHandler(c *gin.Context) {
	var req struct {
		Message   string `json:"message" binding:"required"`
		Code      string `json:"code"`
		SessionID string `json:"session_id"`
		Stream    bool   `json:"stream"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.Code = strings.ToUpper(req.Code)

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}

	// Build pre-fetched stock context
	var stockContext string
	if req.Code != "" {
		stockContext = h.buildStockContext(req.Code)
	}

	// Build user message with code context
	userMessage := req.Message
	if req.Code != "" {
		userMessage = fmt.Sprintf("我正在分析股票 %s。\n\n%s", req.Code, userMessage)
	}

	systemPrompt := buildAgentSystemPrompt(stockContext)

	if req.Stream {
		h.handleStreamChatWithMemory(c, userMessage, systemPrompt, req.SessionID)
	} else {
		reply := h.runAgentLoopWithMemory(userMessage, systemPrompt, req.SessionID)
		c.JSON(http.StatusOK, gin.H{"reply": reply, "session_id": req.SessionID})
	}
}

// handleStreamChat handles streaming with tool status events
func (h *Handler) handleStreamChat(c *gin.Context, userMessage, systemPrompt string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flush := func() { c.Writer.Flush() }
	eventChan := make(chan string, 50)

	// Writer goroutine — safely handles client disconnect
	go func() {
		defer func() { recover() }() // ignore panic if client disconnected
		for event := range eventChan {
			if c.IsAborted() {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", event)
			flush()
		}
		if !c.IsAborted() {
			fmt.Fprintln(c.Writer, "data: [DONE]")
			flush()
		}
	}()

	// Run agent loop
	h.streamAgentLoop(userMessage, systemPrompt, c.Writer, flush, eventChan)
	close(eventChan)
}

// handleStreamChatWithMemory is like handleStreamChat but with conversation memory.
func (h *Handler) handleStreamChatWithMemory(c *gin.Context, userMessage, systemPrompt string, sessionID string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flush := func() { c.Writer.Flush() }
	eventChan := make(chan string, 50)

	// Writer goroutine — safely handles client disconnect
	go func() {
		defer func() { recover() }() // ignore panic if client disconnected
		for event := range eventChan {
			if c.IsAborted() {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", event)
			flush()
		}
		if !c.IsAborted() {
			fmt.Fprintln(c.Writer, "data: [DONE]")
			flush()
		}
	}()

	// Run agent loop with memory
	h.streamAgentLoopWithMemory(userMessage, systemPrompt, c.Writer, flush, eventChan, sessionID)
	close(eventChan)
}

// --- Legacy Helpers (used by ChatContextHandler) ---

func (h *Handler) buildStockContext(code string) string {
	var parts []string

	// Get quote
	quote, err := h.DS.GetQuote(code)
	if err == nil {
		parts = append(parts, fmt.Sprintf("## 实时报价\n- 名称: %s (%s)\n- 价格: %.2f\n- 涨跌: %+.2f%%\n- 高: %.2f 低: %.2f",
			quote.Name, code, quote.Price, quote.ChangePct, quote.High, quote.Low))
	}

	// Get K-lines and indicators
	klines, err := h.DS.GetKLines(code, "d", 60)
	if err == nil && len(klines) > 0 {
		ind := indicator.CalcAllIndicators(klines, code)
		sig := h.Engine.Analyze(code, quote.Name, quote.Price, ind)

		parts = append(parts, fmt.Sprintf("## 技术分析\n- 综合评分: %d (%s)\n- 信号: %s",
			sig.Score, sig.Direction, sig.Message))

		ma5 := getLastNonZero(ind.MA.MA5)
		ma20 := getLastNonZero(ind.MA.MA20)
		dif := getLastNonZero(ind.MACD.DIF)
		rsi6 := getLastNonZero(ind.RSI.RSI6)

		parts = append(parts, fmt.Sprintf("## 关键指标\n- MA5: %.2f MA20: %.2f\n- MACD DIF: %.4f\n- RSI6: %.2f",
			ma5, ma20, dif, rsi6))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("股票代码: %s (数据获取失败)", code)
	}

	return strings.Join(parts, "\n\n")
}

func getLastNonZero(arr []float64) float64 {
	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i] != 0 {
			return arr[i]
		}
	}
	return 0
}

// ChatContextHandler returns stock context for frontend
func (h *Handler) ChatContextHandler(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	ctx := h.buildStockContext(code)
	c.JSON(http.StatusOK, gin.H{"context": ctx})
}

// SignalHandler handles POST /api/chat/signal for signal interpretation
func (h *Handler) SignalHandler(c *gin.Context) {
	var req struct {
		Code    string `json:"code" binding:"required"`
		Message string `json:"message"`
		Stream  bool   `json:"stream"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.Code = strings.ToUpper(req.Code)

	// Build signal context
	signalCtx := h.buildSignalContext(req.Code)

	if signalCtx.Quote == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("无法获取 %s 数据", req.Code)})
		return
	}

	// Build user message
	userMessage := req.Message
	if userMessage == "" {
		userMessage = "请解读当前信号"
	}
	userMessage = fmt.Sprintf("我正在分析股票 %s (%s) 的技术信号。\n\n%s", req.Code, signalCtx.Quote.Name, userMessage)

	systemPrompt := buildSignalInterpretPrompt(signalCtx)

	if req.Stream {
		h.handleStreamChat(c, userMessage, systemPrompt)
	} else {
		reply := h.runAgentLoop(userMessage, systemPrompt)
		c.JSON(http.StatusOK, gin.H{"reply": reply})
	}
}

// RegisterChatRoutes adds chat routes
func RegisterChatRoutes(api *gin.RouterGroup, handler *Handler) {
	api.POST("/chat", handler.ChatHandler)
	api.POST("/chat/signal", handler.SignalHandler)
	api.GET("/chat/context/:code", handler.ChatContextHandler)
}

// --- New AI Tools for All-Purpose Capability ---

// compareTool compares multiple stocks
func compareTool(h *Handler, args string) string {
	// Args format: "code1,code2,..." or "code1 code2 ..."
	codes := strings.Split(args, ",")
	if len(codes) < 2 {
		return "请提供至少两个股票代码进行比较"
	}

	var results []string
	for _, code := range codes {
		code = strings.TrimSpace(strings.ToUpper(code))
		if q, err := h.DS.GetQuote(code); err == nil {
			results = append(results, fmt.Sprintf("%s (%s): %.2f 元，涨幅 %.2f%%", q.Name, code, q.Price, q.ChangePct))
		} else {
			results = append(results, fmt.Sprintf("%s: 数据不可用", code))
		}
	}

	return "股票对比:\n" + strings.Join(results, "\n")
}

// sectorAnalysisTool analyzes sector rotation and trends
func sectorAnalysisTool(h *Handler, args string) string {
	// Args: optional sector name or "all" for all sectors
	engine := strategy.NewSectorRotationEngine()

	// Get sector data from datasource
	sectors, _ := h.DS.GetSectorRotation()

	description := engine.AnalyzeSectorRotation(sectors)
	return fmt.Sprintf("板块轮动分析:\n%s", description)
}

// getQuotesForAllPositions is a helper to fetch quotes for all portfolio positions
func (h *Handler) getQuotesForAllPositions() map[string]*model.Quote {
	return h.Portfolio.getQuotesForPositions(h.DS)
}

// portfolioReviewTool reviews portfolio P&L and positions
func portfolioReviewTool(h *Handler, args string) string {
	// Args: optional code filter or "all" for all positions
	portfolio := h.Portfolio.GetPortfolio(h.getQuotesForAllPositions())

	var summary []string
	summary = append(summary, fmt.Sprintf("总资金：%.2f 元", portfolio.TotalValue))
	summary = append(summary, fmt.Sprintf("可用现金：%.2f 元", portfolio.Cash))

	if len(portfolio.Positions) > 0 {
		summary = append(summary, "\n持仓:")
		for _, pos := range portfolio.Positions {
			summary = append(summary, fmt.Sprintf("- %s (%s): %.2f 元，%.1f%%", pos.Name, pos.Code, pos.PnL, pos.PnLPct))
		}

		summary = append(summary, fmt.Sprintf("\n总盈亏：%.2f 元 (%.1f%%)", portfolio.PnL, portfolio.PnLPct))
	}

	return strings.Join(summary, "\n")
}
