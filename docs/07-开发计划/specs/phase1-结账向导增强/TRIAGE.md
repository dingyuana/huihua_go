# Phase 1 TRIAGE — 结账向导增强（交互升级）

## 需求摘要

Phase 0 完成了结账向导的"数据展示"阶段：后端返回 risk_warnings / key_indicators / pending_accruals，前端展示为 7 区域向导布局。Phase 1 的目标是**从"展示"升级到"交互"**——让用户不仅能看到问题，还能在同一个页面内处理和确认。

## 核心判断

| 决策 | 依据 |
|:---|:---|
| **不做**银行余额调节表集成 | 当前 cash 模块只有基础对账（标记交易已对账），没有余额调节表（账面 vs 银行差额分析）。要做一个可用版本至少需要新建模型+API+前端页面，3-5天。投入产出比低，不如让用户去现有的银行模块对账 |
| **不做**封账 PDF 归档 | 现有 export 生成 CSV，不是 PDF。引入 PDF 库需要新依赖（reportlab/weasyprint），或前端 html2pdf.js，违反"不新增依赖"约束 |
| **不做**close_check_service 重构 | 纯架构变动，用户零价值。等有第 4 个 check_* 函数时再一并迁移 |
| **做**断号检测 | 纯查询逻辑，无 DB 变更，估算 120 行后端 + 80 行测试，高性价比 |
| **做**风险预警可交互 | 需要新建 `WarningAcknowledgement` 模型（自动建表，无需迁移），前端加确认/忽略按钮，估算后端 100 行 + 前端 120 行 |
| **做**税金比对增强 | 复用 `check_pending_accruals` 现有结构，比较 tax_declarations 实缴额与科目余额应计提额，估算后端 60 行 |
| **做**折旧计提明细 | 追加预计折旧金额到 `pending_accruals`，前端展示待计提清单，估算后端 30 行 + 前端 60 行 |

## 输出物清单

| 文件 | 内容 |
|:---|:---|
| `TRIAGE.md`（本文件） | 整体决策和优先级 |
| `TASK01-断号检测.md` | 后端：voucher_gaps + 前端提示 |
| `TASK02-预警可交互.md` | 新模型 WarningAcknowledgement + API + 前端操作 |
| `TASK03-税金比对&折旧细化.md` | 后端增强 + 前端展示 |

## 完成标准

- [ ] pre_close_check 返回 voucher_gaps，前端有断号告警
- [ ] 风险预警可确认/忽略，刷新后状态保持
- [ ] check_pending_accruals 返回应计提税金金额
- [ ] check_pending_accruals 折旧项返回预计折旧金额
- [ ] 56 个现有 period 测试全部通过
- [ ] 新加测试覆盖新增逻辑

## 约束（继承自 Phase 0）

1. 不修改现有数据库模型（AccVoucher/AccPeriod/FixedAsset 等）
2. 新增模型 `WarningAcknowledgement` 除外（`Base.metadata.create_all` 自动建表）
3. 不修改凭证状态机、损益结转逻辑、结账锁定逻辑
4. 不修改现有 pre_close_check 返回字段
5. 不引入新的 npm / pip 依赖
