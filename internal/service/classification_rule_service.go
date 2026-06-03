package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ClassificationRuleService handles classification rule business logic.
type ClassificationRuleService struct {
	repo *repository.ClassificationRuleRepository
}

// NewClassificationRuleService creates a new ClassificationRuleService.
func NewClassificationRuleService(repo *repository.ClassificationRuleRepository) *ClassificationRuleService {
	return &ClassificationRuleService{repo: repo}
}

// CreateRule creates a new classification rule.
func (s *ClassificationRuleService) CreateRule(ctx context.Context, tenantID uuid.UUID, req *model.CreateRuleRequest) (*model.ClassificationRule, error) {
	// Validate request
	if req.Name == "" {
		return nil, errors.New("rule name is required")
	}
	if req.Pattern == "" {
		return nil, errors.New("pattern is required")
	}
	if req.Classification == "" {
		return nil, errors.New("classification is required")
	}

	// Auto-assign priority if not specified
	if req.Priority <= 0 {
		maxPriority, err := s.repo.GetMaxPriority(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("get max priority: %w", err)
		}
		req.Priority = maxPriority + 1
	}

	// Set defaults
	if req.RuleType == "" {
		req.RuleType = "keyword"
	}
	if req.MatchField == "" {
		req.MatchField = "description"
	}

	rule := &model.ClassificationRule{
		Name:            req.Name,
		RuleType:        req.RuleType,
		Pattern:         req.Pattern,
		MatchField:      req.MatchField,
		Direction:       req.Direction,
		Classification:  req.Classification,
		Priority:        req.Priority,
		IsActive:        req.IsActive,
		DebitAccountID:  req.DebitAccountID,
		CreditAccountID: req.CreditAccountID,
	}

	return s.repo.Create(ctx, tenantID, rule)
}

// UpdateRule updates an existing classification rule.
func (s *ClassificationRuleService) UpdateRule(ctx context.Context, tenantID, id uuid.UUID, req *model.UpdateRuleRequest) error {
	rule, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.RuleType != nil {
		rule.RuleType = *req.RuleType
	}
	if req.Pattern != nil {
		rule.Pattern = *req.Pattern
	}
	if req.MatchField != nil {
		rule.MatchField = *req.MatchField
	}
	if req.Direction != nil {
		rule.Direction = *req.Direction
	}
	if req.Classification != nil {
		rule.Classification = *req.Classification
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.DebitAccountID != nil {
		rule.DebitAccountID = req.DebitAccountID
	}
	if req.CreditAccountID != nil {
		rule.CreditAccountID = req.CreditAccountID
	}

	return s.repo.Update(ctx, tenantID, id, rule)
}

// DeleteRule removes a classification rule.
func (s *ClassificationRuleService) DeleteRule(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// ListRules returns all classification rules for a tenant.
func (s *ClassificationRuleService) ListRules(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// ReorderPriority updates rule priorities based on ordered rule IDs.
func (s *ClassificationRuleService) ReorderPriority(ctx context.Context, tenantID uuid.UUID, ruleIDs []uuid.UUID) error {
	return s.repo.ReorderPriority(ctx, tenantID, ruleIDs)
}

// MatchTransaction finds the first matching rule for given transaction.
func (s *ClassificationRuleService) MatchTransaction(ctx context.Context, tenantID uuid.UUID, description, counterparty, direction string) (*model.RuleMatchResult, error) {
	rules, err := s.repo.ListActiveRules(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("match transaction: %w", err)
	}

	for _, rule := range rules {
		if s.repo.MatchRule(&rule, description, counterparty, direction) {
			return &model.RuleMatchResult{
				Matched:        true,
				RuleID:         &rule.ID,
				RuleName:       &rule.Name,
				Classification: &rule.Classification,
			}, nil
		}
	}

	return &model.RuleMatchResult{Matched: false}, nil
}

// GetRuleByID returns a single rule including its per-rule account mapping.
func (s *ClassificationRuleService) GetRuleByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ClassificationRule, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// SeedRules creates initial default rules for a tenant.
func (s *ClassificationRuleService) SeedRules(ctx context.Context, tenantID uuid.UUID) error {
	// Check if tenant already has rules
	existingRules, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("check existing rules: %w", err)
	}
	if len(existingRules) > 0 {
		return errors.New("tenant already has rules")
	}

	// Default rules
	defaultRules := []*model.CreateRuleRequest{
		{
			Name:           "银行手续费",
			RuleType:       "keyword_regex",
			Pattern:        "手续费|工本费|年费|账户管理费",
			MatchField:     "description",
			Direction:      "out",
			Classification: "bank_fee",
			Priority:       1,
			IsActive:       true,
		},
		{
			Name:           "利息收入",
			RuleType:       "keyword_regex",
			Pattern:        "利息|结息|存款利息",
			MatchField:     "description",
			Direction:      "in",
			Classification: "interest_income",
			Priority:       2,
			IsActive:       true,
		},
		{
			Name:           "业务收款",
			RuleType:       "keyword",
			Pattern:        "货款",
			MatchField:     "description",
			Direction:      "in",
			Classification: "business_receipt",
			Priority:       3,
			IsActive:       true,
		},
		{
			Name:           "业务付款",
			RuleType:       "keyword",
			Pattern:        "货款",
			MatchField:     "description",
			Direction:      "out",
			Classification: "business_payment",
			Priority:       4,
			IsActive:       true,
		},
		{
			Name:           "内部转账",
			RuleType:       "keyword_regex",
			Pattern:        "转账|转存|调拨|上划|下拨",
			MatchField:     "description",
			Direction:      "",
			Classification: "internal_transfer",
			Priority:       5,
			IsActive:       true,
		},
		{
			Name:           "税务缴费",
			RuleType:       "keyword_regex",
			Pattern:        "税|税务|缴税|税金|税款|增值税|所得税|城建税|教育费附加|国家金库|国库|印花",
			MatchField:     "description",
			Direction:      "out",
			Classification: "tax_payment",
			Priority:       6,
			IsActive:       true,
		},
		{
			Name:           "社保缴费",
			RuleType:       "keyword_regex",
			Pattern:        "社保|公积金|养老|医疗|失业|工伤|生育",
			MatchField:     "description",
			Direction:      "out",
			Classification: "social_security",
			Priority:       7,
			IsActive:       true,
		},
		{
			Name:           "保险费用",
			RuleType:       "keyword_regex",
			Pattern:        "保险|保费|投保|财产险|责任险|雇主责任险|意外险",
			MatchField:     "description",
			Direction:      "out",
			Classification: "insurance_fee",
			Priority:       8,
			IsActive:       true,
		},
	}

	for i, ruleReq := range defaultRules {
		ruleReq.Priority = i + 1
		_, err := s.CreateRule(ctx, tenantID, ruleReq)
		if err != nil {
			return fmt.Errorf("create rule %d: %w", i+1, err)
		}
	}

	return nil
}
