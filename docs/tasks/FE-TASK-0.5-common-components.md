# FE-TASK-0.5 | 公共业务组件

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：8-12h
**前置**：FE-TASK-0.2
**状态**：待开发

---

## 任务描述

开发 5 个核心复用业务组件，供所有页面使用。

### 具体步骤

1. **AccountSelector.vue** — 科目选择器（弹窗模式）

```
Props: modelValue (Account | null), disabledGroup (boolean=true), ledgerOnly (boolean)
Emits: update:modelValue
功能:
  - 点击输入框弹出科目树弹窗
  - 搜索过滤（输入编码或名称即时过滤）
  - 颜色区分：Group 科目灰色+斜体（不可选），Ledger 科目正常（可选）
  - 选择后显示标签：`1001-01 银行存款-工行`
  - support v-model
```

2. **AmountInput.vue** — 金额输入组件

```
Props: modelValue (string), currency (string='CNY'), max (string|null), showBalance (boolean)
Emits: update:modelValue
功能:
  - 输入时实时千分位格式化（1,234.56）
  - 负数显示红色
  - 超过 max 时输入框边框变红 + 提示
  - 前缀显示币种标签（¥/$/€）
  - 支持 Tab 键跳到下一字段
  - 通过 <input type="text"> + 正则实现，不用 type="number"（避免精度问题）
```

3. **DocStatusTag.vue** — 单据状态标签

```
Props: docstatus (0|1|2)
渲染:
  - 0(Draft): 灰色标签「草稿」
  - 1(Submitted): 蓝色标签「已审核」
  - 2(Cancelled): 红色标签「已作废」
  - 可选：新增 < 映射到黄色「待处理」
```

4. **PartySelector.vue** — 客商选择器（远程搜索）

```
Props: modelValue (Party | null), partyType ('customer'|'supplier'|'both')
Emits: update:modelValue
功能:
  - 输入 ≥2 字符触发远程搜索（debounce 300ms）
  - 下拉展示：名称 | 税号 | 开户行
  - 已选后显示 el-tag 可清除
```

5. **PeriodPicker.vue** — 会计期间选择器

```
Props: modelValue (string), yearRange ([number, number])
Emits: update:modelValue
功能:
  - 年份下拉 + 月份下拉联动
  - 只显示已创建的期间（防止选出不存在的期间）
  - 快速选项：本期、上期、本年累计
```

---

## 验收标准

- [ ] 5 个组件均可独立使用，Props 类型完整
- [ ] AccountSelector 搜索后正确过滤树节点，Group 科目点击无效
- [ ] AmountInput 输入 `1234567.89` 显示为 `1,234,567.89`
- [ ] AmountInput 传入 `max="10000"`，输入 `15000` 时提示超限
- [ ] DocStatusTag 正确渲染三种状态颜色
- [ ] PartySelector 搜索 debounce 有效，下拉展示正确
- [ ] PeriodPicker 年份和月份联动正确

---

## 参考

- 架构文档：第 9 章公共业务组件
- Element Plus Tree：https://element-plus.org/zh-CN/component/tree.html
