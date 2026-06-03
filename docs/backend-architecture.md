# 慧财智能财务平台 - 后端架构文档

---

## 一、项目概述

慧财智能财务平台是一款银行流水驱动的业财一体化系统，以银行流水为核心入口，自动生成凭证、自动审批、银行智能对账。

**项目定位**：面向中小企业的财务SaaS平台

**技术栈**：Go 1.24 + Fiber 2.52 + PostgreSQL 15 + Redis 7

---

## 二、整体架构

采用经典的 Repository-Service-Handler 三层架构模式：

```
┌─────────────────────────────────────────────────────────────────────┐
│                        前端层 (Vue 3)                              │
│                  http://localhost:3002                            │
└───────────────────────┬─────────────────────────────────────────────┘
                        │ HTTP/JSON
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Handler Layer (控制器层)                       │
│  auth_handler | account_handler | voucher_handler | ... (21个)    │
└───────────────────────┬─────────────────────────────────────────────┘
                        │ 依赖注入
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Service Layer (业务逻辑层)                      │
│  auth_service | voucher_state_machine | report_service | ...      │
└───────────────────────┬─────────────────────────────────────────────┘
                        │ 调用
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   Repository Layer (数据访问层)                    │
│  user_repo | journal_repo | gl_entry_repo | ... (18个)            │
└───────────────────────┬─────────────────────────────────────────────┘
                        │ SQL
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Database Layer                              │
│          PostgreSQL 15 (主库)           │    Redis 7 (缓存)       │
│  tenants | users | journals | gl_entries |   Session | RateLimit  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 三、目录结构

| 层级 | 目录 | 职责 | 文件数 |
|------|------|------|--------|
| **入口** | `cmd/api/` | HTTP服务启动、路由注册 | 1 |
| **配置** | `internal/config/` | Viper配置加载 | 1 |
| **控制器** | `internal/handler/` | HTTP请求处理、参数校验 | 21 |
| **中间件** | `internal/middleware/` | JWT认证、Tenant注入、日志 | 3 |
| **模型** | `internal/model/` | 数据库实体定义 | 22 |
| **数据访问** | `internal/repository/` | SQL操作封装 | 18 |
| **业务逻辑** | `internal/service/` | 核心业务规则、事务 | 21 |
| **数据库** | `pkg/database/` | Postgres/Redis连接管理 | 2 |
| **工具** | `pkg/jwt/` | JWT签发与验证 | 2 |
| **工具** | `pkg/utils/` | 通用工具函数 | 1 |

---

## 四、核心模块详解

### 4.1 认证模块 (Auth)

| 文件 | 职责 |
|------|------|
| [auth_handler.go](file:///root/data/disk/huihua-finance/internal/handler/auth_handler.go) | 登录接口 `/api/v1/auth/login` |
| [auth_service.go](file:///root/data/disk/huihua-finance/internal/service/auth_service.go) | JWT签发、密码验证 |
| [user_repo.go](file:///root/data/disk/huihua-finance/internal/repository/user_repo.go) | 用户数据访问 |

**API接口**：
- `POST /api/v1/auth/login` - 用户登录

### 4.2 凭证管理模块 (Voucher)

| 文件 | 职责 |
|------|------|
| [voucher_handler.go](file:///root/data/disk/huihua-finance/internal/handler/voucher_handler.go) | 凭证CRUD、状态流转 |
| [voucher_service.go](file:///root/data/disk/huihua-finance/internal/service/voucher_service.go) | 凭证业务逻辑 |
| [voucher_state_machine.go](file:///root/data/disk/huihua-finance/internal/service/voucher_state_machine.go) | 状态机管理 |
| [journal_repo.go](file:///root/data/disk/huihua-finance/internal/repository/journal_repo.go) | 凭证主表操作 |
| [gl_entry_repo.go](file:///root/data/disk/huihua-finance/internal/repository/gl_entry_repo.go) | 分录明细操作 |

**API接口**：
- `GET /api/v1/vouchers` - 凭证列表
- `POST /api/v1/vouchers` - 创建凭证
- `GET /api/v1/vouchers/:id` - 获取凭证详情
- `PUT /api/v1/vouchers/:id` - 更新凭证
- `DELETE /api/v1/vouchers/:id` - 删除凭证
- `POST /api/v1/vouchers/:id/submit` - 提交审批
- `POST /api/v1/vouchers/:id/approve` - 核准凭证
- `POST /api/v1/vouchers/:id/reject` - 驳回凭证
- `POST /api/v1/vouchers/:id/cancel` - 取消凭证
- `POST /api/v1/vouchers/:id/reverse` - 红冲凭证

**状态流转**：
```
草稿 → 提交 → 待审批 → 核准/驳回 → 已过账
              ↑        |            |
              └────────┘            ↓
                              已红冲/已取消
```

### 4.3 银行流水模块 (Bank Transaction)

| 文件 | 职责 |
|------|------|
| [bank_transaction_handler.go](file:///root/data/disk/huihua-finance/internal/handler/bank_transaction_handler.go) | 流水导入、分类、匹配 |
| [bank_transaction_service.go](file:///root/data/disk/huihua-finance/internal/service/bank_transaction_service.go) | 流水业务处理 |
| [classification_rule_service.go](file:///root/data/disk/huihua-finance/internal/service/classification_rule_service.go) | 智能分类规则引擎 |

**API接口**：
- `GET /api/v1/bank-transactions` - 流水列表
- `POST /api/v1/bank-transactions/import` - 导入流水
- `POST /api/v1/bank-transactions/:id/classify` - 分类流水
- `POST /api/v1/bank-transactions/:id/mark-matched` - 标记已匹配

### 4.4 审批流程模块 (Approval)

| 文件 | 职责 |
|------|------|
| [approval_handler.go](file:///root/data/disk/huihua-finance/internal/handler/approval_handler.go) | 审批任务管理 |
| [approval_service.go](file:///root/data/disk/huihua-finance/internal/service/approval_service.go) | 审批流程逻辑 |

**API接口**：
- `POST /api/v1/approvals/submit` - 提交审批
- `POST /api/v1/approvals/:id/approve` - 批准
- `POST /api/v1/approvals/:id/reject` - 拒绝
- `GET /api/v1/approvals/pending` - 待处理任务
- `POST /api/v1/approval-flows` - 创建审批流程

### 4.5 财务报表模块 (Report)

| 文件 | 职责 |
|------|------|
| [report_handler.go](file:///root/data/disk/huihua-finance/internal/handler/report_handler.go) | 报表接口 |
| [report_service.go](file:///root/data/disk/huihua-finance/internal/service/report_service.go) | 报表计算逻辑 |

**API接口**：
- `GET /api/v1/reports/trial-balance` - 试算平衡表
- `GET /api/v1/reports/income-statement` - 利润表
- `GET /api/v1/reports/balance-sheet` - 资产负债表
- `GET /api/v1/reports/cash-flow` - 现金流量表

### 4.6 会计期间模块 (Period)

| 文件 | 职责 |
|------|------|
| [period_handler.go](file:///root/data/disk/huihua-finance/internal/handler/period_handler.go) | 期间管理接口 |
| [period_service.go](file:///root/data/disk/huihua-finance/internal/service/period_service.go) | 期间业务逻辑 |

**API接口**：
- `GET /api/v1/periods` - 期间列表
- `GET /api/v1/periods/current` - 当前期间
- `GET /api/v1/periods/pre-close-check` - 结账预检
- `POST /api/v1/periods/:period_no/close` - 结账
- `POST /api/v1/periods/:period_no/unclose` - 反结账

---

## 五、技术栈详细说明

| 分类 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.24.0 | 后端主语言 |
| HTTP框架 | Fiber | 2.52.0 | 高性能Web框架 |
| 数据库 | PostgreSQL | 15 | 主数据库，RLS多租户 |
| 缓存 | Redis | 7 | 缓存、会话 |
| 配置 | Viper | 1.21.0 | 环境变量、.env |
| 认证 | JWT | 5.3.1 | HS256算法 |
| 金额 | decimal | 1.4.0 | 精确小数计算 |
| Excel | excelize | 2.10.1 | 文件读写 |
| UUID | google/uuid | 1.6.0 | UUID生成 |

---

## 六、关键设计特点

### 6.1 RLS多租户隔离

通过 PostgreSQL 行级安全策略实现多租户隔离：

```sql
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON <table> FOR ALL
  USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
```

中间件 `tenant.go` 在每个请求中执行：
```go
SET app.current_tenant = '<tenant_id>'
```

### 6.2 状态机模式

凭证状态流转集中管理，支持以下状态：
- 草稿 (draft)
- 待审批 (pending)
- 已核准 (approved)
- 已驳回 (rejected)
- 已取消 (cancelled)
- 已红冲 (reversed)

### 6.3 依赖注入

采用链式依赖注入模式：
```go
repo := repository.NewXxxRepository(db)
svc := service.NewXxxService(repo, otherSvc)
handler := handler.NewXxxHandler(svc)
```

### 6.4 中间件链

```
CORS → Logger → Recover → Auth → Tenant
```

---

## 七、API路由总览

### 7.1 公开接口
| 接口 | 方法 | Handler | 说明 |
|------|------|---------|------|
| `/health` | GET | health.go | 健康检查 |
| `/api/v1/auth/login` | POST | auth_handler.go | 用户登录 |

### 7.2 认证接口

| 模块 | 前缀 | 文件 | 接口数 |
|------|------|------|--------|
| 审计日志 | `/api/v1/audit-logs` | audit_handler.go | 2 |
| 科目 | `/api/v1/accounts` | account_handler.go | 2 |
| 汇率 | `/api/v1/exchange-rates` | exchange_rate_handler.go | 5 |
| 银行账户 | `/api/v1/bank-accounts` | bank_handler.go | 4 |
| 客商档案 | `/api/v1/parties` | party_handler.go | 6 |
| 账套设置 | `/api/v1/account-setup` | setup_handler.go | 2 |
| 折旧 | `/api/v1/assets/depreciation` | asset_depreciation_handler.go | 4 |
| 发票 | `/api/v1/invoices` | invoice_handler.go | 6 |
| 分类规则 | `/api/v1/classification-rules` | classification_rule_handler.go | 7 |
| 银行流水 | `/api/v1/bank-transactions` | bank_transaction_handler.go | 8 |
| 凭证模板 | `/api/v1/voucher-templates` | voucher_template_handler.go | 8 |
| 凭证 | `/api/v1/vouchers` | voucher_handler.go | 10 |
| 开办余额 | `/api/v1/opening-balances` | opening_balance_handler.go | 5 |
| 会计期间 | `/api/v1/periods` | period_handler.go | 5 |
| 核销 | `/api/v1/reconciliation` | reconciliation_handler.go | 4 |
| 银企对账 | `/api/v1/bank-reconciliation` | bank_reconciliation_handler.go | 4 |
| 报表 | `/api/v1/reports` | report_handler.go | 4 |
| 审批 | `/api/v1/approvals` / `/api/v1/approval-flows` | approval_handler.go | 9 |
| 凭证自动生成 | `/api/v1/bank-transactions/:id/generate-voucher` | voucher_auto_generate_handler.go | 3 |

---

## 八、配置说明

### 8.1 .env 文件格式

```bash
# 服务器配置
server.host=0.0.0.0
server.port=8080

# 数据库配置
database.host=localhost
database.port=5432
database.user=huihua_app
database.password=hfpwd_app
database.dbname=huihua_finance
database.sslmode=disable

# Redis配置
redis.host=localhost
redis.port=6379
redis.password=
redis.db=0

# JWT配置
jwt.secret=your-secret-key
jwt.expiry=30m
```

### 8.2 启动方式

```bash
# 开发模式
cd /root/data/disk/huihua-finance
go run ./cmd/api

# 生产构建
go build -o /tmp/huihua-api ./cmd/api
/tmp/huihua-api
```

---

## 九、核心数据流

```
[银行流水导入] → [分类规则匹配] → [自动生成凭证] → [提交审批] → [GL过账]
       ↓                ↓                 ↓               ↓            ↓
  bank_import      classify        auto_generate      approval      gl_entry
       ↓                ↓                 ↓               ↓            ↓
  Excel解析      规则引擎匹配      模板映射          多级审批       科目余额更新
```

---

## 十、数据库迁移

数据库迁移文件位于 `migrations/` 目录，共23个迁移文件：

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
| 022 | seed_data — 种子数据 |

---

## 十一、测试

### 11.1 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/service/...
```

### 11.2 集成测试

集成测试位于 `tests/test_api.py`，使用 Python 编写，共807行测试代码。

---

## 十二、部署

### 12.1 Docker Compose

```bash
docker-compose up -d
```

包含以下服务：
- PostgreSQL 15
- Redis 7-alpine
- 后端API

---

## 十三、版本历史

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-30 | v1.0.0 | 全栈构建通过，90+ API路由接通 |
| 2026-05-28 | v0.9.0 | 财务报表、期间结账、审批流完成 |

---

**文档生成日期**：2026-06-01  
**项目版本**：v1.0.0