# SPEC: TASK01 — 后端：扩展 pre_close_check 返回值（已完成）

## 基本信息

- **任务 ID**: phase0-close-001
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **状态**: ✅ 已完成

## 背景

当前 pre_close_check 只返回3项布尔值，缺少风险预警、关键指标环比、计提检查。后端已实现。

## 变更文件

- `backend/app/api/v1/accounting.py` — 新增3个函数 + 返回值扩展

## 验收标准（已满足）

- [x] risk_warnings 在应收贷方余额时返回结构正确的预警
- [x] key_indicators 能在有数据时计算环比
- [x] pending_accruals 能判断折旧是否已计提
- [x] 现有返回字段不变
- [x] 零数据库变更，零新路由
