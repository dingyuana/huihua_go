package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"huihua/finance/internal/config"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewPostgres(cfg *config.Config) (*DB, error) {
	host := os.Getenv("HF_DATABASE_HOST")
	if host == "" {
		host = cfg.Database.Host
	}
	port := os.Getenv("HF_DATABASE_PORT")
	if port == "" {
		port = cfg.Database.Port
	}
	user := os.Getenv("HF_DATABASE_USER")
	if user == "" {
		user = cfg.Database.User
	}
	password := os.Getenv("HF_DATABASE_PASSWORD")
	if password == "" {
		password = cfg.Database.Password
	}
	dbname := os.Getenv("HF_DATABASE_DBNAME")
	if dbname == "" {
		dbname = cfg.Database.DBName
	}
	sslmode := os.Getenv("HF_DATABASE_SSLMODE")
	if sslmode == "" {
		sslmode = cfg.Database.SSLMode
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)
	log.Printf("[DEBUG] DSN: postgres://%s:***@%s:%s/%s?sslmode=%s", user, host, port, dbname, sslmode)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{pool: pool}, nil
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const tenantCtxKey contextKey = "tenant_id"

// ContextWithTenant wraps a context with tenant information.
// Repository methods should use pgxpool.Acquire + SET app.current_tenant
// or pass tenant via this context.
func ContextWithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantCtxKey, tenantID)
}

// TenantFromContext extracts tenant ID from a context.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantCtxKey).(uuid.UUID)
	return id, ok
}

func (db *DB) SetTenant(tenantID uuid.UUID) error {
	query := fmt.Sprintf("SET app.current_tenant = '%s'", tenantID)
	_, err := db.pool.Exec(context.Background(), query)
	if err != nil {
		log.Printf("[WARN] SetTenant failed for %s: %v", tenantID, err)
		return fmt.Errorf("set tenant %s: %w", tenantID, err)
	}
	return nil
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Close() {
	db.pool.Close()
}

type RedisClient struct {
	client *redis.Client
}

func NewRedis(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	return &RedisClient{client: client}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}