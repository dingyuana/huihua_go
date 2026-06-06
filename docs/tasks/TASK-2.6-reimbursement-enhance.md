# TASK-2.6 | 费用报销单增强（附件上传+驳回持久化+发票关联）

**版本**：V1.0
**优先级**：P0
**工时**：20-28h
**前置**：TASK-2.3（发票管理页面基础）
**状态**：待开发

---

## 任务目标

完善费用报销单（BusReimbursement）三大缺失：实际上传附件、驳回原因持久化、进项发票关联，对应需求分析书V6.1第十三章P0优先级。

---

## 1. 附件实际上传功能

### 1.1 新建表

```sql
CREATE TABLE reimbursement_attachments (
  id VARCHAR(36) PRIMARY KEY,
  reimbursement_id VARCHAR(36) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(500) NOT NULL,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(100),
  uploaded_by VARCHAR(36),
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_reimbursement_id (reimbursement_id)
);
```

### 1.2 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/reimbursements/{id}/attachments` | 附件列表 |
| POST | `/reimbursements/{id}/attachments` | 上传附件（multipart，≤10MB，jpg/png/pdf） |
| DELETE | `/reimbursements/{id}/attachments/{file_id}` | 删除附件 |
| GET | `/reimbursements/{id}/attachments/{file_id}/download` | 下载附件 |

### 1.3 存储路径

`uploads/attachments/{reimbursement_no}/`

### 1.4 字段替换

- 原 `attachment_count`（仅数字）→ 保留兼容，新功能使用 `reimbursement_attachments` 表计数

---

## 2. 驳回原因持久化

### 2.1 现状

`reject_reason` 字段在模型中已存在，但 `POST /reimbursements/{id}/reject` 接口仅接受参数未写入DB。

### 2.2 修改

- `BusReimbursement` 模型已有 `reject_reason` 字段（无需新增）
- 修改 `Reject` service方法：将 `reject_reason` 参数写入 DB
- 修改数据库字段：从 VARCHAR(500) 改为 NOT NULL 或加索引

```sql
ALTER TABLE bus_reimbursements MODIFY COLUMN reject_reason VARCHAR(500) NOT NULL DEFAULT '';
```

### 2.3 驳回接口

```
PUT /reimbursements/{id}/reject
Body: { "reason": "xxx" }

→ 更新 docstatus=rejected，reject_reason=reason，updated_at=NOW()
```

---

## 3. 进项发票关联

### 3.1 新建关联表

```sql
CREATE TABLE reimbursement_invoice_links (
  id VARCHAR(36) PRIMARY KEY,
  reimbursement_id VARCHAR(36) NOT NULL,
  invoice_id VARCHAR(36) NOT NULL,
  invoice_type VARCHAR(20) NOT NULL,  -- expense_invoice
  linked_amount DECIMAL(18,2) NOT NULL,
  linked_by VARCHAR(36),
  linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_reimbursement_id (reimbursement_id),
  UNIQUE KEY uk_reim_invoice (reimbursement_id, invoice_id)
);
```

### 3.2 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/reimbursements/{id}/invoices` | 获取可关联的进项发票列表 |
| POST | `/reimbursements/{id}/invoices/{invoice_id}` | 关联发票（支持部分金额） |
| DELETE | `/reimbursements/{id}/invoices/{invoice_id}` | 取消关联 |
| GET | `/reimbursements/{id}/invoices/linked` | 获取已关联发票列表 |

### 3.3 关联逻辑

- 报销单提交审核时，检查是否关联发票（可选）
- 发票金额 > 报销金额时，允许部分报销（linked_amount < invoice.total_amount）
- 无票报销：允许不关联发票（需填写 description 说明原因）

---

## 4. 费用类型细分（同步实施）

### 4.1 现有类型（5种）

`travel / entertain / office / transport / other`

### 4.2 扩展为8种+子类型

```go
const (
    ExpenseTypeTravel     = "travel"     // 差旅费
    ExpenseTypeOffice     = "office"     // 办公费
    ExpenseTypeEntertain  = "entertain"  // 业务招待费
    ExpenseTypeTransport  = "transport"  // 交通费
    ExpenseTypeCommunication = "communication" // 通讯费
    ExpenseTypeTraining   = "training"   // 培训费
    ExpenseTypeWelfare    = "welfare"    // 福利费
    ExpenseTypeOther      = "other"      // 其他
)
```

### 4.3 子类型字段

`BusReimbursement` 新增 `sub_expense_type` VARCHAR(50) 字段。

---

## 5. 凭证生成规则

### 5.1 科目映射配置

| 费用类型 | 借方科目 | 贷方科目 |
|---------|---------|---------|
| travel | 6602.03 管理费用-差旅费 | 银行存款/应付账款 |
| office | 6602.01 管理费用-办公费 | 银行存款/应付账款 |
| entertain | 6602.02 管理费用-业务招待费 | 银行存款 |
| transport | 6602.04 管理费用-交通费 | 银行存款 |
| communication | 6602.05 管理费用-通讯费 | 银行存款 |
| training | 6602.06 管理费用-培训费 | 银行存款 |
| welfare | 6602.07 管理费用-福利费 | 银行存款 |
| other | 6602.99 管理费用-其他 | 银行存款 |

---

## 验收标准

- [ ] 附件上传：POST /reimbursements/{id}/attachments 可上传jpg/png/pdf≤10MB，写入reimbursement_attachments表
- [ ] 附件列表：GET /reimbursements/{id}/attachments 返回附件列表+下载链接
- [ ] 附件删除：DELETE 可删除附件+物理文件
- [ ] 驳回持久化：POST /reimbursements/{id}/reject 的reason字段写入DB
- [ ] 发票关联：可关联InvoiceIn（进项发票），支持部分金额
- [ ] 费用类型：下拉显示8种+子类型
- [ ] 凭证生成：按费用类型正确映射科目