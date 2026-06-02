package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// DeepSeekClient is a minimal client for the DeepSeek Chat API.
type DeepSeekClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewDeepSeekClient creates a DeepSeekClient from environment variables.
// Env: DEEPSEEK_API_KEY, DEEPSEEK_BASE_URL (default https://api.deepseek.com),
// DEEPSEEK_MODEL (default deepseek-chat).
func NewDeepSeekClient() *DeepSeekClient {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// IsEnabled returns true when a non-empty API key is configured.
func (c *DeepSeekClient) IsEnabled() bool {
	return c.apiKey != ""
}

// Analyze sends a prompt to DeepSeek and returns the raw response text.
// It returns an error when the API key is missing, the request fails, or the
// response cannot be parsed.
func (c *DeepSeekClient) Analyze(ctx context.Context, prompt string) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("deepseek client: API key not configured")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepseek request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek responded with status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("deepseek returned no choices")
	}
	return result.Choices[0].Message.Content, nil
}

// BankTxnAnalyzeInput carries the transaction fields passed to the AI prompt.
type BankTxnAnalyzeInput struct {
	Date           string
	Direction      string // "in" or "out"
	Amount         decimal.Decimal
	Description    string
	Counterparty   string
	Classification string // rule-based fallback classification
}

// bankTxnAnalyzePrompt is the system + user prompt sent to DeepSeek for a
// single bank transaction.
func bankTxnAnalyzePrompt(txn BankTxnAnalyzeInput) string {
	return `你是一个专业的财务助手。请分析以下银行流水的特征，判断其业务场景和处理方式。

银行流水信息：
- 日期：` + txn.Date + `
- 借贷方向：` + txn.Direction + `
- 金额：` + txn.Amount.String() + `
- 摘要：` + txn.Description + `
- 对方账户名：` + txn.Counterparty + `
- 规则引擎分类（参考）：` + txn.Classification + `

请以JSON格式返回分析结果：
{
  "business_scene": "业务场景描述",
  "suggested_action": "auto_voucher|generate_payment|manual_pending",
  "confidence": 置信度(0-100),
  "reasoning": "推理过程"
}

判断规则：
- A类（auto_voucher）：银行手续费、利息收入、税务扣款、社保扣款、保险费、内部转账
- B类（generate_payment）：货款收款、货款付款
- C类（manual_pending）：无法确定的款项

只返回JSON，不要包含其他文字。`
}

// DeepSeekAnalyzeResponse is the struct decoded from DeepSeek's JSON reply.
type DeepSeekAnalyzeResponse struct {
	BusinessScene   string `json:"business_scene"`
	SuggestedAction string `json:"suggested_action"`
	Confidence      int    `json:"confidence"`
	Reasoning       string `json:"reasoning"`
}

// ParseBankTxnResponse parses a JSON response body from DeepSeek into a
// DeepSeekAnalyzeResponse. It returns an error if the body is not valid JSON.
func ParseBankTxnResponse(raw string) (*DeepSeekAnalyzeResponse, error) {
	raw = trimMarkdown(raw)
	var r DeepSeekAnalyzeResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("parse bank txn response: %w", err)
	}
	return &r, nil
}

// trimMarkdown removes optional triple-backtick wrappers.
func trimMarkdown(s string) string {
	s = trimPrefix(s, "```json")
	s = trimPrefix(s, "```")
	return trimPrefix(s, "\n")
}

func trimPrefix(s, prefix string) string {
	for len(s) > 0 && (s[:1] == "\n" || s[:1] == " ") {
		s = s[1:]
	}
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		s = s[len(prefix):]
	}
	for len(s) > 0 && (s[len(s)-1:] == "\n" || s[len(s)-1:] == " ") {
		s = s[:len(s)-1]
	}
	if len(s) >= 2 && s[len(s)-2:] == "``" {
		s = s[:len(s)-2]
	}
	return s
}

// AIAnalysisResult is returned by BankTxnAIService.AnalyzeBankTxn.
type AIAnalysisResult struct {
	BusinessScene   string `json:"business_scene"`
	SuggestedAction string `json:"suggested_action"`
	Confidence      int    `json:"confidence"`
	Reasoning       string `json:"reasoning"`
}

// AIFieldUpdater is the subset of BankTransactionRepository needed by the AI service.
type AIFieldUpdater interface {
	UpdateAIFields(ctx context.Context, id uuid.UUID, scene, action string, confidence int) error
}

// BankTxnAIService analyses bank transactions using the DeepSeek API.
type BankTxnAIService struct {
	deepseek *DeepSeekClient
	updater  AIFieldUpdater
}

// NewBankTxnAIService creates a BankTxnAIService.
func NewBankTxnAIService(deepseek *DeepSeekClient, updater AIFieldUpdater) *BankTxnAIService {
	return &BankTxnAIService{
		deepseek: deepseek,
		updater:  updater,
	}
}

// AnalyzeBankTxn analyses a single bank transaction and returns the AI result.
// The transaction record is updated with the AI fields on success.
func (s *BankTxnAIService) AnalyzeBankTxn(ctx context.Context, txn model.BankTransaction) (*AIAnalysisResult, error) {
	direction := ""
	if txn.Direction != nil {
		direction = *txn.Direction
	}
	desc := ""
	if txn.Description != nil {
		desc = *txn.Description
	}
	counterparty := ""
	if txn.CounterpartyName != nil {
		counterparty = *txn.CounterpartyName
	}
	classification := ""
	if txn.Classification != nil {
		classification = *txn.Classification
	}

	amount := txn.Debit
	if txn.Credit.GreaterThan(amount) {
		amount = txn.Credit
	}

	prompt := bankTxnAnalyzePrompt(BankTxnAnalyzeInput{
		Date:           txn.TxnDate.Format("2006-01-02"),
		Direction:      direction,
		Amount:         amount,
		Description:    desc,
		Counterparty:   counterparty,
		Classification: classification,
	})

	raw, err := s.deepseek.Analyze(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("deepseek analyze: %w", err)
	}

	result, err := ParseBankTxnResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Persist AI fields back to the transaction row.
	if s.updater != nil {
		_ = s.updater.UpdateAIFields(ctx, txn.ID, result.BusinessScene, result.SuggestedAction, result.Confidence)
	}

	return &AIAnalysisResult{
		BusinessScene:   result.BusinessScene,
		SuggestedAction: result.SuggestedAction,
		Confidence:      result.Confidence,
		Reasoning:       result.Reasoning,
	}, nil
}

// BatchAnalyzeBankTxns analyses multiple transactions sequentially.
// It stops on the first error and returns all results collected before that point.
func (s *BankTxnAIService) BatchAnalyzeBankTxns(ctx context.Context, txns []model.BankTransaction) ([]AIAnalysisResult, error) {
	results := make([]AIAnalysisResult, 0, len(txns))
	for i := range txns {
		r, err := s.AnalyzeBankTxn(ctx, txns[i])
		if err != nil {
			return results, fmt.Errorf("batch analyze txn %d: %w", i, err)
		}
		results = append(results, *r)
	}
	return results, nil
}