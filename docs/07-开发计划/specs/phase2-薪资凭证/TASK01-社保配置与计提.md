# SPEC: TASK01 — 社保配置与自动计提

## 基本信息

- **任务 ID**: phase2-wage-001
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **执行者**: OpenCode

## 背景

WageEmployee 表中已有 `social_base` 和 `housing_fund_base` 字段，但缺少社保公积金的**比例配置**和**自动计提计算**逻辑。

## 目标

创建社保公积金比例配置表 + 按月自动计提计算 + 生成计提凭证。

## 技术约束

- 新建模型 `WageSocialConfig` 放入 `app/models/wage.py`
- 新建路由放入 `app/api/v1/wage.py`
- 不修改现有 WageEmployee 表

## 详细设计

### 社保公积金配置表

```python
class WageSocialConfig(Base):
    __tablename__ = "wage_social_config"

    id = Column(String(36), primary_key=True)
    insurance_type = Column(String(20), nullable=False, comment="险种：pension养老/medical医疗/unemployment失业/injury工伤/maternity生育/housing公积金")
    company_rate = Column(String(10), nullable=False, comment="公司比例，如0.16表示16%")
    personal_rate = Column(String(10), nullable=False, comment="个人比例，如0.08表示8%")
    is_active = Column(Boolean, default=True)
    remark = Column(String(200))
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
```

预置数据（按典型比例，允许管理员修改）：
```
pension   公司0.16 个人0.08
medical   公司0.08 个人0.02
unemployment 公司0.005 个人0.005
injury    公司0.004 个人0.0
maternity 公司0.008 个人0.0
housing   公司0.12  个人0.12
```

### API 路由

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/wage/social-config` | 查询配置列表 |
| PUT | `/wage/social-config/{id}` | 修改比例 |
| POST | `/wage/social-config/init` | 初始化预置数据 |

### 计提计算逻辑

```python
def calculate_social_charges(db, register_id: str) -> dict:
    """根据工资台账的期间，计算该月社保公积金公司部分"""
    register = db.query(WageRegister).get(register_id)
    details = db.query(WageDetail).filter(WageDetail.register_id == register_id).all()
    configs = db.query(WageSocialConfig).filter(WageSocialConfig.is_active == True).all()
    
    config_map = {c.insurance_type: c for c in configs}
    
    total_company = Decimal("0")
    total_personal = Decimal("0")
    items = []
    
    for detail in details:
        employee = db.query(WageEmployee).get(detail.employee_id)
        social_base = Decimal(employee.social_base or "0")
        housing_base = Decimal(employee.housing_fund_base or "0")
        
        for ins_type, config in config_map.items():
            base = housing_base if ins_type == "housing" else social_base
            company_amt = base * Decimal(config.company_rate)
            personal_amt = base * Decimal(config.personal_rate)
            total_company += company_amt
            total_personal += personal_amt
            items.append({...})
    
    return {"total_company": total_company, "total_personal": total_personal, ...}
```

### 生成计提凭证

```python
POST /wage/registers/{id}/generate-social-voucher
```
生成：
```
借：管理费用-社保/公积金（公司部分）
借：销售费用-社保/公积金（按部门拆分，简化版可先全入管理费用）
贷：应付职工薪酬-社保
贷：应付职工薪酬-公积金
```

## 验收标准

- [ ] 社保配置表可管理（查看+修改比例）
- [ ] 选择工资台账 → 点击计提社保 → 系统按员工基数和比例计算出公司应缴金额
- [ ] 生成计提凭证，借贷平衡
- [ ] 计提记录在工资台账状态中标记

## OpenCode 指令

**目标**：在薪资模块中新增社保公积金比例配置和自动计提功能。

**约束**：
- 社保配置表模型放入 `app/models/wage.py`
- 计提计算逻辑写在 `app/services/wage_service.py`（新建）
- 路由放在 `app/api/v1/wage.py`
- 凭证生成调用 `create_voucher_core()`

**上下文**：
- repo: `/root/huihua-financial-master`
- 参考 `WageEmployee` 中的 `social_base`、`housing_fund_base` 字段
- 参考 `WageRegister` 的状态流转（draft→calculated→confirmed→posted）

**验收**：
- 社保配置CRUD可用
- 计提计算结果正确
- 凭证生成成功
