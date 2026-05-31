# Huihua Finance - Test & Build commands

.PHONY: test test-go test-go-unit test-go-integration test-frontend test-frontend-unit build build-go build-frontend

# ─── Full test suite ───

test: test-go test-frontend-unit

# ─── Backend (Go) ───

test-go:
	go test ./... -timeout 120s -count=1

test-go-unit:
	go test ./internal/middleware/ ./pkg/... -timeout 30s -count=1

test-go-integration:
	go test ./internal/repository/ ./internal/service/ ./internal/handler/ -timeout 120s -count=1

# ─── Frontend ───

test-frontend-unit:
	cd frontend && npx vitest run

test-frontend-e2e:
	cd frontend && npx playwright test

# ─── Build ───

build: build-go build-frontend

build-go:
	go build ./...

build-frontend:
	cd frontend && npx vue-tsc --noEmit && npx vite build
