package service

import (
	"context"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	poolOnce sync.Once
	pool     *pgxpool.Pool
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	poolOnce.Do(func() {
		var err error
		pool, err = pgxpool.New(context.Background(), "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable")
		if err != nil {
			t.Fatalf("connect to test db: %v", err)
		}
	})
	return pool
}
