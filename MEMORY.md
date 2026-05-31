---
## 📌 Changelog

### 🔴 Major Change — 2026-05-30
- **Change type**: documentation update / build fix / testing + assessment
- **Summary**: 
  - 修复 2 个 handler 中 `GetTenantID` 签名不匹配问题（`audit_handler.go`, `classification_rule_handler.go`）
  - 修复 `SetupWizard.vue` 中 `instanceof Date` 类型错误（初始值 `''` → `null as Date | null`）
  - 前端构建通过：`vue-tsc --noEmit` ✅ → `vite build` ✅
  - 后端构建通过：`go build ./...` ✅ → `go test ./...` ✅（7/7 PASS）
  - 更新 MEMORY.md 技术栈文档（版本号、文件计数、API 路由表）
  - 创建 `docs/01-项目总览/功能完成度矩阵-Go版.md` — 基于 Go 版源码的完整评估
  - 标注 3 个 P0 开发缺口（经营分析仪表盘、费用报销、结账向导后端聚合端点）

### 🔴 Major Change — 2026-05-28
- **Change type**: architecture adjustment / scope expansion
- **Summary**: huihua-finance Go版完成度从~83%上调，MEMORY.md架构文档今日新增，所有核心模块全部接通
- **Trigger**: 2026-05-28 提交，18条commit，涵盖报表、期间结账、审批流CRUD、对手方/银行账户CRUD
- **Impact**: F8财务报表、F10会计期间结账、F9审批流全部功能完成，API路由达99条全部接通
- **Details**:
  - 新增 MEMORY.md 架构文档（2026-05-28）
  - 完成财务报表（试算平衡表+利润表+资产负债表+期间GL合并）
  - 完成会计期间结账API + 结账凭证自动生成
  - 完成审批流 Update + Delete + 2条新路由
  - 完成对手方 GetByID/Create/Update/Delete CRUD
  - 完成银行账户 GetByID/Update/Delete CRUD
  - 补全 VoucherTemplate 语义化响应
  - 集成测试套件今日建立
- **Corresponding commit**: `980959f Add integration test suite for core API endpoints`
---

# huihua-finance (Go版) — MEMORY

> AI财务SaaS · Go + PostgreSQL + Fiber v2
> Remote: git@github.com:dingyuana/huihua_go.git

## 项目概述

银行流水驱动业财一体化平台。以银行流水为核心入口，自动生成凭证、自动审批、银行智能对账。
面向中小企业的第二套财务系统技术验证（独立于 Python FastAPI 版的 `huihua-financial-master`）。

## 技术栈

| Layer | Tech |
|-------|------|
| Language | **Go 1.24** (`go.mod: go 1.24.0`) |
| HTTP Framework | **Fiber v2.52** (`github.com/gofiber/fiber/v2 v2.52.0`) |
| DB Driver | **pgx v5 / pgxpool** (`github.com/jackc/pgx/v5 v5.6.0`) |
| ORM | **原生 SQL**（无 ORM，全部手写 SQL） |
| Amount Type | **shopspring/decimal v1.4**（`.CoefficientInt64()` 传前端） |
| Config | **Viper v1.21**（`.env` 嵌套格式 `database.host=...`） |
| Cache | **Redis 7** via go-redis v9 (`github.com/redis/go-redis/v9 v9.19.0`) |
| JWT | **golang-jwt v5** (HS256) |
| Auth Middleware | Fiber middleware.Auth(cfg) — JWT 验证 + Tenant 上下文注入 |
| Multi-tenant | **PostgreSQL RLS** — `SET app.current_tenant` 策略 |
| Deploy | **Docker Compose** — PostgreSQL 15 + Redis 7-alpine |
| Database | **PostgreSQL 15** + RLS 多租户隔离 |
| Excel | **xuri/excelize v2.10** — 发票/流水 Excel 导入 |

## 项目结构

```
huihua-finance/
├── cmd/api/main.go          # 入口，90+ API 路由，Fiber 注册
├── internal/
│   ├── config/config.go     # Viper 配置加载
│   ├── middleware/
│   │   ├── auth.go          # JWT 验证（HS256）
│   │   ├── tenant.go        # RLS: SET app.current_tenant
│   │   └── log.go
│   ├── model/               # 22 个数据模型
│   ├── handler/             # 21 个 HTTP Handler（不含 health）
│   ├── service/             # 21 个业务逻辑 Service
│   └── repository/          # 18 个数据访问 Repository
├── pkg/database/postgres.go  # pgxpool 连接管理
├── pkg/jwt/                 # JWT 签发与验证
├── pkg/utils/               # 通用工具
├── migrations/               # 23 个 SQL 迁移文件（含种子数据）
├── tests/test_api.py        # Python 集成测试（807行）
├── frontend/                # Vue 3 SPA
│   └── src/
│       ├── api/             # Axios + 9 个业务模块
│       ├── components/check/# 4 个检测通用组件
│       └── views/           # 23 个页面
└── docs/                    # 需求/设计/开发计划/API 文档
```

## API 路由（90+条，全部接通）

| 路由前缀 | Handler 文件 | 说明 |
|---------|-------------|------|
| `/health` | `health.go` | 健康检查（公开） |
| `/api/v1/auth` | `auth_handler.go` | 登录（公开） |
| `/api/v1/audit-logs` | `audit_handler.go` | 审计日志查询 |
| `/api/v1/accounts` | `account_handler.go` | 科目树 / 种子数据 |
| `/api/v1/exchange-rates` | `exchange_rate_handler.go` | 汇率 CRUD + 换算 |
| `/api/v1/bank-accounts` | `bank_handler.go` | 银行账户 CRUD |
| `/api/v1/parties` | `party_handler.go` | 往来单位 CRUD + Excel 导入 |
| `/api/v1/account-setup` | `setup_handler.go` | 账套初始状态 / 创建公司 |
| `/api/v1/assets` + `/depreciation` | `asset_depreciation_handler.go` | 折旧计划 / 运行 |
| `/api/v1/invoices` | `invoice_handler.go` | 发票 CRUD + Excel 导入 + 解析 |
| `/api/v1/classification-rules` | `classification_rule_handler.go` | 分类规则 CRUD + 排序 + 匹配 |
| `/api/v1/bank-transactions` | `bank_transaction_handler.go` | 流水导入/分类/匹配/查询 |
| `/api/v1/voucher-templates` | `voucher_template_handler.go` | 模板 CRUD + 编号规则 |
| `/api/v1/vouchers` | `voucher_handler.go` | 凭证 CRUD + 状态机 |
| `/api/v1/opening-balances` | `opening_balance_handler.go` | 开办余额导入/试算/校验 |
| `/api/v1/periods` | `period_handler.go` | 期间管理 + 断号检测 + 结账 |
| `/api/v1/reconciliation` | `reconciliation_handler.go` | 核销配对/确认 |
| `/api/v1/bank-reconciliation` | `bank_reconciliation_handler.go` | 银企对账/报告/状态 |
| `/api/v1/reports` | `report_handler.go` | 试算/利润/资产负债表 |
| `/api/v1/approval-flows` + `/approvals` | `approval_handler.go` | 审批流 CRUD + 审批任务 |
| `/api/v1/bank-transactions/:id/generate-voucher` | `voucher_auto_generate_handler.go` | 流水→凭证 自动生成 |

## 数据库（23个 Migration）

| 文件 | 内容 |
|------|------|
| 001 | init.sql — tenants, users, departments, roles, menus |
| 002 | journal_gl — gl_entries 表 |
| 003 | invoice_payment — 发票/应收应付 |
| 004 | bank — 银行账户/对手方 |
| 005 | asset — 固定资产 |
| 006 | budget — 预算 |
| 007 | audit — 审计日志 |
| 008 | rls_force — RLS 策略强制 |
| 009 | app_user — 应用用户 |
| 010 | account_setup — 科目初始数据 |
| 011 | depreciation_run — 折旧 |
| 012a | classification_rules — 银行流水分类规则 |
| 012b | voucher_template — 凭证模板 |
| 013 | voucher_state_machine — 凭证状态机 |
| 014 | bank_transactions — 银行流水 |
| 015 | opening_balance — 开办余额 |
| 016 | exchange_rates — 汇率 |
| 017 | bank_reconciliation — 银行对账 |
| 018 | reconciliation — 对账记录 |
| 019 | approval — 审批流 + 审批任务 |
| 020 | voucher_template_approval_bind — 模板绑定审批流 |
| 021 | voucher_state_transitions — 状态转换日志 |
| 022 | seed_data — 种子数据（测试用户 + 科目初始数据） |

## 核心模块进度

| 模块 | 状态 | 说明 |
|------|------|------|
| F0 基础设施 | ✅ | JWT/RLS/审计日志/应用用户 |
| F1 账套初始化 | ✅ | 科目/客商/会计期间/开办余额/试算平衡 |
| F2 发票管理 | ✅ | 发票CRUD+Excel解析+分类规则引擎 |
| F3 银行流水 | ✅ | 导入（CSV/Excel）+智能分类（规则引擎）+匹配 |
| F4 凭证状态机 | ✅ | 提交/核准/驳回/取消/红冲+GL过账+审计追踪 |
| F5 凭证自动生成 | ✅ | 银行流水→凭证+批量生成+auto-submit审批 |
| F6 固定资产折旧 | ✅ | 折旧计划生成+运行执行+凭证生成 |
| F7 银行对账 | ✅ | 5级匹配策略+余额调节表+未达账项 |
| F8 财务报表 | ✅ | 试算平衡表+利润表+资产负债表（3个端点） |
| F9 审批流 | ✅ | 审批流CRUD+审批任务+金额阈值DB化+模板绑定 |
| F10 会计期间结账 | ✅ | 期间开启/关闭+结账凭证生成+断号检测 |
| 前端检测组件 | ✅ | CheckResultPanel/CheckSummaryCard/BlockingGuard/StatusBadge |
| 前端结账向导 | ✅ | 7区域布局：基础检查+风险预警+关键指标+人工确认+损益结转+结账操作 |

**构建状态（2026-05-30）：** `go build ./...` ✅ | `go test ./...` ✅（7/7 PASS）| `vue-tsc --noEmit` ✅ | `vite build` ✅

**剩余低优功能：**
- 凭证模板 Clone
- 银行交易 Update
- 银行对账 AutoMatch 入口
- E2E Playwright 测试套件
- 经营分析仪表盘（Dashboard）
- 费用报销模块

## 关键实现细节

### decimal 处理
```go
// 必须用 .CoefficientInt64() 不能用 .Int64()
amount := entry.Amount.CoefficientInt64()
// 不要用 entry.Amount.Int64() — 会丢失小数位
```

### Viper 配置（重要）
```bash
# .env 文件必须用嵌套格式，不能用 HF_DATABASE_HOST
database.host=127.0.0.1
database.port=5432
database.user=huihua_app
database.password=hfpwd_app
database.dbname=huihua_finance
database.sslmode=disable
```

**SetEnvKeyReplacer 只对环境变量生效，不对 .env 文件键生效。**
因此 `.env` 必须是嵌套格式 `database.host`，不能是 `HF_DATABASE_HOST`。

### pgx 连接行为
若 DSN 的 Host 为空（配置未读取），pgx 会 fallback 到 Unix socket `/tmp/.s.PGSQL.5432`，导致连接的用户变 `root` 而非 `huihua_app`。

### JWT 认证
- 算法：HS256
- Claims：sub(user_id), tenant_id, username, role, exp, iat
- Token 示例（有效期24h）：
  ```
  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODAwNDUwMjUsImlhdCI6MTc3OTk1ODYyNSwicm9sZSI6ImFkbWluIiwic3ViIjoiMzk0YWE2YzgtMGY5Ny00YTM1LWJhM2ItNDFjNjUxYzI3OWNkIiwidGVuYW50X2lkIjoiYTAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAxIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciJ9.Ei1A7J6N3JcLcXWWB2iINugLcVh1P1-6SdwIC6CBJec
  ```

### RLS 多租户隔离
```sql
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON <table> FOR ALL
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
```
中间件 `tenant.go` 在每个请求中执行 `SET app.current_tenant = '<tenant_id>'`。

### 审批流阈值（从代码硬编码迁移到 DB）
- Level2: 1,000,000（100万）
- Level3: 5,000,000（500万）
- 存储在 `approval_flows.threshold_amount_level2/level3`
- 币种：`approval_flows.currency`

## 当前状态（2026-05-30）

- 全栈构建通过：Go 后端 + Vue 前端均正常编译
- 后端测试：**7 个单元测试全部 PASS**（middleware 4 个 + jwt 3 个）
- 前端测试：**vitest 无配置文件**，**Playwright 无配置文件**（测试框架已安装但未配置）
- 测试 token 有效期至 2026-05-29（已过期，需重新生成或替换硬编码 Token）

## 本地启动

```bash
cd /root/data/disk/huihua-finance
go build -o /tmp/huihua-api ./cmd/api
/tmp/huihua-api
# API: http://localhost:8080
# Health: http://localhost:8080/health
```

## 相关文档

- 任务计划：`docs/plans/2026-05-27-task-f0.1-scaffold.md`
- 参考：`~/.hermes/skills/devops/project-design-sync/references/huihua-finance-go.md`
