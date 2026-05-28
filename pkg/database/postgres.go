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

func (db *DB) SetTenant(tenantID uuid.UUID) error {
	_, err := db.pool.Exec(context.Background(),
		fmt.Sprintf("SET app.current_tenant = '%s'", tenantID))
	return err
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