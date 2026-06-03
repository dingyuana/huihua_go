package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherTemplateService handles voucher template business logic.
type VoucherTemplateService struct {
	repo        *repository.VoucherTemplateRepository
	accountRepo *repository.AccountRepository
}

// NewVoucherTemplateService creates a new VoucherTemplateService.
func NewVoucherTemplateService(repo *repository.VoucherTemplateRepository, accountRepo *repository.AccountRepository) *VoucherTemplateService {
	return &VoucherTemplateService{repo: repo, accountRepo: accountRepo}
}

// CreateTemplate creates a new voucher template with validation.
func (s *VoucherTemplateService) CreateTemplate(ctx context.Context, tenantID uuid.UUID, req *model.CreateTemplateRequest) (*model.VoucherTemplate, error) {
	// Validate accounts exist
	for _, line := range req.Lines {
		account, err := s.accountRepo.GetByID(ctx, tenantID, line.AccountID)
		if err != nil {
			return nil, fmt.Errorf("validate account %s: %w", line.AccountID, err)
		}
		if account == nil {
			return nil, fmt.Errorf("account %s not found", line.AccountID)
		}
	}

	numberPrefix := req.NumberPrefix
	if numberPrefix == "" {
		numberPrefix = "PZ"
	}

	t := &model.VoucherTemplate{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		NumberPrefix:   numberPrefix,
		IsActive:       req.IsActive,
		Classification: req.Classification,
		ApprovalFlowID: req.ApprovalFlowID,
	}

	for i, lineReq := range req.Lines {
		line := model.VoucherTemplateLine{
			ID:               uuid.New(),
			AccountID:        lineReq.AccountID,
			DrAmountTemplate: lineReq.DrAmountTemplate,
			CrAmountTemplate: lineReq.CrAmountTemplate,
			SummaryTemplate:  lineReq.SummaryTemplate,
			DimensionType:    lineReq.DimensionType,
			DimensionValue:   lineReq.DimensionValue,
			LineOrder:        lineReq.LineOrder,
		}
		if line.LineOrder == 0 {
			line.LineOrder = i + 1
		}
		t.Lines = append(t.Lines, line)
	}

	return s.repo.CreateTemplate(ctx, tenantID, t)
}

// GetTemplateByID returns a template with its lines.
func (s *VoucherTemplateService) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*model.VoucherTemplate, error) {
	return s.repo.GetTemplateByID(ctx, tenantID, id)
}

// ListTemplates returns all templates for a tenant.
func (s *VoucherTemplateService) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]model.VoucherTemplate, error) {
	return s.repo.ListTemplates(ctx, tenantID)
}

// UpdateTemplate updates a voucher template.
func (s *VoucherTemplateService) UpdateTemplate(ctx context.Context, tenantID, id uuid.UUID, req *model.UpdateTemplateRequest) error {
	// Validate accounts exist
	for _, lineReq := range req.Lines {
		account, err := s.accountRepo.GetByID(ctx, tenantID, lineReq.AccountID)
		if err != nil {
			return fmt.Errorf("validate account %s: %w", lineReq.AccountID, err)
		}
		if account == nil {
			return fmt.Errorf("account %s not found", lineReq.AccountID)
		}
	}

	existing, err := s.repo.GetTemplateByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("template not found")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.NumberPrefix != "" {
		existing.NumberPrefix = req.NumberPrefix
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.ApprovalFlowID != nil {
		existing.ApprovalFlowID = req.ApprovalFlowID
	}
	if req.Classification != nil {
		existing.Classification = req.Classification
	}

	// Replace lines
	existing.Lines = nil
	for i, lineReq := range req.Lines {
		line := model.VoucherTemplateLine{
			ID:               uuid.New(),
			AccountID:        lineReq.AccountID,
			DrAmountTemplate: lineReq.DrAmountTemplate,
			CrAmountTemplate: lineReq.CrAmountTemplate,
			SummaryTemplate:  lineReq.SummaryTemplate,
			DimensionType:    lineReq.DimensionType,
			DimensionValue:   lineReq.DimensionValue,
			LineOrder:        lineReq.LineOrder,
		}
		if line.LineOrder == 0 {
			line.LineOrder = i + 1
		}
		existing.Lines = append(existing.Lines, line)
	}

	return s.repo.UpdateTemplate(ctx, tenantID, id, existing)
}

// DeleteTemplate performs a soft delete.
func (s *VoucherTemplateService) DeleteTemplate(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.DeleteTemplate(ctx, tenantID, id)
}

// GetOrCreateNumberingRule returns the numbering rule or creates a default one.
func (s *VoucherTemplateService) GetOrCreateNumberingRule(ctx context.Context, tenantID uuid.UUID) (*model.VoucherNumberingRule, error) {
	rule, err := s.repo.GetNumberingRule(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if rule != nil {
		return rule, nil
	}

	// Create default rule
	defaultReq := &model.NumberingRuleRequest{
		Prefix:     "PZ",
		DateFormat: "20060102",
		ResetRule:  "daily",
	}
	return s.repo.CreateOrUpdateNumberingRule(ctx, tenantID, defaultReq)
}

// UpdateNumberingRule updates or creates the numbering rule.
func (s *VoucherTemplateService) UpdateNumberingRule(ctx context.Context, tenantID uuid.UUID, req *model.NumberingRuleRequest) (*model.VoucherNumberingRule, error) {
	return s.repo.CreateOrUpdateNumberingRule(ctx, tenantID, req)
}

// GenerateVoucherNumber generates a unique voucher number.
func (s *VoucherTemplateService) GenerateVoucherNumber(ctx context.Context, tenantID uuid.UUID) (*model.VoucherNumberResponse, error) {
	rule, err := s.GetOrCreateNumberingRule(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get numbering rule: %w", err)
	}

	// Check if we need to reset based on reset_rule
	if !s.shouldReset(rule) {
		// Reset the next_number if needed
		_, err = s.UpdateNumberingRule(ctx, tenantID, &model.NumberingRuleRequest{
			Prefix:     rule.Prefix,
			DateFormat: rule.DateFormat,
			ResetRule:  rule.ResetRule,
		})
		if err != nil {
			return nil, err
		}
	}

	seq, err := s.repo.GenerateNextNumber(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("generate next number: %w", err)
	}

	voucherNumber := FormatVoucherNumber(rule.Prefix, time.Now(), seq)
	return &model.VoucherNumberResponse{
		VoucherNumber: voucherNumber,
		Sequence:      seq,
	}, nil
}

// shouldReset checks if the sequence should be reset based on the reset rule.
func (s *VoucherTemplateService) shouldReset(rule *model.VoucherNumberingRule) bool {
	if rule == nil {
		return true
	}
	switch rule.ResetRule {
	case "yearly":
		return false // We don't track yearly resets without more state
	case "monthly":
		return false
	case "daily":
		return false
	default:
		return false
	}
}

// FormatVoucherNumber formats a voucher number as PREFIX-YYYYMMDD-SEQ.
func FormatVoucherNumber(prefix string, date time.Time, seq int) string {
	dateStr := date.Format("20060102")
	return fmt.Sprintf("%s-%s-%03d", prefix, dateStr, seq)
}
