# 慧财智能财务平台 (Huihua Finance) — Code Wiki

> 银行流水驱动的业财一体化 SaaS 平台
> Go + Vue 3 全栈实现 | PostgreSQL 15 + Redis 7

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈](#2-技术栈)
3. [整体架构](#3-整体架构)
4. [后端模块详解](#4-后端模块详解)
5. [前端模块详解](#5-前端模块详解)
6. [数据库设计](#6-数据库设计)
7. [核心数据流](#7-核心数据流)
8. [API 路由总览](#8-api-路由总览)
9. [关键类与函数说明](#9-关键类与函数说明)
10. [依赖关系](#10-依赖关系)
11. [项目运行方式](#11-项目运行方式)

---

## 1. 项目概述

慧财智能财务平台是一套面向中小企业的**第二代财务系统技术验证**。以**银行流水为核心入口**，自动生成凭证、自动审批、智能对账，打通业财税全链路。

### 核心理念

- **流水驱动** — 银行流水进入系统即触发后续全部流程
- **业财一体** — 发票、收付款、凭证、报表全链路自动流转
- **多租户 SaaS** — 基于 PostgreSQL RLS 的原生多租户隔离
- **国产化适配** — 遵循中国会计准则与发票制度

### 目标用户

| 角色 | 职责 |
|------|------|
| 出纳 | 银行流水导入、银企对账 |
| 往来会计 | 发票管理、核销处理 |
| 财务主管 | 凭证审核、期末结账、报表 |
| 老板/经理 | 经营分析、财务报表查看 |
| 普通员工 | 费用报销 |
| 代账会计 | 多家企业账务处理 |

### 项目状态

| 维度 | 状态 |
|------|------|
| 后端 API | ✅ **99+ 条路由全部接通** |
| 前端页面 | ✅ **22+ 个页面可用** |
| 数据库迁移 | ✅ **64 个迁移文件** |
| 功能模块 | ✅ **F0-F10 全部完成** |

---

## 2. 技术栈

### 后端技术栈

| 层次 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.24 | 高性能编译型语言 |
| HTTP 框架 | Fiber | 2.52.0 | 类 Express 的高速路由框架 |
| 数据库驱动 | pgx | v5 (pgxpool) | 原生 PostgreSQL 驱动 |
| 数据访问 | 原生 SQL | — | 无 ORM，全部手写 SQL |
| 配置管理 | Viper | 1.21.0 | `.env` 文件 + 环境变量 |
| 缓存 | Redis | 7 (go-redis v9) | 会话/缓存层 |
| 认证 | JWT | golang-jwt v5 | HS256 算法 |
| 金额处理 | shopspring/decimal | 1.4.0 | 精确十进制运算 |
| Excel 处理 | xuri/excelize | 2.10.1 | 发票/流水 Excel 导入 |

### 前端技术栈

| 层次 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 框架 | Vue | 3.4 | Composition API (`<script setup>`) |
| 构建工具 | Vite | 5 | 开发服务器 + 生产构建 |
| 语言 | TypeScript | 5.4 (strict) | 全栈类型安全 |
| UI 组件库 | Element Plus | 2.7 | 中后台组件体系 |
| 状态管理 | Pinia | 2 | Composition API 风格 Store |
| 路由 | Vue Router | 4 | 路由守卫 + 角色权限 |
| HTTP 客户端 | Axios | 1.7 | 拦截器注入 JWT |
| 图表 | ECharts | 5 + vue-echarts | 财务报表可视化 |
| CSS 预处理器 | SCSS | — | 全局样式变量体系 |
| Mock 服务 | MSW | 2 | 开发模式 API Mock |
| 包管理 | pnpm | — | 依赖管理 |

### 基础设施

| 组件 | 版本 | 用途 |
|------|------|------|
| PostgreSQL | 15 | 主数据库 |
| Redis | 7-alpine | 缓存 |
| Docker Compose | — | 本地基础设施编排 |

---

## 3. 整体架构

### 3.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        前端层 (Vue 3 SPA)                          │
│                  http://localhost:3002                              │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ HTTP/JSON
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     API 网关层 (Nginx)                              │
│              /api/* → Go API (:8080)  /* → Vue SPA                │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────────────┐
│                     Handler Layer (控制器层)                        │
│  auth_handler | account_handler | voucher_handler | ... (22个)     │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ 依赖注入
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Service Layer (业务逻辑层)                       │
│  auth_service | voucher_state_machine | report_service | ...       │
│                    (27 个 Service 文件)                             │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ 调用
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   Repository Layer (数据访问层)                     │
│  user_repo | journal_repo | gl_entry_repo | ... (22个)             │
└───────────────────────────┬─────────────────────────────────────────┘
                            │ SQL (pgx)
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Database Layer                                │
│          PostgreSQL 15 (主库)           │    Redis 7 (缓存)        │
│  tenants | users | journals | gl_entries │   Session | RateLimit  │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 分层架构原则

```
Handler (HTTP) → Service (业务) → Repository (数据)
```

- **Handler** — 参数解析、响应格式化，不含业务逻辑
- **Service** — 业务编排、事务管理、权限校验
- **Repository** — 原生 SQL，参数绑定，结果扫描

### 3.3 项目目录结构

```
huihua-finance/
├── cmd/api/
│   └── main.go                    # 应用入口（路由注册 + 依赖注入）
│
├── internal/
│   ├── config/
│   │   └── config.go              # Viper 配置加载
│   │
│   ├── middleware/
│   │   ├── auth.go                # JWT 认证中间件 (HS256)
│   │   ├── tenant.go               # PostgreSQL RLS 多租户中间件
│   │   └── auth_test.go           # 认证中间件单元测试
│   │
│   ├── model/                     # 22 个数据模型
│   │   ├── user.go                # 用户模型
│   │   ├── tenant.go              # 租户模型
│   │   ├── company.go              # 公司模型
│   │   ├── account.go              # 科目表模型（嵌套集 lft/rgt）
│   │   ├── accounting_period.go    # 会计期间模型
│   │   ├── journal.go              # 凭证模型 (JournalEntry + JournalEntryLine)
│   │   ├── gl_entry.go             # 总账分录模型
│   │   ├── bank.go                 # 银行账户模型
│   │   ├── bank_transaction.go     # 银行流水模型
│   │   ├── invoice.go              # 销售发票模型
│   │   ├── ar_invoice.go           # 应收发票模型
│   │   ├── ap_invoice.go           # 应付发票模型
│   │   ├── expense_invoice.go      # 费用发票模型
│   │   ├── payment.go              # 付款模型
│   │   ├── party.go                # 客商档案模型
│   │   ├── reconciliation.go       # 核销模型
│   │   ├── approval.go             # 审批流模型
│   │   ├── asset.go                # 固定资产模型
│   │   ├── asset_depreciation.go   # 资产折旧模型
│   │   ├── classification_rule.go   # 分类规则模型
│   │   ├── voucher_template.go     # 凭证模板模型
│   │   ├── opening_balance.go      # 开办余额模型
│   │   ├── exchange_rate.go         # 汇率模型
│   │   ├── audit.go                # 审计日志模型
│   │   ├── reimbursement.go        # 报销模型
│   │   ├── payroll.go              # 薪资模型
│   │   ├── advance_receipt.go      # 预收账款模型
│   │   ├── advance_payment.go      # 预付款项模型
│   │   ├── advance_allocation.go   # 预收预付核销模型
│   │   └── settlement_log.go        # 结算日志模型
│   │
│   ├── handler/                    # 22 个 HTTP Handler
│   │   ├── auth_handler.go         # 登录认证
│   │   ├── account_handler.go       # 科目管理
│   │   ├── bank_handler.go         # 银行账户
│   │   ├── bank_transaction_handler.go  # 银行流水
│   │   ├── bank_txn_review_handler.go    # 流水审核工作台
│   │   ├── invoice_handler.go       # 发票管理
│   │   ├── ar_invoice_handler.go   # 应收发票
│   │   ├── ap_invoice_handler.go   # 应付发票
│   │   ├── expense_invoice_handler.go    # 费用发票
│   │   ├── expense_invoice_import_handler.go  # 费用发票导入
│   │   ├── expense_invoice_ocr_handler.go     # 费用发票 OCR
│   │   ├── payment_handler.go      # 收付款单
│   │   ├── voucher_handler.go      # 凭证管理
│   │   ├── voucher_template_handler.go  # 凭证模板
│   │   ├── voucher_auto_generate_handler.go  # 凭证自动生成
│   │   ├── reconciliation_handler.go  # 核销管理
│   │   ├── bank_reconciliation_handler.go  # 银企对账
│   │   ├── period_handler.go        # 会计期间
│   │   ├── report_handler.go         # 财务报表
│   │   ├── approval_handler.go      # 审批流
│   │   ├── party_handler.go          # 往来单位
│   │   ├── company_handler.go       # 公司设置
│   │   ├── setup_handler.go         # 账套初始化
│   │   ├── opening_balance_handler.go  # 开办余额
│   │   ├── exchange_rate_handler.go  # 汇率管理
│   │   ├── asset_depreciation_handler.go  # 固定资产折旧
│   │   ├── payroll_handler.go       # 薪资管理
│   │   ├── reimbursement_handler.go  # 报销管理
│   │   ├── reimbursement_attachment_handler.go  # 报销附件
│   │   ├── reimbursement_invoice_link_handler.go  # 报销发票关联
│   │   ├── advance_receipt_handler.go   # 预收账款
│   │   ├── advance_payment_handler.go   # 预付款项
│   │   ├── advance_allocation_handler.go  # 预收预付核销
│   │   ├── classification_rule_handler.go  # 分类规则
│   │   ├── audit_handler.go         # 审计日志
│   │   ├── dashboard_handler.go     # 仪表盘统计
│   │   ├── health.go                 # 健康检查
│   │   ├── clear_data_handler.go    # 数据清理
│   │   ├── credit_control_handler.go  # 信用控制
│   │   └── aging_handler.go         # 账龄分析
│   │
│   ├── service/                     # 27 个业务逻辑 Service
│   │   ├── auth_service.go          # 认证服务
│   │   ├── account_service.go       # 科目服务
│   │   ├── bank_service.go          # 银行账户服务
│   │   ├── bank_transaction_service.go  # 银行流水服务
│   │   ├── bank_txn_review_service.go  # 流水审核服务
│   │   ├── classification_rule_service.go  # 分类规则服务
│   │   ├── invoice_service.go       # 发票服务
│   │   ├── ar_invoice_service.go    # 应收发票服务
│   │   ├── ap_invoice_service.go    # 应付发票服务
│   │   ├── expense_invoice_service.go  # 费用发票服务
│   │   ├── expense_invoice_import_service.go  # 费用发票导入服务
│   │   ├── payment_service.go       # 收付款服务
│   │   ├── payment_state_machine.go  # 收付款状态机
│   │   ├── voucher_service.go       # 凭证服务
│   │   ├── voucher_state_machine.go  # 凭证状态机
│   │   ├── voucher_template_service.go  # 凭证模板服务
│   │   ├── voucher_auto_generate_service.go  # 凭证自动生成服务
│   │   ├── reconciliation_service.go  # 核销服务
│   │   ├── bank_reconciliation_service.go  # 银企对账服务
│   │   ├── period_service.go        # 会计期间服务
│   │   ├── report_service.go        # 报表服务
│   │   ├── approval_service.go      # 审批流服务
│   │   ├── party_service.go         # 往来单位服务
│   │   ├── setup_service.go         # 账套初始化服务
│   │   ├── opening_balance_service.go  # 开办余额服务
│   │   ├── exchange_rate_service.go  # 汇率服务
│   │   ├── asset_depreciation_service.go  # 折旧服务
│   │   ├── payroll_service.go       # 薪资服务
│   │   ├── reimbursement_service.go  # 报销服务
│   │   ├── reimbursement_attachment_service.go  # 报销附件服务
│   │   ├── reimbursement_invoice_link_service.go  # 报销发票关联服务
│   │   ├── advance_receipt_service.go  # 预收账款服务
│   │   ├── advance_payment_service.go  # 预付款项服务
│   │   ├── advance_allocation_service.go  # 预收预付核销服务
│   │   ├── credit_control_service.go  # 信用控制服务
│   │   ├── aging_service.go         # 账龄分析服务
│   │   ├── audit_service.go         # 审计日志服务
│   │   ├── ai_feedback_service.go   # AI 反馈服务
│   │   ├── ocr_service.go          # OCR 服务
│   │   ├── clear_data_service.go    # 数据清理服务
│   │   └── invoice_state_machine.go  # 发票状态机
│   │
│   ├── repository/                  # 22 个数据访问 Repository
│   │   ├── user_repo.go
│   │   ├── account_repo.go
│   │   ├── bank_repo.go
│   │   ├── bank_transaction_repo.go
│   │   ├── bank_txn_review_repo.go
│   │   ├── classification_rule_repo.go
│   │   ├── invoice_repo.go
│   │   ├── ar_invoice_repo.go
│   │   ├── ap_invoice_repo.go
│   │   ├── expense_invoice_repo.go
│   │   ├── payment_repo.go
│   │   ├── journal_repo.go
│   │   ├── gl_entry_repo.go
│   │   ├── voucher_template_repo.go
│   │   ├── reconciliation_repo.go
│   │   ├── period_repo.go
│   │   ├── party_repo.go
│   │   ├── company_repo.go
│   │   ├── opening_balance_repo.go
│   │   ├── exchange_rate_repo.go
│   │   ├── asset_depreciation_repo.go
│   │   ├── payroll_repo.go
│   │   ├── reimbursement_repo.go
│   │   ├── reimbursement_attachment_repo.go
│   │   ├── reimbursement_invoice_link_repo.go
│   │   ├── advance_receipt_repo.go
│   │   ├── advance_payment_repo.go
│   │   ├── advance_allocation_repo.go
│   │   ├── approval_repo.go
│   │   ├── audit_repo.go
│   │   ├── social_config_repo.go
│   │   ├── bank_journal_repo.go
│   │   ├── ai_feedback_log_repo.go
│   │   ├── bus_doc_mapping_repository.go
│   │   └── settlement_log_repo.go
│   │
│   └── event/                       # 事件总线
│       ├── eventbus.go              # 事件总线实现
│       ├── events.go                # 事件定义
│       ├── audit_subscriber.go      # 审计日志订阅者
│       └── settlement_subscriber.go # 结算日志订阅者
│
├── pkg/
│   ├── database/
│   │   ├── postgres.go             # pgxpool 连接管理
│   │   └── redis.go                # Redis 客户端
│   ├── jwt/
│   │   └── utils.go                # JWT 签发与验证
│   └── utils/
│       └── ptr.go                  # 通用工具函数
│
├── migrations/                      # 64 个 SQL 迁移文件
├── frontend/                        # Vue 3 SPA
│   ├── src/
│   │   ├── api/                    # Axios 封装 + 模块化 API
│   │   │   ├── request.ts          # HTTP 请求拦截器
│   │   │   └── modules/            # 业务模块 API
│   │   │       ├── auth.ts
│   │   │       ├── account.ts
│   │   │       ├── bank.ts
│   │   │       ├── invoice.ts
│   │   │       ├── voucher.ts
│   │   │       ├── payment.ts
│   │   │       ├── period.ts
│   │   │       ├── reconciliation.ts
│   │   │       ├── classification-rule.ts
│   │   │       ├── expense-invoice.ts
│   │   │       ├── reimbursement.ts
│   │   │       ├── payroll.ts
│   │   │       ├── advance_payment.ts
│   │   │       ├── advance_receipt.ts
│   │   │       ├── advance_allocation.ts
│   │   │       ├── aging.ts
│   │   │       ├── credit_control.ts
│   │   │       ├── opening-balance.ts
│   │   │       ├── asset.ts
│   │   │       ├── ar_invoice.ts
│   │   │       ├── ap_invoice.ts
│   │   │       ├── bank_txn_review.ts
│   │   │       └── clearData.ts
│   │   ├── components/             # Vue 组件
│   │   │   ├── app/                # 应用级组件
│   │   │   │   ├── AppLayout.vue
│   │   │   │   ├── AppHeader.vue
│   │   │   │   ├── AppSidebar.vue
│   │   │   │   └── PageLayout.vue
│   │   │   ├── business/           # 业务组件
│   │   │   │   ├── AccountSelector.vue
│   │   │   │   ├── AmountInput.vue
│   │   │   │   ├── DocStatusTag.vue
│   │   │   │   ├── PartySelector.vue
│   │   │   │   └── PeriodPicker.vue
│   │   │   └── check/             # 检测组件
│   │   │       ├── BlockingGuard.vue
│   │   │       ├── CheckResultPanel.vue
│   │   │       ├── CheckStatusBadge.vue
│   │   │       └── CheckSummaryCard.vue
│   │   ├── config/                 # 应用配置
│   │   │   ├── app.config.ts
│   │   │   └── menu.config.ts
│   │   ├── directives/             # 自定义指令
│   │   │   └── permission.ts       # v-permission 权限指令
│   │   ├── router/                 # Vue Router
│   │   │   ├── index.ts
│   │   │   └── routes/
│   │   │       └── base.ts
│   │   ├── stores/                 # Pinia Store
│   │   │   ├── app.store.ts
│   │   │   ├── auth.store.ts
│   │   │   └── tenant.store.ts
│   │   ├── styles/                 # SCSS 样式
│   │   ├── types/                  # TypeScript 类型
│   │   │   ├── models/
│   │   │   ├── api.ts
│   │   │   ├── enums.ts
│   │   │   └── router.ts
│   │   └── views/                  # 页面组件
│   │       ├── login/
│   │       ├── dashboard/
│   │       ├── setup/
│   │       ├── bank/
│   │       ├── invoices/
│   │       ├── reconciliation/
│   │       ├── reconciliation-bank/
│   │       ├── voucher/
│   │       ├── period/
│   │       ├── reimbursement/
│   │       ├── advance-payments/
│   │       ├── advance-receipts/
│   │       ├── ap-invoices/
│   │       ├── ar-invoices/
│   │       ├── payroll/
│   │       ├── asset/
│   │       ├── depreciation/
│   │       ├── reports/
│   │       ├── opening-balance/
│   │       └── error/
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
│
├── tests/
│   ├── run.sh
│   └── test_api.py                 # Python 集成测试 (807 行)
│
├── docker-compose.yml              # PostgreSQL + Redis 本地环境
├── Dockerfile                      # Go 多阶段构建
├── Makefile                        # 构建命令
├── go.mod / go.sum               # Go 模块依赖
└── README.md                       # 项目说明文档
```

---

## 4. 后端模块详解

### 4.1 核心模块概览

| 模块 | 状态 | 说明 |
|------|------|------|
| F0 基础设施 | ✅ | JWT 认证 + RLS 多租户 + 审计日志 |
| F1 账套初始化 | ✅ | 公司创建 + 科目表 + 往来单位 + 会计期间 |
| F2 发票管理 | ✅ | 销售发票 + 采购发票 + 费用发票 + Excel 导入 |
| F3 银行流水 | ✅ | 多格式导入 + 智能分类 + 重复检测 |
| F4 凭证状态机 | ✅ | 草稿→提交→核准/驳回→过账 + 红字冲销 |
| F5 凭证自动生成 | ✅ | 银行流水/收付款/发票 → 凭证智能转换 |
| F6 固定资产折旧 | ✅ | 折旧计划 + 执行 + 凭证生成 |
| F7 银行对账 | ✅ | 5 级匹配策略 + 余额调节表 |
| F8 财务报表 | ✅ | 试算平衡表 + 利润表 + 资产负债表 |
| F9 审批流 | ✅ | 多级审批 + 金额阈值 + 模板绑定 |
| F10 会计期间结账 | ✅ | 结账前检查 + 结转凭证 + 期间开关 |

### 4.2 Handler 层

Handler 层负责 HTTP 请求处理，共 22 个 Handler 文件。

#### 认证模块 (AuthHandler)

| 文件 | 路由 | 说明 |
|------|------|------|
| `auth_handler.go` | `/api/v1/auth/login`, `/api/v1/auth/logout` | 登录认证 |

#### 基础设置模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `account_handler.go` | `/api/v1/accounts` | 科目表管理 |
| `party_handler.go` | `/api/v1/parties` | 往来单位管理 |
| `bank_handler.go` | `/api/v1/bank-accounts` | 银行账户管理 |
| `company_handler.go` | `/api/v1/company-settings` | 公司设置 |
| `setup_handler.go` | `/api/v1/account-setup` | 账套初始化向导 |
| `exchange_rate_handler.go` | `/api/v1/exchange-rates` | 汇率管理 |
| `opening_balance_handler.go` | `/api/v1/opening-balances` | 开办余额 |

#### 银行流水模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `bank_transaction_handler.go` | `/api/v1/bank-transactions` | 流水导入/分类/匹配 |
| `bank_txn_review_handler.go` | `/api/v1/bank-transactions/review-*` | 流水审核工作台 |

#### 发票模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `invoice_handler.go` | `/api/v1/invoices` | 销售/采购发票 |
| `ar_invoice_handler.go` | `/api/v1/ar-invoices` | 应收发票 |
| `ap_invoice_handler.go` | `/api/v1/ap-invoices` | 应付发票 |
| `expense_invoice_handler.go` | `/api/v1/expense-invoices` | 费用发票 |
| `expense_invoice_import_handler.go` | `/api/v1/expense-invoices/import` | 费用发票导入 |
| `expense_invoice_ocr_handler.go` | `/api/v1/expense-invoices/ocr` | 费用发票 OCR |

#### 凭证模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `voucher_handler.go` | `/api/v1/vouchers` | 凭证 CRUD + 状态机 |
| `voucher_template_handler.go` | `/api/v1/voucher-templates` | 凭证模板 |
| `voucher_auto_generate_handler.go` | `/api/v1/*/generate-voucher` | 凭证自动生成 |

#### 收付款模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `payment_handler.go` | `/api/v1/payment-entries` | 收付款单 |
| `advance_receipt_handler.go` | `/api/v1/advance-receipts` | 预收账款 |
| `advance_payment_handler.go` | `/api/v1/advance-payments` | 预付款项 |
| `advance_allocation_handler.go` | `/api/v1/advance-allocations` | 预收预付核销 |

#### 对账模块

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `reconciliation_handler.go` | `/api/v1/reconciliation` | 发票/收付款核销 |
| `bank_reconciliation_handler.go` | `/api/v1/bank-reconciliation` | 银企对账 |

#### 审批与审核

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `approval_handler.go` | `/api/v1/approvals`, `/api/v1/approval-flows` | 审批流 |

#### 报表与期间

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `report_handler.go` | `/api/v1/reports` | 财务报表 |
| `period_handler.go` | `/api/v1/periods` | 会计期间 + 结账 |

#### 薪资与报销

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `payroll_handler.go` | `/api/v1/payroll` | 薪资管理 |
| `reimbursement_handler.go` | `/api/v1/reimbursements` | 报销管理 |
| `reimbursement_attachment_handler.go` | `/api/v1/reimbursements/:id/attachments` | 报销附件 |
| `reimbursement_invoice_link_handler.go` | `/api/v1/reimbursements/:id/invoices` | 报销发票关联 |

#### 固定资产

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `asset_depreciation_handler.go` | `/api/v1/assets`, `/api/v1/depreciation` | 折旧管理 |

#### 其他

| Handler | 路由前缀 | 说明 |
|---------|---------|------|
| `audit_handler.go` | `/api/v1/audit-logs` | 审计日志 |
| `dashboard_handler.go` | `/api/v1/dashboard/stats` | 仪表盘统计 |
| `health.go` | `/health` | 健康检查 |
| `clear_data_handler.go` | `/api/v1/setup/clear-*` | 数据清理 |
| `credit_control_handler.go` | `/api/v1/credit-control` | 信用控制 |
| `aging_handler.go` | `/api/v1/aging-analysis` | 账龄分析 |
| `classification_rule_handler.go` | `/api/v1/classification-rules` | 分类规则 |

### 4.3 Service 层

#### 核心 Service

| Service | 依赖 | 职责 |
|---------|------|------|
| `auth_service.go` | UserRepository | JWT 签发、密码验证 |
| `account_service.go` | AccountRepository | 科目表 CRUD、树形结构 |
| `bank_service.go` | BankRepository | 银行账户管理 |
| `bank_transaction_service.go` | BankTransactionRepository, ClassificationRuleService | 流水业务处理 |
| `invoice_service.go` | InvoiceRepository | 发票 CRUD、状态管理 |
| `voucher_service.go` | JournalRepository, VoucherTemplateService | 凭证业务逻辑 |
| `voucher_state_machine.go` | JournalRepository, AuditRepository, GLEntryRepository | 凭证状态流转 |
| `voucher_auto_generate_service.go` | 8 个依赖 ★ | 银行流水/发票→凭证 |
| `reconciliation_service.go` | BankTransactionRepository, PaymentRepository, InvoiceRepository | 核销处理 |
| `bank_reconciliation_service.go` | BankTransactionRepository, JournalRepository | 银企对账 |
| `period_service.go` | PeriodRepository, JournalRepository, GLEntryRepository | 期间管理、结账 |
| `report_service.go` | GLEntryRepository, OpeningBalanceRepository | 财务报表计算 |

#### 业务 Service

| Service | 职责 |
|---------|------|
| `ar_invoice_service.go` | 应收发票业务逻辑 |
| `ap_invoice_service.go` | 应付发票业务逻辑 |
| `expense_invoice_service.go` | 费用发票业务逻辑 |
| `expense_invoice_import_service.go` | 费用发票导入业务逻辑 |
| `payment_service.go` | 收付款业务逻辑 |
| `payment_state_machine.go` | 收付款状态机 |
| `voucher_template_service.go` | 凭证模板业务逻辑 |
| `approval_service.go` | 审批流业务逻辑 |
| `party_service.go` | 往来单位业务逻辑 |
| `setup_service.go` | 账套初始化业务逻辑 |
| `opening_balance_service.go` | 开办余额业务逻辑 |
| `exchange_rate_service.go` | 汇率业务逻辑 |
| `asset_depreciation_service.go` | 折旧业务逻辑 |
| `payroll_service.go` | 薪资业务逻辑 |
| `reimbursement_service.go` | 报销业务逻辑 |
| `advance_receipt_service.go` | 预收账款业务逻辑 |
| `advance_payment_service.go` | 预付款项业务逻辑 |
| `advance_allocation_service.go` | 预收预付核销业务逻辑 |
| `classification_rule_service.go` | 分类规则业务逻辑 |
| `audit_service.go` | 审计日志业务逻辑 |
| `credit_control_service.go` | 信用控制业务逻辑 |
| `aging_service.go` | 账龄分析业务逻辑 |

### 4.4 Repository 层

| Repository | 核心表 | 职责 |
|-----------|--------|------|
| `user_repo.go` | `users` | 用户数据访问 |
| `account_repo.go` | `accounts` | 科目表数据访问（嵌套集模型） |
| `bank_repo.go` | `bank_accounts` | 银行账户数据访问 |
| `bank_transaction_repo.go` | `bank_transactions` | 银行流水数据访问 |
| `journal_repo.go` | `journal_entries`, `journal_entry_lines` | 凭证主表/分录数据访问 |
| `gl_entry_repo.go` | `gl_entries` | 总账分录数据访问 |
| `invoice_repo.go` | `sales_invoices` | 销售发票数据访问 |
| `ar_invoice_repo.go` | `ar_invoices` | 应收发票数据访问 |
| `ap_invoice_repo.go` | `ap_invoices` | 应付发票数据访问 |
| `expense_invoice_repo.go` | `expense_invoices` | 费用发票数据访问 |
| `payment_repo.go` | `payment_entries` | 收付款单数据访问 |
| `party_repo.go` | `parties` | 往来单位数据访问 |
| `period_repo.go` | `accounting_periods` | 会计期间数据访问 |
| `reconciliation_repo.go` | `reconciliation_pairs` | 核销数据访问 |
| `voucher_template_repo.go` | `voucher_templates` | 凭证模板数据访问 |
| `approval_repo.go` | `approval_flows`, `approval_tasks` | 审批流数据访问 |
| `audit_repo.go` | `audit_logs` | 审计日志数据访问 |
| `opening_balance_repo.go` | `opening_balances` | 开办余额数据访问 |
| `classification_rule_repo.go` | `classification_rules` | 分类规则数据访问 |
| `payroll_repo.go` | `payroll` | 薪资数据访问 |
| `reimbursement_repo.go` | `reimbursements` | 报销数据访问 |
| `advance_receipt_repo.go` | `advance_receipts` | 预收账款数据访问 |
| `advance_payment_repo.go` | `advance_payments` | 预付款项数据访问 |
| `advance_allocation_repo.go` | `advance_allocations` | 预收预付核销数据访问 |

---

## 5. 前端模块详解

### 5.1 项目结构

```
frontend/src/
├── main.ts                    # Vue 应用入口
├── App.vue                    # 根组件
├── api/
│   ├── request.ts             # Axios 实例配置
│   └── modules/               # API 模块封装
│       ├── auth.ts            # 认证 API
│       ├── account.ts          # 科目 API
│       ├── bank.ts             # 银行 API
│       └── ...
├── components/
│   ├── app/                   # 应用级组件
│   │   ├── AppLayout.vue      # 主布局
│   │   ├── AppHeader.vue       # 顶部导航
│   │   └── AppSidebar.vue     # 侧边栏
│   ├── business/              # 业务组件
│   │   ├── AccountSelector.vue # 科目选择器
│   │   ├── AmountInput.vue    # 金额输入
│   │   └── ...
│   └── check/                 # 检测组件
│       └── ...
├── config/
│   ├── app.config.ts          # 应用配置
│   └── menu.config.ts         # 菜单配置
├── directives/
│   └── permission.ts          # 权限指令
├── router/
│   └── index.ts               # 路由配置
├── stores/
│   ├── auth.store.ts          # 认证状态
│   ├── app.store.ts           # 应用状态
│   └── tenant.store.ts        # 租户状态
├── styles/
│   ├── variables.scss         # 样式变量
│   └── index.scss             # 全局样式
├── types/
│   ├── models/                # 数据模型
│   ├── api.ts                 # API 类型
│   └── enums.ts               # 枚举
└── views/                     # 页面组件
    ├── login/
    ├── dashboard/
    ├── setup/
    ├── bank/
    ├── invoices/
    ├── reconciliation/
    ├── voucher/
    ├── period/
    └── ...
```

### 5.2 路由结构

| 路径 | 页面 | 角色 |
|------|------|------|
| `/login` | LoginView | 公开 |
| `/` | DashboardView | 全部 |
| `/setup/company` | SetupWizard | admin |
| `/setup/accounts` | AccountChart | admin, agent |
| `/setup/bank-accounts` | BankAccountList | admin, cashier, agent |
| `/setup/parties` | PartyList | admin, accountant_ar, agent |
| `/setup/rules` | RuleLibrary | admin |
| `/bank/import` | ImportView | cashier, admin, agent |
| `/bank/workbench` | CashierWorkbench | cashier, admin, agent |
| `/invoices` | InvoiceList | accountant_ar, admin, agent |
| `/reconciliation/*` | PreCheckView, MatchView, ManualMatch | accountant_ar, admin |
| `/vouchers` | VoucherList | admin, agent |
| `/vouchers/create` | VoucherEdit | admin, agent |
| `/vouchers/review` | ReviewWorkbench | admin |
| `/bank-reconciliation/*` | MatchingView, BalanceSheet | cashier, admin |
| `/period/health-check` | HealthCheck | admin |
| `/period/reports` | FinancialReports | admin, boss |
| `/payroll/*` | PayrollList, PayrollForm | admin |
| `/reimbursement/*` | ReimbursementList, ReimbursementForm | employee, admin |

### 5.3 API 层

#### 统一请求封装 (request.ts)

```typescript
// 核心功能
- Axios 实例配置 (baseURL: /api/v1, timeout: 30000)
- 请求拦截器: JWT Token 注入
- 响应拦截器: 统一错误处理 (401→登录, 403→权限提示)
```

#### 业务 API 模块

| 模块 | 文件 | 说明 |
|------|------|------|
| 认证 | `auth.ts` | 登录、登出 |
| 科目 | `account.ts` | 科目树、科目列表 |
| 银行 | `bank.ts` | 银行账户、银行流水 |
| 发票 | `invoice.ts` | 发票 CRUD、导入 |
| 凭证 | `voucher.ts` | 凭证 CRUD、状态流转 |
| 收付款 | `payment.ts` | 收付款单 |
| 期间 | `period.ts` | 期间管理、结账 |
| 核销 | `reconciliation.ts` | 核销处理 |

### 5.4 状态管理 (Pinia)

#### auth.store.ts

```typescript
// 核心状态
- token: JWT 令牌
- user: 当前用户信息 { id, name, email, role }
- permissions: 权限列表

// 核心方法
- login(credentials): 登录
- logout(): 登出
- hasPermission(permission): 权限检查
- hasRole(role): 角色检查
```

#### tenant.store.ts

```typescript
// 核心状态
- currentTenantId: 当前租户 ID
- currentCompany: 当前公司信息
- tenantList: 租户列表（代账会计用）

// 核心方法
- switchTenant(tenantId): 切换租户 → 重置所有业务 Store
```

#### app.store.ts

```typescript
// 核心状态
- sidebarCollapsed: 侧边栏折叠状态
- globalLoading: 全局加载状态
- theme: 主题配置
```

---

## 6. 数据库设计

### 6.1 核心表结构

#### 用户与认证

| 表名 | 说明 |
|------|------|
| `tenants` | 租户表 |
| `users` | 用户表 |
| `companies` | 公司表 |

#### 基础档案

| 表名 | 说明 |
|------|------|
| `accounts` | 科目表（嵌套集 lft/rgt） |
| `parties` | 往来单位（客户/供应商） |
| `bank_accounts` | 银行账户 |

#### 业务单据

| 表名 | 说明 |
|------|------|
| `sales_invoices` | 销售发票 |
| `ar_invoices` | 应收发票 |
| `ap_invoices` | 应付发票 |
| `expense_invoices` | 费用发票 |
| `payment_entries` | 收付款单 |
| `advance_receipts` | 预收账款 |
| `advance_payments` | 预付款项 |
| `reimbursements` | 报销单 |

#### 凭证与账务

| 表名 | 说明 |
|------|------|
| `journal_entries` | 凭证主表 |
| `journal_entry_lines` | 凭证分录 |
| `gl_entries` | 总账分录 |
| `voucher_templates` | 凭证模板 |
| `opening_balances` | 开办余额 |

#### 对账与审核

| 表名 | 说明 |
|------|------|
| `bank_transactions` | 银行流水 |
| `reconciliation_pairs` | 核销配对 |
| `approval_flows` | 审批流程 |
| `approval_tasks` | 审批任务 |

#### 固定资产

| 表名 | 说明 |
|------|------|
| `fixed_assets` | 固定资产卡片 |
| `asset_depreciation_runs` | 折旧运行记录 |

#### 配置与日志

| 表名 | 说明 |
|------|------|
| `accounting_periods` | 会计期间 |
| `classification_rules` | 分类规则 |
| `exchange_rates` | 汇率 |
| `audit_logs` | 审计日志 |

### 6.2 多租户隔离 (RLS)

```sql
-- 启用行级安全
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;

-- 创建租户隔离策略
CREATE POLICY tenant_isolation ON <table> FOR ALL
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 中间件在每个请求中执行
SET app.current_tenant = '<tenant_id>'
```

---

## 7. 核心数据流

### 7.1 银行流水 → 凭证 → 过账 全链路

```
[银行流水导入]
    │
    ├── 格式解析 (Excel/CSV/Camt053/MT940)
    ├── 重复检测
    └── 存储 bank_transactions
           │
           ▼
[智能分类] (ClassificationRuleService)
    │
    ├── 规则引擎匹配
    └── 更新 bank_transactions.category
           │
           ▼
[自动生成凭证] (VoucherAutoGenerateService)
    │
    ├── 根据 category 匹配凭证模板
    ├── 创建 journal_entries + journal_entry_lines
    └── 关联 source_doc_type/id
           │
           ▼
[提交审批] (ApprovalService)
    │
    ├── 创建 approval_tasks
    └── 更新凭证状态 → submitted
           │
           ▼
[审批核准] (ApprovalService)
    │
    ├── 更新凭证状态 → approved
    └── 触发 GL 过账
           │
           ▼
[GL 过账] (GLEntryRepository)
    │
    └── 创建 gl_entries，实时更新科目余额
           │
           ▼
[银企对账] (BankReconciliationService)
    │
    ├── 5 级智能匹配
    └── 生成余额调节表
```

### 7.2 发票 → 核销 → 凭证

```
[发票录入/导入]
    │
    └── 存储 sales_invoices (draft)
           │
           ▼
[发票确认] (InvoiceService)
    │
    └── 更新状态 → verified
           │
           ▼
[手工/自动核销] (ReconciliationService)
    │
    ├── 匹配发票与收付款
    └── 创建 reconciliation_pairs
           │
           ▼
[生成凭证] (VoucherAutoGenerateService)
    │
    └── 从发票/收付款生成凭证
```

### 7.3 期末结账流程

```
[结账前检查] (PeriodService.PreCloseCheck)
    │
    ├── 凭证借贷平衡检查
    ├── 凭证编号连续性检查
    ├── 固定资产折旧检查
    ├── 银行日记账一致性检查
    └── 往来核销完成度检查
           │
           ▼
[生成结转凭证]
    │
    ├── 损益结转
    ├── 汇兑损益结转
    └── 其他自动结转
           │
           ▼
[执行结账] (PeriodService.Close)
    │
    ├── 更新期间状态 → closed
    └── 锁定期间数据
```

---

## 8. API 路由总览

### 8.1 公开接口

| 接口 | 方法 | Handler | 说明 |
|------|------|---------|------|
| `/health` | GET | HealthHandler | 健康检查 |
| `/api/v1/auth/login` | POST | AuthHandler | 用户登录 |
| `/api/v1/auth/logout` | POST | AuthHandler | 用户登出 |

### 8.2 认证接口

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/audit-logs` | AuditHandler | 审计日志查询 |

### 8.3 科目与设置

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/accounts` | AccountHandler | 科目表 |
| `/api/v1/parties` | PartyHandler | 往来单位 |
| `/api/v1/bank-accounts` | BankHandler | 银行账户 |
| `/api/v1/account-setup` | SetupHandler | 账套初始化 |
| `/api/v1/company-settings` | CompanyHandler | 公司设置 |
| `/api/v1/exchange-rates` | ExchangeRateHandler | 汇率 |
| `/api/v1/opening-balances` | OpeningBalanceHandler | 开办余额 |

### 8.4 银行流水

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/bank-transactions` | BankTransactionHandler | 银行流水 |
| `/api/v1/bank-transactions/review-*` | BankTxnReviewHandler | 流水审核 |

### 8.5 发票管理

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/invoices` | InvoiceHandler | 销售/采购发票 |
| `/api/v1/ar-invoices` | ArInvoiceHandler | 应收发票 |
| `/api/v1/ap-invoices` | ApInvoiceHandler | 应付发票 |
| `/api/v1/expense-invoices` | ExpenseInvoiceHandler | 费用发票 |

### 8.6 凭证管理

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/vouchers` | VoucherHandler | 凭证 CRUD + 状态机 |
| `/api/v1/voucher-templates` | VoucherTemplateHandler | 凭证模板 |
| `/api/v1/*/generate-voucher` | VoucherAutoGenerateHandler | 凭证自动生成 |

### 8.7 收付款

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/payment-entries` | PaymentHandler | 收付款单 |
| `/api/v1/advance-receipts` | AdvanceReceiptHandler | 预收账款 |
| `/api/v1/advance-payments` | AdvancePaymentHandler | 预付款项 |
| `/api/v1/advance-allocations` | AdvanceAllocationHandler | 预收预付核销 |

### 8.8 对账

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/reconciliation` | ReconciliationHandler | 发票/收付款核销 |
| `/api/v1/bank-reconciliation` | BankReconciliationHandler | 银企对账 |

### 8.9 审批

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/approvals` | ApprovalHandler | 审批任务 |
| `/api/v1/approval-flows` | ApprovalHandler | 审批流程 |

### 8.10 报表与期间

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/periods` | PeriodHandler | 会计期间 + 结账 |
| `/api/v1/reports` | ReportHandler | 财务报表 |

### 8.11 其他

| 路由前缀 | Handler | 说明 |
|---------|---------|------|
| `/api/v1/payroll` | PayrollHandler | 薪资管理 |
| `/api/v1/reimbursements` | ReimbursementHandler | 报销管理 |
| `/api/v1/assets`, `/api/v1/depreciation` | AssetDepreciationHandler | 固定资产折旧 |
| `/api/v1/classification-rules` | ClassificationRuleHandler | 分类规则 |
| `/api/v1/dashboard/stats` | DashboardHandler | 仪表盘统计 |

---

## 9. 关键类与函数说明

### 9.1 后端关键类型

#### 凭证模型 (journal.go)

```go
type JournalEntry struct {
    ID              uuid.UUID       // 凭证ID
    TenantID        uuid.UUID       // 租户ID
    PeriodNo        string          // 会计期间编号
    VoucherNo       string          // 凭证编号
    VoucherDate     time.Time       // 凭证日期
    DocStatus       int             // 文档状态 (0=草稿, 1=已提交, 2=已核准, 3=已驳回, 4=已取消, 5=已红冲)
    SourceDocType   string          // 来源单据类型
    SourceDocID     uuid.UUID       // 来源单据ID
    SourceDocNo     string          // 来源单据编号
    AttachmentCount int             // 附件数
    EntryBy         string          // 制单人
    ApprovedBy      string          // 核准人
    Remark          string          // 备注
    CreatedAt       time.Time       // 创建时间
    UpdatedAt       time.Time       // 更新时间
}

type JournalEntryLine struct {
    ID               uuid.UUID  // 分录ID
    JournalEntryID   uuid.UUID  // 凭证ID
    AccountID        uuid.UUID  // 科目ID
    AccountCode      string     // 科目编码
    AccountName      string     // 科目名称
    Debit            decimal.Decimal  // 借方金额
    Credit           decimal.Decimal  // 贷方金额
    // ... 其他字段
}
```

#### 银行流水模型 (bank_transaction.go)

```go
type BankTransaction struct {
    ID                uuid.UUID       // 流水ID
    TenantID          uuid.UUID       // 租户ID
    BankAccountID     uuid.UUID       // 银行账户ID
    TransactionDate   time.Time       // 交易日期
    Amount            decimal.Decimal // 交易金额
    Balance           decimal.Decimal // 余额
    CounterpartyName  string          // 对方名称
    CounterpartyBank  string          // 对方银行
    CounterpartyAcct  string          // 对方账号
    Remark            string          // 摘要
    Category          string          // 分类
    Status            string          // 状态 (pending/classified/matched)
    VoucherID         *uuid.UUID      // 关联凭证ID
    CreatedAt         time.Time       // 创建时间
}
```

#### 发票模型 (invoice.go)

```go
type SalesInvoice struct {
    ID                uuid.UUID       // 发票ID
    TenantID          uuid.UUID       // 租户ID
    InvoiceNo         string          // 发票号
    InvoiceDate       time.Time       // 发票日期
    CustomerID        uuid.UUID       // 客户ID
    Amount            decimal.Decimal // 价税合计
    TaxAmount         decimal.Decimal // 税额
    NetAmount         decimal.Decimal // 金额（不含税）
    Currency          string          // 币种
    Status            string          // 状态 (draft/verified/paid/cancelled)
    DocStatus         int             // 单据状态
    VoucherID         *uuid.UUID      // 关联凭证ID
    CreatedAt         time.Time       // 创建时间
}
```

### 9.2 凭证状态机 (voucher_state_machine.go)

```go
// 状态常量
const (
    STATUS_DRAFT     = 0  // 草稿
    STATUS_SUBMITTED = 1  // 已提交
    STATUS_APPROVED  = 2  // 已核准
    STATUS_REJECTED  = 3  // 已驳回
    STATUS_CANCELLED = 4  // 已取消
    STATUS_REVERSED  = 5  // 已红冲
)

// 状态流转
// 草稿(0) → 提交 → 待审批 → 核准(2) → 过账
//                ↓            │
//              驳回          红冲(5)
//                ↓            │
//              草稿(0)      取消(4)
```

### 9.3 前端关键类型

#### API 响应格式 (api.ts)

```typescript
interface ApiResponse<T = unknown> {
    code: number      // 0=成功, 非0=业务错误
    message: string
    data: T
}

interface PageQuery {
    page: number
    pageSize: number
    sort?: string
}

interface PageResult<T> {
    list: T[]
    total: number
    page: number
    pageSize: number
}
```

#### 科目模型 (models/account.ts)

```typescript
interface Account {
    id: string
    code: string           // 科目编码 (如 "1001", "1002")
    name: string           // 科目名称
    fullName: string       // 完整名称
    accountType: AccountType  // 科目类型
    isGroup: boolean       // 是否为分组（汇总）科目
    isLedger: boolean     // 是否可记账
    lft: number           // 嵌套集左值
    rgt: number           // 嵌套集右值
    level: number        // 层级深度
    parentId: string | null  // 父级ID
    children?: Account[] // 子科目
}
```

### 9.4 关键工具函数

#### decimal 处理规范 (必须遵守)

```go
// 所有金额必须使用 shopspring/decimal
// 传输到前端时使用 .CoefficientInt64() 而非 .Int64()
amount := entry.Amount.CoefficientInt64()
// 不要用 entry.Amount.Int64() — 会丢失小数位
```

#### JWT Token 结构

```go
// Claims: sub(user_id), tenant_id, username, role, exp, iat
type Claims struct {
    UserID   string `json:"sub"`
    TenantID string `json:"tenant_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

---

## 10. 依赖关系

### 10.1 后端 Service 层依赖图

```
VoucherAutoGenerateService ★ (8 个依赖)
  ├── JournalRepository
  ├── GLEntryRepository
  ├── BankTransactionRepository
  ├── InvoiceRepository
  ├── AccountRepository
  ├── ClassificationRuleService
  ├── VoucherTemplateService
  └── ApprovalService

PeriodService ★ (5 个依赖)
  ├── PeriodRepository
  ├── JournalRepository
  ├── GLEntryRepository
  ├── AccountRepository
  └── AssetDepreciationRepository

BankReconciliationService
  ├── BankTransactionRepository
  ├── JournalRepository
  ├── BankRepository
  └── GLEntryRepository

ReconciliationService
  ├── BankTransactionRepository
  ├── PaymentRepository
  ├── ArInvoiceRepository
  ├── ApInvoiceRepository
  ├── InvoiceRepository
  ├── ReconciliationRepository
  └── JournalRepository

ReportService
  ├── GLEntryRepository
  ├── OpeningBalanceRepository
  ├── AccountRepository
  └── PeriodRepository
```

### 10.2 前端模块依赖

```
main.ts
  ├── router/index.ts
  ├── stores/
  ├── ElementPlus
  └── App.vue

router/index.ts
  ├── stores/auth.store (路由守卫)
  └── views/* (页面组件)

stores/auth.store
  └── types/models/user

api/request.ts
  └── stores/auth.store (JWT 注入)

views/* (页面组件)
  ├── api/modules/* (API 调用)
  ├── stores/* (状态管理)
  ├── components/business/* (业务组件)
  └── types/*
```

### 10.3 数据库迁移依赖顺序

```
001_init.sql                 ← 基础表: tenants, users, companies
002_journal_gl.sql           ← 凭证 + 总账
003_invoice_payment.sql      ← 发票 + 付款
004_bank.sql                 ← 银行账户
005_asset.sql                ← 固定资产
006_budget.sql               ← 预算
007_audit.sql                ← 审计日志
008_rls_force.sql            ← RLS 策略
009_app_user.sql             ← 应用用户
010_account_setup.sql        ← 科目表
011_depreciation_run.sql     ← 折旧
012_*_classification_rules.sql ← 分类规则
012_*_voucher_template.sql   ← 凭证模板
013_voucher_state_machine.sql ← 凭证状态机
014_bank_transactions.sql    ← 银行流水
015_opening_balance.sql      ← 开办余额
016_exchange_rates.sql       ← 汇率
017_bank_reconciliation.sql  ← 银企对账
018_reconciliation.sql      ← 核销
019_approval.sql             ← 审批流
020_*_approval_bind.sql      ← 模板审批绑定
021_voucher_state_transitions.sql  ← 状态变更日志
022_seed_data.sql           ← 种子数据
... (后续迁移持续增量)
```

---

## 11. 项目运行方式

### 11.1 环境要求

| 组件 | 版本要求 |
|------|---------|
| Go | 1.22+ |
| Node.js | 18+ |
| pnpm | 最新版 |
| Docker & Docker Compose | 最新版 |
| PostgreSQL | 15 |
| Redis | 7 |

### 11.2 启动基础设施 (Docker)

```bash
# 启动 PostgreSQL + Redis
docker compose up -d

# 验证服务
docker compose ps
```

### 11.3 数据库初始化

```bash
# 执行所有迁移文件
for f in migrations/*.sql; do
    psql -h localhost -U huihua -d huihua_finance -f "$f"
done
```

### 11.4 启动后端

```bash
# 复制环境变量
cp .env.example .env

# 编辑 .env 配置数据库连接
# database.host=localhost
# database.port=5432
# database.user=huihua
# database.password=hfpwd
# database.dbname=huihua_finance

# 编译并启动
go build -o /tmp/huihua-api ./cmd/api
/tmp/huihua-api

# API 服务: http://localhost:8080
# 健康检查: http://localhost:8080/health
```

### 11.5 启动前端

```bash
cd frontend

# 安装依赖
pnpm install

# 开发模式（Mock 开启）
pnpm dev
# 前端: http://localhost:3002

# 开发模式（连接真实后端）
VITE_ENABLE_MOCK=false pnpm dev

# 生产构建
pnpm build
```

### 11.6 测试用户

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `testuser` | `test123` | admin |

### 11.7 构建命令 (Makefile)

```bash
# 全部测试
make test

# 后端测试
make test-go
make test-go-unit

# 前端测试
make test-frontend-unit

# 构建全部
make build
make build-go
make build-frontend
```

### 11.8 Docker 部署

```bash
# 构建镜像
docker build -t huihua-finance-api .

# 运行容器
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  huihua-finance-api
```

---

## 附录：架构决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| ORM 选型 | 无 ORM，原生 SQL | 财务系统对 SQL 精确度要求极高 |
| 多租户方案 | PostgreSQL RLS | 原生支持，维护成本最低 |
| 金额类型 | `shopspring/decimal` | 避免浮点数精度问题 |
| HTTP 框架 | Fiber v2 | 高性能，类 Express API |
| 前端 Mock | MSW | 真实拦截 Service Worker，贴近生产 |
| 科目表模型 | 嵌套集 (lft/rgt) | 高效树形查询，无需递归 CTE |

---

*文档生成时间：2026-06-11*
*项目版本：v1.0*
