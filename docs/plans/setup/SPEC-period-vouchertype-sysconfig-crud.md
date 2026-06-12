# SPEC: 期间管理 + 凭证类型 + 系统参数 CRUD（P1-3）

**版本**：V1.0
**日期**：2026-06-15
**关联**：TASK-P1-3-period-vouchertype-sysconfig-crud.md

---

## 背景

经过 Phase 0/1 已落地的功能审计，目前基础数据模块（科目/客商/银行账户/汇率）已有完整 CRUD，但还有 3 个基础设置类模块缺失完整 CRUD：

1. **会计期间（Period）**：`internal/{handler,service,repository}/period_*.go` 已存在，但只暴露 List / GetCurrent / Close / Unclose / PreviewClosing / ExecuteClosing / VoucherGaps / PreCloseCheck / CloseCheckSummary。缺少：创建/更新/删除/启用。无法手动新增期间（结账前置依赖），也无法纠错。

2. **凭证类型（VoucherType）**：模型 `model.VoucherType` + DTO 已定义（`internal/model/voucher_type.go`），迁移 `migrations/067_voucher_types.sql` 已存在，**但无任何 service/handler/repo 也无路由**。前端 `frontend/src/views/setup/VoucherTypeList.vue` 已存在但调用不到后端。

3. **系统参数（SysConfig）**：模型 `model.SysConfig` + DTO 已定义（同样在 `voucher_type.go` 里），**无 migration、无 service/handler/repo、无路由**。连数据库表都不存在。

3 个模块共同特征：典型"基础数据"CRUD，代码量小但前端依赖度高（前端页面已写好等接口）。

---

## 目标

完成 3 个基础设置模块的完整 CRUD 接口，使用 RLS 多租户隔离、统一 `fiber.Map` 响应格式（项目惯例，非 `R.ok()`）。

### 1. Period 补全 CRUD

**新增路由：**
- `POST /api/v1/periods` — 创建期间（管理员/初始化场景）
- `PUT /api/v1/periods/:id` — 更新期间元数据（period_name/日期范围）
- `DELETE /api/v1/periods/:id` — 删除期间（仅当无凭证数据引用时允许）
- `POST /api/v1/periods/:id/enable` — 把 `status=closed` 期间重新启用为 `open`（与 unclose 区别：unclose 是反审核结账流程，enable 是基础数据维护）

**约束：**
- `period_no` 全局唯一（同租户下唯一，依赖现有 `UNIQUE(tenant_id, period_no)`）
- 期间日期范围不可重叠：创建/更新前校验 `start_date, end_date` 不与同租户其他期间重叠
- `closed` 状态期间不可删除；`open` 状态但期间内已有 `journal_entries.posting_date` 落入时也禁止删除

### 2. VoucherType 完整 CRUD

**新增迁移**：`migrations/068_voucher_types_crud.sql`（如果 `067` 已有完整 schema 就跳过）

**新增 service/handler/repo：**
- `repository.VoucherTypeRepository`：Create / GetByID / List / Update / Delete（软删，set deleted_at）
- `service.VoucherTypeService`：包装 repo，加业务校验（code 唯一性、删除时若有 voucher 引用则禁止）
- `handler.VoucherTypeHandler`：List / Create / GetByID / Update / Delete

**路由：**
- `GET /api/v1/voucher-types?is_active=&limit=&offset=`
- `POST /api/v1/voucher-types`
- `GET /api/v1/voucher-types/:id`
- `PUT /api/v1/voucher-types/:id`
- `DELETE /api/v1/voucher-types/:id`

### 3. SysConfig 完整 CRUD

**新增迁移**：`migrations/069_sys_configs.sql`
```sql
CREATE TABLE IF NOT EXISTS sys_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL DEFAULT '',
    description VARCHAR(255) DEFAULT '',
    group_name VARCHAR(50) DEFAULT 'default',
    is_system BOOLEAN DEFAULT FALSE,        -- TRUE=系统内置不允许前端删除
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, config_key)
);
ALTER TABLE sys_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sys_configs USING (tenant_id::text = current_setting('app.current_tenant_id', TRUE));
```

**新增 service/handler/repo：**
- `repository.SysConfigRepository`：Create / GetByKey / GetByID / List / ListByKeys / Update / Delete（软删）
- `service.SysConfigService`：业务校验（is_system=true 不可改 key 不可删；config_value 按 key 做格式校验，如果 key 后缀是 `_json` 尝试 json.Valid）
- `handler.SysConfigHandler`：List / Create / GetByID / GetByKey / Update / Delete / BatchGet

**路由：**
- `GET /api/v1/configs?group=&is_system=&limit=&offset=`
- `POST /api/v1/configs`
- `GET /api/v1/configs/:id`
- `GET /api/v1/configs/key/:config_key`
- `POST /api/v1/configs/batch-get` — 批量获取，body `{keys: ["a","b"]}` → `{configs: {a:...,b:...}}`
- `PUT /api/v1/configs/:id`
- `DELETE /api/v1/configs/:id`

---

## 改动范围

### 1. 数据库迁移
- **新增** `migrations/068_sys_configs.sql` — 创建 sys_configs 表 + RLS + 索引

### 2. 后端 - Period 补全
- `internal/repository/period_repo.go` — 新增 `GetByID(ctx, tenantID, id)`、`UpdateMeta(ctx, tenantID, id, fields)`、`Delete(ctx, tenantID, id)`、`CheckOverlap(ctx, tenantID, start, end, excludeID)`
- `internal/service/period_service.go` — 新增 `CreatePeriod(ctx, tenantID, req)`、`UpdatePeriod(ctx, tenantID, id, req)`、`DeletePeriod(ctx, tenantID, id)`、`EnablePeriod(ctx, tenantID, id)`，全部调用 repo + 业务校验（overlap 校验、closed 期间不可删/不可更新）
- `internal/handler/period_handler.go` — 新增 `Create / Update / Delete / Enable` 4 个 method
- `cmd/api/main.go` — 注册新路由

### 3. 后端 - VoucherType 全套
- `internal/repository/voucher_type_repo.go`（新文件）
- `internal/service/voucher_type_service.go`（新文件）
- `internal/handler/voucher_type_handler.go`（新文件）
- `cmd/api/main.go` — 实例化 + 注册路由

### 4. 后端 - SysConfig 全套
- `internal/repository/sys_config_repo.go`（新文件）
- `internal/service/sys_config_service.go`（新文件）
- `internal/handler/sys_config_handler.go`（新文件）
- `cmd/api/main.go` — 实例化 + 注册路由

### 5. 测试
- `internal/handler/period_handler_test.go` — 追加 `TestPeriodHandler_Create/TestPeriodHandler_Update/TestPeriodHandler_Delete/TestPeriodHandler_Enable`（沿用现有 testHandlerPool/testAuthMW 模式）
- `internal/handler/voucher_type_handler_test.go`（新文件）— CRUD happy path + 重复 code 拒绝 + 引用检查
- `internal/handler/sys_config_handler_test.go`（新文件）— CRUD + 批量获取 + is_system 不可删
- `internal/repository/*_test.go` — 已有 *_test.go 的就追加（period_repo_test.go 加 overlap/case）

### 6. 文档
- `docs/api-contracts/v1/openapi.yml` — 追加新端点定义（或新建 `docs/api-contracts/v1/setup-f8.md` 子文档）
- 本 SPEC 末尾追加"已实现 API 列表"章节

---

## 响应约定

**重要勘误**：原任务 body 提到 `R.ok()` 统一响应，但本项目实际惯例是直接返回 `fiber.Map{...}` JSON，**没有 `R.ok()` helper**（`grep R.ok` 0 命中）。OpenCode 应沿用项目现有 `c.JSON(fiber.Map{"data": ...})` 模式，**不要引入新的统一响应包装层**（避免范围蔓延）。如果认为确实有必要，可作为独立后续任务讨论。

成功响应：
```go
return c.JSON(fiber.Map{"data": period})
return c.JSON(fiber.Map{"status": "deleted"})
return c.JSON(fiber.Map{"configs": map[string]string{...}})
```

错误响应：
```go
return c.Status(400).JSON(fiber.Map{"error": "..."})
return c.Status(404).JSON(fiber.Map{"error": "..."})
```

---

## 验证标准

1. `cd /root/data/disk/huihua-finance && go build ./...` 通过
2. `go vet ./...` 0 error
3. `SKIP_INTEGRATION=0 go test ./internal/repository/... ./internal/handler/... ./internal/service/...` — 新增测试全部通过，已有测试不破坏
4. HTTP 端到端验证（必须按 OpenCode 铁律执行）：
   - 启动 API server (port 18080)
   - 调 `POST /api/v1/v1/auth/login` 拿 token
   - 依次 curl：List Periods → Create Period (202607) → Update Period Name → GetCurrent → Enable closed period → Delete future period
   - 调 VoucherType CRUD + 重复 code 拒绝
   - 调 SysConfig CRUD + BatchGet + 尝试删 is_system=true 应失败
5. 数据库对照（4 层验证）：每个创建操作执行后 `SELECT * FROM xxx WHERE id=...` 确认行存在；每个删除操作确认行消失或 deleted_at 设置
6. 不破坏现有 `/periods/current` `/periods/:period_no/close` 等已注册路由

---

## 非目标 / 范围控制

- **不做**前端改动（前端页面 VoucherTypeList.vue 已存在，本任务只通后端 API；前端联通是后续任务）
- **不做**系统参数的"按 group 分组查询 UI"（API 支持 `?group=` 即可）
- **不做**凭证类型的"排序持久化（拖拽保存）" — sort_order 字段已支持但前端拖拽是另一任务
- **不引入**新的依赖（统一用 pgx + uuid）
- **不改**现有 voucher 业务逻辑

---

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| Period Delete 误删已有数据的期间 | 强校验：journal_entries 引用计数 > 0 时禁止 |
| Period Create 期间重叠 | 强校验：overlap SQL 查询（start <= other.end AND end >= other.start）|
| SysConfig 误删 is_system=true | service 层校验，repo 层不需要（防御性） |
| VoucherType code 重复 | 依赖 `UNIQUE(tenant_id, code) WHERE deleted_at IS NULL` 索引；service 层预检 + DB 兜底 |
| 并行 delegate 改同一 main.go | 按上次教训（MEMORY 记录），OpenCode 必须串行改 main.go，或先 git pull feature/period-vouchertype-sysconfig-crud 分支最新版本 |