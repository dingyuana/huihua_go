# SPEC: 凭证关联与前端显示修复

## 基本信息

- **任务 ID**: phase1-bill-007
- **类型**: bugfix + 增强
- **优先级**: high
- **依赖**: 既有凭证生成链路（`voucher_auto_generate_service.go`）
- **执行者**: OpenCode

## 背景

经过数据库直查与 API 联调，发现用户报告的三个问题中，**后端数据写入是正常的**，但 **API 响应未返回这些字段** + **前端表格列定义缺失**，导致 UI 看不见。逐项分析如下。

## 现状诊断（基于数据库直查 + API 实测）

### 数据库事实（2026-06-03 实测）

```sql
-- journal_entries 表：counterparty_name 已正确写入
id  | voucher_no    | counterparty_name          | remark | docstatus
----+---------------+----------------------------+--------+----------
5c6f| PZ-20260603-324| 北京数字认证股份有限公司   | (空)  | 0

-- journal_entry_lines 表：账户已正确写入
account_id                              | code | name   | debit | credit
582f509a-9e74-4894-82dd-efb87363c23a  | 1002 | 银行存款| 0.01  | 0
c769fba5-594c-4feb-bd65-0e00d258db98  | 1122 | 应收账款| 0     | 0.01
```

### API 实测响应（GET /v1/vouchers/5c6f...）

```json
{
  "journal_entry": {
    "id": "5c6f...",
    "voucher_no": "PZ-20260603-324",
    "docstatus": 0,
    // ❌ 没有 counterparty_name
    // ❌ 没有 remark
    "debit_total": "0",
    "credit_total": "0"
  },
  "journal_entry_lines": [
    {"account_id":"...","account_code":"1002","account_name":"银行存款",...}
    // ✅ 科目信息完整
  ]
}
```

### 前端显示事实

| 文件 | 状态 |
|------|------|
| `VoucherList.vue` | 已有"对方名称"列（行 35-36），引用 `row.counterparty_name`，但 API 不返回该字段 → 永远显示 `—` |
| `VoucherList.vue` | 已有"摘要"列（行 38），引用 `row.remark`，但 API 不返回该字段 → 永远显示空 |
| `VoucherList.vue` | **没有"科目"列** → 列表里完全看不到生成的科目 |
| `VoucherEdit.vue` | ✅ 编辑页正确显示科目（AccountSelector、行 32-34） |

## 问题根因分析

### 问题 1：凭证缺少与原始单据的关联

**根因**：
- `payment_entries` 表**没有 `voucher_id` 列**，单据无法反向追溯到凭证
- `journal_entries` 表**没有 `source_doc_type` / `source_doc_id` 字段**，凭证无法反向追溯到单据
- 业务上的关联现在只能通过 `bank_transactions.matched_gl_entry_id` 或 `bank_transactions.matched_payment_entry_id` 间接推导

**影响**：
- 删除凭证时无法回写"原始单据编号"到凭证摘要
- 用户在凭证列表看不到这条凭证来自哪张单据
- 业务可追溯性差

### 问题 2：凭证列表不显示对方名称

**根因**：
- 后端 `voucher_service.go::GetVoucher` 构造 `JournalEntry` 响应时，**没有把 `counterparty_name` 字段写出来**
- 后端 `journal_repo.go::ListVouchers` SELECT 列表里**没有 `counterparty_name` 列**
- 数据库里数据是对的（counterparty_name='北京数字认证股份有限公司'），但响应里没带 → 前端拿到的是 `undefined`

**修复方向**（按改动量从小到大）：

| 方案 | 改动 | 效果 |
|------|------|------|
| A. 仓储 SELECT 加列 | `ListVouchers` / `GetByID` 查询加 `counterparty_name` | 列表/详情能看到对方名称 |
| B. 服务层组装时赋值 | `GetVoucher` / `ListVouchers` 服务方法把字段塞进结构体 | 同上 |
| C. 前端请求时显式带 | 前端调用时拼 `?include=counterparty` | 不推荐，破坏 RESTful |

**采用 A+B**：在仓储 SELECT 加列，服务层无感透传。

### 问题 3：凭证列表不显示科目

**根因**：
- `VoucherList.vue` 的 `<el-table>` 完全没有"科目"列定义
- `ListVouchers` 仓储返回的 `JournalEntry` 不带 lines（lines 是单独查询），所以列表行没有 `account_code`
- 详情页 `VoucherEdit.vue` 是有科目的（拿到 lines 后显示）

**修复方向**：
- **不**改 ListVouchers 把所有 lines 都带出来（N+1 性能问题）
- 给 `journal_entries` 表加**汇总字段**：`first_account_code` / `first_account_name`（借方或贷方第一行）
- 或者用 SQL 子查询把第一行科目带出来
- 前端表格加"科目"列，渲染 `first_account_code` + `first_account_name`

## 详细设计

### D1：凭证与单据的双向关联

#### 1.1 数据库迁移（036 的扩展）

```sql
-- journal_entries 增加来源单据字段
ALTER TABLE journal_entries
  ADD COLUMN source_doc_type VARCHAR(50),     -- 'payment_entry' | 'bank_txn'
  ADD COLUMN source_doc_id UUID,
  ADD COLUMN source_doc_no VARCHAR(50);       -- 单据编号（冗余便于显示）

-- payment_entries 增加 voucher_id 字段
ALTER TABLE payment_entries
  ADD COLUMN voucher_id UUID,
  ADD COLUMN voucher_no VARCHAR(50);           -- 冗余凭证号便于列表显示

-- 索引
CREATE INDEX idx_journal_entries_source_doc
  ON journal_entries(tenant_id, source_doc_type, source_doc_id);
```

#### 1.2 后端写入（voucher_auto_generate_service.go::GenerateFromPaymentEntry）

```go
je := &model.JournalEntry{
    ...
    SourceDocType: stringPtr("payment_entry"),
    SourceDocID:   &pe.ID,
    SourceDocNo:   &pe.PaymentNo,
    Remark:        stringPtr(fmt.Sprintf("[%s] %s", pe.PaymentNo, partyName)),
}

// 同步回写到 payment_entries
pe.VoucherID = &je.ID
pe.VoucherNo = &je.VoucherNo
```

#### 1.3 摘要自动格式

生成凭证时自动填入 `remark`：
- 收款单：`"收款单 REC-000008 对方户名 0.01"`
- 付款单：`"付款单 PAY-000123 对方户名 0.01"`
- 银行流水直生：`"银行流水 txn-uuid 摘要"`

会计可在凭证编辑页修改。

### D2：凭证列表显示对方名称

#### 2.1 仓储层修复

```go
// journal_repo.go::GetByID — 增加扫描列
err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
    &je.ID, &je.VoucherNo, &je.VoucherType, &je.PostingDate, &je.CompanyID,
    &je.TenantID, &je.Remark, &je.DocStatus, &je.ReversedID, &je.ReversalID,
    &je.SubmittedBy, &je.SubmittedAt, &je.CreatedBy, &je.CreatedAt, &je.UpdatedAt,
    &je.CounterpartyName,  // 新增
    &je.SourceDocType,     // 新增（D1 配套）
    &je.SourceDocID,       // 新增
    &je.SourceDocNo,       // 新增
)

// ListVouchers 同样加列
// SELECT id, voucher_no, ..., counterparty_name, source_doc_type, source_doc_id, source_doc_no, ...
```

#### 2.2 模型层扩展

```go
// model/journal.go
type JournalEntry struct {
    ...
    CounterpartyName *string `json:"counterparty_name,omitempty" db:"counterparty_name"`
    SourceDocType    *string `json:"source_doc_type,omitempty" db:"source_doc_type"`
    SourceDocID      *uuid.UUID `json:"source_doc_id,omitempty" db:"source_doc_id"`
    SourceDocNo      *string `json:"source_doc_no,omitempty" db:"source_doc_no"`
}
```

### D3：凭证列表显示科目（汇总）

#### 3.1 仓储层用 SQL 子查询带出"第一行科目"

```sql
-- ListVouchers 查询增加两个字段
(SELECT a.code FROM journal_entry_lines jel
   JOIN accounts a ON a.id = jel.account_id
   WHERE jel.journal_entry_id = je.id
   ORDER BY jel.debit DESC, jel.credit DESC
   LIMIT 1) AS first_account_code,
(SELECT a.name FROM journal_entry_lines jel
   JOIN accounts a ON a.id = jel.account_id
   WHERE jel.journal_entry_id = je.id
   ORDER BY jel.debit DESC, jel.credit DESC
   LIMIT 1) AS first_account_name
```

#### 3.2 前端加列

```vue
<el-table-column label="科目" min-width="180" show-overflow-tooltip>
  <template #default="{ row }">
    <span>{{ row.first_account_code || '—' }} {{ row.first_account_name }}</span>
  </template>
</el-table-column>
```

### D4：删除凭证时清空单据上的 voucher_id

`voucher_service.go::DeleteVoucher` 已有银行流水的回滚逻辑。需补充：找到关联的 `payment_entries.voucher_id = ?`，清空 `voucher_id` 和 `voucher_no`。

```go
// 新增 repository 方法
func (r *PaymentEntryRepository) FindByVoucherID(ctx, tenantID, voucherID) ([]PaymentEntry, error)
func (r *PaymentEntryRepository) UnlinkVoucher(ctx, tenantID, peID) error
```

## 验收标准

### A1：凭证与单据关联
- [ ] 数据库迁移已应用，`source_doc_*` 和 `voucher_id` 列存在
- [ ] 生成凭证后，`payment_entries.voucher_id` 正确写入
- [ ] 生成凭证后，`journal_entries.source_doc_type='payment_entry'`、`source_doc_id=<pe.id>`、`source_doc_no='REC-000008'`
- [ ] 凭证 `remark` 字段自动填入"[REC-000008] 北京数字认证股份有限公司"

### A2：凭证列表显示对方名称
- [ ] `GET /v1/vouchers` 返回的每条记录包含 `counterparty_name`
- [ ] `GET /v1/vouchers/:id` 详情包含 `counterparty_name`
- [ ] 前端 VoucherList 表格"对方名称"列正常显示"北京数字认证股份有限公司"，不再显示"—"

### A3：凭证列表显示科目
- [ ] 列表 SQL 增加子查询带回 `first_account_code` / `first_account_name`
- [ ] 前端表格新增"科目"列，显示"1002 银行存款"或"1122 应收账款"
- [ ] 多分录的凭证只显示借方第一行（或借方金额最大的科目）

### A4：删除凭证时清空单据的 voucher_id
- [ ] 删除凭证后，对应 `payment_entries.voucher_id = NULL`、`voucher_no = NULL`
- [ ] 删除凭证后，对应 `bank_transactions.matched_gl_entry_id = NULL`、`matched = false`
- [ ] 删除凭证后，付款单 docstatus 回退为 0，可重新生成凭证

## 实施顺序

1. 迁移 037：扩展 `journal_entries` 和 `payment_entries` 表
2. 模型 `JournalEntry` + `PaymentEntry` 加字段
3. 仓储 SELECT 加列扫描
4. `GenerateFromPaymentEntry` 写入来源 + 摘要 + 单据 voucher_id
5. `DeleteVoucher` 补充清空 payment_entries.voucher_id
6. `ListVouchers` SQL 加 first_account_code 子查询
7. 前端 VoucherList 加"科目"列（"对方名称"列已存在）
8. 端到端测试：生成凭证 → 列表显示对方名称+科目+来源单据 → 删除凭证 → 单据状态回滚

## 不在本次范围

- AI 智能科目推荐（已列入 §十二 远期规划）
- 凭证明细账穿透到单据详情
- 单据上增加"查看凭证"按钮（属于反向穿透，后续需求）
