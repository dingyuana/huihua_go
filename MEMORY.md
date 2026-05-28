# huihua-finance (Go版) — MEMORY

> AI财务SaaS · Go + PostgreSQL + Fiber v2
> Remote: git@github.com:dingyuana/huihua_go.git

## 项目概述

银行流水驱动业财一体化平台。以银行流水为核心入口，自动生成凭证、自动审批、银行智能对账。
面向中小企业的第二套财务系统技术验证（独立于 Python FastAPI 版的 `huihua-financial-master`）。

## 技术栈

| Layer | Tech |
|-------|------|
| Language | Go 1.22+ |
| HTTP Framework | Fiber v2.52 |
| DB Driver | pgx / pgxpool v5 |
| ORM | 原生 SQL（无 GORM GEN 代码生成） |
| Config | Viper（`.env` 嵌套格式） |
| Cache | Redis 7 |
| Deploy | Docker Compose |
| Database | PostgreSQL 15 + RLS 多租户隔离 |

## 项目结构

```
huihua-finance/
├── cmd/api/main.go          # 入口，142 handlers，路由注册
├── internal/
│   ├── config/config.go     # Viper 配置加载
│   ├── middleware/
│   │   ├── auth.go          # JWT 验证（HS256）
│   │   ├── tenant.go        # RLS: SET app.current_tenant
│   │   └── log.go
│   ├── model/               # 24 个数据模型
│   ├── handler/             # 20 个 HTTP Handler
│   ├── service/             # 20 个业务逻辑 Service
│   └── repository/          # 17 个数据访问 Repository
├── pkg/database/postgres.go  # pgxpool 连接管理
├── migrations/               # 21 个 SQL 迁移文件
└── docs/plans/              # 任务计划文档
```

## API 路由（99条，全部接通）

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/health` | HealthHandler | 健康检查 |
| `/api/v1/setup` | SetupHandler | 公司注册/账套初始化 |
| `/api/v1/tenants` | TenantHandler | 租户管理 |
| `/api/v1/users` | UserHandler | 应用用户管理 |
| `/api/v1/accounts` | AccountHandler | 科目管理（GetTree 等） |
| `/api/v1/parties` | PartyHandler | 往来单位 CRUD |
| `/api/v1/bank-accounts` | BankAccountHandler | 银行账户 CRUD |
| `/api/v1/bank-transactions` | BankTxnHandler | 银行流水（导入/分类/匹配） |
| `/api/v1/vouchers` | VoucherHandler | 手工凭证 CRUD |
| `/api/v1/voucher-templates` | VoucherTemplateHandler | 凭证模板 CRUD |
| `/api/v1/auto-generate` | VoucherAutoGenerateHandler | 银行流水→凭证 自动生成 |
| `/api/v1/reconciliation` | ReconciliationHandler | 银行对账 |
| `/api/v1/invoices` | InvoiceHandler | 发票 CRUD + Excel 解析 |
| `/api/v1/classification-rules` | ClassificationRuleHandler | 分类规则 CRUD |
| `/api/v1/approval-flows` | ApprovalFlowHandler | 审批流程 CRUD |
| `/api/v1/approval-tasks` | ApprovalTaskHandler | 审批任务 |
| `/api/v1/reports` | ReportHandler | 财务报表（试算/利润/资产负债） |
| `/api/v1/periods` | PeriodHandler | 会计期间管理 |
| `/api/v1/fixed-assets` | FixedAssetHandler | 固定资产 CRUD |
| `/api/v1/depreciation` | DepreciationHandler | 折旧计提/执行 |
| `/api/v1/exchange-rates` | ExchangeRateHandler | 汇率 CRUD |

## 数据库（21个 Migration）

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
| 012 | classification_rules — 银行流水分类规则 |
| 012 | voucher_template — 凭证模板 |
| 013 | voucher_state_machine — 凭证状态机 |
| 014 | bank_transactions — 银行流水 |
| 015 | opening_balance — 开办余额 |
| 016 | exchange_rates — 汇率 |
| 017 | bank_reconciliation — 银行对账 |
| 018 | reconciliation — 对账记录 |
| 019 | approval — 审批流 + 审批任务 |
| 020 | voucher_template_approval_bind — 模板绑定审批流 |

## 核心模块进度

| 模块 | 状态 | 说明 |
|------|------|------|
| F0 基础设施 | ✅ | JWT/RLS/审计日志/应用用户 |
| F1 账套初始化 | ✅ | 科目/客商/会计期间/开办余额/试算平衡 |
| F2 发票管理 | ✅ | 发票CRUD+Excel解析+分类规则引擎 |
| F3 银行流水 | ✅ | 导入+智能分类+标记匹配 |
| F4 凭证状态机 | ✅ | 提交/核准/驳回/取消/红冲+GL过账 |
| F5 凭证自动生成 | ✅ | 银行流水→凭证+批量生成+auto-submit审批 |
| F6 固定资产折旧 | ✅ | 折旧计划+运行执行 |
| F7 银行对账 | ✅ | 5级匹配策略+UnmatchedItem |
| F8 财务报表 | ✅ | 试算平衡表+利润表+资产负债表 |
| F9 审批流 | ✅ | 审批流/任务+阈值DB化+模板绑定 |
| F10 会计期间结账 | ✅ | 期间关闭+结账凭证生成 |
| 银行账户 CRUD | ✅ | GetByID/Update/Delete |
| 对手方 CRUD | ✅ | GetByID/Create/Update/Delete |
| SetupHandler | ✅ | GetStatus+CreateCompany |

**完成度：~83%**

剩余低优先功能：
- 凭证模板 Clone
- 银行交易 Update
- 银行对账 AutoMatch 入口
- E2E 集成测试套件

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

## 当前状态

- API 运行在 `localhost:8080`（非 Docker，直接二进制）
- 数据库：Docker PostgreSQL (`huihua-finance_postgres_1`)
- 测试 token 有效期至 2026-05-29

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
