# SPEC: TASK01 — 银行流水导入前端入口

## 基本信息

- **任务 ID**: phase5-opt-001
- **类型**: feature
- **优先级**: medium
- **依赖**: 无（后端API已就绪）
- **估工**: 0.5天

## 背景

功能完成度矩阵记录：后端有 `cash/transactions/import` API，前端缺导入按钮。

## 目标

在 TransactionList 页面添加"导入Excel"按钮，调用后端API上传并解析银行流水。

## 验收

- [ ] TransactionList 页面出现导入按钮
- [ ] 选择Excel文件 → 上传 → 解析 → 显示导入结果
- [ ] 导入成功后流水列表刷新
