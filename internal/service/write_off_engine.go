package service

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

const (
	RuleTypeExactMatch     = "exact_match"
	RuleTypeContractMatch  = "contract_match"
	RuleTypeVoucherMatch   = "voucher_match"
	RuleTypeFuzzyMatch     = "fuzzy_match"
	RuleTypeFIFO           = "fifo"
	RuleTypePrepaidOffset  = "prepaid_offset"
	RuleTypeTolerance      = "tolerance"
)

type WriteOffEngine struct {
	paymentRepo   *repository.PaymentEntryRepository
	arInvoiceRepo *repository.ArInvoiceRepository
	apInvoiceRepo *repository.ApInvoiceRepository
	writeOffRepo  *repository.WriteOffRepository
	rules         []*model.WriteOffRule
}

func NewWriteOffEngine(
	paymentRepo *repository.PaymentEntryRepository,
	arInvoiceRepo *repository.ArInvoiceRepository,
	apInvoiceRepo *repository.ApInvoiceRepository,
	writeOffRepo *repository.WriteOffRepository,
) *WriteOffEngine {
	return &WriteOffEngine{
		paymentRepo:   paymentRepo,
		arInvoiceRepo: arInvoiceRepo,
		apInvoiceRepo: apInvoiceRepo,
		writeOffRepo:  writeOffRepo,
	}
}

func (e *WriteOffEngine) LoadRules(ctx context.Context, tenantID uuid.UUID) error {
	e.rules = []*model.WriteOffRule{
		{RuleType: RuleTypeExactMatch, Priority: 1, Enabled: true},
		{RuleType: RuleTypeContractMatch, Priority: 2, Enabled: true},
		{RuleType: RuleTypeVoucherMatch, Priority: 3, Enabled: true},
		{RuleType: RuleTypeFuzzyMatch, Priority: 4, Enabled: true, DateWindow: 3},
		{RuleType: RuleTypeFIFO, Priority: 5, Enabled: true},
		{RuleType: RuleTypePrepaidOffset, Priority: 6, Enabled: true},
		{RuleType: RuleTypeTolerance, Priority: 7, Enabled: true, ToleranceAmount: "5.00", DiffAccountCode: "6603"},
	}
	return nil
}

func (e *WriteOffEngine) sortRulesByPriority() {
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority < e.rules[j].Priority
	})
}

type MatchResult struct {
	PaymentID        uuid.UUID
	PaymentNo        string
	InvoiceID        uuid.UUID
	InvoiceNo        string
	InvoiceType      string
	Amount           decimal.Decimal
	DiffAmount       decimal.Decimal
	DiffAccountCode  string
	MatchRule        string
	CounterpartyID   uuid.UUID
	CounterpartyName string
}

type MatchFailure struct {
	PaymentID      uuid.UUID
	PaymentNo      string
	InvoiceID      uuid.UUID
	InvoiceNo      string
	CounterpartyID uuid.UUID
	FailureReasons []string
}

func (e *WriteOffEngine) MatchPaymentAR(ctx context.Context, tenantID, paymentID uuid.UUID) ([]MatchResult, []MatchFailure, error) {
	payment, err := e.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, nil, err
	}
	if payment.PaymentType != "receipt" {
		return nil, nil, nil
	}

	invoices, err := e.arInvoiceRepo.ListByTenant(ctx, tenantID, nil)
	if err != nil {
		return nil, nil, err
	}

	var filteredInvoices []*model.ArInvoice
	for _, inv := range invoices {
		if inv.OutstandingAmount.GreaterThan(decimal.Zero) && inv.CustomerID == payment.PartyID {
			filteredInvoices = append(filteredInvoices, inv)
		}
	}

	prepaidBalance, err := e.writeOffRepo.GetPrepaidBalance(ctx, tenantID, payment.PartyID)
	if err != nil {
		return nil, nil, err
	}

	return e.matchPaymentWithInvoices(ctx, payment, filteredInvoices, prepaidBalance, "ar_invoice")
}

func (e *WriteOffEngine) MatchPaymentAP(ctx context.Context, tenantID, paymentID uuid.UUID) ([]MatchResult, []MatchFailure, error) {
	return nil, nil, nil
}

func (e *WriteOffEngine) matchPaymentWithInvoices(
	ctx context.Context,
	payment *model.PaymentEntry,
	invoices []*model.ArInvoice,
	prepaidBalance decimal.Decimal,
	invoiceType string,
) ([]MatchResult, []MatchFailure, error) {
	var results []MatchResult
	var failures []MatchFailure
	remainingAmount := payment.UnallocatedAmount

	if prepaidBalance.GreaterThan(decimal.Zero) && invoiceType == "ar_invoice" {
		prepaidResults, _ := e.matchPrepaidWithInvoices(invoices, prepaidBalance)
		results = append(results, prepaidResults...)
		matchedIDs := make(map[uuid.UUID]bool)
		for _, res := range prepaidResults {
			matchedIDs[res.InvoiceID] = true
		}

		var remainingInvoices []*model.ArInvoice
		for _, inv := range invoices {
			if !matchedIDs[inv.ID] {
				remainingInvoices = append(remainingInvoices, inv)
			}
		}
		invoices = remainingInvoices
	}

	sortedInvoices := e.sortInvoicesByDate(invoices)
	paymentDesc := ""
	if payment.Description != nil {
		paymentDesc = *payment.Description
	}

	for _, invoice := range sortedInvoices {
		if remainingAmount.LessThanOrEqual(decimal.Zero) {
			break
		}

		failureReasons := []string{}
		matchAmount, diffAmount, diffAccount, matchRule := e.evaluateMatch(payment, invoice, paymentDesc, &failureReasons)
		
		if matchAmount.GreaterThan(decimal.Zero) {
			counterpartyName := ""
			if payment.CounterpartyName != nil {
				counterpartyName = *payment.CounterpartyName
			}
			
			results = append(results, MatchResult{
				PaymentID:        payment.ID,
				PaymentNo:        payment.PaymentNo,
				InvoiceID:        invoice.ID,
				InvoiceNo:        invoice.InvoiceNo,
				InvoiceType:      invoiceType,
				Amount:           matchAmount,
				DiffAmount:       diffAmount,
				DiffAccountCode:  diffAccount,
				MatchRule:        matchRule,
				CounterpartyID:   payment.PartyID,
				CounterpartyName: counterpartyName,
			})
			remainingAmount = remainingAmount.Sub(matchAmount)
		} else {
			failures = append(failures, MatchFailure{
				PaymentID:      payment.ID,
				PaymentNo:      payment.PaymentNo,
				InvoiceID:      invoice.ID,
				InvoiceNo:      invoice.InvoiceNo,
				CounterpartyID: payment.PartyID,
				FailureReasons: failureReasons,
			})
		}
	}

	return results, failures, nil
}

func (e *WriteOffEngine) matchPrepaidWithInvoices(
	invoices []*model.ArInvoice,
	prepaidBalance decimal.Decimal,
) ([]MatchResult, decimal.Decimal) {
	var results []MatchResult
	remaining := prepaidBalance

	sortedInvoices := e.sortInvoicesByDate(invoices)

	for _, invoice := range sortedInvoices {
		if remaining.LessThanOrEqual(decimal.Zero) {
			break
		}

		matchAmount := decimal.Min(invoice.OutstandingAmount, remaining)

		results = append(results, MatchResult{
			InvoiceID:       invoice.ID,
			InvoiceNo:       invoice.InvoiceNo,
			InvoiceType:     "ar_invoice",
			Amount:          matchAmount,
			DiffAmount:      decimal.Zero,
			DiffAccountCode: "",
			MatchRule:       RuleTypePrepaidOffset,
		})

		remaining = remaining.Sub(matchAmount)
	}

	return results, remaining
}

func (e *WriteOffEngine) evaluateMatch(
	payment *model.PaymentEntry,
	invoice *model.ArInvoice,
	paymentDesc string,
	failureReasons *[]string,
) (decimal.Decimal, decimal.Decimal, string, string) {
	if payment.PartyID != invoice.CustomerID {
		*failureReasons = append(*failureReasons, "客户不匹配")
		return decimal.Zero, decimal.Zero, "", ""
	}

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		switch rule.RuleType {
		case RuleTypeExactMatch:
			if e.matchExact(paymentDesc, invoice.InvoiceNo, failureReasons) {
				return invoice.OutstandingAmount, decimal.Zero, "", RuleTypeExactMatch
			}
		case RuleTypeContractMatch:
			if e.matchContract(paymentDesc, invoice.InvoiceNo, failureReasons) {
				return invoice.OutstandingAmount, decimal.Zero, "", RuleTypeContractMatch
			}
		case RuleTypeFuzzyMatch:
			if e.matchFuzzy(payment, invoice, rule.DateWindow, failureReasons) {
				return invoice.OutstandingAmount, decimal.Zero, "", RuleTypeFuzzyMatch
			}
		case RuleTypeTolerance:
			tolerance, _ := decimal.NewFromString(rule.ToleranceAmount)
			diff := payment.UnallocatedAmount.Sub(invoice.OutstandingAmount).Abs()
			if diff.LessThanOrEqual(tolerance) {
				return invoice.OutstandingAmount, diff, rule.DiffAccountCode, RuleTypeTolerance
			}
		}
	}

	*failureReasons = append(*failureReasons, "未找到匹配规则")
	return decimal.Zero, decimal.Zero, "", ""
}

func (e *WriteOffEngine) matchExact(paymentDesc, invoiceNo string, failureReasons *[]string) bool {
	if invoiceNo == "" {
		*failureReasons = append(*failureReasons, "发票号为空")
		return false
	}
	if paymentDesc == "" {
		*failureReasons = append(*failureReasons, "付款描述为空")
		return false
	}

	invoiceNoClean := strings.ReplaceAll(strings.ReplaceAll(invoiceNo, "-", ""), " ", "")
	descLower := strings.ToLower(paymentDesc)
	invoiceLower := strings.ToLower(invoiceNoClean)

	return strings.Contains(descLower, invoiceLower) || strings.Contains(invoiceLower, descLower)
}

func (e *WriteOffEngine) matchContract(paymentDesc, invoiceNo string, failureReasons *[]string) bool {
	re := regexp.MustCompile(`(PO|SO|合同|订单)[-_]?(\d{6,})`)
	matches := re.FindStringSubmatch(paymentDesc)
	if len(matches) >= 2 {
		contractNo := matches[2]
		return strings.Contains(strings.ToLower(invoiceNo), strings.ToLower(contractNo))
	}
	return false
}

func (e *WriteOffEngine) matchFuzzy(payment *model.PaymentEntry, invoice *model.ArInvoice, dateWindow int, failureReasons *[]string) bool {
	if !payment.UnallocatedAmount.Equal(invoice.OutstandingAmount) {
		*failureReasons = append(*failureReasons, "金额不匹配")
		return false
	}

	if invoice.DueDate != nil {
		daysDiff := int(time.Since(*invoice.DueDate).Hours() / 24)
		if daysDiff > dateWindow || daysDiff < -dateWindow {
			*failureReasons = append(*failureReasons, "日期差超过"+string(rune(dateWindow+'0'))+"天")
			return false
		}
	}

	return true
}

func (e *WriteOffEngine) sortInvoicesByDate(invoices []*model.ArInvoice) []*model.ArInvoice {
	sorted := make([]*model.ArInvoice, len(invoices))
	copy(sorted, invoices)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].DueDate != nil && sorted[j].DueDate != nil {
			return sorted[i].DueDate.Before(*sorted[j].DueDate)
		}
		if sorted[i].DueDate != nil {
			return true
		}
		return false
	})
	return sorted
}

func (e *WriteOffEngine) GenerateDraftWriteOffRecords(results []MatchResult, operatorID uuid.UUID) []*model.WriteOffRecord {
	var records []*model.WriteOffRecord
	now := time.Now()

	for _, result := range results {
		record := &model.WriteOffRecord{
			TenantID:              uuid.Nil,
			WriteOffNo:            generateWriteOffNo(),
			Type:                  getWriteOffType(result.InvoiceType),
			ReceiptPaymentID:      result.PaymentID,
			ReceivablePayableID:   result.InvoiceID,
			ReceivablePayableType: result.InvoiceType,
			Amount:               result.Amount,
			DiffAmount:           result.DiffAmount,
			DiffAccountCode:      result.DiffAccountCode,
			WriteOffDate:         now,
			Operator:             &operatorID,
			Status:               model.WriteOffStatusDraft,
			MatchRule:            &result.MatchRule,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		records = append(records, record)
	}

	return records
}

func generateWriteOffNo() string {
	return "WO" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
}

func getWriteOffType(invoiceType string) string {
	if invoiceType == "ar_invoice" {
		return model.WriteOffTypeReceiptAR
	}
	return model.WriteOffTypePaymentAP
}