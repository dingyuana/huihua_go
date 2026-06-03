# SPEC: TASK03 — 后端：科目映射规则配置接口

## 基本信息

- **任务 ID**: phase1-bill-003
- **类型**: feature
- **优先级**: medium
- **依赖**: TASK02（因为TASK02写死了映射规则，TASK03改为可配置）
- **执行者**: OpenCode

## 背景

TASK02 的科目映射是写死在代码中的。会计主管应该能在页面上查看和修改"费用报销单-差旅费→管理费用-差旅费"这种映射关系。本任务提供配置化的映射规则存储和API。

## 目标

创建 `BusDocMapping` 配置表 + CRUD API，允许管理员查看和修改单据类型→借贷科目的映射关系。

## 技术约束

- 新建模型 `BusDocMapping`，放入 `app/models/business.py`
- 路由放在 `app/api/v1/business.py`
- 映射规则修改需要管理员权限（`require_roles("company_admin", "finance")`）
- 不修改 TASK02 的业务逻辑（TASK02 从写死的 map 改为读取配置表）

## 详细设计

### 映射配置表

```python
class BusDocMapping(Base):
    __tablename__ = "bus_doc_mapping"
    
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    doc_type = Column(String(20), nullable=False, comment="单据类型：reimbursement/receipt/payment")
    # 条件键：费用报销单用 expense_type，收/付款单用固定值 "default"
    condition_key = Column(String(50), default="default", comment="条件，如 travel/office/entertain 或 default")
    condition_label = Column(String(100), comment="条件中文名，如'差旅费'")
    # 借方科目
    debit_subject_code = Column(String(20), nullable=False)
    debit_subject_name = Column(String(100))
    # 贷方科目
    credit_subject_code = Column(String(20), nullable=False)
    credit_subject_name = Column(String(100))
    
    is_active = Column(Boolean, default=True)
    sort_order = Column(Integer, default=0)
    
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
```

### 预置数据

系统初始化时插入默认映射：

```python
DEFAULT_MAPPINGS = [
    # 费用报销单映射
    ("reimbursement", "travel",    "差旅费",   "6602.03", "管理费用-差旅费",   "1002", "银行存款"),
    ("reimbursement", "office",    "办公费",   "6602.01", "管理费用-办公费",   "1002", "银行存款"),
    ("reimbursement", "entertain", "招待费",   "6602.04", "管理费用-业务招待费", "1002", "银行存款"),
    ("reimbursement", "transport", "交通费",   "6602.05", "管理费用-交通费",   "1002", "银行存款"),
    ("reimbursement", "other",     "其他费用", "6602.99", "管理费用-其他",     "1002", "银行存款"),
    # 收款单映射（固定）
    ("receipt", "default", "默认", "1002", "银行存款", "1122", "应收账款"),
    # 付款单映射（固定）
    ("payment", "default", "默认", "2202", "应付账款", "1002", "银行存款"),
]
```

### API 路由

| 方法 | 路径 | 功能 | 权限 |
|------|------|------|------|
| GET | `/business/mappings` | 查询映射列表（按doc_type筛选） | 会计/管理员 |
| GET | `/business/mappings/{id}` | 获取单条 | 会计/管理员 |
| PUT | `/business/mappings/{id}` | 更新映射 | 管理员/财务主管 |
| POST | `/business/mappings` | 新增映射 | 管理员/财务主管 |
| DELETE | `/business/mappings/{id}` | 删除映射（仅非系统预置） | 管理员/财务主管 |

### TASK02 改造

TASK02 的 `EXPENSE_TYPE_MAP` 等常量改为从 `BusDocMapping` 表中读取：

```python
def get_mapping(doc_type: str, condition_key: str = "default") -> dict:
    """从 BusDocMapping 表获取科目映射"""
    mapping = db.query(BusDocMapping).filter(
        BusDocMapping.doc_type == doc_type,
        BusDocMapping.condition_key == condition_key,
        BusDocMapping.is_active == True
    ).first()
    if not mapping:
        raise ValueError(f"未找到{doc_type}/{condition_key}的科目映射配置")
    return {
        "debit_code": mapping.debit_subject_code,
        "debit_name": mapping.debit_subject_name,
        "credit_code": mapping.credit_subject_code,
        "credit_name": mapping.credit_subject_name,
    }
```

## 验收标准

- [ ] 系统启动后预设的映射数据写入数据库
- [ ] 管理员可以查看所有映射
- [ ] 管理员可以修改某条映射的借贷科目
- [ ] 修改后，新生成的单据按新规则生成凭证
- [ ] TASK02 的生成凭证逻辑从读表改为读配置表

## OpenCode 指令

**目标**：在 TASK02 基础上，将写死的科目映射改为从 `BusDocMapping` 配置表读取，并提供配置管理 API。

**约束**：
- 将 BusDocMapping 模型加入 `app/models/business.py`
- 添加路由到 `app/api/v1/business.py`
- 添加启动时插入预置数据的逻辑（用 `@app.on_event("startup")` 或手动脚本）
- 修改 TASK02 中生成凭证的代码，从读常量改为读表

**上下文**：
- repo: `/root/huihua-financial-master`
- 预置数据参考 `app/api/v1/company.py` 中初始化科目的模式

**验收**：
- 映射表数据可管理
- 修改映射后新单据按新规则走
- 旧单据不受影响（voucher_id 已经写入）
