# SPEC: D-P0.2 — ArInvoice Repository

## 基本信息
- **任务 ID**: D-P0.2
- **类型**: feature
- **优先级**: P0
- **依赖**: D-P0.1
- **负责 Profile**: dev

## 背景
需为 ArInvoice 提供完整的 DB 访问层。Repository 模式与项目中其他 repo 保持一致（pgx/v5）。

## 目标
创建 `internal/repository/ar_invoice_repo.go`，实现以下方法：

```go
// Create — 单条插入
Create(ctx context.Context, ar *model.ArInvoice) error

// GetByID — 按 ID 查询
GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.ArInvoice, error)

// ListByTenant — 按租户查询，支持 status 过滤
ListByTenant(ctx context.Context, tenantID uuid.UUID, status *string) ([]*model.ArInvoice, error)

// ListByCustomer — 按客户查询
ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, status *string) ([]*model.ArInvoice, error)

// ListByInvoiceID — 按发票查应收单（防重复生成）
ListByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*model.ArInvoice, error)

// UpdateStatus — 更新状态
UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error

// Confirm — 确认应收单（更新 confirmedAt/By + status=confirmed）
Confirm(ctx context.Context, tenantID, id, confirmedBy uuid.UUID) error

// BatchCreate — 批量插入
BatchCreate(ctx context.Context, ars []*model.ArInvoice) error
```

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 所有方法签名符合上述定义
- [ ] 使用 `*pgxpool.Pool` 作为构造参数（与其他 repo 一致）
- [ ] `NewArInvoiceRepository(pool *pgxpool.Pool) *ArInvoiceRepository` 构造签名

## 技术约束
- 参照 `invoice_repo.go` 的代码风格
- `decimal.Decimal` 值用 `.CoefficientInt64()` 不要用 `.Int64()`
- 多租户：所有查询必须带 `tenant_id` 条件
- 草稿状态过滤：`WHERE status = 'draft'`（传入 nil 时不过滤）

## OpenCode 指令模板
**目标**：创建 ArInvoice Repository

**约束**：
- 文件路径：`internal/repository/ar_invoice_repo.go`
- 参照 `invoice_repo.go` 的命名和结构风格
- 多租户查询必须带 tenant_id

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- model：`internal/model/ar_invoice.go`（已由 D-P0.1 创建）

**验收**：
- `go build ./...` 无报错
- 8个方法全部实现（非 stub）