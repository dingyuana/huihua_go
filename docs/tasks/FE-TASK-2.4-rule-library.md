# FE-TASK-2.4 | 规则库配置页面

**版本**：V1.0
**优先级**：P1（MVP 核心）
**工时**：8-12h
**前置**：FE-TASK-1.2
**状态**：待开发

---

## 任务描述

实现智能分类规则库和科目映射规则的配置界面。

### 页面路径

`/setup/rules` → `RuleLibrary.vue`
`/setup/mapping-rules` → `MappingRules.vue`

### 具体步骤

1. **API 对接**
   - `GET /classification-rules` — 规则列表
   - `POST /classification-rules` — 新建
   - `PUT /classification-rules/:id` — 更新
   - `DELETE /classification-rules/:id` — 删除
   - `POST /classification-rules/reorder` — 优先级排序
   - `GET /mapping-rules` — 科目映射列表
   - `PUT /mapping-rules/:id` — 更新映射

2. **分类规则库页面**

   **表格列**：规则名称 | 匹配模式（正则/关键词） | 匹配字段 | 金额方向 | 目标分类（标签） | 优先级 | 启用状态 | 操作
   
   **新建/编辑弹窗**：
   - 规则名称（输入）
   - 规则类型：`keyword | keyword_regex | counterparty_match`
   - 匹配模式输入（正则表达式高亮预览）
   - 金额方向：`in / out / both`
   - 目标分类：下拉（6 类）
   - 优先级：数字输入（越小越优先）
   - 启用开关

   **优先级拖拽**：
   - 使用 `el-table` 拖拽排序行
   - 拖拽后自动更新优先级数值

2. **科目映射规则页面**

   **表格列**：单据类型 | 借方科目 | 贷方科目 | 条件表达式 | 备选科目 | 操作

   **编辑弹窗**：
   - 单据类型：下拉（6 类）
   - 借方科目：`AccountSelector`（限 Ledger）
   - 贷方科目：`AccountSelector`（限 Ledger）
   - 条件表达式（可选）：如 `amount > 5000`
   - 备选科目（可选）
   - 条件满足时走备选科目

---

## 验收标准

- [ ] 分类规则列表展示，支持拖拽调整优先级
- [ ] 新建规则时正则表达式输入框实时高亮预览
- [ ] 规则禁用/启用后列表状态实时更新
- [ ] 科目映射规则中科目选择器只显示 Ledger 科目
- [ ] 映射规则修改后自动保存（或保存按钮）

---

## 参考

- API 契约：`api-contracts/v1/setup-f1.md`「智能分类规则库」「科目映射规则」
- FE-TASK-0.5 `AccountSelector.vue`
