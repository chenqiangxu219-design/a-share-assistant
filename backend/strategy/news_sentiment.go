package strategy

import (
	"encoding/json"
	"fmt"
	"strings"

	"a-share-assistant/backend/chat"
	"a-share-assistant/backend/datasource"
)

// SentimentType represents the sentiment direction
type SentimentType string

const (
	SentimentPositive SentimentType = "positive"
	SentimentNegative SentimentType = "negative"
	SentimentNeutral  SentimentType = "neutral"
)

// SentimentResult represents the analysis result for a single news item
type SentimentResult struct {
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	SentimentSentiment SentimentType `json:"sentiment"`
	Score    float64    `json:"score"` // -1.0 to 1.0
	Reason   string     `json:"reason"`
	Impact   string     `json:"impact"` // "high", "medium", "low"
}

// SentimentSummary is the aggregated result for a stock
type SentimentSummary struct {
	Code           string            `json:"code"`
	Name           string            `json:"name"`
	OverallScore   float64           `json:"overall_score"`
	PositiveCount  int               `json:"positive_count"`
	NegativeCount  int               `json:"negative_count"`
	NeutralCount   int               `json:"neutral_count"`
	Results        []SentimentResult `json:"results"`
	Conclusion     string            `json:"conclusion"`
}

// AnalyzeSentiment analyzes sentiment for a batch of news articles
func AnalyzeSentiment(articles []datasource.NewsArticle) ([]SentimentResult, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles to analyze")
	}

	// Build news text for LLM
	var sb strings.Builder
	sb.WriteString("请分析以下 A 股相关新闻的情感倾向，以 JSON 数组格式返回结果。\n\n")
	for i, a := range articles {
		sb.WriteString(fmt.Sprintf("新闻 %d: %s\n", i+1, a.Title))
		if a.Content != "" {
			sb.WriteString(fmt.Sprintf("内容: %s\n", a.Content))
		}
		sb.WriteString(fmt.Sprintf("来源: %s, 时间: %s\n\n", a.Source, a.Time))
	}

	sb.WriteString("返回格式: [{\"title\": \"标题\", \"sentiment\": \"positive/negative/neutral\", \"score\": 0.75, \"reason\": \"原因\", \"impact\": \"high/medium/low\"}]\n")

	messages := []chat.Message{
		{Role: "system", Content: "你是 A 股新闻情感分析助手。请分析每条新闻对股价的影响方向。"},
		{Role: "user", Content: sb.String()},
	}

	response, err := chat.ChatOnce(messages, "")
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// Parse JSON from LLM response
	var results []SentimentResult
	if err := json.Unmarshal([]byte(response), &results); err != nil {
		// Try to extract JSON from the response
		results = parseSentimentFromText(response)
	}

	return results, nil
}

// parseSentimentFromText extracts JSON from LLM response text
func parseSentimentFromText(text string) []SentimentResult {
	// Find JSON array in the text
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start >= 0 && end > start {
		jsonStr := text[start : end+1]
		var results []SentimentResult
		if err := json.Unmarshal([]byte(jsonStr), &results); err == nil {
			return results
		}
	}
	return nil
}

// AnalyzeSentimentWithSummary analyzes news and produces a summary
func AnalyzeSentimentWithSummary(articles []datasource.NewsArticle, stockName string) (*SentimentSummary, error) {
	results, err := AnalyzeSentiment(articles)
	if err != nil {
		return nil, err
	}

	// Calculate summary
	overallScore := 0.0
	positiveCount := 0
	negativeCount := 0
	neutralCount := 0

	for _, r := range results {
		overallScore += r.Score
		switch r.SentimentSentiment {
		case SentimentPositive:
			positiveCount++
		case SentimentNegative:
			negativeCount++
		default:
			neutralCount++
		}
	}

	if len(results) > 0 {
		overallScore /= float64(len(results))
	}

	// Build conclusion
	var conclusion string
	if overallScore > 0.3 {
		conclusion = fmt.Sprintf("整体偏正面，%d 条正面新闻，%d 条负面新闻，%d 条中性新闻", positiveCount, negativeCount, neutralCount)
	} else if overallScore < -0.3 {
		conclusion = fmt.Sprintf("整体偏负面，%d 条正面新闻，%d 条负面新闻，%d 条中性新闻", positiveCount, negativeCount, neutralCount)
	} else {
		conclusion = fmt.Sprintf("整体中性，%d 条正面新闻，%d 条负面新闻，%d 条中性新闻", positiveCount, negativeCount, neutralCount)
	}

	return &SentimentSummary{
		Code:           articles[0].Code,
		Name:           stockName,
		OverallScore:   overallScore,
		PositiveCount:  positiveCount,
		NegativeCount:  negativeCount,
		NeutralCount:   neutralCount,
		Results:        results,
		Conclusion:     conclusion,
	}, nil
}
