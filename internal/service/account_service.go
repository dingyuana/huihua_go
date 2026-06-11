package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/pkg/utils"
)

// AccountService handles account operations with automatic code generation.
type AccountService struct {
	repo     *repository.AccountRepository
	seedPool *pgxpool.Pool
}

// NewAccountService creates a new AccountService.
func NewAccountService(repo *repository.AccountRepository, seedPool *pgxpool.Pool) *AccountService {
	return &AccountService{repo: repo, seedPool: seedPool}
}

// seedAccountRow is the in-memory shape used by both InitFromSeed and
// InitFromSeedWithTx. Materializing rows into a slice closes the cursor
// before any INSERT — required because pgx refuses concurrent use of the
// same pool while rows are still open (returns "conn busy").
type seedAccountRow struct {
	code, name, accountType, rootType string
	parentCode                        *string
	isGroup                           *bool
	lft, rgt                          *int
}

// fetchSeedAccounts reads all rows from standard_accounts_seed into memory.
func (s *AccountService) fetchSeedAccounts(ctx context.Context) ([]seedAccountRow, error) {
	rows, err := s.seedPool.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return nil, fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	var out []seedAccountRow
	for rows.Next() {
		var sd seedAccountRow
		if err := rows.Scan(&sd.code, &sd.name, &sd.accountType, &sd.rootType, &sd.parentCode, &sd.isGroup, &sd.lft, &sd.rgt); err != nil {
			return nil, fmt.Errorf("scan seed account: %w", err)
		}
		out = append(out, sd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seed accounts: %w", err)
	}
	return out, nil
}

// fetchSeedAccountsInTx is the transaction-bound variant. Use this when the
// caller is already inside a transaction (e.g. SetupService.CreateCompany) so
// the SELECT participates in the same RLS context as the subsequent INSERTs.
func (s *AccountService) fetchSeedAccountsInTx(ctx context.Context, tx pgx.Tx) ([]seedAccountRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT code, name, account_type, root_type, parent_code, is_group, lft, rgt
		FROM standard_accounts_seed ORDER BY lft`)
	if err != nil {
		return nil, fmt.Errorf("fetch seed accounts: %w", err)
	}
	defer rows.Close()

	var out []seedAccountRow
	for rows.Next() {
		var sd seedAccountRow
		if err := rows.Scan(&sd.code, &sd.name, &sd.accountType, &sd.rootType, &sd.parentCode, &sd.isGroup, &sd.lft, &sd.rgt); err != nil {
			return nil, fmt.Errorf("scan seed account: %w", err)
		}
		out = append(out, sd)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seed accounts: %w", err)
	}
	return out, nil
}

func (s *AccountService) buildAccountFromSeed(sd seedAccountRow, tenantID, companyID uuid.UUID, parentCodeMap map[string]uuid.UUID) *model.Account {
	ig := false
	if sd.isGroup != nil {
		ig = *sd.isGroup
	}
	lft, rgt := 0, 0
	if sd.lft != nil {
		lft = *sd.lft
	}
	if sd.rgt != nil {
		rgt = *sd.rgt
	}
	acc := &model.Account{
		ID:             uuid.New(),
		Code:           sd.code,
		Name:           sd.name,
		AccountType:    utils.StrPtr(sd.accountType),
		RootType:       utils.StrPtr(sd.rootType),
		IsGroup:        ig,
		Lft:            lft,
		Rgt:            rgt,
		CompanyID:      companyID,
		TenantID:       tenantID,
		Currency:       "CNY",
		IsActive:       true,
		OpeningBalance: decimal.Zero,
	}
	if sd.parentCode != nil && *sd.parentCode != "" {
		if pid, ok := parentCodeMap[*sd.parentCode]; ok {
			acc.ParentID = &pid
		}
	}
	return acc
}

// InitFromSeed initializes accounts from the standard_accounts_seed table.
// Creates a copy of all seed accounts for the given tenant/company.
// Skips accounts that already exist (same tenant_id + code).
func (s *AccountService) InitFromSeed(ctx context.Context, tenantID, companyID uuid.UUID) error {
	seeds, err := s.fetchSeedAccounts(ctx)
	if err != nil {
		return err
	}

	parentCodeMap := make(map[string]uuid.UUID)
	for _, sd := range seeds {
		acc := s.buildAccountFromSeed(sd, tenantID, companyID, parentCodeMap)
		parentCodeMap[sd.code] = acc.ID

		if _, err := s.seedPool.Exec(ctx, `
			INSERT INTO accounts (id, code, name, account_type, root_type, parent_id, lft, rgt, is_group, company_id, tenant_id, currency, is_active, opening_balance)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			ON CONFLICT (tenant_id, code) DO NOTHING`,
			acc.ID, acc.Code, acc.Name, acc.AccountType, acc.RootType, acc.ParentID,
			acc.Lft, acc.Rgt, acc.IsGroup, acc.CompanyID, acc.TenantID, acc.Currency,
			acc.IsActive, acc.OpeningBalance,
		); err != nil {
			return fmt.Errorf("insert account %s: %w", sd.code, err)
		}
	}
	return nil
}

// InitFromSeedWithTx initializes accounts from the standard_accounts_seed table within a transaction.
func (s *AccountService) InitFromSeedWithTx(ctx context.Context, tx pgx.Tx, tenantID, companyID uuid.UUID) error {
	// Materialize seed rows into a slice first so the rows cursor is closed
	// before we issue INSERTs. Required because pgx refuses concurrent use of
	// the same pool while rows are still open ("conn busy").
	seeds, err := s.fetchSeedAccountsInTx(ctx, tx)
	if err != nil {
		return err
	}

	parentCodeMap := make(map[string]uuid.UUID)
	for _, sd := range seeds {
		acc := s.buildAccountFromSeed(sd, tenantID, companyID, parentCodeMap)
		parentCodeMap[sd.code] = acc.ID

		if _, err := s.repo.CreateWithTx(ctx, tx, tenantID, acc); err != nil {
			return fmt.Errorf("create account %s: %w", sd.code, err)
		}
	}
	return nil
}

// GetTree returns accounts as a nested tree wrapped in a single synthetic root
// so the frontend can treat the root's children as level-1 options.
func (s *AccountService) GetTree(ctx context.Context, tenantID uuid.UUID) ([]model.Account, error) {
	flat, err := s.repo.GetTree(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*model.Account, len(flat))
	for i := range flat {
		byID[flat[i].ID] = &flat[i]
	}
	var roots []*model.Account
	for i := range flat {
		a := &flat[i]
		if a.ParentID == nil {
			roots = append(roots, a)
			continue
		}
		parent, ok := byID[*a.ParentID]
		if !ok {
			continue
		}
		parent.Children = append(parent.Children, a)
	}
	synthetic := model.Account{
		Name:     "全部科目",
		IsGroup:  true,
		Children: roots,
	}
	return []model.Account{synthetic}, nil
}

// List returns a paginated slice of accounts for the tenant.
// `code` empty lists all; non-empty filters by exact code match.
func (s *AccountService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int, code string) ([]model.Account, int, error) {
	return s.repo.ListPaginated(ctx, tenantID, limit, offset, code)
}

// Create creates a new account. If parentID is provided, the account is inserted
// as a child in the nested set. If parentID is nil, it becomes a root-level node.
func (s *AccountService) Create(ctx context.Context, tenantID uuid.UUID, a *model.Account) (*model.Account, error) {
	// Validate code uniqueness
	existing, err := s.repo.GetByCode(ctx, tenantID, a.Code)
	if err == nil && existing != nil && existing.ID != uuid.Nil {
		return nil, fmt.Errorf("科目编码 %s 已存在", a.Code)
	}

	a.ID = uuid.New()
	a.TenantID = tenantID
	a.IsActive = true

	// Calculate nested set values (lft, rgt)
	if a.ParentID != nil && *a.ParentID != uuid.Nil {
		parent, err := s.repo.GetByID(ctx, tenantID, *a.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父科目不存在: %w", err)
		}
		// Compute max sibling rgt inside parent
		maxSiblingRgt, err := s.repo.GetMaxSiblingRgt(ctx, tenantID, *a.ParentID)
		if err != nil {
			return nil, fmt.Errorf("获取同级最大 rgt: %w", err)
		}
		start := maxSiblingRgt
		if start <= parent.Lft {
			start = parent.Lft + 1
		}
		a.Lft = start
		a.Rgt = start + 1
		a.Level = parent.Level + 1
		a.Path = parent.Path
		if a.Path != "" {
			a.Path += "-" + a.Code
		} else {
			a.Path = a.Code
		}
	} else {
		// Root node
		maxRgt, err := s.repo.GetMaxRgt(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("获取最大 rgt: %w", err)
		}
		a.Lft = maxRgt + 1
		a.Rgt = maxRgt + 2
		a.Level = 0
		a.Path = a.Code
	}

	return s.repo.Create(ctx, tenantID, a)
}

// Update updates an account's editable fields.
func (s *AccountService) Update(ctx context.Context, tenantID uuid.UUID, a *model.Account) error {
	// Ensure account exists
	existing, err := s.repo.GetByID(ctx, tenantID, a.ID)
	if err != nil {
		return fmt.Errorf("科目不存在: %w", err)
	}
	// Preserve non-editable fields
	a.Code = existing.Code
	a.ParentID = existing.ParentID
	a.Lft = existing.Lft
	a.Rgt = existing.Rgt
	a.Level = existing.Level
	a.Path = existing.Path
	a.TenantID = existing.TenantID
	a.CompanyID = existing.CompanyID
	return s.repo.Update(ctx, a)
}

// Delete removes an account after checking it has no children.
func (s *AccountService) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	hasChildren, err := s.repo.HasChildren(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("检查子科目: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("该科目存在下级科目，无法删除")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// GetByID retrieves an account by ID.
func (s *AccountService) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.Account, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// AutoCode generates a suggested code for a new child of the given parent.
func (s *AccountService) AutoCode(ctx context.Context, tenantID uuid.UUID, parentID uuid.UUID) (string, error) {
	parent, err := s.repo.GetByID(ctx, tenantID, parentID)
	if err != nil {
		return "", fmt.Errorf("父科目不存在: %w", err)
	}
	maxRgt, err := s.repo.GetMaxSiblingRgt(ctx, tenantID, parentID)
	if err != nil {
		return "", fmt.Errorf("获取同级编码: %w", err)
	}
	// Get all siblings to compute the next sequence number
	// Simple heuristic: next sibling = parent.code + "-" + next_seq
	// We use maxRgt as indicator of presence — actually need to find the max code suffix
	// Better approach: query all children and find max suffix
	children, err := s.repo.ListByParent(ctx, tenantID, parentID)
	if err != nil {
		return "", fmt.Errorf("获取子科目列表: %w", err)
	}
	maxSeq := 0
	for _, child := range children {
		// Parse the last segment of the code after the parent's prefix
		childCode := child.Code
		if len(childCode) > len(parent.Code) && childCode[:len(parent.Code)] == parent.Code {
			// Try to parse the suffix as a number
			suffix := childCode[len(parent.Code):]
			// Remove any separator
			if len(suffix) > 0 && (suffix[0] == '-' || suffix[0] == '.') {
				suffix = suffix[1:]
			}
			// Try parsing as int
			var seq int
			n, _ := fmt.Sscanf(suffix, "%d", &seq)
			if n == 1 && seq > maxSeq {
				maxSeq = seq
			}
		}
	}
	nextSeq := maxSeq + 1
	suggested := fmt.Sprintf("%s-%02d", parent.Code, nextSeq)
	return suggested, nil
}
