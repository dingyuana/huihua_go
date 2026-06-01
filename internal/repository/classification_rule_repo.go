package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// ClassificationRuleRepository handles classification_rules table operations.
type ClassificationRuleRepository struct {
	pool *pgxpool.Pool
}

// NewClassificationRuleRepository creates a new ClassificationRuleRepository.
func NewClassificationRuleRepository(pool *pgxpool.Pool) *ClassificationRuleRepository {
	return &ClassificationRuleRepository{pool: pool}
}

// Create inserts a new classification rule.
func (r *ClassificationRuleRepository) Create(ctx context.Context, tenantID uuid.UUID, rule *model.ClassificationRule) (*model.ClassificationRule, error) {
	rule.ID = uuid.New()
	rule.TenantID = tenantID

	query := `
		INSERT INTO classification_rules (id, tenant_id, name, rule_type, pattern, match_field, direction, classification, priority, is_active, debit_account_id, credit_account_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		rule.ID, tenantID, rule.Name, rule.RuleType, rule.Pattern,
		rule.MatchField, rule.Direction, rule.Classification,
		rule.Priority, rule.IsActive, rule.DebitAccountID, rule.CreditAccountID,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create classification rule: %w", err)
	}
	return rule, nil
}

// ListByTenant retrieves all classification rules for a tenant ordered by priority ASC.
func (r *ClassificationRuleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, name, rule_type, pattern, match_field, direction, classification, priority, is_active, debit_account_id, credit_account_id, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1
		ORDER BY priority ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list classification rules: %w", err)
	}
	defer rows.Close()

	var rules []model.ClassificationRule
	for rows.Next() {
		var rule model.ClassificationRule
		if err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.RuleType, &rule.Pattern,
			&rule.MatchField, &rule.Direction, &rule.Classification, &rule.Priority,
			&rule.IsActive, &rule.DebitAccountID, &rule.CreditAccountID,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan classification rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// GetByID retrieves a classification rule by ID.
func (r *ClassificationRuleRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, name, rule_type, pattern, match_field, direction, classification, priority, is_active, debit_account_id, credit_account_id, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1 AND id = $2`

	rule := &model.ClassificationRule{}
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&rule.ID, &rule.TenantID, &rule.Name, &rule.RuleType, &rule.Pattern,
		&rule.MatchField, &rule.Direction, &rule.Classification, &rule.Priority,
		&rule.IsActive, &rule.DebitAccountID, &rule.CreditAccountID,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get classification rule by id: %w", err)
	}
	return rule, nil
}

// Update updates a classification rule.
func (r *ClassificationRuleRepository) Update(ctx context.Context, tenantID, id uuid.UUID, rule *model.ClassificationRule) error {
	query := `
		UPDATE classification_rules
		SET name = $3, rule_type = $4, pattern = $5, match_field = $6, direction = $7, classification = $8, priority = $9, is_active = $10, debit_account_id = $11, credit_account_id = $12, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	_, err := r.pool.Exec(ctx, query,
		tenantID, id, rule.Name, rule.RuleType, rule.Pattern,
		rule.MatchField, rule.Direction, rule.Classification, rule.Priority, rule.IsActive,
		rule.DebitAccountID, rule.CreditAccountID,
	)
	if err != nil {
		return fmt.Errorf("update classification rule: %w", err)
	}
	return nil
}

// Delete removes a classification rule.
func (r *ClassificationRuleRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM classification_rules WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("delete classification rule: %w", err)
	}
	return nil
}

// ReorderPriority updates priorities based on ordered rule IDs.
func (r *ClassificationRuleRepository) ReorderPriority(ctx context.Context, tenantID uuid.UUID, ruleIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, id := range ruleIDs {
		priority := i + 1
		_, err := tx.Exec(ctx, `
			UPDATE classification_rules SET priority = $3, updated_at = NOW()
			WHERE tenant_id = $1 AND id = $2`,
			tenantID, id, priority)
		if err != nil {
			return fmt.Errorf("reorder priority: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit reorder: %w", err)
	}
	return nil
}

// GetMaxPriority returns the current maximum priority for a tenant.
func (r *ClassificationRuleRepository) GetMaxPriority(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(priority), 0) FROM classification_rules WHERE tenant_id = $1`
	var maxPriority int
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&maxPriority)
	if err != nil {
		return 0, fmt.Errorf("get max priority: %w", err)
	}
	return maxPriority, nil
}

// ListActiveRules retrieves all active rules ordered by priority.
func (r *ClassificationRuleRepository) ListActiveRules(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, name, rule_type, pattern, match_field, direction, classification, priority, is_active, debit_account_id, credit_account_id, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1 AND is_active = TRUE
		ORDER BY priority ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list active rules: %w", err)
	}
	defer rows.Close()

	var rules []model.ClassificationRule
	for rows.Next() {
		var rule model.ClassificationRule
		if err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.RuleType, &rule.Pattern,
			&rule.MatchField, &rule.Direction, &rule.Classification, &rule.Priority,
			&rule.IsActive, &rule.DebitAccountID, &rule.CreditAccountID,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// MatchRule tests if a rule matches given criteria.
func (r *ClassificationRuleRepository) MatchRule(rule *model.ClassificationRule, description, counterparty, direction string) bool {
	// Check direction filter if specified
	if rule.Direction != "" && rule.Direction != direction {
		return false
	}

	// Get the text to match based on match_field
	var matchText string
	if rule.MatchField == "counterparty" {
		matchText = counterparty
	} else {
		matchText = description
	}

	// Perform matching based on rule_type
	switch rule.RuleType {
	case "keyword_regex":
		return regexMatch(matchText, rule.Pattern)
	case "counterparty_match":
		return containsString(matchText, rule.Pattern)
	case "keyword":
		fallthrough
	default:
		return containsString(matchText, rule.Pattern)
	}
}

func containsString(text, pattern string) bool {
	if text == "" || pattern == "" {
		return false
	}
	textLower := strings.ToLower(text)
	patternLower := strings.ToLower(pattern)
	return strings.Contains(textLower, patternLower)
}

func regexMatch(text, pattern string) bool {
	if text == "" || pattern == "" {
		return false
	}
	// Simple implementation - split by | and check any match
	patterns := strings.Split(pattern, "|")
	for _, p := range patterns {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if containsString(text, strings.TrimSpace(p)) {
			return true
		}
	}
	return false
}
