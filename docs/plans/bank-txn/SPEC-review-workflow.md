# SPEC: 银行流水智能处理与对账模块 · 核心审核流程

> 任务 ID: TASK-BANK-01
> 类型: feature
> 优先级: P0
> 依赖: 无
> 状态: 🟡 SPEC（待老丁审核）

---

## 1. 背景

当前系统银行流水导入后只能做到"规则引擎分类 + 自动制证"，缺少人工审核节点、状态机、草稿机制和原子性提交。导致：
1. 自动生成的凭证直接进入账务系统，不符合财务规范
2. 无法处理"不明确"的流水（C 类）
3. 银行流水状态与下游单据状态不同步
4. AI/规则分析结果无反馈日志，无法优化

**目标**：补全完整的「导入 → AI分析 → 人工审核 → 原子性提交 → 分流」闭环。

---

## 2. 目标

### 2.1 整体目标

在银行流水模块中建立完整的"草稿 + 人工审核 + 原子性提交"机制，覆盖：
- **第一类**（银行/税务/社保/利息/保险）→ 直接生成凭证（草稿→正式）
- **第二类**（收款/付款/货款/内部转账）→ 生成收款单/付款单（草稿→正式）
- **C类**（其他不明）→ 待人工处理工作台

### 2.2 本次 SPEC 范围（核心，不含 AI 服务接入）

| 范围 | 说明 |
|------|------|
| ✅ 后端状态机 + 字段变更 | bank_transactions 新增 status / AI字段 / draft_id |
| ✅ 原子性提交审核 API | POST /api/v1/bank-transactions/submit-review |
| ✅ 草稿生成 Service 增强 | 支持草稿模式 vs 正式模式 |
| ✅ 驳回至待处理 API | POST /api/v1/bank-transactions/reject-manual |
| ✅ AI 反馈日志 Service | 记录人工修改历史 |
| ✅ 审核工作台 API | 列表 / 统计 / 预览草稿 |
| ✅ 前端审核工作台页面 | BankTxnReviewView.vue（新建） |
| ✅ 前端草稿预览弹窗 | DraftPreviewDialog.vue（新建） |
| ✅ 前端待人工处理工作台 | ManualPendingView.vue（新建） |
| ❌ AI 分析服务接入 | 下一阶段（外部 API 或 DeepSeek） |
| ❌ 发票核销联动 | 后续 Phase |

---

## 3. 技术约束

### 3.1 后端约束

**项目路径**：`/root/data/disk/huihua-finance`

**约束**：
- 不修改已有的 `ClassificationRuleService` 匹配逻辑（只扩展字段）
- 不修改已有的 `VoucherAutoGenerateService` 生成逻辑（只增加草稿模式参数）
- `SubmitReview` 必须在数据库事务中执行（PgSQL transaction）
- 科目预检查在事务外做，避免不必要的事务占用

**允许修改的文件**：
- `internal/model/bank_transaction.go` — 新增字段
- `internal/repository/bank_transaction_repo.go` — 新增 Query
- `internal/service/bank_transaction_service.go` — 新增 SubmitReview / RejectManual
- `internal/service/voucher_auto_generate_service.go` — 增加 draft_mode 参数
- `internal/service/payment_entry_service.go` — 增加草稿/正式模式
- `internal/handler/bank_transaction_handler.go` — 新增 API
- `internal/model/ai_feedback_log.go` — 新建模型
- `internal/repository/ai_feedback_log_repo.go` — 新建
- `internal/service/ai_feedback_service.go` — 新建
- `cmd/api/main.go` — 注册新路由

**新增文件**：
- `internal/handler/bank_txn_review_handler.go`
- `internal/service/bank_txn_review_service.go`
- `internal/model/bank_txn_status.go`（状态枚举）

### 3.2 前端约束

**项目路径**：`/root/data/disk/huihua-finance/frontend`

**约束**：
- 使用 Element Plus UI 组件（已有）
- 使用 TypeScript（已有）
- 不新增状态管理库（已有 Pinia）
- 草稿预览弹窗复用已有的 el-dialog

**允许修改/新建的文件**：
- `frontend/src/views/bank/BankTxnReviewView.vue`（新建）
- `frontend/src/views/bank/DraftPreviewDialog.vue`（新建）
- `frontend/src/views/bank/ManualPendingView.vue`（新建）
- `frontend/src/api/modules/bank_transaction.ts`（扩展）
- `frontend/src/api/modules/payment.ts`（扩展）

### 3.3 数据库 Migration

```sql
-- bank_transactions 新增字段
ALTER TABLE bank_transactions ADD COLUMN status varchar(30) DEFAULT 'pending';
ALTER TABLE bank_transactions ADD COLUMN ai_confidence int DEFAULT 0;
ALTER TABLE bank_transactions ADD COLUMN ai_suggested_action varchar(50);
ALTER TABLE bank_transactions ADD COLUMN ai_business_scene varchar(100);
ALTER TABLE bank_transactions ADD COLUMN ai_feedback_log jsonb;
ALTER TABLE bank_transactions ADD COLUMN draft_voucher_id uuid;
ALTER TABLE bank_transactions ADD COLUMN draft_payment_id uuid;

-- 存量数据兼容迁移
UPDATE bank_transactions SET status = 'voucher_generated' WHERE matched = true;
UPDATE bank_transactions SET status = 'pending' WHERE matched = false;

-- AI 反馈日志表
CREATE TABLE ai_feedback_logs (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    bank_txn_id uuid NOT NULL,
    ai_suggested_action text,
    ai_confidence int,
    ai_business_scene text,
    human_action text,
    human_modified_fields jsonb,
    created_by uuid,
    created_at timestamp
);
```

---

## 4. 状态机详细定义

### 4.1 BankTransaction.status

| 状态值 | 含义 | 前端显示 |
|--------|------|----------|
| `pending` | 待审核（初始） | 🔵 待审核 |
| `classified` | 已分类（AI/规则已分析） | 🟡 AI已分析 |
| `approved` | 已审核（人工确认） | 🟢 已审核 |
| `voucher_generated` | 已生成凭证 | 🟢 已制证 |
| `payment_created` | 已生成收付款单 | 🟢 已生成单据 |
| `manual_pending` | 待人工处理 | 🔴 待处理 |

### 4.2 状态流转规则

```
pending → classified     （AI分析或规则引擎分析后）
classified → approved    （人工提交审核，通过 submit-review）
classified → manual_pending （人工驳回）
approved → voucher_generated （第一类自动制证，在 submit-review 中完成）
approved → payment_created  （第二类生成收付款单，在 submit-review 中完成）
approved → manual_pending   （第二类无法生成收付款单时降级）
```

---

## 5. 核心 API 设计

### 5.1 审核工作台列表

```
GET /api/v1/bank-transactions/review-list
Query: status, start_date, end_date, page, page_size
Response:
{
  "data": [
    {
      "id": "uuid",
      "txn_date": "2026-06-01",
      "description": "转账手续费",
      "counterparty_name": "招商银行",
      "debit": "25.00",
      "credit": "0.00",
      "direction": "out",
      "status": "classified",
      "classification": "bank_fee",
      "ai_confidence": 85,
      "ai_suggested_action": "auto_voucher",
      "ai_business_scene": "银行手续费",
      "draft_voucher_id": "uuid",      // 有草稿时填充
      "draft_payment_id": null,
      "has_draft": true
    }
  ],
  "total": 120,
  "page": 1,
  "page_size": 50
}
```

### 5.2 审核工作台统计

```
GET /api/v1/bank-transactions/review-stats
Response:
{
  "monthly_txns": 156,
  "pending_count": 12,
  "ai_processed_count": 89,
  "manual_pending_count": 7
}
```

### 5.3 预览草稿

```
POST /api/v1/bank-transactions/preview-draft/{id}
Response:
{
  "bank_txn": { ...流水原始信息... },
  "ai_result": {
    "business_scene": "银行手续费",
    "suggested_action": "auto_voucher",
    "confidence": 85
  },
  "draft_voucher": {
    "id": "uuid",
    "lines": [
      {"account_name": "财务费用", "debit": "25.00", "credit": "0.00"},
      {"account_name": "银行活期存款", "debit": "0.00", "credit": "25.00"}
    ],
    "summary": "转账手续费"
  },
  "or_draft_payment": null
}
```

### 5.4 原子性提交审核（核心事务 API）

```
POST /api/v1/bank-transactions/submit-review
Request:
{
  "txn_ids": ["uuid1", "uuid2"],    // 批量
  "human_modified_drafts": {        // 可选，人工修正的草稿内容
    "uuid1": {
      "lines": [...],
      "suggested_action": "auto_voucher"
    }
  }
}

Response (成功):
{
  "data": {
    "approved_count": 2,
    "results": [
      { "txn_id": "uuid1", "outcome": "voucher_generated", "voucher_id": "uuid" },
      { "txn_id": "uuid2", "outcome": "payment_created", "payment_id": "uuid" }
    ]
  }
}

Response (失败):
{
  "error": "科目 5602 不存在，生成凭证失败",
  "failed_txn_ids": ["uuid1"]
}
```

### 事务执行步骤（伪代码）：
```go
func (s *BankTxnReviewService) SubmitReview(ctx, tenantID, txnIDs) {
    tx, _ := s.pool.Begin(ctx)
    defer tx.Rollback(ctx)

    for _, txnID := range txnIDs {
        txn := s.repo.GetByID(tx, txnID)

        // 1. 状态变更
        if txn.status == BankTxnReviewStatusClassified {
            txn.status = BankTxnReviewStatusApproved
        }

        // 2. 草稿生成 + 下游单据（按第一类/第二类分流）
        switch classify(txn.classification) {
        case "第一类":
            // 第一类：生成凭证草稿（docstatus=0），人审后正式生效
            voucher := s.voucherAutoSvc.GenerateFromBankTxn(tx, tenantID, txnID, docstatus=0)
            txn.draft_voucher_id = voucher.ID
            txn.matched_journal_entry_id = voucher.ID
            txn.status = BankTxnReviewStatusVoucherGenerated
        case "第二类":
            // 第二类：生成收款单/付款单草稿（status=draft），人确认后核销发票再生成凭证
            payment := s.paymentSvc.GenerateFromBankTxn(tx, tenantID, txnID, status=draft)
            txn.draft_payment_id = payment.ID
            txn.matched_document_id = payment.ID
            txn.status = BankTxnReviewStatusPaymentCreated
        }

        // 3. 银企勾对
        txn.matched = true

        // 4. 更新流水
        s.repo.UpdateStatus(tx, txnID, txn)

        // 5. 记录AI反馈日志
        s.aiFeedbackSvc.Log(txnID, human_action="submit_review")
    }

    tx.Commit(ctx)
}
```

### 5.5 驳回至待处理

```
POST /api/v1/bank-transactions/reject-manual
Request: { "txn_ids": ["uuid1"] }
Response: { "data": { "rejected_count": 1 } }
```

### 5.6 修正分类

```
PATCH /api/v1/bank-transactions/{id}/classification
Request: { "classification": "business_receipt" }
Response: { "data": { "updated": true } }
```

---

## 6. 前端页面详细设计

### 6.1 BankTxnReviewView.vue（审核工作台）

**顶部统计卡片**（4格）：
```
┌────────────┬────────────┬────────────┬────────────┐
│ 本月流水   │ 待审核     │ AI已处理    │ 待人工处理  │
│ 156笔     │ 12笔 🟡    │ 89笔 🟢     │ 7笔 🔴     │
└────────────┴────────────┴────────────┴────────────┘
```

**操作按钮**：
- 导入银行流水（跳转 ImportView）
- 刷新
- 筛选下拉（状态 / 日期范围）

**流水列表**（关键列）：
| 多选 | 日期 | 摘要 | 对方 | 金额 | 分类 | 置信度 | 草稿状态 | 操作 |
|------|------|------|------|------|------|--------|----------|------|
| ☐ | 06-01 | 转账手续费 | 招商银行 | ¥25.00 | 🟢银行费用 | 85% | 凭证草稿 | 预览 提交 驳回 |

**批量操作**：
- 选中 → 「批量提交审核」（同类型才可批量）
- 选中 → 「批量驳回」

**多选时的金额合计**：底部显示「已选 N 笔，合计 ¥X」

### 6.2 DraftPreviewDialog.vue（草稿预览弹窗）

**弹窗内容**：
```
┌──────────────────────────────────────────────┐
│ 流水详情                                      │
│ 日期: 2026-06-01  对方: 招商银行              │
│ 金额: ¥25.00（借）摘要: 转账手续费            │
├──────────────────────────────────────────────┤
│ AI 分析结果                                    │
│ 业务场景: 银行手续费  置信度: 85%             │
│ 建议动作: 生成凭证                            │
├──────────────────────────────────────────────┤
│ 凭证草稿预览                                   │
│ ──────────────────────────────────           │
│ 借: 财务费用     ¥25.00                      │
│ 贷: 银行活期存款  ¥25.00                      │
│ 摘要: 转账手续费                              │
│ ──────────────────────────────────           │
│ 科目可修改 / 摘要可修改                       │
├──────────────────────────────────────────────┤
│              [驳回] [确认并提交审核]          │
└──────────────────────────────────────────────┘
```

### 6.3 ManualPendingView.vue（待人工处理）

**列表**（仅 status=manual_pending）：
| 日期 | 摘要 | 对方 | 金额 | 无匹配原因 | 操作 |
|------|------|------|------|------------|------|
| 06-01 | 款项 | 深圳市xx公司 | ¥5000 | 无分类规则匹配 | 做凭证 做收款单 做付款单 |

**操作**：
- 「做凭证」→ 跳转 VoucherEditView 带入流水信息
- 「做收款单」→ 生成 PaymentEntry（receipt）
- 「做付款单」→ 生成 PaymentEntry（payment）
- 「标记已完成」→ 仅变更状态为 approved（无需生成单据时使用）

---

## 7. 验收标准（Acceptance Criteria）

- [ ] **AC1** 流水导入后查询 `status='pending'` 初始值正确
- [ ] **AC2** AI 分析后（模拟）流水 `ai_confidence` / `ai_suggested_action` / `ai_business_scene` 字段更新
- [ ] **AC3** 高置信度流水（≥80）自动创建凭证草稿（`draft_voucher_id` 非空，`docstatus=0`）
- [ ] **AC4** 低置信度（<60）流水无草稿，`status='manual_pending'`
- [ ] **AC5** `POST /submit-review` 事务成功：流水状态变更 + 生成正式单据 + `matched=true` 在同一事务中
- [ ] **AC6** `POST /submit-review` 失败时事务回滚（模拟科目不存在场景），无脏数据
- [ ] **AC7** `POST /reject-manual` 将流水状态更新为 `manual_pending`，清空草稿关联
- [ ] **AC8** `GET /review-list?status=classified` 返回 AI 已分析流水列表
- [ ] **AC9** `GET /review-stats` 返回 4 个统计数字
- [ ] **AC10** AI 反馈日志记录每次人工操作（修改科目/变更处理方式）
- [ ] **AC11** 前端 BankTxnReviewView 展示真实 API 数据（4 个统计卡片 + 分页列表）
- [ ] **AC12** 前端草稿预览弹窗可修改科目/摘要后提交
- [ ] **AC13** 前端 ManualPendingView 展示待人工处理列表 + 3 种处理方式按钮
- [ ] **AC14** `go build ./...` 编译通过
- [ ] **AC15** Migration 脚本执行后存量数据正确迁移

---

## 8. 执行顺序（TDD · 测试先行）

> **TDD 原则**：每个子任务先写对应 `_test.go`（验证失败）→ 实现代码 → 测试通过 → commit → push

```
Step 1:  Migration（新增字段 + AI反馈日志表）  [先跑]
        ↓
Step 2:  Model 层
        ├─ model/bank_txn_status.go（枚举）
        └─ 同步写 model 测试（字段/枚举）[TASK-BANK-01.1]

Step 3:  Repository 层
        ├─ bank_transaction_repo.go（ListByStatus / UpdateStatus）
        └─ 同步写 repo 测试[TASK-BANK-01.2]

Step 4:  Service 层
        ├─ bank_txn_review_service.go（SubmitReview 核心）
        └─ 同步写 service 测试（含事务回滚/AC5-AC6）[TASK-BANK-01.3]

Step 5:  Handler 层
        ├─ bank_txn_review_handler.go（5个API）
        └─ 同步写 handler 测试（HTTP响应码）[TASK-BANK-01.4]

Step 6:  AI Feedback 层
        ├─ ai_feedback_service.go + repo
        └─ 同步写测试（AC10）[TASK-BANK-01.5]

Step 7:  前端 API 层
        ├─ bank_transaction.ts 扩展
        └─ 手动验证 API 连通[前端自验]

Step 8:  前端页面
        BankTxnReviewView + DraftPreviewDialog + ManualPendingView[TASK-BANK-01.6]

Step 9:  集成测试 + go build 验证（AC14-AC15）[TASK-BANK-01.7]
```

**每个 Step commit 节奏**：
1. 写测试（RED）→ commit `test: add X_test.go for TASK-BANK-01.N [RED]`
2. 实现代码 → commit `feat: implement TASK-BANK-01.N`
3. `go test ./...` 通过 → `git push` → 才开始下一步

---

## 9. 参考文档

- `docs/plans/银行流水智能处理与对账-需求分析-v2.0.md` — 完整需求分析
- `docs/plans/银行流水全链路改进计划.md` — 业务单据层设计参考
- `docs/requirements/classification-engine.md` — 分类规则引擎参考
- `internal/service/voucher_auto_generate_service.go` — 现有自动制证逻辑
- `internal/service/bank_transaction_service.go` — 现有导入逻辑