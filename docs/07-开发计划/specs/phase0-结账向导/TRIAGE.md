# Phase 0 TRIAGE — 结账向导增强（已完成）

## 需求摘要

将慧话财务的期末结账页面从"3项检查卡片"升级为结构化的结账向导，包含风险预警、关键指标环比、计提检查、人工确认清单。

## 状态

**✅ 已实现（代码已完成）**

- 后端：`backend/app/api/v1/accounting.py` — 新增 `scan_reclassification_risks()`、`compute_key_indicators()`、`check_pending_accruals()` 三个函数，扩展 `pre_close_check` 返回值追加3个字段
- 前端：`frontend/src/pages/accounting/PeriodClose.vue` — 601行→907行，4卡片→7区域上下向导布局

## 变更清单

### 后端变更

| 函数 | 功能 | 行数 |
|------|------|:----:|
| `scan_reclassification_risks()` | 扫描应收/应付科目异常方向，返回重分类预警 | ~55 |
| `compute_key_indicators()` | 计算毛利率+期间费用率，环比上月，变化>5%标红 | ~110 |
| `check_pending_accruals()` | 检查折旧计提+税金申报是否完成 | ~70 |

返回值追加 `risk_warnings`、`key_indicators`、`pending_accruals`，现有字段不变。

### 前端变更

| 步骤 | 区域 | 说明 |
|:----:|------|------|
| 1 | 期间选择 | 原有，不变 |
| 2 | 凭证统计 | 大数字展示 |
| 3 | 结账前检查 | 原3项 + 关键指标子区域（毛利率/费用率环比） |
| 4 | 风险预警 | risk_warnings 按严重度展示 |
| 5 | 人工确认清单 | **新增** — 4个勾选框，全部勾选后结账按钮才可用 |
| 6 | 损益结转 | 原有 |
| 7 | 月末结账 | 原有 + 结账按钮受人工确认清单约束 |

## 验收标准（已满足）

- [x] 存在应收贷方余额时，risk_warnings 返回重分类提示
- [x] 毛利率环比下降超过5%时 alert=true
- [x] 上月数据不存在时 last_value=null
- [x] 有使用中资产但无折旧记录时，pending_accruals 标记 missing=true
- [x] 现有 pre_close_check 字段不变
- [x] 人工确认4项全部勾选后结账按钮可用
- [x] 损益结转预览/执行/反结账功能正常
