package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
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
	repo            *repository.InvoiceRepository
	partyRepo       *repository.PartyRepository
	arInvoiceRepo   *repository.ArInvoiceRepository
	voucherAutoSvc  *VoucherAutoGenerateService
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(
	repo *repository.InvoiceRepository,
	partyRepo *repository.PartyRepository,
	arInvoiceRepo *repository.ArInvoiceRepository,
	voucherAutoSvc *VoucherAutoGenerateService,
) *InvoiceService {
	return &InvoiceService{
		repo:           repo,
		partyRepo:      partyRepo,
		arInvoiceRepo:  arInvoiceRepo,
		voucherAutoSvc: voucherAutoSvc,
	}
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

// GetByInvoiceNo retrieves an invoice by its invoice number (used for red→blue linkage).
func (s *InvoiceService) GetByInvoiceNo(ctx context.Context, tenantID uuid.UUID, invoiceNo string) (*model.SalesInvoice, error) {
	return s.repo.GetByInvoiceNo(ctx, tenantID, invoiceNo)
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

func (s *InvoiceService) Update(ctx context.Context, tenantID, id uuid.UUID, fields map[string]interface{}) error {
	inv, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invoice not found")
	}
	return s.repo.UpdateFields(ctx, tenantID, id, fields)
}

func (s *InvoiceService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
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
	// Red-letter (credit_note) invoices carry negative amounts to indicate reversal.
	// Only enforce non-negative for normal sales/purchase invoices.
	if inv.InvoiceType != "credit_note" {
		if inv.TotalAmount.LessThan(decimal.Zero) {
			return errors.New("total_amount cannot be negative")
		}
		if inv.NetAmount.LessThan(decimal.Zero) {
			return errors.New("net_amount cannot be negative")
		}
		if inv.TaxAmount.LessThan(decimal.Zero) {
			return errors.New("tax_amount cannot be negative")
		}
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
			IsReturn:          item.IsReturn,
		}

		// Apply remark if provided
		if item.Remark != "" {
			r := item.Remark
			inv.Remark = &r
		}

		// If this is a red-letter invoice, link it to the original blue invoice.
		// SourceRedInvoiceNo carries the original invoice_no from import.
		if item.IsReturn && item.SourceRedInvoiceNo != "" {
			src := item.SourceRedInvoiceNo
			inv.SourceRedInvoiceNo = &src
			if srcInv, lerr := s.repo.GetByInvoiceNo(ctx, tenantID, src); lerr == nil && srcInv != nil {
				srcID := srcInv.ID
				inv.ReturnAgainst = &srcID
			}
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
type parsedInvoiceRow struct {
	RowNum              int
	InvoiceNo           string
	PostingDate         time.Time
	NetAmount           float64
	TaxAmount           float64
	TotalAmount         float64
	CustomerID          uuid.UUID
	BuyerName           string
	InvoiceType         string
	BuyerTaxID          string
	InvoiceCode         string
	InvoiceCategory     string
	Status              string
	Remark              string
	IsReturn            bool
	SourceRedInvoiceNo  string
	ItemDescription string
	ItemCode        string
	Unit            string
	Quantity        float64
	UnitPrice       float64
	TaxRate         float64
}

type invoiceGroup struct {
	Rows   []parsedInvoiceRow
	Header parsedInvoiceRow
}

type headerWithItems struct {
	Header model.SalesInvoice
	Items  []model.InvoiceLineItem
}

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

	headerIdx := findExcelHeader(rows)
	cols := newColumnIndex(rows[headerIdx])

	if !cols.has("开票日期", "日期", "posting_date", "date", "发票日期") {
		return nil, errors.New("未找到日期列，请确保Excel包含'开票日期'或'日期'列")
	}
	if !cols.has("数电发票号码", "发票号码", "发票号", "invoice_no", "invoice number") &&
		!cols.has("价税合计", "合计", "total_amount", "total amount", "total", "总金额") {
		return nil, errors.New("未找到发票号或金额列，请确保Excel包含'发票号码'和'价税合计'列")
	}

	var defaultCompanyID uuid.UUID
	if cid, cerr := s.repo.GetDefaultCompanyID(ctx, tenantID); cerr == nil {
		defaultCompanyID = cid
	}

	parsed, failedRows := s.parseExcelRows(rows, headerIdx, cols, ctx, tenantID)
	if len(parsed) == 0 {
		return &model.InvoiceFileImportResult{
			TotalRows:  len(rows) - headerIdx - 1,
			Imported:   0,
			Failed:     len(failedRows),
			FailedRows: failedRows,
		}, nil
	}

	groups, groupKeys := groupByInvoiceNo(parsed)
	validGroups, moreFailed := filterConflicts(ctx, s.repo, tenantID, groups, groupKeys)
	failedRows = append(failedRows, moreFailed...)

	toInsert := buildInvoicesFromGroups(validGroups, defaultCompanyID, s.repo, ctx, tenantID, &failedRows)

	invoices := make([]model.SalesInvoice, len(toInsert))
	allItems := make([][]model.InvoiceLineItem, len(toInsert))
	for i, hwi := range toInsert {
		invoices[i] = hwi.Header
		allItems[i] = hwi.Items
	}

	importedResult, err := s.repo.ImportBatchWithItems(ctx, tenantID, invoices, allItems)
	if err != nil {
		return nil, fmt.Errorf("import batch: %w", err)
	}

	return &model.InvoiceFileImportResult{
		TotalRows:  len(parsed),
		Imported:   len(importedResult),
		Failed:     len(failedRows),
		FailedRows: failedRows,
	}, nil
}

func (s *InvoiceService) resolveCustomer(ctx context.Context, tenantID uuid.UUID, buyerName, buyerTaxID string) (uuid.UUID, error) {
	if s.partyRepo == nil {
		return uuid.Nil, errors.New("party repository not available")
	}

	// 1. Try to find by tax_id first
	if buyerTaxID != "" {
		parties, err := s.partyRepo.List(ctx, tenantID)
		if err != nil {
			return uuid.Nil, err
		}
		for _, p := range parties {
			if p.TaxNumber != nil && *p.TaxNumber == buyerTaxID {
				return p.ID, nil
			}
		}
	}

	// 2. Fall back to name match
	party, err := s.partyRepo.GetByName(ctx, tenantID, buyerName)
	if err == nil && party != nil {
		return party.ID, nil
	}

	// 3. If not found, auto-create via upsert (handles concurrent duplicates atomically)
	newParty := &model.Party{
		PartyType: "customer",
		Name:      buyerName,
		Code:      nil, // upsert auto-generates no code — that's fine
		Source:    "auto_import",
	}
	if buyerTaxID != "" {
		newParty.TaxNumber = &buyerTaxID
	}
	id, err := s.partyRepo.Upsert(ctx, tenantID, newParty)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Upsert failed: %w", err)
	}
	return id, nil
}

func findExcelHeader(rows [][]string) int {
	for i, row := range rows {
		nonEmpty := 0
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty++
			}
		}
		if nonEmpty >= 3 {
			return i
		}
	}
	return 0
}

type columnIndex struct {
	idx   map[string]int
	names []string
}

func newColumnIndex(header []string) *columnIndex {
	ci := &columnIndex{idx: make(map[string]int)}
	for i, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		ci.idx[key] = i
		ci.names = append(ci.names, col)
	}
	return ci
}

func (ci *columnIndex) get(names ...string) (int, bool) {
	for _, name := range names {
		key := strings.ToLower(strings.TrimSpace(name))
		if i, ok := ci.idx[key]; ok {
			return i, true
		}
	}
	for origIdx, col := range ci.names {
		colLower := strings.ToLower(col)
		for _, name := range names {
			if strings.Contains(colLower, strings.ToLower(name)) {
				return origIdx, true
			}
		}
	}
	return 0, false
}

func (ci *columnIndex) has(names ...string) bool {
	_, ok := ci.get(names...)
	return ok
}

func (s *InvoiceService) parseExcelRows(rows [][]string, headerIdx int, cols *columnIndex, ctx context.Context, tenantID uuid.UUID) ([]parsedInvoiceRow, []model.FailedRowDetail) {
	invNoIdx, _ := cols.get("数电发票号码", "发票号码", "发票号", "invoice_no", "invoice number")
	invTypeIdx, _ := cols.get("发票票种", "发票类型", "类型", "票种", "invoice_type", "type")
	dateIdx, _ := cols.get("开票日期", "日期", "posting_date", "date", "发票日期")
	customerNameIdx, _ := cols.get("购买方名称", "购方名称", "对方单位", "客户名称", "购方", "购方识别号")
	buyerTaxIdIdx, _ := cols.get("对方税号", "对方识别号", "购方识别号", "购方税号", "购买方税号", "购买方识别号", "购方纳税人识别号", "购买方纳税人识别号", "购方统一社会信用代码", "购买方统一社会信用代码", "对方纳税人识别号", "对方统一社会信用代码", "统一社会信用代码", "纳税人识别号", "购方纳税识别号", "购买方纳税识别号")
	invCodeIdx, _ := cols.get("发票代码", "invoice_code")
	statusIdx, _ := cols.get("发票状态", "状态", "status")
	remarkIdx, _ := cols.get("备注", "remark", "说明")
	invoiceCategoryIdx, _ := cols.get("发票票种", "发票种类", "票种", "invoice_category")
	sourceRedNoIdx, _ := cols.get("对应蓝字发票号", "原蓝字发票号", "红冲发票号", "source_red_invoice_no")
	isPositiveIdx, isPositiveFound := cols.get("是否正数发票")
	isReturnIdx, isReturnFound := cols.get("是否红字", "是否红冲", "is_return", "红字")
	itemDescIdx, _ := cols.get("货物或应税劳务名称", "货物名称", "商品名称", "名称", "description", "摘要")
	itemCodeIdx, _ := cols.get("规格型号", "规格", "型号", "规格型號")
	unitIdx, _ := cols.get("单位", "unit")
	qtyIdx, _ := cols.get("数量", "quantity", "qty")
	unitPriceIdx, _ := cols.get("单价", "unit_price", "unit price", "price")
	taxRateIdx, _ := cols.get("税率", "tax_rate", "tax rate")
	lineNetIdx, _ := cols.get("金额", "net_amount", "net amount", "net")
	lineTaxIdx, _ := cols.get("税额", "tax_amount", "tax amount", "tax", "税金")
	lineTotalIdx, _ := cols.get("价税合计", "合计", "total_amount", "total amount", "total", "总金额")

	var parsed []parsedInvoiceRow
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

		getCell := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		invNo := getCell(invNoIdx)
		if invNo == "" {
			for _, altName := range []string{"数电发票号码", "发票号码", "invoice_no"} {
				if altIdx, ok := cols.get(altName); ok && altIdx < len(row) {
					if v := strings.TrimSpace(row[altIdx]); v != "" {
						invNo = v
						break
					}
				}
			}
		}
		if invNo == "" {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Reason: "发票号为空"})
			continue
		}

		postingDate := time.Now()
		dateStr := getCell(dateIdx)
		if dateStr != "" {
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

		invType := "sale"
		if v := getCell(invTypeIdx); v != "" {
			t := strings.ToLower(v)
			if strings.Contains(t, "购入") || strings.Contains(t, "采购") || t == "purchase" {
				invType = "purchase"
			} else if strings.Contains(t, "红字") || strings.Contains(t, "红冲") || t == "credit_note" {
				invType = "credit_note"
			}
		}

		buyerName := getCell(customerNameIdx)
		if buyerName == "" {
			buyerName = "默认客户"
		}

		customerID, cerr := s.resolveCustomer(ctx, tenantID, buyerName, getCell(buyerTaxIdIdx))
		if cerr != nil {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowNum, Date: invNo, Reason: fmt.Sprintf("客户解析失败 [%s]: %v", buyerName, cerr)})
			continue
		}

		itemNet := parseFloat(lineNetIdx)
		itemTax := parseFloat(lineTaxIdx)
		itemTotal := parseFloat(lineTotalIdx)
		if itemTotal == 0 && itemNet > 0 {
			itemTotal = itemNet + itemTax
		}

		r := parsedInvoiceRow{
			RowNum:      rowNum,
			InvoiceNo:   invNo,
			PostingDate: postingDate,
			NetAmount:   itemNet,
			TaxAmount:   itemTax,
			TotalAmount: itemTotal,
			CustomerID:  customerID,
			BuyerName:   buyerName,
			InvoiceType: invType,
			BuyerTaxID:  getCell(buyerTaxIdIdx),
			InvoiceCode: getCell(invCodeIdx),
			InvoiceCategory: getCell(invoiceCategoryIdx),
			Status:      getCell(statusIdx),
			Remark:      getCell(remarkIdx),
			ItemDescription: getCell(itemDescIdx),
			ItemCode:        getCell(itemCodeIdx),
			Unit:            getCell(unitIdx),
			Quantity:        parseFloat(qtyIdx),
			UnitPrice:       parseFloat(unitPriceIdx),
			TaxRate:         parseFloat(taxRateIdx),
		}

		if isPositiveFound {
			v := getCell(isPositiveIdx)
			r.IsReturn = (v == "否" || v == "false" || v == "0")
		} else if isReturnFound {
			v := strings.ToLower(getCell(isReturnIdx))
			r.IsReturn = (v == "是" || v == "yes" || v == "true" || v == "1" || v == "红字" || v == "红冲")
		}
		if srcNo := getCell(sourceRedNoIdx); srcNo != "" {
			r.SourceRedInvoiceNo = srcNo
		}

		parsed = append(parsed, r)
	}
	return parsed, failedRows
}

func groupByInvoiceNo(rows []parsedInvoiceRow) (map[string]*invoiceGroup, []string) {
	groups := make(map[string]*invoiceGroup)
	var keys []string
	for i := range rows {
		r := &rows[i]
		if g, ok := groups[r.InvoiceNo]; ok {
			g.Rows = append(g.Rows, *r)
		} else {
			groups[r.InvoiceNo] = &invoiceGroup{
				Rows:   []parsedInvoiceRow{*r},
				Header: *r,
			}
			keys = append(keys, r.InvoiceNo)
		}
	}
	return groups, keys
}

func filterConflicts(ctx context.Context, repo *repository.InvoiceRepository, tenantID uuid.UUID, groups map[string]*invoiceGroup, groupKeys []string) ([]struct {
	Key  string
	Data *invoiceGroup
}, []model.FailedRowDetail) {
	var allInvoiceNos []string
	for _, k := range groupKeys {
		allInvoiceNos = append(allInvoiceNos, k)
	}
	conflicts, err := repo.ValidateDuplicateBatch(ctx, tenantID, allInvoiceNos)
	if err != nil {
		return nil, nil
	}
	conflictSet := make(map[string]bool, len(conflicts))
	for _, c := range conflicts {
		conflictSet[c] = true
	}

	var valid []struct {
		Key  string
		Data *invoiceGroup
	}
	var failedRows []model.FailedRowDetail
	for _, k := range groupKeys {
		if conflictSet[k] {
			for _, r := range groups[k].Rows {
				failedRows = append(failedRows, model.FailedRowDetail{
					Row:    r.RowNum,
					Date:   r.InvoiceNo,
					Reason: fmt.Sprintf("发票号 %s 已存在，导入被拒绝", r.InvoiceNo),
				})
			}
		} else {
			valid = append(valid, struct {
				Key  string
				Data *invoiceGroup
			}{k, groups[k]})
		}
	}
	return valid, failedRows
}

func buildInvoicesFromGroups(validGroups []struct {
	Key  string
	Data *invoiceGroup
}, defaultCompanyID uuid.UUID, repo *repository.InvoiceRepository, ctx context.Context, tenantID uuid.UUID, failedRows *[]model.FailedRowDetail) []headerWithItems {
	var toInsert []headerWithItems

	for _, g := range validGroups {
		h := g.Data.Header
		items := g.Data.Rows

		var sumNet, sumTax, sumTotal float64
		for i := range items {
			it := &items[i]
			sumNet += it.NetAmount
			sumTax += it.TaxAmount
			sumTotal += it.TotalAmount
		}
		if sumTotal == 0 && len(items) > 0 {
			sumNet = items[0].NetAmount
			sumTax = items[0].TaxAmount
			sumTotal = items[0].TotalAmount
		}
		if sumTotal == 0 {
			for _, it := range items {
				*failedRows = append(*failedRows, model.FailedRowDetail{Row: it.RowNum, Date: it.InvoiceNo, Reason: "金额为空"})
			}
			continue
		}

		header := model.SalesInvoice{
			InvoiceNo:         h.InvoiceNo,
			InvoiceType:       h.InvoiceType,
			CustomerID:        h.CustomerID,
			CompanyID:         defaultCompanyID,
			PostingDate:       h.PostingDate,
			TotalAmount:       decimal.NewFromFloat(sumTotal),
			TaxAmount:         decimal.NewFromFloat(sumTax),
			NetAmount:         decimal.NewFromFloat(sumNet),
			OutstandingAmount: decimal.NewFromFloat(sumTotal),
			Status:            "unpaid",
		}
		if h.BuyerTaxID != "" {
			header.TaxID = &h.BuyerTaxID
		}
		if h.InvoiceCode != "" {
			header.InvoiceCode = &h.InvoiceCode
		}
		if h.InvoiceCategory != "" {
			header.InvoiceCategory = &h.InvoiceCategory
		}
		// Imported invoices always start as draft (unpaid).
		// The "状态" column in Excel is informational only; manual 确认 is required to advance status.
		if h.Remark != "" {
			header.Remark = &h.Remark
		}
		header.IsReturn = h.IsReturn

		srcNo := h.SourceRedInvoiceNo
		if srcNo == "" && h.Remark != "" {
			srcNo = extractBlueInvoiceNo(h.Remark)
		}
		if srcNo != "" {
			header.SourceRedInvoiceNo = &srcNo
			if srcInv, lerr := repo.GetByInvoiceNo(ctx, tenantID, srcNo); lerr == nil && srcInv != nil {
				id := srcInv.ID
				header.ReturnAgainst = &id
			}
		}

		var lineItems []model.InvoiceLineItem
		for i := range items {
			it := &items[i]
			lineItems = append(lineItems, model.InvoiceLineItem{
				Description: it.ItemDescription,
				ItemCode:    it.ItemCode,
				Unit:        it.Unit,
				Quantity:    decimal.NewFromFloat(it.Quantity),
				UnitPrice:   decimal.NewFromFloat(it.UnitPrice),
				TaxRate:     decimal.NewFromFloat(it.TaxRate),
				NetAmount:   decimal.NewFromFloat(it.NetAmount),
				TaxAmount:   decimal.NewFromFloat(it.TaxAmount),
				TotalAmount: decimal.NewFromFloat(it.TotalAmount),
			})
		}
		toInsert = append(toInsert, headerWithItems{Header: header, Items: lineItems})
	}
	return toInsert
}

func extractBlueInvoiceNo(remark string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`被红冲蓝字数电票号码[：:]\s*(\d{8,20})`),
		regexp.MustCompile(`被红冲蓝字数电发票号码[：:]\s*(\d{8,20})`),
		regexp.MustCompile(`对应蓝字发票号[：:]\s*(\d{8,20})`),
		regexp.MustCompile(`对应正数发票号码[：:]\s*(\d{8,20})`),
		regexp.MustCompile(`红冲发票[：:号]?\s*(\d{8,20})`),
		regexp.MustCompile(`蓝字发票[：:号]?\s*(\d{8,20})`),
	}
	for _, p := range patterns {
		if m := p.FindStringSubmatch(remark); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

// parseDateInvoice tries common date formats for invoice dates.
func parseDateInvoice(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02", "2006/01/02", "20060102",
		"2006-01-02 15:04:05", "2006/01/02 15:04:05",
		"02/01/2006", "01/02/2006",
		"2006年01月02日", "2006年1月2日",
		"1/2/06 15:04", "01/02/06 15:04",
		"2/1/06 15:04", "02/01/06 15:04",
		"1/2/06", "01/02/06",
		"2/1/06", "02/01/06",
		"1/2/06 15:08", "01/02/06 15:04:05",
		"2006-01-02 15:04", "2006/01/02 15:04",
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}

// BatchImportPreview parses Excel file and returns preview with AI deduplication check.
func (s *InvoiceService) BatchImportPreview(ctx context.Context, tenantID uuid.UUID, data []byte) (*model.InvoiceBatchPreviewResult, error) {
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

	invNoIdx, _ := findCol("发票号码", "发票号", "invoice_no", "invoice number", "invoice")
	nameIdx, _ := findCol("购买方名称", "购方名称", "客户名称", "customer_name", "customer", "客户")
	dateIdx, dateFound := findCol("开票日期", "日期", "posting_date", "date")
	netIdx, _ := findCol("不含税金额", "金额", "net_amount", "net")
	taxIdx, _ := findCol("税额", "tax_amount", "tax")
	totalIdx, _ := findCol("价税合计", "合计", "total_amount", "total")
	_, _ = findCol("购方识别号", "税号", "tax_id", "taxid")

	if !dateFound {
		return nil, errors.New("未找到日期列，请确保Excel包含'开票日期'或'日期'列")
	}

	var details []model.InvoicePreviewDetail
	var validCount, errorCount, duplicateCount int

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

		detail := model.InvoicePreviewDetail{RowIndex: rowNum}

		if invNoIdx < len(row) {
			detail.InvoiceNo = strings.TrimSpace(row[invNoIdx])
		}
		if nameIdx < len(row) {
			detail.CustomerName = strings.TrimSpace(row[nameIdx])
		}
		if dateIdx < len(row) {
			detail.PostingDate = strings.TrimSpace(row[dateIdx])
		}

		var validationErrs []string

		if detail.InvoiceNo == "" {
			validationErrs = append(validationErrs, "发票号为空")
		}

		if detail.PostingDate != "" {
			if _, err := parseDateInvoice(detail.PostingDate); err != nil {
				validationErrs = append(validationErrs, "日期格式无效")
			}
		} else {
			validationErrs = append(validationErrs, "日期为空")
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

		detail.NetAmount = parseFloat(netIdx)
		detail.TaxAmount = parseFloat(taxIdx)
		detail.TotalAmount = parseFloat(totalIdx)

		if detail.TotalAmount == 0 && detail.NetAmount > 0 {
			detail.TotalAmount = detail.NetAmount + detail.TaxAmount
		}

		if detail.TotalAmount == 0 {
			validationErrs = append(validationErrs, "金额为空")
		} else if detail.TotalAmount < 0 {
			validationErrs = append(validationErrs, "金额为负数")
		}

		if len(validationErrs) > 0 {
			detail.Status = "error"
			detail.ValidationErr = strings.Join(validationErrs, "; ")
			errorCount++
		} else {
			if detail.InvoiceNo != "" {
				exists, err := s.repo.ValidateDuplicateInvoiceNo(ctx, tenantID, detail.InvoiceNo)
				if err != nil {
					detail.Status = "warning"
					detail.ValidationErr = "查重失败: " + err.Error()
					errorCount++
				} else if exists {
					detail.Status = "duplicate"
					detail.IsDuplicate = true
					detail.DuplicateInfo = "发票号已存在于系统中"
					duplicateCount++
				} else {
					detail.Status = "valid"
					validCount++
				}
			} else {
				detail.Status = "valid"
				validCount++
			}
		}

		details = append(details, detail)
	}

	// Build customer matches info
	customerMatches := make([]model.CustomerMatchInfo, 0, len(details))
	for i, detail := range details {
		status := "matched"
		warning := ""
		if detail.Status == "error" || detail.Status == "warning" {
			status = "error"
			if detail.ValidationErr != "" {
				warning = detail.ValidationErr
			}
		} else if detail.Status == "duplicate" {
			status = "fuzzy"
			warning = detail.DuplicateInfo
		}
		customerMatches = append(customerMatches, model.CustomerMatchInfo{
			RowIndex:       i + 1,
			Status:         status,
			CustomerName:   detail.CustomerName,
			WarningMessage: warning,
		})
	}

	return &model.InvoiceBatchPreviewResult{
		BatchID:          uuid.New().String(),
		TotalRows:        len(details),
		ValidRows:        validCount,
		ErrorRows:        errorCount,
		DuplicateRows:    duplicateCount,
		Details:          details,
		CustomerMatches:  customerMatches,
		WillGenerateAs: model.WillGenerateSummary{
			InvoicesWillCreate: validCount,
			ARWillCreate:       0,
			VouchersWillCreate: 0,
		},
	}, nil
}

// BatchImportConfirm confirms and imports the batch invoices.
func (s *InvoiceService) BatchImportConfirm(ctx context.Context, tenantID uuid.UUID, req *model.InvoiceBatchConfirmRequest) (*model.InvoiceBatchConfirmResult, error) {
	if req == nil || len(req.SelectedIDs) == 0 {
		return nil, errors.New("no invoices selected for import")
	}

	var failedRows []model.FailedRowDetail
	importedCount := 0

	for _, rowID := range req.SelectedIDs {
		rowIdx, err := strconv.Atoi(rowID)
		if err != nil {
			failedRows = append(failedRows, model.FailedRowDetail{Row: rowIdx, Reason: "无效的行ID"})
			continue
		}

		failedRows = append(failedRows, model.FailedRowDetail{Row: rowIdx, Reason: "请使用完整的导入流程"})
	}

	return &model.InvoiceBatchConfirmResult{
		Imported:   importedCount,
		Skipped:    len(req.SelectedIDs) - importedCount - len(failedRows),
		Errors:     len(failedRows),
		FailedRows: failedRows,
	}, nil
}

// VoucherAutoGenerateService getter for injection (used by main.go after initialization)
func (s *InvoiceService) InjectAutoGenSvc(svc *VoucherAutoGenerateService) {
	s.voucherAutoSvc = svc
}

// ConfirmSalesInvoice confirms a sales invoice, generates an ArInvoice draft, and triggers voucher auto-generation.
func (s *InvoiceService) ConfirmSalesInvoice(ctx context.Context, tenantID, invoiceID, userID uuid.UUID) error {
	// 1. Get invoice
	inv, err := s.repo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return err
	}
	if inv == nil {
		return errors.New("invoice not found")
	}

	// 2. Check status
	// Normalize status for confirmation eligibility check (support legacy Chinese values)
	normalizedStatus := inv.Status
	if normalizedStatus == "正常" || normalizedStatus == "unpaid" {
		normalizedStatus = "draft"
	} else if normalizedStatus == "已确认" {
		normalizedStatus = "verified"
	} else if normalizedStatus == "部分核销" {
		normalizedStatus = "partially_paid"
	}
	if normalizedStatus != "draft" && normalizedStatus != "submitted" && normalizedStatus != "verified" {
		return errors.New("invalid invoice status for confirmation")
	}

	// 3. Check if ArInvoice already exists (prevent duplicate)
	if s.arInvoiceRepo != nil {
		existing, err := s.arInvoiceRepo.ListByInvoiceID(ctx, tenantID, invoiceID)
		if err != nil {
			return err
		}
		if existing != nil {
			return errors.New("ar_invoice already exists for this invoice")
		}
	}

	// 4. Generate ArInvoice draft
	if s.arInvoiceRepo != nil {
		ar := &model.ArInvoice{
			ID:         uuid.New(),
			TenantID:   tenantID,
			CompanyID:  inv.CompanyID,
			CustomerID: inv.CustomerID,
			InvoiceID:  invoiceID,
			InvoiceNo:  inv.InvoiceNo,
			Amount:     inv.TotalAmount,
			DueDate:    inv.DueDate,
			Status:     string(model.ArInvoiceStatusDraft),
			SourceType: "auto_import",
			CreatedBy:  &userID,
			CreatedAt:  time.Now(),
		}
		if err := s.arInvoiceRepo.Create(ctx, ar); err != nil {
			return fmt.Errorf("failed to create ar_invoice: %v", err)
		}
	}

	// 5. Trigger voucher auto-generation (non-critical, log on failure)
	if s.voucherAutoSvc != nil {
		if _, err := s.voucherAutoSvc.GenerateFromInvoice(ctx, tenantID, invoiceID, userID); err != nil {
			fmt.Printf("[WARN] failed to generate voucher for invoice %s: %v\n", invoiceID, err)
		}
	}

	// 6. Update invoice status to verified
	return s.repo.UpdateStatus(ctx, tenantID, invoiceID, string(model.InvoiceStatusVerified))
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

// ListUnmatchedInvoices returns invoices with outstanding balance for a given party.
func (s *InvoiceService) ListUnmatchedInvoices(ctx context.Context, tenantID uuid.UUID, partyID *uuid.UUID) ([]model.SalesInvoice, error) {
	return s.repo.ListInvoicesForMatching(ctx, tenantID, partyID)
}

// AllocationRequest represents a single invoice allocation within a payment entry.
type AllocationRequest struct {
	InvoiceID       uuid.UUID
	AllocatedAmount decimal.Decimal
}

// AllocateToPaymentEntry creates payment allocations and updates invoice outstanding amounts.
func (s *InvoiceService) AllocateToPaymentEntry(ctx context.Context, tenantID uuid.UUID, paymentEntryID uuid.UUID, allocations []AllocationRequest) ([]model.PaymentAllocation, error) {
	var result []model.PaymentAllocation
	for _, a := range allocations {
		inv, err := s.repo.GetByID(ctx, tenantID, a.InvoiceID)
		if err != nil {
			return nil, err
		}
		if inv == nil {
			return nil, fmt.Errorf("invoice %s not found", a.InvoiceID)
		}
		if a.AllocatedAmount.GreaterThan(inv.OutstandingAmount) {
			return nil, fmt.Errorf("allocation amount %s exceeds outstanding %s for invoice %s",
				a.AllocatedAmount.String(), inv.OutstandingAmount.String(), a.InvoiceID)
		}

		invoiceType := "sale"
		alloc := &model.PaymentAllocation{
			PaymentEntryID:  paymentEntryID,
			InvoiceID:       a.InvoiceID,
			InvoiceType:     &invoiceType,
			AllocatedAmount: a.AllocatedAmount,
			TenantID:        tenantID,
		}

		if err := s.repo.CreateAllocation(ctx, alloc); err != nil {
			return nil, err
		}
		newOutstanding := inv.OutstandingAmount.Sub(a.AllocatedAmount)
		if err := s.repo.UpdateOutstandingAmount(ctx, tenantID, a.InvoiceID, newOutstanding.String()); err != nil {
			return nil, err
		}
		result = append(result, *alloc)
	}
	return result, nil
}
