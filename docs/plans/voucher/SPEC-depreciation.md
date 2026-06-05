# SPEC: 固定资产折旧/无形资产摊销自动生成凭证

## 背景

`AssetDepreciationService.GenerateMonthlyDepreciation` 已实现月折旧凭证生成逻辑，创建 `JournalEntry`（借：折旧费用科目，贷：累计折旧科目）。

**缺口：**
1. 凭证直接 `DocStatus=1`（posted），违反"人是审核唯一主体"原则——应先生成草稿（`docstatus=0`），人审核后才过账
2. 缺少按期间触发折旧的 API（结账时自动检查是否有未生成凭证的折旧）
3. 无形资产摊销（intangible asset amortization）逻辑未实现

## 目标

1. **修正**：折旧生成的凭证 `DocStatus=0`（草稿），经人审核后才 `posted`
2. **新增 API**：`POST /api/v1/depreciation/generate?period_no=202506` — 手动触发月折旧凭证生成（草稿状态）
3. **无形资产**：复用现有 `AssetDepreciationService` 框架，新增摊销逻辑（与折旧并行，借：无形资产摊销费，贷：累计摊销）
4. **结账联动**：`PeriodService` 月末检查时，对未执行折旧/摊销的资产给出警告

## 改动范围

### 1. `AssetDepreciationService.GenerateMonthlyDepreciation` 修正 DocStatus

文件：`internal/service/asset_depreciation_service.go`

将：
```go
DocStatus: 1, // posted
```
改为：
```go
DocStatus: 0, // 草稿，等人审核
```

并更新注释说明"生成的凭证为草稿状态，人审核后通过 VoucherStateMachine 过账"。

### 2. `AssetDepreciationHandler` 新增 `GenerateDepreciation` 端点

文件：`internal/handler/asset_depreciation_handler.go`

```go
// GenerateDepreciation POST /api/v1/depreciation/generate?period_no=202506
// 生成指定期间的折旧凭证（草稿状态）。
// 人审核后在凭证列表点击"核准"过账。
func (h *AssetDepreciationHandler) GenerateDepreciation(c *fiber.Ctx) error {
    tenantID := c.Locals("tenant_id").(uuid.UUID)
    periodNo := c.QueryInt("period_no", 0)
    if periodNo == 0 {
        return c.Status(400).JSON(fiber.Map{"error": "period_no required"})
    }
    run, err := h.svc.GenerateMonthlyDepreciation(c.Context(), tenantID, periodNo)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(fiber.Map{"data": run})
}
```

### 3. `AssetDepreciationService` 新增 `GenerateMonthlyAmortization` 方法

文件：`internal/service/asset_depreciation_service.go`

在文件末尾添加：
```go
// GenerateMonthlyAmortization generates journal entries for intangible asset amortization.
// 借：6603 无形资产摊销费，贷：1702 累计摊销
// 复用 depreciation_run 记录，但 source_type='amortization' 以区分。
func (s *AssetDepreciationService) GenerateMonthlyAmortization(
    ctx context.Context,
    tenantID uuid.UUID,
    periodNo int,
) (*model.DepreciationRun, error) {
    // 逻辑同 GenerateMonthlyDepreciation，但：
    // - 查 intangible_assets 表（或 asset.asset_type='intangible'）
    // - 科目：借 6603，贷 1702
    // - voucher_type = "Amortization"
}
```

### 4. `AssetDepreciationHandler` 新增 `GenerateAmortization` 端点

```go
// GenerateAmortization POST /api/v1/depreciation/generate-amortization?period_no=202506
func (h *AssetDepreciationHandler) GenerateAmortization(c *fiber.Ctx) error {
    tenantID := c.Locals("tenant_id").(uuid.UUID)
    periodNo := c.QueryInt("period_no", 0)
    run, err := h.svc.GenerateMonthlyAmortization(c.Context(), tenantID, periodNo)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(fiber.Map{"data": run})
}
```

### 5. 路由注册

`cmd/api/main.go` 确保以下路由已注册：
- `POST /api/v1/depreciation/generate`
- `POST /api/v1/depreciation/generate-amortization`

## 验证

1. `go build ./...` 通过
2. 调用 `POST /depreciation/generate?period_no=202506` → 返回凭证草稿（docstatus=0）
3. 在凭证列表找到该凭证 → 人点击核准 → docstatus 变为 1（posted）
4. 调用 `POST /depreciation/generate-amortization?period_no=202506` → 无形资产摊销凭证草稿