package repository

import (
	"context"

	"huihua/finance/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, tenant_id, username, password_hash, role, is_active, created_at, updated_at
	          FROM users WHERE username = $1`
	var u model.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.TenantID, &u.Username, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, tenant_id, username, password_hash, role, is_active, created_at, updated_at
	          FROM users WHERE id = $1`
	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Username, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
