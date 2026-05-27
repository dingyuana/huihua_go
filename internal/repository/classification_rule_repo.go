package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

	keywordsJSON, err := json.Marshal(rule.Keywords)
	if err != nil {
		return nil, fmt.Errorf("marshal keywords: %w", err)
	}

	query := `
		INSERT INTO classification_rules (id, tenant_id, rule_name, priority, keywords, account_id, party_type, debit_direction, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err = r.pool.QueryRow(ctx, query,
		rule.ID, tenantID, rule.RuleName, rule.Priority, keywordsJSON, rule.AccountID,
		rule.PartyType, rule.DebitDirection, rule.IsActive,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create classification rule: %w", err)
	}
	return rule, nil
}

// ListByTenant retrieves all classification rules for a tenant ordered by priority ASC.
func (r *ClassificationRuleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, rule_name, priority, keywords, account_id, party_type, debit_direction, is_active, created_at, updated_at
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
			&rule.ID, &rule.TenantID, &rule.RuleName, &rule.Priority, &rule.Keywords,
			&rule.AccountID, &rule.PartyType, &rule.DebitDirection, &rule.IsActive,
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
		SELECT id, tenant_id, rule_name, priority, keywords, account_id, party_type, debit_direction, is_active, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1 AND id = $2`

	rule := &model.ClassificationRule{}
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&rule.ID, &rule.TenantID, &rule.RuleName, &rule.Priority, &rule.Keywords,
		&rule.AccountID, &rule.PartyType, &rule.DebitDirection, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get classification rule by id: %w", err)
	}
	return rule, nil
}

// Update updates a classification rule.
func (r *ClassificationRuleRepository) Update(ctx context.Context, tenantID, id uuid.UUID, rule *model.ClassificationRule) error {
	var keywordsJSON []byte
	var err error
	if rule.Keywords != nil {
		keywordsJSON, err = json.Marshal(rule.Keywords)
		if err != nil {
			return fmt.Errorf("marshal keywords: %w", err)
		}
	}

	query := `
		UPDATE classification_rules
		SET rule_name = $3, priority = $4, keywords = $5, account_id = $6, party_type = $7, debit_direction = $8, is_active = $9, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`

	_, err = r.pool.Exec(ctx, query,
		tenantID, id, rule.RuleName, rule.Priority, keywordsJSON, rule.AccountID,
		rule.PartyType, rule.DebitDirection, rule.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update classification rule: %w", err)
	}
	return nil
}

// Delete soft-deletes a classification rule by setting is_active = false.
func (r *ClassificationRuleRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `UPDATE classification_rules SET is_active = FALSE, updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, id)
	return err
}

// FindMatch finds the first matching rule for a given keywords text and amount.
// keywords: the transaction description text to match against
// amount: the transaction amount
// direction: "debit" or "credit"
func (r *ClassificationRuleRepository) FindMatch(ctx context.Context, tenantID uuid.UUID, keywords string, amount decimal.Decimal, direction string) (*model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, rule_name, priority, keywords, account_id, party_type, debit_direction, is_active, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1 AND is_active = TRUE
		ORDER BY priority ASC
		LIMIT 1`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("find match: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rule model.ClassificationRule
		if err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.RuleName, &rule.Priority, &rule.Keywords,
			&rule.AccountID, &rule.PartyType, &rule.DebitDirection, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rule for match: %w", err)
		}

		// Check direction filter
		if rule.DebitDirection != nil && *rule.DebitDirection != "both" && *rule.DebitDirection != direction {
			continue
		}

		// Parse keywords JSON array and check OR matching
		var kwArray []string
		if err := json.Unmarshal(rule.Keywords, &kwArray); err != nil {
			continue // skip invalid JSON
		}

		matched := false
		for _, kw := range kwArray {
			if kw != "" && containsKeyword(keywords, kw) {
				matched = true
				break
			}
		}
		if matched {
			return &rule, nil
		}
	}

	return nil, nil // no match found
}

// FindMatchBatch returns all rules ordered by priority for a tenant (caller filters in service).
func (r *ClassificationRuleRepository) FindMatchBatch(ctx context.Context, tenantID uuid.UUID) ([]model.ClassificationRule, error) {
	query := `
		SELECT id, tenant_id, rule_name, priority, keywords, account_id, party_type, debit_direction, is_active, created_at, updated_at
		FROM classification_rules
		WHERE tenant_id = $1 AND is_active = TRUE
		ORDER BY priority ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("find match batch: %w", err)
	}
	defer rows.Close()

	var rules []model.ClassificationRule
	for rows.Next() {
		var rule model.ClassificationRule
		if err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.RuleName, &rule.Priority, &rule.Keywords,
			&rule.AccountID, &rule.PartyType, &rule.DebitDirection, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
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

// containsKeyword checks if keyword is contained in text (case-insensitive substring match).
func containsKeyword(text, keyword string) bool {
	if len(keyword) == 0 {
		return false
	}
	// Simple substring match (case-insensitive)
	textLower := toLower(text)
	keywordLower := toLower(keyword)
	for i := 0; i <= len(textLower)-len(keywordLower); i++ {
		if textLower[i:i+len(keywordLower)] == keywordLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
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