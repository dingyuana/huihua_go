# FE-TASK-1.3 | 资金账户管理页面

**版本**：V1.0
**优先级**：P1（MVP 核心）
**工时**：8-12h
**前置**：FE-TASK-0.4
**状态**：待开发

---

## 任务描述

实现银行账户/现金账户的列表展示、增删改、关联 GL 科目。

### 页面路径

`/setup/bank-accounts` → `BankAccountList.vue`

### 具体步骤

1. **API 对接**
   - `GET /bank-accounts` — 列表
   - `POST /bank-accounts` — 创建
   - `PUT /bank-accounts/:id` — 更新
   - `DELETE /bank-accounts/:id` — 停用（软删除）
   - `GET /bank-accounts/:id/balance` — 余额查询

2. **表格列**
   - 银行名称 | 账号（脱敏展示中间**) | 类型（银行/现金） | 关联科目 | 币种 | 状态（启用/停用） | 余额 | 操作

3. **新建/编辑弹窗**
   - 银行名称（输入）
   - 账号（输入，16-19 位数字校验）
   - 类型（`Bank` | `Cash` 单选）
   - 关联 GL 科目（使用 `AccountSelector` 组件，自动过滤 `account_type=asset` 的 Ledger 科目）
   - 币种（下拉：CNY/USD/HKD）
   - 创建时校验：`clearing_account_id` 必选，选中后不可修改

4. **余额卡片**
   - 表格上方展示各账户汇总卡片
   - 每卡片：账户名 + 当前余额 + 最后更新时间

---

## 验收标准

- [ ] 列表展示所有资金账户，余额列格式化为千分位
- [ ] 新增时 GL 科目选择器只显示资产类 Ledger 科目
- [ ] 创建后关联科目不可编辑（置灰）
- [ ] 删除（停用）后有业务引用的账户提示「该账户存在未结流水」
- [ ] 余额卡片刷新后更新

---

## 参考

- API 契约：`api-contracts/v1/setup-f1.md`「资金账户」
- FE-TASK-0.5 `AccountSelector.vue`
