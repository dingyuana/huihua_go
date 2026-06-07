package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

type ExpenseInvoiceImportService struct {
	repo       *repository.ExpenseInvoiceRepository
	batchCache map[string]*model.ImportBatch
}

func NewExpenseInvoiceImportService(repo *repository.ExpenseInvoiceRepository) *ExpenseInvoiceImportService {
	return &ExpenseInvoiceImportService{repo: repo, batchCache: make(map[string]*model.ImportBatch)}
}

// Upload parses an Excel file and returns a batch preview with validation results
func (s *ExpenseInvoiceImportService) Upload(ctx context.Context, tenantID uuid.UUID, file io.Reader, fileName string) (*model.ImportUploadResponse, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open excel: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}
	if len(rows) < 2 {
		return nil, errors.New("empty file: no data rows found")
	}

	headerIdx := findExpenseExcelHeader(rows)
	cols := newExpenseColumnIndex(rows[headerIdx])

	if !cols.has("开票日期", "日期", "invoice_date", "date") {
		return nil, errors.New("未找到日期列，请确保Excel包含'开票日期'或'日期'列")
	}
	if !cols.has("发票号码", "发票号", "invoice_no", "invoice number") {
		return nil, errors.New("未找到发票号码列")
	}
	if !cols.has("价税合计", "合计", "total_amount", "total amount", "total", "总金额") {
		return nil, errors.New("未找到价税合计列")
	}

	batchID := uuid.New().String()
	now := time.Now()
	batch := &model.ImportBatch{
		BatchID:   batchID,
		Status:    "pending",
		Rows:      []model.ImportRow{},
		CreatedAt: now,
	}

	var validCount, errorCount int
	seenInvoiceNos := make(map[string]bool)

	invNoIdx, _ := cols.get("发票号码", "发票号", "invoice_no", "invoice number")
	invCodeIdx, _ := cols.get("发票代码", "invoice_code")
	dateIdx, _ := cols.get("开票日期", "日期", "invoice_date", "date")
	totalIdx, _ := cols.get("价税合计", "合计", "total_amount", "total amount", "total", "总金额")
	taxIdx, _ := cols.get("税额", "tax_amount")
	vendorIdx, _ := cols.get("供应商名称", "vendor_name", "供应商")

	for i := headerIdx + 1; i < len(rows); i++ {
		row := rows[i]
		rowIndex := i + 1

		getCell := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		invoiceNo := getCell(invNoIdx)
		invoiceCode := getCell(invCodeIdx)
		invoiceDateStr := getCell(dateIdx)
		totalAmountStr := getCell(totalIdx)
		taxAmountStr := getCell(taxIdx)
		vendorName := getCell(vendorIdx)

		importRow := model.ImportRow{
			RowIndex:    rowIndex,
			InvoiceNo:   invoiceNo,
			InvoiceCode: invoiceCode,
			InvoiceDate: invoiceDateStr,
			VendorName:  vendorName,
			Status:      "valid",
		}

		// Required field checks
		if invoiceNo == "" {
			importRow.Status = "error"
			importRow.ErrorMsg = "发票号码为必填字段"
		} else if invoiceDateStr == "" {
			importRow.Status = "error"
			importRow.ErrorMsg = "开票日期为必填字段"
		} else if _, err := time.Parse("2006-01-02", invoiceDateStr); err != nil {
			importRow.Status = "error"
			importRow.ErrorMsg = fmt.Sprintf("开票日期格式错误，应为YYYY-MM-DD，实际：%s", invoiceDateStr)
		} else if totalAmountStr == "" {
			importRow.Status = "error"
			importRow.ErrorMsg = "价税合计为必填字段"
		} else {
			totalAmount, err := strconv.ParseFloat(totalAmountStr, 64)
			if err != nil {
				importRow.Status = "error"
				importRow.ErrorMsg = fmt.Sprintf("价税合计金额格式错误：%s", totalAmountStr)
			} else if taxAmountStr == "" {
				importRow.Status = "error"
				importRow.ErrorMsg = "税额为必填字段"
			} else {
				taxAmount, err := strconv.ParseFloat(taxAmountStr, 64)
				if err != nil {
					importRow.Status = "error"
					importRow.ErrorMsg = fmt.Sprintf("税额格式错误：%s", taxAmountStr)
				} else {
					// All validations passed, check for duplicates
					importRow.TotalAmount = totalAmount
					importRow.TaxAmount = taxAmount

					if seenInvoiceNos[invoiceNo] {
						importRow.Status = "duplicate"
						importRow.ErrorMsg = "发票号码在本批次中重复"
					} else {
						seenInvoiceNos[invoiceNo] = true
						// Check against database
						existing, dbErr := s.repo.GetByInvoiceNo(ctx, tenantID, invoiceNo)
						if dbErr != nil {
							importRow.Status = "error"
							importRow.ErrorMsg = fmt.Sprintf("查询已存在发票失败：%v", dbErr)
						} else if existing != nil {
							importRow.Status = "duplicate"
							importRow.ErrorMsg = "发票号码已存在"
						}
					}
				}
			}
		}

		if importRow.Status == "valid" {
			validCount++
		} else {
			errorCount++
		}
		batch.Rows = append(batch.Rows, importRow)
	}

	batch.TotalRows = len(batch.Rows)
	batch.ValidRows = validCount
	batch.ErrorRows = errorCount
	s.batchCache[batchID] = batch

	return &model.ImportUploadResponse{
		BatchID:    batchID,
		TotalRows:  batch.TotalRows,
		ValidRows:  validCount,
		ErrorRows:  errorCount,
		ValidCount: validCount,
		ErrorCount: errorCount,
		Details:    batch.Rows,
	}, nil
}

// Preview returns the batch preview for a given batch ID
func (s *ExpenseInvoiceImportService) Preview(ctx context.Context, batchID string) (*model.ImportBatch, error) {
	batch, ok := s.batchCache[batchID]
	if !ok {
		return nil, errors.New("batch not found or expired")
	}
	return batch, nil
}

// Confirm imports the selected rows from a batch
func (s *ExpenseInvoiceImportService) Confirm(ctx context.Context, tenantID uuid.UUID, req *model.ImportConfirmRequest) (*model.ImportConfirmResponse, error) {
	batch, ok := s.batchCache[req.BatchID]
	if !ok {
		return nil, errors.New("batch not found or expired")
	}

	if batch.Status == "confirmed" {
		return nil, errors.New("batch already confirmed")
	}

	selectedSet := make(map[int]bool)
	for _, idStr := range req.SelectedIDs {
		idx, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		selectedSet[idx] = true
	}

	var importedCount, failedCount int

	for i, row := range batch.Rows {
		if row.Status != "valid" {
			continue
		}
		if !selectedSet[i] {
			continue
		}

		invoiceDate, _ := time.Parse("2006-01-02", row.InvoiceDate)
		inv := &model.ExpenseInvoice{
			ID:              uuid.New(),
			TenantID:        tenantID,
			InvoiceNo:       row.InvoiceNo,
			InvoiceCode:     &row.InvoiceCode,
			InvoiceDate:     invoiceDate,
			InvoiceKind:     "electronic_normal",
			TaxAmount:       decimal.NewFromFloat(row.TaxAmount),
			TotalAmount:     decimal.NewFromFloat(row.TotalAmount),
			VendorName:      &row.VendorName,
			VerifyStatus:    model.ExpenseVerifyStatusUnverified,
			DeductionStatus: model.ExpenseDeductionStatusUndeducted,
			Status:          "pending",
			CreatedAt:       time.Now(),
		}

		if err := s.repo.Create(ctx, tenantID, inv); err != nil {
			failedCount++
			continue
		}
		importedCount++
	}

	batch.Status = "confirmed"
	return &model.ImportConfirmResponse{
		Imported: importedCount,
		Failed:   failedCount,
	}, nil
}

// Helper functions

func findExpenseExcelHeader(rows [][]string) int {
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if cell != "" {
				nonEmpty++
			}
		}
		if nonEmpty >= 3 {
			return i
		}
	}
	return 0
}

type expenseColumnIndex struct {
	idx   map[string]int
	names []string
}

func newExpenseColumnIndex(header []string) *expenseColumnIndex {
	ci := &expenseColumnIndex{idx: make(map[string]int)}
	for i, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		ci.idx[key] = i
		ci.names = append(ci.names, col)
	}
	return ci
}

func (ci *expenseColumnIndex) get(names ...string) (int, bool) {
	for _, name := range names {
		key := strings.ToLower(strings.TrimSpace(name))
		if i, ok := ci.idx[key]; ok {
			return i, true
		}
	}
	return -1, false
}

func (ci *expenseColumnIndex) has(names ...string) bool {
	_, ok := ci.get(names...)
	return ok
}
