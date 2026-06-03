# SPEC: TASK02 — 前端：重构 PeriodClose 结账向导页面（已完成）

## 基本信息

- **任务 ID**: phase0-close-002
- **类型**: feature
- **优先级**: high
- **依赖**: phase0-close-001
- **状态**: ✅ 已完成

## 背景

PeriodClose 页面从4卡片网格改为7区域上下向导布局。

## 变更文件

- `frontend/src/pages/accounting/PeriodClose.vue` — 601行→907行

## 验收标准（已满足）

- [x] 页面按7区域上下布局展示
- [x] 有风险预警时，预警区域以颜色区分展示
- [x] 关键指标区域展示按环比的指标对比
- [x] 人工确认4项全部勾选后结账按钮可用
- [x] 损益结转预览/执行/反结账功能正常
- [x] 现有功能不受影响
