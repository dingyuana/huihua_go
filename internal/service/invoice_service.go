package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// InvoiceService handles invoice operations.
type InvoiceService struct {
	repo *repository.InvoiceRepository
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(repo *repository.InvoiceRepository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

// Create creates a new invoice.
func (s *InvoiceService) Create(ctx context.Context, tenantID uuid.UUID, req *model.SalesInvoice) (*model.SalesInvoice, error) {
	// Validate invoice data
	if err := s.ValidateInvoice(req); err != nil {
		return nil, err
	}

	// Check for duplicate invoice number
	exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, req.InvoiceNo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("invoice number already exists")
	}

	req.Status = string(model.InvoiceStatusDraft)
	return s.repo.Create(ctx, tenantID, req)
}

// List returns invoices for a tenant with optional filters.
func (s *InvoiceService) List(ctx context.Context, tenantID uuid.UUID, filters model.InvoiceFilter) ([]model.SalesInvoice, error) {
	return s.repo.ListByTenant(ctx, tenantID, filters)
}

// GetByID retrieves an invoice by ID.
func (s *InvoiceService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.SalesInvoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// UpdateStatus updates the status of an invoice.
func (s *InvoiceService) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	// Validate status transition
	validStatuses := map[string]bool{
		"draft": true, "submitted": true, "verified": true, "invalid": true,
		"unpaid": true, "partially_paid": true, "paid": true,
	}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invoice not found")
	}

	return s.repo.UpdateStatus(ctx, tenantID, id, status)
}

// ValidateInvoice validates invoice data.
func (s *InvoiceService) ValidateInvoice(inv *model.SalesInvoice) error {
	if inv.InvoiceNo == "" {
		return errors.New("invoice_no is required")
	}
	if inv.CustomerID == uuid.Nil {
		return errors.New("customer_id is required")
	}
	if inv.CompanyID == uuid.Nil {
		return errors.New("company_id is required")
	}
	if inv.TotalAmount.LessThan(decimal.Zero) {
		return errors.New("total_amount cannot be negative")
	}
	if inv.NetAmount.LessThan(decimal.Zero) {
		return errors.New("net_amount cannot be negative")
	}
	if inv.TaxAmount.LessThan(decimal.Zero) {
		return errors.New("tax_amount cannot be negative")
	}
	return nil
}

// ValidateLineItems validates that line item amounts are consistent.
func (s *InvoiceService) ValidateLineItems(items []model.InvoiceLineItem) error {
	for _, item := range items {
		// Calculate expected net amount: quantity * unit_price
		expectedNet := item.Quantity.Mul(item.UnitPrice)

		// Verify net amount matches
		if !expectedNet.Equal(item.NetAmount) {
			return fmt.Errorf("line item net amount mismatch: expected %s, got %s",
				expectedNet.String(), item.NetAmount.String())
		}

		// Calculate expected tax amount: net_amount * tax_rate
		taxRate := item.TaxRate.Div(decimal.NewFromInt(100))
		expectedTax := item.NetAmount.Mul(taxRate)

		// Verify tax amount (allow small floating point differences)
		diff := expectedTax.Sub(item.TaxAmount).Abs()
		if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
			return fmt.Errorf("line item tax amount mismatch: expected %s, got %s",
				expectedTax.String(), item.TaxAmount.String())
		}

		// Calculate total amount: net_amount + tax_amount
		expectedTotal := item.NetAmount.Add(item.TaxAmount)
		if !expectedTotal.Equal(item.TotalAmount) {
			return fmt.Errorf("line item total amount mismatch: expected %s, got %s",
				expectedTotal.String(), item.TotalAmount.String())
		}
	}
	return nil
}

// ImportFromExcel imports invoices from Excel data.
func (s *InvoiceService) ImportFromExcel(ctx context.Context, tenantID uuid.UUID, req *model.InvoiceImportRequest) ([]model.SalesInvoice, error) {
	if req == nil || len(req.Invoices) == 0 {
		return nil, errors.New("no invoices to import")
	}

	var invoices []model.SalesInvoice
	for _, item := range req.Invoices {
		// Parse posting date
		postingDate, err := time.Parse("2006-01-02", item.PostingDate)
		if err != nil {
			return nil, fmt.Errorf("invalid posting_date format for invoice %s: %v", item.InvoiceNo, err)
		}

		// Parse optional due date
		var dueDate *time.Time
		if item.DueDate != "" {
			t, err := time.Parse("2006-01-02", item.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date format for invoice %s: %v", item.InvoiceNo, err)
			}
			dueDate = &t
		}

		// Parse customer_id
		customerID, err := uuid.Parse(item.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("invalid customer_id for invoice %s: %v", item.InvoiceNo, err)
		}

		// Create invoice model
		inv := &model.SalesInvoice{
			InvoiceNo:         item.InvoiceNo,
			InvoiceType:       item.InvoiceType,
			CustomerID:        customerID,
			PostingDate:       postingDate,
			DueDate:           dueDate,
			TotalAmount:       decimal.NewFromFloat(item.TotalAmount),
			TaxAmount:         decimal.NewFromFloat(item.TaxAmount),
			NetAmount:         decimal.NewFromFloat(item.NetAmount),
			OutstandingAmount: decimal.NewFromFloat(item.TotalAmount),
			Status:            item.Status,
		}

		// Validate
		if err := s.ValidateInvoice(inv); err != nil {
			return nil, fmt.Errorf("validation failed for invoice %s: %v", item.InvoiceNo, err)
		}

		// Check for duplicate
		exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, item.InvoiceNo)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("invoice number %s already exists", item.InvoiceNo)
		}

		invoices = append(invoices, *inv)
	}

	// Batch import
	return s.repo.ImportBatch(ctx, tenantID, invoices)
}

// ImportFromExcelFile parses an Excel/CSV file and imports invoice rows.
func (s *InvoiceService) ImportFromExcelFile(ctx context.Context, tenantID uuid.UUID, data []byte) (*model.InvoiceFileImportResult, error) {
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

	// Find header row
	headerIdx := 0
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty++
			}
		}
		if nonEmpty >= 3 {
			headerIdx = i
			break
		}
	}

	headerMap := make(map[string]int)
	for i, col := range rows[headerIdx] {
		key := strings.ToLower(strings.TrimSpace(col))
		headerMap[key] = i
	}

	findCol := func(names ...string) (int, bool) {
		for _, name := range names {
			key := strings.ToLower(strings.TrimSpace(name))
			if idx, ok := headerMap[key]; ok {
				return idx, true
			}
		}
		// substring fallback
		for ci, col := range rows[headerIdx] {
			colLower := strings.ToLower(col)
			for _, name := range names {
				if strings.Contains(colLower, strings.ToLower(name)) {
					return ci, true
				}
			}
		}
		return 0, false
	}

	invNoIdx, _ := findCol("发票号码", "发票号", "invoice_no", "invoice number", "invoice", "发票代码")
	invTypeIdx, _ := findCol("发票类型", "类型", "invoice_type", "type")
	dateIdx, dateFound := findCol("开票日期", "日期", "posting_date", "date", "发票日期")
	netIdx, _ := findCol("不含税金额", "金额", "net_amount", "net amount", "net", "净额")
	taxIdx, _ := findCol("税额", "tax_amount", "tax amount", "tax", "税金")
	totalIdx, _ := findCol("价税合计", "合计", "total_amount", "total amount", "total", "总金额")

	if !dateFound {
		return nil, errors.New("未找到日期列，请确保Excel包含'开票日期'或'日期'列")
	}

	if invNoIdx == 0 && totalIdx == 0 {
		return nil, errors.New("未找到发票号或金额列，请确保Excel包含'发票号码'和'价税合计'列")
	}

	var imported []model.SalesInvoice
	var failedRows []model.FailedRowDetail

	for rowIdx, row := range rows[headerIdx+1:] {
		rowNum := rowIdx + headerIdx + 2
		if len(row) == 0 {
			continue
		}
		allEmpty := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		invNo := ""
		if invNoIdx < len(row) {
			invNo = strings.TrimSpace(row[invNoIdx])
		}
		if invNo == "" {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Reason: "发票号为空"})
			continue
		}

		postingDate := time.Now()
		if dateIdx < len(row) {
			dateStr := strings.TrimSpace(row[dateIdx])
			if t, err := parseDateInvoice(dateStr); err == nil {
				postingDate = t
			} else {
				failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: dateStr, Reason: "日期格式无法解析: " + dateStr})
				continue
			}
		}

		parseFloat := func(idx int) float64 {
			if idx < len(row) {
				v, err := strconv.ParseFloat(strings.ReplaceAll(row[idx], ",", ""), 64)
				if err == nil {
					return v
				}
			}
			return 0
		}

		netAmount := parseFloat(netIdx)
		taxAmount := parseFloat(taxIdx)
		totalAmount := parseFloat(totalIdx)
		if totalAmount == 0 && netAmount > 0 {
			totalAmount = netAmount + taxAmount
		}
		if totalAmount == 0 {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: invNo, Reason: "金额为空"})
			continue
		}

		invType := "purchase"
		if invTypeIdx < len(row) {
			t := strings.ToLower(strings.TrimSpace(row[invTypeIdx]))
			if strings.Contains(t, "销项") || strings.Contains(t, "销售") || t == "sale" {
				invType = "sale"
			} else if strings.Contains(t, "红字") || strings.Contains(t, "红冲") || t == "credit_note" {
				invType = "credit_note"
			}
		}

		inv := &model.SalesInvoice{
			InvoiceNo:         invNo,
			InvoiceType:       invType,
			PostingDate:       postingDate,
			TotalAmount:       decimal.NewFromFloat(totalAmount),
			TaxAmount:         decimal.NewFromFloat(taxAmount),
			NetAmount:         decimal.NewFromFloat(netAmount),
			OutstandingAmount: decimal.NewFromFloat(totalAmount),
			Status:            "unpaid",
		}

		// lightweight validation for file import — skip customer_id/company_id UUID checks since Excel has names, not UUIDs
		if inv.InvoiceNo == "" {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Reason: "发票号为空"})
			continue
		}
		if inv.TotalAmount.LessThan(decimal.Zero) {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: invNo, Reason: "金额不能为负数"})
			continue
		}

		exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, invNo)
		if err != nil {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: invNo, Reason: "查重错误: " + err.Error()})
			continue
		}
		if exists {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: invNo, Reason: "发票号已存在"})
			continue
		}

		imported = append(imported, *inv)
	}

	importedResult, err := s.repo.ImportBatch(ctx, tenantID, imported)
	if err != nil {
		return nil, fmt.Errorf("import batch: %w", err)
	}
	_ = importedResult // batch slice returned

	return &model.InvoiceFileImportResult{
		TotalRows:   len(rows) - headerIdx - 1,
		Imported:    len(imported),
		Failed:      len(failedRows),
		FailedRows:  failedRows,
	}, nil
}

// parseDateInvoice tries common date formats for invoice dates.
func parseDateInvoice(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02", "2006/01/02", "20060102",
		"2006-01-02 15:04:05", "2006/01/02 15:04:05",
		"02/01/2006", "01/02/2006",
		"2006年01月02日", "2006年1月2日",
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}

// ParseInvoicePDF is a placeholder for PDF OCR parsing.
func (s *InvoiceService) ParseInvoicePDF(ctx context.Context, tenantID uuid.UUID, fileURL string) (*model.SalesInvoice, error) {
	return nil, errors.New("OCR not implemented, use manual import")
}

// ParseInvoiceImage is a placeholder for image OCR parsing.
func (s *InvoiceService) ParseInvoiceImage(ctx context.Context, tenantID uuid.UUID, fileURL string) (*model.SalesInvoice, error) {
	return nil, errors.New("OCR not implemented, use manual import")
}

// MatchToBankTxn matches an invoice to a bank transaction.
func (s *InvoiceService) MatchToBankTxn(ctx context.Context, tenantID, invoiceID, bankTxnID uuid.UUID, amount decimal.Decimal) error {
	// Get the invoice to verify it exists
	inv, err := s.repo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %v", err)
	}
	if inv == nil {
		return errors.New("invoice not found")
	}

	// Check that amount doesn't exceed outstanding
	if amount.GreaterThan(inv.OutstandingAmount) {
		return errors.New("allocation amount exceeds outstanding invoice amount")
	}

	// Create the payment allocation
	return s.repo.MatchToBankTxn(ctx, tenantID, invoiceID, bankTxnID, amount.String())
}

// GetLineItems retrieves line items for an invoice.
func (s *InvoiceService) GetLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]model.InvoiceLineItem, error) {
	return s.repo.GetLineItems(ctx, tenantID, invoiceID)
}

// ListInvoicesForMatching returns invoices eligible for bank matching.
func (s *InvoiceService) ListInvoicesForMatching(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]model.SalesInvoice, error) {
	return s.repo.ListInvoicesForMatching(ctx, tenantID, customerID)
}