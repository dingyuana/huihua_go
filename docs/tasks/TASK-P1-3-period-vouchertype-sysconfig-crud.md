# TASK-P1-3 | 期间管理 + 凭证类型 + 系统参数 CRUD

**版本**：V1.0
**日期**：2026-06-15
**优先级**：P1（基础数据完善）
**状态**：待开发
**关联 SPEC**：`docs/plans/setup/SPEC-period-vouchertype-sysconfig-crud.md`

---

## 1. 任务描述

完成 3 个基础设置类模块的 CRUD 接口：

1. **会计期间 Period**：补 Create / Update / Delete / Enable
2. **凭证类型 VoucherType**：从模型到完整 service/handler/repo + 路由
3. **系统参数 SysConfig**：从无到有（migration + 模型 + service/handler/repo + 路由 + 批量获取）

详细 API 规格、字段、校验、错误码见 SPEC 第 3 节。

---

## 2. 任务拆分

### 2.1 数据库迁移
- [ ] `migrations/068_sys_configs.sql` — sys_configs 表 + RLS + 索引（voucher_types 表 067 已存在，不重复）

### 2.2 Period 补全
- [ ] `internal/repository/period_repo.go` — 加 `GetByID / UpdateMeta / Delete / CheckOverlap / Enable`
- [ ] `internal/service/period_service.go` — 加 `CreatePeriod / UpdatePeriod / DeletePeriod / EnablePeriod`
- [ ] `internal/handler/period_handler.go` — 加 `Create / Update / Delete / Enable` handler
- [ ] `internal/handler/period_handler_test.go` — 追加 4 个 test
- [ ] `cmd/api/main.go` — 注册 `POST/PUT/DELETE /periods/:id` 和 `POST /periods/:id/enable`

### 2.3 VoucherType 全套
- [ ] `internal/repository/voucher_type_repo.go` — 新建（含软删）
- [ ] `internal/service/voucher_type_service.go` — 新建（code 唯一性 + 引用检查）
- [ ] `internal/handler/voucher_type_handler.go` — 新建
- [ ] `internal/handler/voucher_type_handler_test.go` — 新建 CRUD + 重复 code 测试
- [ ] `cmd/api/main.go` — 实例化 + 5 个路由

### 2.4 SysConfig 全套
- [ ] `internal/repository/sys_config_repo.go` — 新建
- [ ] `internal/service/sys_config_service.go` — 新建（is_system 保护 + json 后缀校验）
- [ ] `internal/handler/sys_config_handler.go` — 新建
- [ ] `internal/handler/sys_config_handler_test.go` — 新建 CRUD + batch-get + is_system 保护测试
- [ ] `cmd/api/main.go` — 实例化 + 7 个路由

### 2.5 文档
- [ ] `docs/api-contracts/v1/setup-f8.md` — 新建子文档，列出所有新端点 + curl 示例
- [ ] SPEC 末尾追加"已实现 API 列表"

---

## 3. 字段规格

### 3.1 Period Create/Update
```
POST /api/v1/periods
{
  "period_no": 202607,                  // 必填，YYYYMM 格式
  "period_name": "2026年7月",            // 必填
  "start_date": "2026-07-01",           // 必填
  "end_date": "2026-07-31"              // 必填
}

PUT /api/v1/periods/:id
{
  "period_name": "...",                 // 可选
  "start_date": "...",                  // 可选
  "end_date": "..."                     // 可选
}
```

### 3.2 VoucherType Create/Update
```
POST /api/v1/voucher-types
{
  "code": "JZ",                          // 必填，同租户唯一
  "name": "记账凭证",                     // 必填
  "description": "通用记账",              // 可选
  "sort_order": 10                       // 可选，默认 0
}

PUT /api/v1/voucher-types/:id
{
  "code": "...",                         // 可选
  "name": "...",                         // 可选
  "description": "...",                  // 可选
  "sort_order": ...,                     // 可选
  "is_active": true                      // 可选
}
```

### 3.3 SysConfig Create/Update
```
POST /api/v1/configs
{
  "config_key": "auto_match_threshold",  // 必填，同租户唯一
  "config_value": "0.95",                // 必填
  "description": "自动核销阈值",          // 可选
  "group": "reconciliation",             // 可选，默认 'default'
  "is_system": false                     // 可选，默认 false
}

PUT /api/v1/configs/:id
{
  "config_value": "...",                 // 可选
  "description": "...",                  // 可选
  "group": "..."                         // 可选（不允许 is_system 通过 update 改 true）
}

POST /api/v1/configs/batch-get
{ "keys": ["k1", "k2", "k3"] }

→ 200
{ "configs": { "k1": "v1", "k2": "v2", "k3": null } }    // 不存在的 key 返回 null
```

---

## 4. 业务校验

| 校验 | 模块 | 实现位置 |
|---|---|---|
| period_no 唯一 | Period | 依赖 DB UNIQUE + service 预检 |
| 期间日期范围不重叠 | Period | service.CheckOverlap() → SQL: `WHERE start_date <= $end AND end_date >= $start`（排除自身 id）|
| closed 状态期间不可删 | Period | service.DeletePeriod() 校验 |
| 期间内已有 journal_entries → 禁止删 | Period | service 查 `SELECT COUNT(*) FROM journal_entries WHERE tenant_id=$1 AND posting_date BETWEEN $start AND $end` |
| voucher_type code 唯一（软删范围内） | VoucherType | 依赖 DB partial unique index + service 预检 |
| 删除 voucher_type 时若有 voucher 引用 → 禁止 | VoucherType | service 查 `journal_entries.voucher_type = code` 计数；或保留灵活度：先 soft delete + warning；这里采用 hard delete 拒绝策略 |
| is_system=true 不可改 key | SysConfig | service.Update 校验 |
| is_system=true 不可删 | SysConfig | service.Delete 校验 |
| config_key 以 `_json` 结尾 → value 必须是合法 JSON | SysConfig | service.Create/Update 用 `json.Valid` 校验 |

---

## 5. 验证清单（OpenCode 必须全部勾选才能 claim 完成）

### 5.1 编译
- [ ] `cd /root/data/disk/huihua-finance && go build ./...` 通过
- [ ] `go vet ./...` 0 error

### 5.2 测试
- [ ] `go test ./internal/repository/... -count=1 -run 'TestPeriod|TestVoucherType|TestSysConfig'` 通过
- [ ] `go test ./internal/handler/... -count=1 -run 'TestPeriod|TestVoucherType|TestSysConfig'` 通过
- [ ] `go test ./internal/service/... -count=1` 通过（确保不破坏现有 service 测试）

### 5.3 HTTP 端到端（必须）
- [ ] 启动 API server：`go build -o /tmp/huihua-api ./cmd/api && /tmp/huihua-api &`（或 `make dev`）
- [ ] Login: `curl -X POST http://localhost:18080/api/v1/auth/login -d '{"username":"admin","password":"admin"}' -H 'Content-Type: application/json'`
- [ ] 期间 CRUD：
  - [ ] `GET /periods` → 200 列表
  - [ ] `POST /periods {period_no:202607,...}` → 200 返回新建
  - [ ] 重叠区间 `POST /periods {period_no:202608,start:2026-07-15,end:2026-08-15}` → 400 "overlap"
  - [ ] `PUT /periods/:id` → 200
  - [ ] 尝试 `DELETE /periods/:id` 已 closed 期间 → 400
  - [ ] `POST /periods/:id/enable` 已 closed 期间 → 200 status=open
- [ ] 凭证类型 CRUD：
  - [ ] `GET /voucher-types` → 200 列表
  - [ ] `POST /voucher-types {code:"JZ",name:"记账凭证"}` → 200
  - [ ] 重 code 再 POST → 400 "code already exists"
  - [ ] `PUT /voucher-types/:id` → 200
  - [ ] `DELETE /voucher-types/:id` → 200
- [ ] 系统参数 CRUD：
  - [ ] `GET /configs` → 200 列表
  - [ ] `POST /configs {config_key:"test.k1",config_value:"v1",group:"test"}` → 200
  - [ ] `POST /configs/batch-get {keys:["test.k1","nonexistent"]}` → 200 `{configs:{"test.k1":"v1","nonexistent":null}}`
  - [ ] 创建 is_system=true → `DELETE` → 400 "is_system cannot be deleted"
  - [ ] `PUT /configs/:id` is_system → 拒绝改 key（如果业务允许 value 修改）
  - [ ] `DELETE /configs/:id` → 200

### 5.4 数据库对照（铁律：每个写操作验证行存在/消失）
- [ ] `psql -d huihua_finance -c "SELECT * FROM accounting_periods WHERE period_no=202607"` → 1 行
- [ ] `psql -d huihua_finance -c "SELECT * FROM voucher_types WHERE code='JZ'"` → 1 行
- [ ] `psql -d huihua_finance -c "SELECT * FROM sys_configs WHERE config_key='test.k1'"` → 1 行
- [ ] 删除后行消失或 deleted_at 设置（按模块语义）

### 5.5 回归
- [ ] `GET /periods/current` 仍工作
- [ ] `POST /periods/:period_no/close` 仍工作
- [ ] 现有 `go test` 全套通过（不只是新模块）

---

## 6. Git 工作流（按项目惯例）

1. 在 `main` 分支拉新分支：
   ```
   cd /root/data/disk/huihua-finance
   git checkout main && git pull
   git checkout -b feature/period-vouchertype-sysconfig-crud
   ```
2. 开发：每完成一个模块（period / voucher_type / sys_config）一次 commit
3. 测试：本地跑 `go build` + `go test`，确保通过
4. **不 push**，留给老丁 review 后合并（按 MEMORY 规则：未测试代码禁止合入 main，重大功能 feature 分支 → 测试 → PR）

commit message 风格：
```
feat(period): add Create/Update/Delete/Enable endpoints
feat(voucher-type): add full CRUD service+handler+repo
feat(sys-config): add migration + full CRUD + batch-get
```

---

## 7. 范围红线（禁止蔓延）

- 不要改 frontend
- 不要改 voucher 业务逻辑
- 不要改既有的 period close/unclose/preview-closing/execute-closing 逻辑（只新增 endpoint）
- 不要新增依赖（保持 pgx + uuid）
- 不要做 system config 的 UI 分组（API 支持 `?group=` 即可）
- 不要做 voucher type 的拖拽排序持久化（前端任务）
- 不要引入新的统一响应包装（沿用 `fiber.Map` 惯例）
- **不要在 `git add -A`**（项目根目录有 docs/、frontend/node_modules/ 等不要的目录，必须指定文件）

---

## 8. 参考资料

- 模型：`internal/model/accounting_period.go`、`internal/model/voucher_type.go`（含 SysConfig）
- 迁移：`migrations/010_account_setup.sql`（period 表 schema）、`migrations/067_voucher_types.sql`
- 最近 CRUD 先例：`internal/handler/account_handler.go`、`internal/service/account_service.go`、`internal/handler/classification_rule_handler.go`
- 测试模式：`internal/handler/period_handler_test.go`（testHandlerPool / testAuthMW）
- 路由注册：`cmd/api/main.go`（搜索 `periodHandler.List` 看上下文）
- 项目惯例（MEMORY）：
  - 重大功能 feature 分支 → 测试 → PR
  - `git add` 必须指定路径，不用 `-A`
  - 测试铁律：每个写操作 4 层验证（HTTP + DB 行存在 + DB 行消失 + 边界）

---

## 9. 交付物清单

OpenCode 完成后必须产出（每项都要 git status 确认存在）：

- [ ] `migrations/068_sys_configs.sql`
- [ ] 修改后的 `internal/repository/period_repo.go`
- [ ] 修改后的 `internal/service/period_service.go`
- [ ] 修改后的 `internal/handler/period_handler.go`
- [ ] 修改后的 `internal/handler/period_handler_test.go`
- [ ] 新建的 `internal/repository/voucher_type_repo.go`
- [ ] 新建的 `internal/service/voucher_type_service.go`
- [ ] 新建的 `internal/handler/voucher_type_handler.go`
- [ ] 新建的 `internal/handler/voucher_type_handler_test.go`
- [ ] 新建的 `internal/repository/sys_config_repo.go`
- [ ] 新建的 `internal/service/sys_config_service.go`
- [ ] 新建的 `internal/handler/sys_config_handler.go`
- [ ] 新建的 `internal/handler/sys_config_handler_test.go`
- [ ] 修改后的 `cmd/api/main.go`（新路由注册）
- [ ] 新建的 `docs/api-contracts/v1/setup-f8.md`
- [ ] 修改后的 `docs/plans/setup/SPEC-period-vouchertype-sysconfig-crud.md`（追加 API 列表章节）

---

**注**：本任务由 Hermes（PM/架构师）拆解并交付给 OpenCode（唯一执行层）执行。OpenCode 完成后 Hermes 负责审核，必要时要求修改，最终由老丁（Human Manager）合并到 main。