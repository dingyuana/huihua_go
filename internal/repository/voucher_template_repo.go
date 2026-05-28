package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
)

// VoucherTemplateRepository provides data access for voucher templates.
type VoucherTemplateRepository struct {
	pool *pgxpool.Pool
}

// NewVoucherTemplateRepository creates a new VoucherTemplateRepository.
func NewVoucherTemplateRepository(pool *pgxpool.Pool) *VoucherTemplateRepository {
	return &VoucherTemplateRepository{pool: pool}
}

// CreateTemplate creates a new voucher template with its lines.
func (r *VoucherTemplateRepository) CreateTemplate(ctx context.Context, tenantID uuid.UUID, t *model.VoucherTemplate) (*model.VoucherTemplate, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now()

	// Insert template
	_, err = tx.Exec(ctx, `
		INSERT INTO voucher_templates (id, tenant_id, name, description, number_prefix, is_active, approval_flow_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		t.ID, tenantID, t.Name, t.Description, t.NumberPrefix, t.IsActive, t.ApprovalFlowID, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert template: %w", err)
	}

	// Insert lines
	for i := range t.Lines {
		if t.Lines[i].ID == uuid.Nil {
			t.Lines[i].ID = uuid.New()
		}
		t.Lines[i].TemplateID = t.ID
		t.Lines[i].CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO voucher_template_lines (id, template_id, account_id, dr_amount_template, cr_amount_template, summary_template, dimension_type, dimension_value, line_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			t.Lines[i].ID, t.Lines[i].TemplateID, t.Lines[i].AccountID,
			t.Lines[i].DrAmountTemplate, t.Lines[i].CrAmountTemplate,
			t.Lines[i].SummaryTemplate, t.Lines[i].DimensionType,
			t.Lines[i].DimensionValue, t.Lines[i].LineOrder, t.Lines[i].CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert template line: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	t.TenantID = tenantID
	t.CreatedAt = now
	t.UpdatedAt = now
	return t, nil
}

// ListTemplates returns all voucher templates for a tenant.
func (r *VoucherTemplateRepository) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]model.VoucherTemplate, error) {
	query := `
		SELECT id, tenant_id, name, description, number_prefix, is_active, approval_flow_id, created_at, updated_at
		FROM voucher_templates
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var templates []model.VoucherTemplate
	for rows.Next() {
		var t model.VoucherTemplate
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Description, &t.NumberPrefix, &t.IsActive, &t.ApprovalFlowID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// GetTemplateByID returns a voucher template with its lines.
func (r *VoucherTemplateRepository) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*model.VoucherTemplate, error) {
	query := `
		SELECT id, tenant_id, name, description, number_prefix, is_active, approval_flow_id, created_at, updated_at
		FROM voucher_templates
		WHERE id = $1 AND tenant_id = $2`

	t := &model.VoucherTemplate{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.NumberPrefix, &t.IsActive, &t.ApprovalFlowID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get template by id: %w", err)
	}

	// Get lines
	lineQuery := `
		SELECT l.id, l.template_id, l.account_id, a.code, a.name,
		       l.dr_amount_template, l.cr_amount_template, l.summary_template,
		       l.dimension_type, l.dimension_value, l.line_order, l.created_at
		FROM voucher_template_lines l
		JOIN accounts a ON a.id = l.account_id
		WHERE l.template_id = $1
		ORDER BY l.line_order`

	rows, err := r.pool.Query(ctx, lineQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get template lines: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var line model.VoucherTemplateLine
		if err := rows.Scan(&line.ID, &line.TemplateID, &line.AccountID, &line.AccountCode, &line.AccountName,
			&line.DrAmountTemplate, &line.CrAmountTemplate, &line.SummaryTemplate,
			&line.DimensionType, &line.DimensionValue, &line.LineOrder, &line.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan template line: %w", err)
		}
		t.Lines = append(t.Lines, line)
	}

	return t, rows.Err()
}

// UpdateTemplate updates a voucher template and replaces its lines.
func (r *VoucherTemplateRepository) UpdateTemplate(ctx context.Context, tenantID, id uuid.UUID, t *model.VoucherTemplate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	// Update template
	_, err = tx.Exec(ctx, `
		UPDATE voucher_templates 
		SET name = $1, description = $2, number_prefix = $3, is_active = $4, approval_flow_id = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		t.Name, t.Description, t.NumberPrefix, t.IsActive, t.ApprovalFlowID, now, id, tenantID)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	// Delete existing lines
	_, err = tx.Exec(ctx, `DELETE FROM voucher_template_lines WHERE template_id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete old lines: %w", err)
	}

	// Insert new lines
	for i := range t.Lines {
		if t.Lines[i].ID == uuid.Nil {
			t.Lines[i].ID = uuid.New()
		}
		t.Lines[i].TemplateID = id
		t.Lines[i].CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO voucher_template_lines (id, template_id, account_id, dr_amount_template, cr_amount_template, summary_template, dimension_type, dimension_value, line_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			t.Lines[i].ID, t.Lines[i].TemplateID, t.Lines[i].AccountID,
			t.Lines[i].DrAmountTemplate, t.Lines[i].CrAmountTemplate,
			t.Lines[i].SummaryTemplate, t.Lines[i].DimensionType,
			t.Lines[i].DimensionValue, t.Lines[i].LineOrder, t.Lines[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("insert template line: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// DeleteTemplate performs a soft delete on a voucher template.
func (r *VoucherTemplateRepository) DeleteTemplate(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE voucher_templates SET is_active = FALSE, updated_at = $1
		WHERE id = $2 AND tenant_id = $3`,
		time.Now(), id, tenantID)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

// GetNumberingRule returns the numbering rule for a tenant.
func (r *VoucherTemplateRepository) GetNumberingRule(ctx context.Context, tenantID uuid.UUID) (*model.VoucherNumberingRule, error) {
	query := `
		SELECT id, tenant_id, prefix, next_number, date_format, reset_rule, created_at, updated_at
		FROM voucher_numbering_rules
		WHERE tenant_id = $1`

	rule := &model.VoucherNumberingRule{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&rule.ID, &rule.TenantID, &rule.Prefix, &rule.NextNumber,
		&rule.DateFormat, &rule.ResetRule, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get numbering rule: %w", err)
	}
	return rule, nil
}

// CreateOrUpdateNumberingRule creates or updates a numbering rule.
func (r *VoucherTemplateRepository) CreateOrUpdateNumberingRule(ctx context.Context, tenantID uuid.UUID, req *model.NumberingRuleRequest) (*model.VoucherNumberingRule, error) {
	now := time.Now()
	rule := &model.VoucherNumberingRule{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Prefix:     req.Prefix,
		NextNumber: 1,
		DateFormat: req.DateFormat,
		ResetRule:  req.ResetRule,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if rule.DateFormat == "" {
		rule.DateFormat = "20060102"
	}
	if rule.ResetRule == "" {
		rule.ResetRule = "daily"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO voucher_numbering_rules (id, tenant_id, prefix, next_number, date_format, reset_rule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id) DO UPDATE SET
			prefix = EXCLUDED.prefix,
			date_format = EXCLUDED.date_format,
			reset_rule = EXCLUDED.reset_rule,
			updated_at = EXCLUDED.updated_at
		RETURNING id, next_number, created_at`,
		rule.ID, tenantID, rule.Prefix, rule.NextNumber, rule.DateFormat, rule.ResetRule, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert numbering rule: %w", err)
	}

	// Fetch updated rule
	return r.GetNumberingRule(ctx, tenantID)
}

// GenerateNextNumber increments and returns the next sequence number.
func (r *VoucherTemplateRepository) GenerateNextNumber(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var nextNum int
	err := r.pool.QueryRow(ctx, `
		UPDATE voucher_numbering_rules 
		SET next_number = next_number + 1, updated_at = NOW()
		WHERE tenant_id = $1
		RETURNING next_number`,
		tenantID).Scan(&nextNum)
	if err != nil {
		return 0, fmt.Errorf("generate next number: %w", err)
	}
	return nextNum, nil
}
