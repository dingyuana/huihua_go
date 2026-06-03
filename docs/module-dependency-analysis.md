# 慧财智能财务平台 — 模块依赖结构分析

> 生成时间：2026-05-31
> 分析范围：Go 后端 + Vue 3 前端全栈

---

## 目录

1. [项目整体架构](#1-项目整体架构)
2. [后端依赖图 (Go)](#2-后端依赖图-go)
   - [2.1 严格分层](#21-严格分层)
   - [2.2 层间依赖验证](#22-层间依赖验证)
   - [2.3 业务模块间的 Service 级依赖](#23-业务模块间的-service-级依赖)
   - [2.4 核心模型实体](#24-核心模型实体)
   - [2.5 数据库迁移依赖顺序](#25-数据库迁移依赖顺序)
3. [前端依赖图 (Vue 3 + TypeScript)](#3-前端依赖图-vue-3--typescript)
   - [3.1 前端整体架构](#31-前端整体架构)
   - [3.2 API 层模块](#32-api-层模块)
   - [3.3 Store 层](#33-store-层)
   - [3.4 路由 → 页面映射](#34-路由--页面映射)
   - [3.5 类型依赖](#35-类型依赖)
   - [3.6 组件依赖](#36-组件依赖)
4. [跨端映射](#4-跨端映射)
5. [关键发现](#5-关键发现)
   - [5.1 架构优点](#51-架构优点)
   - [5.2 架构风险](#52-架构风险)
   - [5.3 数据流全景](#53-数据流全景)

---

## 1. 项目整体架构

```
huihua-finance/
├── cmd/api/              # Go 入口 (main.go)
│   └── main.go          # 路由注册、依赖注入
├── internal/             # Go 业务逻辑层
│   ├── config/          # Viper 配置加载 (1 file)
│   ├── middleware/       # JWT 认证 + RLS 租户中间件 (2 files)
│   ├── model/           # 22 个数据模型
│   ├── handler/         # 22 个 HTTP Handler
│   ├── service/         # 27 个 service 文件 (含测试)
│   └── repository/      # 22 个 repository 文件 (含测试)
├── pkg/                 # Go 共享基础设施
│   ├── database/        # pgxpool + Redis 客户端
│   ├── jwt/             # JWT 签发与验证
│   └── utils/           # 通用工具函数
├── migrations/          # 23 个 SQL 迁移文件
├── tests/               # Python 集成测试套件
└── frontend/            # Vue 3 SPA
    └── src/
        ├── api/         # 8 个业务模块 API 调用封装
        │   ├── modules/ # auth, account, bank, invoice, payment, period, reconciliation, voucher
        │   └── mock/    # MSW Mock Service Worker
        ├── stores/      # 3 个 Pinia Store (auth, app, tenant)
        ├── router/      # 22 个路由（带角色守卫）
        ├── views/       # 11 个目录，22+ 页面组件
        ├── components/  # 5 个业务组件 + 布局/图表/通用组件
        ├── types/       # 8 个模型类型 + 枚举 + API 类型
        └── config/      # 应用配置 + 菜单配置
```

---

## 2. 后端依赖图 (Go)

### 2.1 严格分层

```
cmd/main.go
  │
  ├──→ pkg/database       (PostgreSQL pgxpool + Redis go-redis)
  ├──→ internal/config    (Viper 配置加载)
  ├──→ internal/middleware (JWT 认证 + RLS 租户中间件)
  │     ├──→ internal/config
  │     ├──→ pkg/jwt
  │     └──→ pkg/database
  ├──→ internal/handler   (HTTP 层：参数解析、响应格式化)
  │     └──→ internal/service  ← 唯一的下游依赖
  ├──→ internal/service   (业务逻辑层：编排、事务、校验)
  │     ├──→ internal/repository
  │     ├──→ internal/model
  │     └──→ internal/service  ← service 间交叉引用
  └──→ internal/repository (数据访问层：原生 SQL)
        └──→ internal/model
```

### 2.2 层间依赖验证

| 层 | 导入项目内包 | 说明 | 是否合规 |
|---|---|---|---|
| `handler` | `service` + `model` | Handler 仅调用 Service，偶尔引用 model 类型 | ✅ |
| `service` | `repository` + `model` | Service 编排 Repository，使用 model | ✅ |
| `repository` | `model` | 纯 SQL，仅扫描到 model struct | ✅ |
| `middleware` | `config` + `pkg/jwt` + `pkg/database` | 基础设施依赖 | ✅ |
| `config` | viper (外部) | 零内部依赖，纯配置 | ✅ |
| `pkg/database` | pgx + go-redis (外部) | 零内部依赖 | ✅ |
| `pkg/jwt` | golang-jwt (外部) | 零内部依赖 | ✅ |

**结论：分层严格，无违规反向依赖。** 所有依赖方向都从上层指向下层。

### 2.3 业务模块间的 Service 级依赖

```
BankService
  └──→ AccountRepository  (校验 clearing_account 类型为 asset)

BankTransactionService
  ├──→ ClassificationRuleService  (自动智能分类)
  └──→ BankRepository             (查询银行账户)

VoucherService
  └──→ VoucherTemplateService  (生成凭证编号)

VoucherStateMachine                             ← 凭证状态机
  ├──→ JournalRepository
  ├──→ AuditRepository
  └──→ GLEntryRepository

VoucherAutoGenerateService  ★ (耦合度最高，8个依赖)
  ├──→ JournalRepository
  ├──→ GLEntryRepository
  ├──→ BankTransactionRepository
  ├──→ InvoiceRepository
  ├──→ AccountRepository
  ├──→ ClassificationRuleService
  ├──→ VoucherTemplateService
  └──→ ApprovalService

SetupService
  ├──→ CompanyRepository
  ├──→ PeriodRepository
  └──→ AccountService

PeriodService  ★ (第二高耦合，5个依赖)
  ├──→ PeriodRepository
  ├──→ JournalRepository
  ├──→ GLEntryRepository
  ├──→ AccountRepository
  └──→ AssetDepreciationRepository

ReconciliationService (核销)
  ├──→ BankTransactionRepository
  ├──→ InvoiceRepository
  ├──→ ReconciliationRepository
  └──→ JournalRepository

BankReconciliationService (银企对账)
  ├──→ BankTransactionRepository
  ├──→ JournalRepository
  ├──→ BankRepository
  └──→ GLEntryRepository

ReportService (财务报表)
  ├──→ GLEntryRepository
  ├──→ OpeningBalanceRepository
  ├──→ AccountRepository
  └──→ PeriodRepository

ApprovalService (审批流)
  ├──→ ApprovalRepository
  └──→ JournalRepository

OpeningBalanceService
  ├──→ OpeningBalanceRepository
  └──→ AccountRepository

ClassificationRuleService
  ├──→ ClassificationRuleRepository
  └──→ AccountRepository

VoucherTemplateService
  ├──→ VoucherTemplateRepository
  └──→ AccountRepository

ExchangeRateService
  └──→ ExchangeRateRepository

AuthService
  ├──→ UserRepository
  └──→ config

AuditService
  └──→ AuditRepository

AssetDepreciationService
  ├──→ AssetDepreciationRepository
  └──→ JournalRepository

InvoiceService
  └──→ InvoiceRepository

PartyService
  └──→ PartyRepository
```

### 2.4 核心模型实体

```
model/ (22 个文件)
├── user.go                  # 用户
├── tenant.go                # 租户
├── company.go               # 公司
├── account.go               # 科目表（嵌套集 lft/rgt 模型）
├── accounting_period.go     # 会计期间
├── journal.go               # 凭证 (JournalEntry + JournalEntryLine)
├── gl_entry.go              # 总账分录
├── bank.go                  # 银行账户 (BankAccount)
├── bank_transaction.go      # 银行流水
├── invoice.go               # 销售发票 (SalesInvoice)
├── payment.go               # 付款
├── party.go                 # 客商档案
├── reconciliation.go        # 核销
├── approval.go              # 审批流
├── asset.go                 # 固定资产
├── asset_depreciation.go    # 资产折旧
├── classification_rule.go   # 分类规则
├── exchange_rate.go         # 汇率
├── voucher_template.go      # 凭证模板
├── audit.go                 # 审计日志
├── budget.go                # 预算
└── opening_balance.go       # 期初余额
```

### 2.5 数据库迁移依赖顺序

迁移文件按编号顺序执行，反映了表的构建依赖链：

```
001_init.sql                ← 基础表: tenants, users, companies, roles
002_journal_gl.sql          ← 凭证 + 总账 (依赖 001)
003_invoice_payment.sql     ← 发票 + 付款 (依赖 001)
004_bank.sql                ← 银行账户 (依赖 001)
005_asset.sql               ← 固定资产
006_budget.sql              ← 预算
007_audit.sql               ← 审计日志
008_rls_force.sql           ← PostgreSQL RLS 行级安全策略
009_app_user.sql            ← 应用用户
010_account_setup.sql       ← 科目表 (嵌套集)
011_depreciation_run.sql    ← 折旧运行记录
012_classification_rules.sql / voucher_template.sql  ← 规则+模板
013_voucher_state_machine.sql  ← 凭证状态机
014_bank_transactions.sql      ← 银行流水 (依赖 bank_accounts)
015_opening_balance.sql         ← 期初余额 (依赖 accounts)
016_exchange_rates.sql          ← 汇率
017_bank_reconciliation.sql     ← 银企对账
018_reconciliation.sql          ← 核销
019_approval.sql                ← 审批流
020_voucher_template_approval_bind.sql  ← 模板与审批绑定
021_voucher_state_transitions.sql      ← 状态变更审计追踪
022_seed_data.sql               ← 种子数据
```

---

## 3. 前端依赖图 (Vue 3 + TypeScript)

### 3.1 前端整体架构

```
frontend/src/main.ts (Vite 启动入口)
│
├──→ createApp(App.vue)
│     └──→ <router-view />  (根组件)
│
├──→ router/index.ts
│     ├──→ routes/base.ts (22 个路由声明)
│     │     ├──→ views/           (11 个目录的页面组件)
│     │     ├──→ components/app/AppLayout.vue (布局 Shell)
│     │     └──→ stores/auth.store (路由守卫)
│     └──→ 路由守卫逻辑:
│           ├── 认证守卫 (isLoggedIn → /login)
│           └── 权限守卫 (meta.roles → /403)
│
├──→ createPinia()
│     ├──→ auth.store    (JWT Token + 用户信息 + 角色权限)
│     ├──→ app.store     (UI 状态：侧边栏、全局加载)
│     └──→ tenant.store  (租户切换、公司选择)
│
├──→ ElementPlus (UI 组件库, zh-cn 语言包)
├──→ ElementPlusIconsVue (图标库, 全局注册)
├──→ permissionDirective (v-permission 自定义指令)
├──→ SCSS 全局样式 (styles/index.scss)
│
└──→ [条件式] MSW Mock Service Worker
      └──→ api/mock/browser.ts → handlers.ts
            ├── accounts.ts
            ├── auth.ts
            ├── bank-transactions.ts
            ├── invoices.ts
            ├── periods.ts
            └── vouchers.ts
```

### 3.2 API 层模块

| 模块 | 文件 | 依赖 | 对应后端路由前缀 |
|------|------|------|-----------------|
| `auth` | `api/modules/auth.ts` | `request.ts` | `/auth/login` |
| `account` | `api/modules/account.ts` | `request.ts` + `types/models/account` | `/accounts/*` |
| `bank` | `api/modules/bank.ts` | `request.ts` + `types/models/bank` | `/bank-accounts/*`, `/bank-transactions/*` |
| `invoice` | `api/modules/invoice.ts` | `request.ts` + `types/models/bank` | `/invoices/*` |
| `voucher` | `api/modules/voucher.ts` | `request.ts` + `types/models/journal` | `/vouchers/*` |
| `payment` | `api/modules/payment.ts` | `request.ts` | `/payments/*` |
| `period` | `api/modules/period.ts` | `request.ts` | `/periods/*` |
| `reconciliation` | `api/modules/reconciliation.ts` | `request.ts` | `/reconciliation/*` |

**统一 HTTP 客户端** `api/request.ts` 提供：
- Axios 实例 + JWT 注入拦截器
- 统一错误处理（401 → 跳转登录，403 → 权限提示）
- 后端错误格式兼容（`{ error }` / `{ code, message }`）

### 3.3 Store 层

```
stores/
├── auth.store.ts      ← 依赖: types/models/user
│   ├── token (localStorage 持久化)
│   ├── user (当前用户信息)
│   ├── isLoggedIn (computed)
│   ├── permissions (computed)
│   ├── setAuth() / logout()
│   └── hasPermission() / hasRole()
│
├── app.store.ts       ← 零依赖
│   ├── sidebarCollapsed
│   ├── globalLoading
│   └── toggleSidebar()
│
├── tenant.store.ts    ← 依赖: types/models/tenant
│   ├── currentTenantId
│   ├── currentCompany
│   ├── tenantList
│   └── switchTenant()
│
└── plugins/
      └── tenant-reset.ts  ← 租户切换时重置所有业务 store
```

### 3.4 路由 → 页面映射

| 路径 | 页面组件 | 布局 | 角色 |
|------|---------|------|------|
| `/login` | `views/login/LoginView.vue` | blank | 公开 |
| `/403` | `views/error/403.vue` | blank | 公开 |
| `/` | `AppLayout` → redirect `/dashboard` | Shell | 受保护 |
| `/dashboard` | `views/dashboard/DashboardView.vue` | Shell | 全部 |
| `/setup/company` | `views/setup/SetupWizard.vue` | Shell | admin |
| `/setup/accounts` | `views/setup/AccountChart.vue` | Shell | admin, agent |
| `/setup/bank-accounts` | `views/setup/BankAccountList.vue` | Shell | admin, cashier, agent |
| `/setup/parties` | `views/setup/PartyList.vue` | Shell | admin, accountant_ar, agent |
| `/setup/rules` | `views/setup/RuleLibrary.vue` | Shell | admin |
| `/bank/import` | `views/bank/ImportView.vue` | Shell | cashier, admin, agent |
| `/bank/workbench` | `views/bank/CashierWorkbench.vue` | Shell | cashier, admin, agent |
| `/invoices` | `views/invoices/InvoiceList.vue` | Shell | accountant_ar, admin, agent |
| `/reconciliation/precheck` | `views/reconciliation/PreCheckView.vue` | Shell | accountant_ar, admin |
| `/reconciliation/match` | `views/reconciliation/MatchView.vue` | Shell | accountant_ar, admin |
| `/reconciliation/manual` | `views/reconciliation/ManualMatch.vue` | Shell | accountant_ar, admin |
| `/vouchers` | `views/voucher/VoucherList.vue` | Shell | admin, agent |
| `/vouchers/create` | `views/voucher/VoucherEdit.vue` | Shell | admin, agent |
| `/vouchers/review` | `views/voucher/ReviewWorkbench.vue` | Shell | admin |
| `/bank-reconciliation/match` | `views/reconciliation-bank/MatchingView.vue` | Shell | cashier, admin |
| `/bank-reconciliation/balance` | `views/reconciliation-bank/BalanceSheet.vue` | Shell | cashier, admin |
| `/period/health-check` | `views/period/HealthCheck.vue` | Shell | admin |
| `/period/voucher-gaps` | `views/period/VoucherGapView.vue` | Shell | admin |
| `/period/reports` | `views/period/FinancialReports.vue` | Shell | admin, boss |
| `/analytics` | `views/analytics/` (待完善) | Shell | boss |

### 3.5 类型依赖

```
types/
├── api.ts            ← 所有 api/modules/* 使用 (ApiResponse<T>, PageResult<T>, PageQuery)
├── enums.ts          ← 所有 types/models/* 使用 (DocStatus, VoucherType, Role, PartyType 等)
├── enums.ts          ← stores/auth.store 使用 (Role)
├── router.ts         ← router/index.ts 使用 (路由元信息接口)
├── check.ts          ← components/check/ 使用
├── models/
│   ├── account.ts    → enums
│   ├── bank.ts       → enums
│   ├── journal.ts    → enums
│   ├── invoice.ts    → enums
│   ├── party.ts      → enums
│   ├── payment.ts    → enums
│   ├── tenant.ts     → (无额外依赖)
│   └── user.ts       → (无额外依赖)
└── store/            ← store 类型定义
```

### 3.6 组件依赖

```
components/
├── app/
│   └── AppLayout.vue          ← 所有受保护页面的 Shell（侧边栏+顶栏+内容区）
├── business/                   ← 被 views/* 引用
│   ├── AccountSelector.vue    ← 科目表树形选择器
│   ├── AmountInput.vue        ← 金额输入组件（decimal 格式化）
│   ├── DocStatusTag.vue       ← 凭证状态标签
│   ├── PartySelector.vue      ← 客商选择器
│   └── PeriodPicker.vue       ← 会计期间选择器
├── charts/                    ← ECharts 图表组件
│   └── ...
├── check/                     ← 结账体检组件
│   └── ...
└── common/                    ← 通用 UI 组件
    └── ...
```

---

## 4. 跨端映射

| 业务模块 | 前端页面 | 前端 API 模块 | 后端 Handler | 后端 Service | Repository | 核心表 |
|---------|---------|--------------|-------------|-------------|-----------|--------|
| 认证登录 | LoginView | auth | AuthHandler | AuthService | UserRepository | users |
| 科目表 | AccountChart | account | AccountHandler | AccountService | AccountRepository | accounts |
| 银行账户 | BankAccountList | bank | BankHandler | BankService | BankRepository | bank_accounts |
| 银行流水 | ImportView, CashierWorkbench | bank | BankTransactionHandler | BankTransactionService | BankTransactionRepository | bank_transactions |
| 发票 | InvoiceList | invoice | InvoiceHandler | InvoiceService | InvoiceRepository | sales_invoices |
| 凭证 | VoucherList, VoucherEdit, ReviewWorkbench | voucher | VoucherHandler | VoucherService + VoucherStateMachine | JournalRepository + GLEntryRepository | journal_entries, gl_entries |
| 核销 | PreCheckView, MatchView, ManualMatch | reconciliation | ReconciliationHandler | ReconciliationService | ReconciliationRepository | reconciliation_pairs |
| 银企对账 | MatchingView, BalanceSheet | (无独立模块) | BankReconciliationHandler | BankReconciliationService | (跨 repo) | bank_reconciliation* |
| 审批流 | ReviewWorkbench 内嵌 | (无独立模块) | ApprovalHandler | ApprovalService | ApprovalRepository | approval_flows, approval_tasks |
| 会计期间 | HealthCheck, VoucherGapView, FinancialReports | period | PeriodHandler + ReportHandler | PeriodService + ReportService | PeriodRepository + GLEntryRepository | accounting_periods |
| 分类规则 | RuleLibrary | (无独立模块) | ClassificationRuleHandler | ClassificationRuleService | ClassificationRuleRepository | classification_rules |
| 凭证模板 | (凭证创建时使用) | (无独立模块) | VoucherTemplateHandler | VoucherTemplateService | VoucherTemplateRepository | voucher_templates |
| 期初余额 | SetupWizard 内嵌 | (无独立模块) | OpeningBalanceHandler | OpeningBalanceService | OpeningBalanceRepository | opening_balances |
| 汇率管理 | (无独立页面) | (无独立模块) | ExchangeRateHandler | ExchangeRateService | ExchangeRateRepository | exchange_rates |
| 固定资产 | (待完善) | (无独立模块) | AssetDepreciationHandler | AssetDepreciationService | AssetDepreciationRepository | fixed_assets |
| 审计日志 | (无独立页面) | (无独立模块) | AuditHandler | AuditService | AuditRepository | audit_logs |
| 账套创建 | SetupWizard | (无独立模块) | SetupHandler | SetupService | CompanyRepository + PeriodRepository | companies |

---

## 5. 关键发现

### 5.1 架构优点

1. **严格分层**：Handler → Service → Repository → Model，所有依赖方向从上到下，无违规反向依赖
2. **无 ORM**：全部手写原生 SQL，财务系统对 SQL 精确度的要求得到最大保障，SQL 完全可控
3. **依赖注入**：`cmd/api/main.go` 中通过构造函数手动组装依赖，依赖关系清晰可见，便于替换mock
4. **多租户隔离**：PostgreSQL RLS 行级安全策略 + `middleware/tenant.go` 的 `SetTenant()`，数据层天然隔离
5. **前端路由守卫**：双层权限控制（`router.beforeEach` 检查 `meta.roles` + `v-permission` 指令），不同角色看到不同的菜单和页面
6. **迁移文件版本化**：23 个 SQL 文件按编号顺序执行，依赖链清晰，可增量迁移

### 5.2 架构风险

1. **`VoucherAutoGenerateService` 耦合度最高**（依赖 8 个 Repository/Service），是未来重构的首要目标。如果改动了任一依赖模块的接口，后续影响面较大
2. **`PeriodService` 耦合度次高**（依赖 5 个 Repository），结账逻辑牵涉凭证、总账、科目、折旧等多个模块，结账事务边界需要特别注意
3. **`handler/bank_handler.go` 泛型复用问题**：该文件的 List/Create/Update/Delete 方法被多个实体（bank-accounts, parties, invoices, classification-rules, exchange-rates）共用一个 Handler struct，导致路由注册容易混淆
4. **Service 层交叉引用**：Service → Service 的调用形成了隐式依赖网，不利于单元测试（需要 mock 其他 Service），典型例子是 `VoucherAutoGenerateService` 调用了 `ClassificationRuleService`、`VoucherTemplateService`、`ApprovalService`
5. **前端 Mock 条件式加载**：MSW 通过 `VITE_ENABLE_MOCK` 环境变量控制，增加了开发时的心智负担和构建复杂度
6. **repository 层缺少接口抽象**：当前是直接 struct 引用，无法在单元测试中轻松 mock repository（不过 Go 可以通过定义接口解决此问题）
7. **部分业务模块前端无独立 API 封装**：分类规则、凭证模板、期初余额、汇率等模块在前端没有独立的 `api/modules/*.ts` 文件，调用分散在页面组件中

### 5.3 核心数据流全景

```
银行流水导入 (BankTransactionHandler)
  │
  ▼
BankTransactionService
  ├──→ 分类: ClassificationRuleService (规则引擎自动匹配)
  ├──→ 去重检测
  │
  ▼
VoucherAutoGenerateService  ★ (核心编排点)
  ├──→ JournalRepository (创建凭证)
  ├──→ GLEntryRepository (过总账)
  ├──→ 关联发票/流水
  ├──→ 使用 VoucherTemplateService
  └──→ 提交 ApprovalService
        │
        ▼
  VoucherStateMachine (凭证状态机)
    ├── 草稿(0) → 提交 → 核准(1) → 过账
    ├── 驳回 → 草稿
    ├── 取消 → 作废
    └── 红字冲销 → 负数凭证
        │
        ▼
  GLEntry (总账实时过账)
        │
        ▼
  ┌── BankReconciliationService (银企对账)
  │     ├── 5级智能匹配
  │     └── 余额调节表
  │
  ├── ReportService (财务报表)
  │     ├── 试算平衡表
  │     ├── 利润表
  │     ├── 资产负债表
  │     └── 现金流量表
  │
  └── PeriodService (会计期间结账)
        ├── 结账前健康检查
        ├── 自动结转凭证
        └── 期间开/关
```
