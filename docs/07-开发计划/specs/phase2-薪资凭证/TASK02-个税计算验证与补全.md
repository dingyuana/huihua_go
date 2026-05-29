# SPEC: TASK02 — 个税计算验证与补全

## 基本信息

- **任务 ID**: phase2-wage-002
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **执行者**: OpenCode

## 背景

`WageDetail` 表中已有 `individual_tax` 字段，`IndividualTax` 表有明细记录，路由 `/wage/registers/{id}/calculate` 中有累计预扣法逻辑。但需要验证逻辑完整性、补全导出功能。

## 目标

1. 验证并确保个税累计预扣法逻辑完整
2. 提供个税明细导出（兼容个税申报系统格式）

## 技术约束

- 不重写已有计算逻辑，只验证和补全
- 导出用 `openpyxl`（已有依赖）

## 详细设计

### 1. 个税计算验证

现有逻辑在 `backend/app/api/v1/wage.py` 中 `calculate_register` 函数里：

```python
# 已有逻辑：
# 1. 取该员工本月税前工资(gross_wage)
# 2. 扣除社保个人部分(social_insurance)
# 3. 扣除公积金个人部分(housing_fund_amt)
# 4. 扣除专项附加扣除(children_education等)
# 5. 计算应纳税所得额
# 6. 查税率表，用累计预扣法计算个税
# 7. 写入 WageDetail.individual_tax 和 IndividualTax 表
```

需要补全：
- 确保 `WageEmployee` 中的5项专项附加扣除全部参与计算
- 确保累计预扣法中的"累计减除费用"=5000×月数
- 确保跨年时累计数据重置

### 2. 个税明细导出API

```python
GET /wage/registers/{id}/export-tax
```

导出Excel格式：
| 序号 | 姓名 | 身份证号 | 税前工资 | 社保个人 | 公积金个人 | 专项附加扣除 | 应纳税所得额 | 税率 | 速算扣除数 | 应缴个税 |
|------|------|---------|---------|---------|----------|------------|-----------|------|----------|--------|
| 1    | 张三 | 110101..| 15,000  | 1,200   | 1,800    | 2,000      | 10,000    | 10%  | 210      | 790    |

## 验收标准

- [ ] 现有个税计算逻辑完整：累计预扣法→应纳税所得额→税率→扣除数→实缴个税
- [ ] 专项附加扣除全部参与计算
- [ ] 导出Excel格式正确，字段完整

## OpenCode 指令

**目标**：验证并补全薪资模块的个税计算逻辑，新增个税明细导出API。

**约束**：
- 不重写已有计算逻辑
- 导出写入 `backend/app/services/wage_service.py`
- 路由放在 `app/api/v1/wage.py`

**上下文**：
- repo: `/root/huihua-financial-master`
- 现有计算逻辑位置：`app/api/v1/wage.py` 中大约 580-670 行的 `calculate_register` 函数
- 导出Excel参考 `app/utils/report_export.py` 的导出模式

**验收**：
- 个税计算结果正确
- 导出Excel可打开、字段完整
