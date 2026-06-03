# TASK-F1.1 | F1 | 账套与科目表管理

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P1（MVP 核心）
**状态**：待开发

---

## 任务描述

实现账套创建向导和科目表管理模块：

### 1.1 账套创建向导

- 财务年度配置（1 月 1 日起或自定义起始月）
- 会计期间配置（自然月度 12 期，或自定义期间数）
- 启用日期设置
- 账套创建后，期初数据修改须主管审批

### 1.2 科目表 CRUD

- 树形结构展示（前端 Tree 组件）
- is_group 严格区分：Group（汇总，不可记账）vs Ledger（明细，可记账）
- 自动生成 4-2-2-2 编码（如 1001-01-00-00）
- 科目类型（account_type）决定默认余额方向：Asset/Expense → 借方，Liability/Income/Equity → 贷方
- 支持为科目绑定辅助核算维度（成本中心/项目/部门）
- 内置《小企业会计准则》标准科目表，账套创建时一键初始化

### 1.3 资金账户管理

- Bank Account 主数据：bank_name / account_number / clearing_account（GL 关联科目） / iban / swift_code
- 区分"银行存款"（Bank）与"库存现金"（Cash）类型
- 每个 Bank Account 关联唯一的 GL Account，支持多币种

### 1.4 客商/部门/员工档案

- 客商档案：名称 / 税号 / 开户行 / 账号 / 信用额度 / 账期策略
- 支持 Excel 批量导入
- 客商类型：customer / supplier / both

---

## 验收标准

- [ ] 账套创建向导完成，创建后基础科目表自动生成
- [ ] 科目 Tree 支持展开/折叠/新增/编辑/删除
- [ ] 尝试为 Group 科目创建记账凭证时，系统报错并阻止保存
- [ ] Bank Account 绑定 GL clearing_account，提交后数据一致
- [ ] 客商 Excel 批量导入，100 条数据 < 5 秒完成
- [ ] 所有 API 查询自动携带 tenant_id 过滤（RLS）

---

## 前置依赖

TASK-F0.2（核心数据模型），需要先有 accounts 表和 bank_accounts 表

---

## 预计工时

- 最小：32h
- 最大：64h

---

## 技术提示

### 科目 Tree 前端实现

```javascript
// 使用 Element Plus Tree，懒加载子节点
<el-tree
  :props="{ label: 'name', children: 'children' }"
  :load="loadAccountNodes"
  lazy
>
</el-tree>

// 后端：GET /api/v1/accounts/tree?parent_id=nil
// 返回所有根节点（含 direct_children），前端按需展开
```

### 编码自动生成

```
规则：父编码 + "-NN"
父科目 1001（银行存款）下的第一个子科目：1001-01
子科目 1001-01 下的第一个孙科目：1001-01-00
深度限制：4 层（编码共 4 段）

实现：在 Service 层，插入时查询 MAX(code) WHERE parent_id = $parent_id
```

### Bank Account 关联 GL Account

```go
// Bank Account 创建时必须指定 clearing_account_id
type BankAccount struct {
    ClearingAccountID uuid.UUID `json:"clearing_account_id"`
    // 校验：clearing_account.account_type 必须是 asset 且 root_type = debit
}
```

### 参考资料

- ERPNext：accounts/doctype/account/account.json
- Element Plus Tree：https://element-plus.org/zh-CN/component/tree.html

---

## 上下文信息（架构师决策记录）

- **决策**：科目表内置《小企业会计准则》而非空白账套，降低用户初始化成本
- **决策**：Bank Account 的 clearing_account 必须在创建时绑定，不允许后续修改（因为历史凭证依赖该科目）
- **风险**：Excel 批量导入时，若客商税号重复，以最后一个为准还是报错？建议报错并列出重复行号，让用户修正后重新导入