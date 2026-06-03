# SPEC: TASK01 — 后端：业务单据数据模型 + CRUD API

## 基本信息

- **任务 ID**: phase1-bill-001
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **执行者**: OpenCode

## 背景

慧话财务当前只有手工凭证入口，缺费用报销单、收款单、付款单三类高频业务单据。需要新建数据表和API。

## 目标

创建 `BusReimbursement`（费用报销单）、`BusReceipt`（收款单）、`BusPayment`（付款单）三个数据模型 + 完整的CRUD API。

## 技术约束

- 新建文件 `backend/app/models/business.py`，不修改现有模型文件
- 新建文件 `backend/app/schemas/business.py`
- 新建文件 `backend/app/api/v1/business.py`
- 路由前缀统一为 `/api/v1/business`
- 使用 UUID 主键，时间戳用 `server_default=func.now()`
- 状态字段用字符串枚举，不用 Integer
- 金额统一用 `Numeric(18, 2)`，存储为 Decimal

## 详细设计

### 费用报销单 (BusReimbursement)

```python
class BusReimbursement(Base):
    __tablename__ = "bus_reimbursement"
    
    id = Column(String(36), primary_key=True)
    doc_no = Column(String(20), unique=True, comment="单据编号，自动生成格式: FY+YYYYMM+NNNN")
    doc_date = Column(Date, nullable=False, comment="报销日期")
    
    # 费用信息
    expense_type = Column(String(50), nullable=False, comment="费用类型：travel差旅/office办公/entertain招待/transport交通/other其他")
    amount = Column(Numeric(18, 2), nullable=False, comment="报销金额")
    summary = Column(String(500), comment="报销事由")
    
    # 关联信息
    dept_id = Column(String(36), ForeignKey("sys_dept.id"), comment="部门")
    dept_name = Column(String(100))
    payee = Column(String(100), comment="收款人")
    payment_method = Column(String(20), comment="支付方式：cash现金/bank银行/transfer转账")
    bank_account_id = Column(String(36), ForeignKey("ca_bank.id"), comment="支付银行账户")
    
    # 附件
    attachment_count = Column(Integer, default=0, comment="附件张数")
    
    # 状态
    status = Column(String(20), default="draft", comment="draft/pending/approved/rejected/posted")
    
    # 生成凭证信息
    voucher_id = Column(String(36), ForeignKey("acc_voucher.id"), comment="关联凭证ID")
    voucher_no = Column(String(20), comment="关联凭证号")
    
    # 审计
    created_by = Column(String(36))
    created_by_name = Column(String(100))
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
```

### 收款单 (BusReceipt)

```python
class BusReceipt(Base):
    __tablename__ = "bus_receipt"
    
    id = Column(String(36), primary_key=True)
    doc_no = Column(String(20), unique=True, comment="自动生成: SK+YYYYMM+NNNN")
    doc_date = Column(Date, nullable=False)
    
    # 往来信息
    customer_id = Column(String(36), ForeignKey("aux_customer.id"))
    customer_name = Column(String(200))
    amount = Column(Numeric(18, 2), nullable=False)
    receipt_method = Column(String(20), comment="收款方式：bank/ cash/ wechat/ alipay/ other")
    bank_account_id = Column(String(36), ForeignKey("ca_bank.id"))
    
    summary = Column(String(500))
    status = Column(String(20), default="draft")
    
    voucher_id = Column(String(36), ForeignKey("acc_voucher.id"))
    voucher_no = Column(String(20))
    
    created_by = Column(String(36))
    created_by_name = Column(String(100))
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
```

### 付款单 (BusPayment)

基本同收款单，但关联 `supplier_id` 和 `supplier_name`。字段同收款单结构，将 customer 替换为 supplier。

### CRUD API 路由

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/reimbursements` | 创建费用报销单 |
| GET | `/reimbursements` | 查询列表（按日期/状态/费用类型筛选） |
| GET | `/reimbursements/{id}` | 获取详情 |
| PUT | `/reimbursements/{id}` | 更新草稿状态单据 |
| DELETE | `/reimbursements/{id}` | 删除草稿状态单据 |
| POST | `/reimbursements/{id}/submit` | 提交审核 |
| POST | `/reimbursements/{id}/approve` | 审核通过 |
| POST | `/reimbursements/{id}/reject` | 驳回 |

收款单和付款单同理，路径分别用 `/receipts` 和 `/payments`。

### 单据编号生成规则

```
费用报销单: FY{year}{month}-{4位序号}  如 FY202605-0001
收款单:     SK{year}{month}-{4位序号}  如 SK202605-0001
付款单:     FK{year}{month}-{4位序号}  如 FK202605-0001
```

序号按月重置，从 0001 开始。

## 验收标准

- [ ] 三个表都可以正常创建/查询/更新/删除（草稿状态）
- [ ] 单据编号自动生成且不重复
- [ ] 费用报销单可按日期、费用类型、状态筛选
- [ ] 收款单/付款单可按日期、客户/供应商、状态筛选
- [ ] 提交审核和审核流程的状态流转正常
- [ ] 金额字段精度为 Decimal(18,2)

## OpenCode 指令

**目标**：在 `/root/huihua-financial-master/backend` 中新建3个数据模型 + 对应的 CRUD API。

**约束**：
- 新建文件：`app/models/business.py`、`app/schemas/business.py`、`app/api/v1/business.py`
- 不修改现有模型文件
- 路由前缀 `/api/v1/business`
- 在 `app/main.py` 中注册新路由
- 使用现有的 `get_db`、`require_roles`、`get_current_user`

**上下文**：
- repo: `/root/huihua-financial-master`
- 现有模型参考: `app/models/cash.py`（CaTransaction 的写法风格）
- 科目表：`app/models/ledger.py` 的 AccSubject
- 辅助核算：`app/models/auxiliary.py` 的 AuxCustomer / AuxSupplier

**验收**：
- 三个表数据模型创建成功
- API 路由可访问
- 单据编号自动生成格式正确
- 草稿可CRUD，提交后状态流转正常
