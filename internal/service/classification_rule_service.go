package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ClassificationRuleService handles classification rule business logic.
type ClassificationRuleService struct {
	repo        *repository.ClassificationRuleRepository
	accountRepo *repository.AccountRepository
}

// NewClassificationRuleService creates a new ClassificationRuleService.
func NewClassificationRuleService(repo *repository.ClassificationRuleRepository, accountRepo *repository.AccountRepository) *ClassificationRuleService {
	return &ClassificationRuleService{repo: repo, accountRepo: accountRepo}
}

// MatchTransaction finds the first matching rule for given keywords and amount.
// Returns the matched rule details or nil if no match.
func (s *ClassificationRuleService) MatchTransaction(ctx context.Context, tenantID uuid.UUID, keywords string, amount decimal.Decimal, direction string) (*model.RuleMatchResult, error) {
	rules, err := s.repo.FindMatchBatch(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("match transaction: %w", err)
	}

	for _, rule := range rules {
		// Direction filter
		if rule.DebitDirection != nil && *rule.DebitDirection != "both" && *rule.DebitDirection != direction {
			continue
		}

		// Keywords OR matching
		var kwArray []string
		if err := json.Unmarshal(rule.Keywords, &kwArray); err != nil {
			continue
		}

		matched := false
		for _, kw := range kwArray {
			if kw != "" && containsSubstring(keywords, kw) {
				matched = true
				break
			}
		}
		if matched {
			account, _ := s.accountRepo.GetByID(ctx, tenantID, rule.AccountID)
			result := &model.RuleMatchResult{
				Matched:  true,
				RuleID:   &rule.ID,
				RuleName: &rule.RuleName,
				Priority: &rule.Priority,
			}
			if account != nil {
				result.AccountID = &rule.AccountID
				result.AccountCode = &account.Code
				result.AccountName = &account.Name
			}
			if rule.PartyType != nil {
				result.PartyType = rule.PartyType
			}
			return result, nil
		}
	}

	return &model.RuleMatchResult{Matched: false}, nil
}

// CreateRule creates a new classification rule with auto-assigned priority.
func (s *ClassificationRuleService) CreateRule(ctx context.Context, tenantID uuid.UUID, req *model.CreateRuleRequest) (*model.ClassificationRule, error) {
	// Validate account exists and type matches
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errors.New("invalid account_id format")
	}
	if err := s.ValidateAccount(ctx, tenantID, accountID); err != nil {
		return nil, err
	}

	// Validate keywords
	if len(req.Keywords) == 0 {
		return nil, errors.New("keywords cannot be empty")
	}

	// Auto-assign priority
	maxPriority, err := s.repo.GetMaxPriority(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get max priority: %w", err)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	debitDirection := "both"
	if req.DebitDirection != nil {
		debitDirection = *req.DebitDirection
	}

	keywordsJSON, err := json.Marshal(req.Keywords)
	if err != nil {
		return nil, fmt.Errorf("marshal keywords: %w", err)
	}

	rule := &model.ClassificationRule{
		RuleName:       req.RuleName,
		Priority:       maxPriority + 1,
		Keywords:       keywordsJSON,
		AccountID:      accountID,
		PartyType:      req.PartyType,
		DebitDirection: &debitDirection,
		IsActive:       isActive,
	}

	return s.repo.Create(ctx, tenantID, rule)
}

// UpdateRule updates an existing classification rule.
func (s *ClassificationRuleService) UpdateRule(ctx context.Context, tenantID, id uuid.UUID, req *model.UpdateRuleRequest) error {
	rule, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	if req.RuleName != nil {
		rule.RuleName = *req.RuleName
	}
	if req.Keywords != nil {
		keywordsJSON, err := json.Marshal(req.Keywords)
		if err != nil {
			return fmt.Errorf("marshal keywords: %w", err)
		}
		rule.Keywords = keywordsJSON
	}
	if req.AccountID != nil {
		accountID, err := uuid.Parse(*req.AccountID)
		if err != nil {
			return errors.New("invalid account_id format")
		}
		if err := s.ValidateAccount(ctx, tenantID, accountID); err != nil {
			return err
		}
		rule.AccountID = accountID
	}
	if req.PartyType != nil {
		rule.PartyType = req.PartyType
	}
	if req.DebitDirection != nil {
		rule.DebitDirection = req.DebitDirection
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	return s.repo.Update(ctx, tenantID, id, rule)
}

// DeleteRule soft-deletes a classification rule.
func (s *ClassificationRuleService) DeleteRule(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// ListRules returns all classification rules for a tenant.
func (s *ClassificationRuleService) ListRules(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// ReorderPriority reorders rule priorities based on the provided ordered list.
func (s *ClassificationRuleService) ReorderPriority(ctx context.Context, tenantID uuid.UUID, ruleIDs []uuid.UUID) error {
	return s.repo.ReorderPriority(ctx, tenantID, ruleIDs)
}

// ValidateAccount validates that the account exists and has an appropriate type for classification rules.
func (s *ClassificationRuleService) ValidateAccount(ctx context.Context, tenantID, accountID uuid.UUID) error {
	account, err := s.accountRepo.GetByID(ctx, tenantID, accountID)
	if err != nil {
		return errors.New("account_id not found")
	}
	// Accept any account type for classification rules (service level decides appropriateness)
	if account.AccountType == nil {
		return errors.New("account must have a type")
	}
	return nil
}

// ValidateRule performs full rule validation.
func (s *ClassificationRuleService) ValidateRule(ctx context.Context, tenantID uuid.UUID, req *model.CreateRuleRequest) error {
	if req.RuleName == "" {
		return errors.New("rule_name is required")
	}
	if len(req.Keywords) == 0 {
		return errors.New("keywords cannot be empty")
	}
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return errors.New("invalid account_id format")
	}
	return s.ValidateAccount(ctx, tenantID, accountID)
}

// containsSubstring checks if keyword exists in text (case-insensitive).
func containsSubstring(text, keyword string) bool {
	if len(keyword) == 0 {
		return false
	}
	textLower := toLowerString(text)
	keywordLower := toLowerString(keyword)
	for i := 0; i <= len(textLower)-len(keywordLower); i++ {
		if textLower[i:i+len(keywordLower)] == keywordLower {
			return true
		}
	}
	return false
}

func toLowerString(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}