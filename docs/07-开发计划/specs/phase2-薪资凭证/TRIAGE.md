# Phase 2 TRIAGE — 薪资→凭证闭环 & 个税社保计提

## 需求摘要

慧话财务已有工资项目配置(WageProject)、员工工资信息(WageEmployee)、月度工资台账(WageRegister+WageDetail)、个税计算记录(IndividualTax)的完整数据模型，还有个税累计预扣法的计算逻辑。但数据录完后走不到凭证——缺社保计提、缺一键生成凭证，薪资和凭证是两张皮。

## 识别到的约束

1. **不重建薪资模块**：已有 WageEmployee/WageRegister/WageDetail 完整模型，只补"生成凭证"这一步
2. **个税计算已有**：`WageDetail` 中有 `individual_tax` 字段，`IndividualTax` 表记录明细，已有累计预扣法计算
3. **社保基础数据已有**：`WageEmployee` 中有 `social_base` / `housing_fund_base` 字段，但缺比例配置
4. **薪资科目映射**：复用 Phase 1.3 的科目映射模式（但可独立开发，不依赖Phase 1）

## 假设

- 社保公积金的比例存在系统配置中（`social_insurance_rate` 公司部分比例、`housing_fund_rate` 公司部分比例等）
- 个税的Excel导出格式参照主流个税申报系统的模板
- 薪资凭证的科目映射固定（借：管理费用/销售费用-工资/社保/公积金  贷：应付职工薪酬）

## 依赖

- 无（可独立开发，与Phase 1并行）

## 输出物

| 文件 | 内容 |
|------|------|
| `docs/specs/phase2-薪资凭证/TRIAGE.md` | 本文件 |
| `docs/specs/phase2-薪资凭证/TASK01-社保配置与计提.md` | 社保配置表+自动计提逻辑 |
| `docs/specs/phase2-薪资凭证/TASK02-个税计算验证与补全.md` | 验证个税逻辑+补全导出 |
| `docs/specs/phase2-薪资凭证/TASK03-一键生成凭证.md` | 薪资→凭证 |
| `docs/specs/phase2-薪资凭证/TASK04-薪资报表导出.md` | 工资单+个税明细导出Excel |
