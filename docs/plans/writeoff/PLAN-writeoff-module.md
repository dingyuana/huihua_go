# PLAN: 核销模块 — 开发执行计划

> 对应 SPEC: `SPEC-writeoff-module.md`
> 优先级: P0/P1
> 目标: 实现完整的核销模块，支持自动核销、手工核销、反核销功能

---

## 执行顺序

按依赖关系排序，先完成数据层再做业务层：

```
P0-1  数据库表结构设计与 Migration
         ↓
P0-2  核销记录模型与 Repository
         ↓
P0-3  自动核销规则引擎
         ↓
P0-4  核销核心服务（自动核销 + 手工核销）
         ↓
P1-1  反核销功能
         ↓
P1-2  HTTP Handler 与 API 接口
         ↓
P1-3  定时任务与批量导入
```

---

## P0-1: 数据库表结构设计与 Migration

### 任务范围

创建核销相关的数据库表结构迁移脚本。

### 改动文件

| 文件 | 改动 |
|------|------|
| `db/migrations/20240101000000_write_off_records.sql` | 新增核销记录表 |
| `db/migrations/20240101000001_write_off_rules.sql` | 新增核销规则配置表 |
| `db/migrations/20240101000002_payment_entry_write_off_fields.sql` | payment_entries 表新增核销字段 |
| `db/migrations/20240101000003_ar_invoice_write_off_fields.sql` | ar_invoices 表新增核销字段 |
| `db/migrations/20240101000004_ap_invoice_write_off_fields.sql` | ap_invoices 表新增核销字段 |

### 验证条件

- [ ] Migration 可独立运行（`migrate up`）
- [ ] 所有新增字段有默认值，不影响现有数据
- [ ] 外键约束正确

---

## P0-2: 核销记录模型与 Repository

### 任务范围

定义核销记录模型和仓储接口实现。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/model/write_off_record.go` | 新增 WriteOffRecord 模型 |
| `internal/model/write_off_rule.go` | 新增 WriteOffRule 模型 |
| `internal/repository/write_off_repo.go` | 新增 WriteOffRepository 接口与实现 |
| `internal/repository/write_off_rule_repo.go` | 新增 WriteOffRuleRepository 接口与实现 |

### 验证条件

- [ ] `go build ./...` 通过
- [ ] 模型字段与数据库表结构一致
- [ ] Repository 包含基本 CRUD 方法

---

## P0-3: 自动核销规则引擎

### 任务范围

实现自动核销的核心匹配逻辑，包括容差处理和多笔匹配。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/write_off_engine.go` | 新增 WriteOffEngine |
| `internal/service/write_off_engine_test.go` | 新增单元测试 |

### 核心逻辑

```go
// 匹配优先级：往来单位 → 金额 → 单据编号 → 日期
func (e *WriteOffEngine) Match(ctx context.Context, payment *model.PaymentEntry) ([]*model.WriteOffRecord, error) {
    // 1. 按往来单位筛选候选单据
    // 2. 按金额排序
    // 3. 应用容差规则
    // 4. 生成核销记录
}
```

### 验证条件

- [ ] 单位相同+金额相等的单据能正确匹配
- [ ] 容差范围内的单据可自动核销
- [ ] 一对多、多对一匹配逻辑正确
- [ ] 单元测试覆盖主要场景

---

## P0-4: 核销核心服务

### 任务范围

实现 WriteOffService，包含自动核销和手工核销功能。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/write_off_service.go` | 新增 WriteOffService 接口与实现 |
| `internal/service/write_off_service_test.go` | 新增单元测试 |

### 验证条件

- [ ] AutoWriteOff 方法正确调用引擎执行匹配
- [ ] ManualWriteOff 方法正确创建核销记录
- [ ] 核销后正确更新单据余额字段
- [ ] `go build ./...` 通过

---

## P1-1: 反核销功能

### 任务范围

实现反核销功能，恢复单据状态和金额。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/write_off_service.go` | 新增 ReverseWriteOff 方法 |

### 验证条件

- [ ] 反核销后核销记录状态变为"已反核销"
- [ ] 反核销后单据未核销金额恢复
- [ ] 操作日志记录反核销操作人

---

## P1-2: HTTP Handler 与 API 接口

### 任务范围

实现核销模块的 HTTP 接口。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/handler/write_off_handler.go` | 新增 WriteOffHandler |
| `internal/router/router.go` | 注册核销相关路由 |
| `docs/api-contracts/v1/writeoff.yaml` | 新增 API 文档 |

### 接口列表

| 接口 | 方法 | 路径 |
|-----|------|------|
| 自动核销 | POST | `/api/writeoff/auto` |
| 手工核销 | POST | `/api/writeoff/manual` |
| 反核销 | POST | `/api/writeoff/reverse/{record_id}` |
| 查询核销记录 | GET | `/api/writeoff/records` |
| 查询未核销汇总 | GET | `/api/writeoff/unmatched-summary` |

### 验证条件

- [ ] 所有接口返回正确的 HTTP 状态码
- [ ] 请求参数校验正确
- [ ] 响应格式符合 API 规范

---

## P1-3: 定时任务与批量导入

### 任务范围

实现定时自动核销和批量导入核销方案功能。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/scheduler/write_off_scheduler.go` | 新增定时任务 |
| `internal/service/write_off_service.go` | 新增 BatchImport 方法 |

### 验证条件

- [ ] 定时任务可配置执行时间
- [ ] Excel 模板导入功能正常
- [ ] 批量导入支持错误处理和回滚

---

## 依赖关系

```
P0-1 (Migration)
    └── 依赖: 无
    └── 风险: 需确保与现有表结构兼容

P0-2 (模型与 Repository)
    └── 依赖: P0-1（表结构存在）
    └── 风险: 模型字段与数据库字段映射正确

P0-3 (规则引擎)
    └── 依赖: P0-2（Repository 存在）
    └── 风险: 匹配逻辑正确性，需充分测试

P0-4 (核心服务)
    └── 依赖: P0-2, P0-3
    └── 风险: 事务边界处理，确保数据一致性

P1-1 (反核销)
    └── 依赖: P0-4（服务层存在）
    └── 风险: 金额计算正确性

P1-2 (API Handler)
    └── 依赖: P0-4, P1-1
    └── 风险: 参数校验和错误处理

P1-3 (定时任务)
    └── 依赖: P0-4
    └── 风险: 定时任务并发安全
```

---

## 预计工作量

| 任务 | 估计 |
|------|------|
| P0-1 | 2 小时（编写 Migration 脚本） |
| P0-2 | 3 小时（模型 + Repository） |
| P0-3 | 4 小时（规则引擎 + 单元测试） |
| P0-4 | 3 小时（核心服务） |
| P1-1 | 1.5 小时（反核销功能） |
| P1-2 | 2.5 小时（Handler + 路由） |
| P1-3 | 2 小时（定时任务 + 批量导入） |

---

## 里程碑

| 阶段 | 完成标志 |
|------|---------|
| 第一阶段 | P0-1 + P0-2 完成，数据层就绪 |
| 第二阶段 | P0-3 + P0-4 完成，核心功能可用 |
| 第三阶段 | P1-1 + P1-2 完成，API 接口就绪 |
| 第四阶段 | P1-3 完成，全部功能上线 |