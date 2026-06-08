---
## ⚠️ 项目边界（强约束）

> **永不再操作** `/root/data/disk/huihua-financial-master`（Python FastAPI 版）。**所有"财务软件" = 本仓库 Go 项目**，除非用户显式说 Python 才考虑。

| 路径 | 状态 | 操作 |
|---|---|---|
| `/root/data/disk/huihua-finance` (Go + Fiber + pgx) | 🟢 **活跃** | 唯一允许修改 |
| `/root/data/disk/huihua-financial-master` (Python FastAPI) | 🔴 **停用** | 永不操作 |
| `/root/.hermes/skills/projects/huihua-financial` 等 | 🔴 Python 技能 | 忽略 |
| `/root/.hermes/skills/projects/huihua-go-backend` | 🟢 Go 技能 | 可参考 |

## 🛠️ Dev Server 维护责任

> **本地 = 开发服务器**。Sisyphus 负责保持本地与 origin 最新一致。

- **跟踪分支**：`feature/expense-invoice`（当前唯一活跃分支）
- **同步节奏**：每次任务开始前 `git fetch origin && git status`，如有漂移立即 pull / rebase
- **构建验证**：每次改完代码跑 `go build ./... && go vet ./... && (cd frontend && pnpm exec vite build)` 三件套
- **不上生产**：本地 dev server ≠ 生产 `129.211.7.254`。部署到生产是用户手动操作
- **MEMORY.md 是真理**：未来 Sisyphus 必须先读本文件 + 仓库 README.md 了解项目状态

---
## 📌 Changelog

### 🔴 Major Change — 2026-06-03
- **Change type**: scope expansion / business flow change
- **Summary**: 业务单据凭证智能生成模块上线，完成从收付款单→凭证的全自动链路，支持科目映射配置、双向关联、税务场景自动检测
- **Trigger**: commit `a545f8eb` — 业务单据凭证智能生成 + 科目映射配置 + 凭证关联显示
- **Impact**: F5凭证自动生成模块增强（F5完成度100%），前端发票列表+凭证编辑完整闭环
- **Details**:
  - 新增 `bus_doc_mapping` 科目映射配置表（收款/付款/转账/费用/利息预置数据）
  - 收付款单生成凭证时通过映射表自动确定借贷科目
  - 凭证与单据双向关联（`source_doc_type/id/no` + `voucher_id/no`）
  - 凭证分录行自动填写 `user_remark`（收款/付款/应收/应付 + 对方名称）
  - 发票核销附加分录支持（按 `payment_allocations` 拆分多行）
  - 税务场景自动检测（对方/描述含税务关键词时切换到应交税费科目）
  - 前端凭证列表显示对方名称、来源单据、科目标签
  - 新增迁移 036-038（counterparty_name, voucher_source_doc_link, bus_doc_mapping）
- **Corresponding commit**: `a545f8eb feat: 业务单据凭证智能生成 + 科目映射配置 + 凭证关联显示`

### 🔴 Major Change — 2026-06-08
- **Change type**: bug fix / style alignment
- **Summary**: 修复银行对账页面500错误 + 凭证审核金额显示为零 + 凭证审核页样式对齐
- **Trigger**: 用户反馈 http://129.211.7.254:3002/bank-reconciliation/match 报500，http://129.211.7.254:3002/vouchers/review 金额显示0
- **Impact**: 银行对账模块可用性修复，凭证审核金额正确显示
- **Details**:
  - **银行对账500**: `bank_accounts.clearing_account_id` 未配置，`ReconcileBankAccount` 中解引用 nil 指针 panic。添加了 nil guard 返回明确错误，并在数据库中设置了 `clearing_account_id = 'ef72ff1d-3cbd-4cd0-9ce2-18430106bd2e'`（科目1002银行存款）
  - **凭证金额为零**: `approval_tasks.amount` 在旧版 `SubmitForApproval` 中计算错误，现有4条pending记录存储了 `0.0000`。使用 SQL `UPDATE ... SET amount = (SELECT SUM(jel.debit) FROM journal_entry_lines jel WHERE jel.journal_entry_id = at.journal_entry_id)` 修复存量数据。修正了 `journal_repo.go:GetByID` SQL（移除了不存在的 `debit_total, credit_total` 列）
  - **构建验证**: 每次修改后跑 `go build ./...` + `go vet ./...` + `(cd frontend && pnpm exec vite build)` 三件套
  - **关键排查技巧**: API 服务器的全局 ErrorHandler（`cmd/api/main.go`）将所有 handler 返回的错误统一包装为 `{"error":"internal server error"}`，导致真实错误信息被隐藏。排查时需查看服务器日志或直接调用 API 绕过中间件
  - **dev server**: 运行中的 API 进程在 `/tmp/huihua-api`，源码修改后必须 `go build -o /tmp/huihua-api ./cmd/api && kill <pid> && /tmp/huihua-api` 才能生效
  - **凭证审核页样式**: ReviewWorkbench.vue 对齐 VoucherList.vue 风格——添加 filter-card（日期范围选择器+查询按钮），保持相同的 `.page-header` + `.filter-card` CSS 类和表格式样
  - **凭证审核页字段补齐**: ReviewWorkbench.vue 表格列从"凭证号/日期/摘要/金额/制单人/AI风控"改为"凭证号/日期/对方名称/科目/来源单据/摘要/借方合计/贷方合计/状态"，与凭证列表页完全一致
    - 后端改动: `ApprovalTaskWithVoucher` 模型新增 7 个字段、`ListPending` SQL 增加 7 列子查询（含从 `journal_entry_lines` 汇总 `debit_total`/`credit_total` + 联表取首行 `first_account_code`/`first_account_name`）、`PendingReview` handler 透传字段
    - `journal_entry_lines` 无 `line_no` 列，排序用 `id`
- **Corresponding session**: Sisyphus 调试会话 2026-06-08

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

## 项目核心规则

### 单据状态管理（单据状态机原则）

所有业务单据（发票、应收单、应付单、凭证、收付款单等）必须拥有 `docstatus` 或 `status` 字段，状态变化遵循以下原则：

1. **正向变化**：单据状态随业务流程推进而变化，每个步骤将状态往前推。例如：
   - 发票：草稿(draft) → 已确认(verified) → 已核销(paid)
   - 凭证：草稿(draft, docstatus=0) → 已提交(submitted, docstatus=1) → 已核准(approved, docstatus=2)
   - 收付款单：草稿(draft) → 已提交(submitted) → 已核准(approved)

2. **反向变化**：当上游单据被删除/取消时，下游关联单据的状态必须同步回退。例如：
   - 删除凭证 → 发票 `docstatus` 回退为 0（可重新生成凭证）✓
   - 删除凭证 → 收付款单状态回退，`voucher_id/voucher_no` 置空 ✓
   - 删除凭证 → 银行流水解绑凭证 ✓

3. **防止重复生成**：任何"生成凭证"操作必须先检查源单据状态：
   - 源单据 `docstatus != 0` → 拒绝生成，返回错误
   - 生成成功后必须锁定源单据状态（`docstatus = 1`）

4. **凭证退回 = 作废**：退回意味着凭证作废（docstatus → 3 cancelled），不再显示且不可修改：
   - 草稿(0) → 作废(3)：调用 Cancel API
   - 已提交(1) → 作废(3)：调用 Cancel API（状态机已允许 posted 状态下 cancel）
   - 退回后源发票 `docstatus` 回退为 0，可以重新生成凭证
   - 收付款单状态与凭证状态互相独立，退回不影响收付款单状态 ✓

5. **发票确认不自动生成凭证**：发票确认（status → verified）与生成凭证是两个独立步骤：
   - 确认仅设置 `status = verified`，不触发自动生成
   - 用户手动点击"生成凭证"后调用 `GenerateFromInvoice`
   - 生成成功后发票 `docstatus = 1` 锁定，防止重复生成

### 会计科目编码规范

- 主营业务收入：**5001**（非 6001）
- 主营业务成本：**5401**（非 6003）
- 应收账款：1122
- 应付账款：2202
- 库存现金：1001
- 银行存款：1002

发票生成凭证时通过 `findAccountByCode` 按科目编码查找，必须使用正确的科目编码。

### 凭证平衡规则

- 凭证的借方总额必须严格等于贷方总额
- 生成分录时必须确保至少有一个借方行和一个贷方行
- `len(lines) == 0` 的检查不足以保证平衡，需要确保 `total_debit == total_credit`

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
