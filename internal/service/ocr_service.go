package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"huihua/finance/internal/model"
)

// OcrService provides invoice OCR recognition
// Supports multiple OCR providers: Baidu, Tencent, Alibaba
type OcrService struct {
	provider string // "baidu", "tencent", "alibaba", "mock"
}

func NewOcrService(provider string) *OcrService {
	if provider == "" {
		provider = "mock"
	}
	return &OcrService{provider: provider}
}

// RecognizeInvoice performs OCR on an invoice image and returns structured data
func (s *OcrService) RecognizeInvoice(ctx context.Context, fileURL string) (*model.OcrInvoiceResponse, error) {
	switch s.provider {
	case "baidu":
		return s.recognizeWithBaidu(ctx, fileURL)
	case "tencent":
		return s.recognizeWithTencent(ctx, fileURL)
	case "alibaba":
		return s.recognizeWithAlibaba(ctx, fileURL)
	default:
		return s.mockRecognize(fileURL)
	}
}

// mockRecognize returns mock OCR data for testing
func (s *OcrService) mockRecognize(fileURL string) (*model.OcrInvoiceResponse, error) {
	rawData, _ := json.Marshal(map[string]interface{}{
		"source":     "mock",
		"file_url":   fileURL,
		"timestamp":  time.Now().Format(time.RFC3339),
		"confidence": 0.95,
	})

	return &model.OcrInvoiceResponse{
		InvoiceNo:   fmt.Sprintf("INV%d", time.Now().Unix()%100000),
		InvoiceCode: "1100231410",
		InvoiceDate: time.Now().Format("2006-01-02"),
		TotalAmount: 1000.00,
		TaxAmount:   100.00,
		VendorName:  "测试供应商",
		InvoiceKind: "electronic_normal",
		RawData:     string(rawData),
	}, nil
}

// recognizeWithBaidu uses Baidu OCR API
func (s *OcrService) recognizeWithBaidu(ctx context.Context, fileURL string) (*model.OcrInvoiceResponse, error) {
	// TODO: Implement Baidu OCR integration
	// Baidu OCR API: https://cloud.baidu.com/doc/OCRAPI/9d7ej3j3r
	// 1. Download image if URL
	// 2. Call Baidu OCR API with AK/SK for token
	// 3. Parse response and map to OcrInvoiceResponse
	return s.mockRecognize(fileURL)
}

// recognizeWithTencent uses Tencent OCR API
func (s *OcrService) recognizeWithTencent(ctx context.Context, fileURL string) (*model.OcrInvoiceResponse, error) {
	// TODO: Implement Tencent OCR integration
	// Tencent OCR API: https://cloud.tencent.com/document/product/866/49525
	return s.mockRecognize(fileURL)
}

// recognizeWithAlibaba uses Alibaba OCR API
func (s *OcrService) recognizeWithAlibaba(ctx context.Context, fileURL string) (*model.OcrInvoiceResponse, error) {
	// TODO: Implement Alibaba OCR integration
	// Alibaba OCR API: https://help.aliyun.com/document_detail/300066.html
	return s.mockRecognize(fileURL)
}
