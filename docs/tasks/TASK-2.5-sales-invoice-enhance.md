# TASK-2.5 | 销售发票增强（数电票+部分红冲）

**版本**：V1.0
**优先级**：P1
**工时**：16-24h
**前置**：TASK-2.3（发票管理页面基础）
**状态**：待开发

---

## 任务目标

增强销售发票（SalesInvoice）模型，支持数电发票字段、部分红冲功能，补齐与需求分析书V6.1第十四章的差距。

---

## 1. 模型层增强

### 1.1 新增字段（SalesInvoice）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| invoice_kind | VARCHAR(20) | 否 | 纸票/数电普票/数电专票（paper_special/paper_normal/electronic_special/electronic_normal） |
| electronic_url | VARCHAR(500) | 否 | 数电发票版式文件URL |
| red_letter_info_id | VARCHAR(50) | 否 | 红字信息表编号（关联开具） |
| red_letter_reason | VARCHAR(200) | 否 | 开具红字发票原因 |
| original_invoice_id | VARCHAR(36) | 否 | 原蓝字发票ID（红冲时填） |
| is_part_red | BOOLEAN | 否 | 是否部分红冲 |
| red_amount | DECIMAL(18,2) | 否 | 红冲金额（部分红冲时填写） |
| tax_authority_code | VARCHAR(20) | 否 | 主管税务机关代码 |
| confirm_status | VARCHAR(20) | 否 | 确认状态：unconfirmed/confirmed/invalid |
| confirm_date | DATE | 否 | 确认日期 |

### 1.2 migration SQL

```sql
ALTER TABLE sales_invoices
  ADD COLUMN invoice_kind VARCHAR(20) DEFAULT NULL,
  ADD COLUMN electronic_url VARCHAR(500) DEFAULT NULL,
  ADD COLUMN red_letter_info_id VARCHAR(50) DEFAULT NULL,
  ADD COLUMN red_letter_reason VARCHAR(200) DEFAULT NULL,
  ADD COLUMN original_invoice_id VARCHAR(36) DEFAULT NULL,
  ADD COLUMN is_part_red BOOLEAN DEFAULT FALSE,
  ADD COLUMN red_amount DECIMAL(18,2) DEFAULT NULL,
  ADD COLUMN tax_authority_code VARCHAR(20) DEFAULT NULL,
  ADD COLUMN confirm_status VARCHAR(20) DEFAULT 'unconfirmed',
  ADD COLUMN confirm_date DATE DEFAULT NULL;
```

### 1.3 模型对象更新

- `internal/model/invoice.go` — SalesInvoice 结构体新增上述字段
- 类型常量新增：
  - `InvoiceKindPaperSpecial = "paper_special"`
  - `InvoiceKindPaperNormal = "paper_normal"`
  - `InvoiceKindElectronicSpecial = "electronic_special"`
  - `InvoiceKindElectronicNormal = "electronic_normal"`
  - `ConfirmStatusUnconfirmed = "unconfirmed"`
  - `ConfirmStatusConfirmed = "confirmed"`
  - `ConfirmStatusInvalid = "invalid"`

---

## 2. 后端API增强

### 2.1 新增接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/invoices/sales/{id}/confirm` | 确认销售发票（更新confirm_status=confirmed） |
| POST | `/invoices/sales/{id}/red` | 红冲销售发票（新建红字发票，关联原发票） |
| POST | `/invoices/sales/{id}/red/part` | 部分红冲（is_part_red=true, red_amount>0） |
| POST | `/invoices/sales/{id}/void` | 作废销售发票 |
| GET | `/invoices/sales/{id}` | 详情（返回新字段） |

### 2.2 红冲业务逻辑

```
1. 校验原发票状态（仅confirmed状态可红冲）
2. 提取备注中的原蓝字发票号（如有）存入source_red_invoice_no
3. 生成红字发票（is_return=true，关联原发票ID）
4. 原发票标记为reversed
5. 自动生成红字凭证：借 主营业务收入（红字）/ 应交税费-销项税额（红字） 贷 应收账款（红字）
```

### 2.3 部分红冲业务逻辑

```
1. 接收 red_amount（必须 ≤ 原发票未核销金额）
2. 生成红字发票，is_part_red=true，red_amount=申请红冲金额
3. 原发票 outstanding_amount -= red_amount
4. 原发票状态更新为 PART_RED
5. 生成部分红冲凭证
```

---

## 3. 前端页面增强

### 3.1 列表页（InvoiceList.vue）

- 新增「发票类型」筛选（全部/纸票/数电普票/数电专票）
- 新增「数电票」角标标签

### 3.2 详情页（InvoiceDetail.vue）

- 新增数电票字段展示区域
- 新增「部分红冲」入口按钮
- 新增「红字信息表编号」字段展示
- 确认状态标签（未确认/已确认/异常）

### 3.3 红冲弹窗

- 普通红冲：原因下拉 + 确认
- 部分红冲：金额输入框（≤未核销金额）+ 原因下拉 + 确认

---

## 4. 凭证规则

| 发票类型 | 借方 | 贷方 |
|---------|------|------|
| 正常销售发票 | 应收账款 | 主营业务收入 / 应交税费-销项税额 |
| 红冲发票 | 主营业务收入（红字）/ 应交税费-销项税额（红字） | 应收账款（红字） |
| 部分红冲 | 主营业务收入（红字部分）/ 应交税费-销项税额（红字部分） | 应收账款（红字部分） |

---

## 验收标准

- [ ] sales_invoices 表新增10个字段，migration可执行
- [ ] GET /invoices/sales/{id} 返回新字段
- [ ] POST /invoices/sales/{id}/confirm 更新confirm_status
- [ ] POST /invoices/sales/{id}/red 生成红字发票+凭证，原发票标记reversed
- [ ] POST /invoices/sales/{id}/red/part 支持部分红冲，原发票outstanding_amount正确扣减
- [ ] 前端详情页展示所有新字段
- [ ] 前端红冲弹窗区分普通红冲和部分红冲