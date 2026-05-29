# SPEC: TASK02 — 后端：单据生成凭证逻辑

## 基本信息

- **任务 ID**: phase1-bill-002
- **类型**: feature
- **优先级**: high
- **依赖**: TASK01（单据模型+API）
- **执行者**: OpenCode

## 背景

TASK01 建立了三类业务单据的CRUD，但单据本身只是数据记录。会计填完报销单/收款单/付款单后，需要一键生成记账凭证，自动填写借贷科目和金额。

## 目标

为三类单据各实现一个"生成凭证"API，调用 `create_voucher_core()` 自动创建凭证并提交审核。

## 技术约束

- 调用 `app/services/accounting.py` 中现有的 `create_voucher_core()`
- 凭证默认状态为"待审核"（调用 `create_voucher_core` + 自动 submit）
- 不修改 `create_voucher_core` 现有逻辑
- 科目映射规则写死在代码中，Phase 1.3改为可配置

## 详细设计

### 凭证生成逻辑

#### 费用报销单 → 凭证

```
借：费用类科目（按 expense_type 映射）
   - "travel" → 6601.01 销售费用-差旅费 或 6602.03 管理费用-差旅费
   - "office" → 6602.01 管理费用-办公费
   - "entertain" → 6602.04 管理费用-业务招待费
   - "transport" → 6602.05 管理费用-交通费
   - "other" → 6602.99 管理费用-其他
贷：银行存款（取 payment_method + bank_account_id 对应的科目）
    - "cash" → 1001 库存现金
    - "bank" → 1002 银行存款
```

摘要格式：`{员工姓名}报销{费用类型}{摘要}`

**映射配置表**（预置默认值，Phase 1.3提供管理界面）：

```python
# 在数据库或代码中维护
EXPENSE_TYPE_MAP = {
    "travel":   {"debit_code": "6602.03", "debit_name": "管理费用-差旅费"},
    "office":   {"debit_code": "6602.01", "debit_name": "管理费用-办公费"},
    "entertain": {"debit_code": "6602.04", "debit_name": "管理费用-业务招待费"},
    "transport": {"debit_code": "6602.05", "debit_name": "管理费用-交通费"},
    "other":    {"debit_code": "6602.99", "debit_name": "管理费用-其他"},
}
PAYMENT_METHOD_MAP = {
    "cash":  {"credit_code": "1001", "credit_name": "库存现金"},
    "bank":  {"credit_code": "1002", "credit_name": "银行存款"},
    "transfer": {"credit_code": "1002", "credit_name": "银行存款"},
}
```

#### 收款单 → 凭证

```
借：银行存款（按 receipt_method + bank_account_id）
贷：应收账款-{客户名称}（1122 + 辅助核算客户ID）

摘要格式："{客户名称}收款{金额}"
```

#### 付款单 → 凭证

```
借：应付账款-{供应商名称}（2202 + 辅助核算供应商ID）
贷：银行存款（按 payment_method + bank_account_id）

摘要格式："{供应商名称}付款{金额}"
```

### API 路由

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/business/reimbursements/{id}/generate-voucher` | 费用报销单→凭证 |
| POST | `/business/receipts/{id}/generate-voucher` | 收款单→凭证 |
| POST | `/business/payments/{id}/generate-voucher` | 付款单→凭证 |

**响应结构**：
```json
{
  "success": true,
  "voucher_id": "uuid",
  "voucher_no": "记-202605-0001",
  "message": "已生成凭证并提交审核"
}
```

### 生成后的状态流转

1. 调用 `create_voucher_core()` 创建凭证（状态=draft）
2. 调用 submit 逻辑提交审核（状态=pending）
3. 将单据的 `voucher_id` 和 `voucher_no` 写入该条记录
4. 单据状态变为 `posted`

## 验收标准

- [ ] 填完费用报销单 → 点击生成凭证 → 系统自动创建借费用贷银行凭证
- [ ] 填完收款单 → 点击生成凭证 → 系统自动创建借银行贷应收凭证
- [ ] 填完付款单 → 点击生成凭证 → 系统自动创建借应付贷银行凭证
- [ ] 生成的凭证状态为"待审核"
- [ ] 生成的凭证摘要格式正确
- [ ] 单据状态更新为 posted，关联凭证号显示
- [ ] 如果单据已有凭证（voucher_id 非空），拒绝重复生成

## OpenCode 指令

**目标**：在 TASK01 基础上，为三类业务单据各实现一个"生成凭证"API。

**约束**：
- 调用 `app/services/accounting.py` 中的 `create_voucher_core()`
- 科目映射写死在 `app/services/business_service.py`（新建文件）
- 在 `business.py` 中注册路由
- 不修改 accounting.py

**上下文**：
- repo: `/root/huihua-financial-master`
- 参考 `app/services/accounting.py` 中 `create_voucher_core` 的入参格式
- 可参考 `app/services/period_service.py` 中的结转逻辑（也是调用 create_voucher_core + post_voucher_core 的模式）

**验收**：
- 三类单据生成凭证后，数据库中有对应 voucher 记录
- 凭证借贷平衡
- 科目映射符合预定规则
- 重复生成被拒绝
