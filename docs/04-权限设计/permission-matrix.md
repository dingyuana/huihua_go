# 慧话财务 · 权限矩阵设计文档

> 本文档记录慧话财务 V6.0 版本的 RBAC 角色权限设计，作为权限变更的权威参考。
> 建立时间：2026-05-23 | 最后更新：2026-05-26

---

## 一、角色定义

### 1.1 角色类型总览

| 前端角色名 | 后端 `role_type` | 数据库 `role_code` | 说明 |
|---|---|---|---|
| `system_admin` | `isSystemAdmin=true` | `superadmin` | 系统管理员（最高权限，独立于 RBAC） |
| `company_admin` | `setup` | `admin` | 公司管理员 — 负责系统初始化、科目映射、用户管理 |
| `finance_manager` | `finance` | `finance_manager` | 财务经理 — 审批、查看所有报表、税务管理 |
| `accountant` | `accounting` | `accountant` | 会计 — 凭证录入、审核、发票、工资、资产等核心记账工作 |
| `cashier` | `cash` | `cashier` | 出纳 — 资金收付、银行账户、发票（收票/开票）、日记账 |
| `company_manager` | `query` | `company_manager` | 公司经理 — 查看报表、凭证总览，无录入权限 |

### 1.2 角色职责分工

| 职责 | system_admin | company_admin | finance_manager | accountant | cashier | company_manager |
|---|---|---|---|---|---|---|
| 系统初始化/配置 | ✅ 全权 | ✅ 公司级 | | | | |
| 用户/角色管理 | | ✅ | | | | |
| **资金收付** | | | 审批 | | ✅ 录入/操作 | |
| **业务单据**（收款/付款/报销） | | | 审批 | ✅ 审核凭证 | ✅ 录入/操作 | |
| **凭证**（制单/审核/过账） | | | 审批 | ✅ 录入/审核 | | |
| **发票**（领购/开具） | | | | ✅ 审核 | ✅ 开具/领购 | |
| **账簿**（科目/余额表/明细账） | | | 审批 | ✅ 录入 | 🔲 查看（不允许） | |
| **报表**（资产负债表/利润表） | | | ✅ 查看 | ✅ 查看 | 🔲 查看（不允许） | ✅ 查看 |
| **税务** | | | ✅ 管理 | | | |
| **工资/资产/预算/成本** | | | 审批 | ✅ 录入 | | |

> ✅ = 有权限  🔲 = 明确无权限（后端拒绝）

### 1.3 前端角色标识映射（`MainLayout.vue`）

```javascript
const role = computed(() => {
  const map = {
    accounting: 'accountant',
    cash: 'cashier',
    query: 'company_manager',
    setup: 'company_admin',
    finance: 'finance_manager',
    finance_manager: 'finance_manager',
    company_admin: 'company_admin'
  }
  return map[roleType.value] || roleType.value || ''
})

const isAccountant = computed(() => role.value === 'accountant')
const isCashier     = computed(() => role.value === 'cashier')
const isFinanceManager = computed(() => role.value === 'finance_manager')
const isCompanyManager = computed(() => role.value === 'company_manager')
const isCompanyAdmin   = computed(() => role.value === 'company_admin')
const isSystemAdmin   = computed(() => localStorage.getItem('isSystemAdmin') === '1' ...)
```

---

## 二、API 权限矩阵

### 2.1 路径规范说明

- `main.py` 中的 `prefix` 为实际 HTTP 路由前缀
- 部分文件内部路径（如 `@router.get("/assets")`）需结合 `main.py` 前缀才能拼出完整路径
- **重要**：`fixed_asset.py` 和 `archive.py` 的前端导航路径（`/fixed_asset/list`、`/archive/list`）与 API 实际路径（`/fixed-assets/assets`、`/archive/documents`）**不一致**，但前端 API 封装已正确使用实际路径，功能正常——此为历史遗留，前端导航 URL 无需改动。

### 2.2 各模块权限（2026-05-23 实测）

| 模块 | API 路径 | finance | manager(query) | accountant | cashier | 前端导航条件 |
|---|---|:---:|:---:|:---:|:---:|:---|
| **报表** 资产负债表 | `/api/v1/reports/balance/calculate` | 200 | 200 | 200 | 403 | `isAccountant \|\| isFinanceManager \|\| isCompanyManager` |
| **报表** 利润表 | `/api/v1/reports/profit/calculate` | 200 | 200 | 200 | 403 | 同上 |
| **报表** 现金流量表 | `/api/v1/reports/cashflow/calculate` | 200 | 200 | 200 | 403 | 同上 |
| **业务** 收款单列表 | `/api/v1/business/receipts` | 200 | **200✅** | 200 | 200 | `isCashier \|\| isFinanceManager` |
| **业务** 付款单列表 | `/api/v1/business/payments` | 200 | **200✅** | 200 | 200 | `isCashier \|\| isFinanceManager` |
| **业务** 费用报销单列表 | `/api/v1/business/reimbursements` | 200 | **200✅** | 200 | 200 | `isCashier \|\| isFinanceManager` |
| **业务** 科目映射 | `/api/v1/business/mappings` | 200 | 403 | 200 | 403 | `isCompanyAdmin \|\| isFinanceManager` |
| **资金** 银行账户 | `/api/v1/cash/banks` | 200 | **403❌** | 200 | 200 | `isCashier \|\| isFinanceManager` |
| **资金** 资金流水 | `/api/v1/cash/transactions` | 200 | **403❌** | 200 | 200 | 同上 |
| **发票** 发票领购 | `/api/v1/invoice/purchases` | 200 | **403❌** | 200 | 200 | `isCashier \|\| isFinanceManager \|\| isAccountant` |
| **发票** 发票开具 | `/api/v1/invoice/issues` | 200 | **403❌** | 200 | 200 | 同上 |
| **凭证** 凭证列表 | `/api/v1/accounting/vouchers` | 200 | 200 | 200 | 403 | `isAccountant \|\| isFinanceManager \|\| isCompanyManager` |
| **凭证** 新增凭证 | `/api/v1/accounting/vouchers/create` | 403 | 403 | 200 | 403 | `isAccountant` |
| **凭证** 审核/过账/反过账 | `/api/v1/accounting/vouchers/{id}/approve\|post\|reverse` | 200 | 403 | 200 | 403 | 凭证管理内 |
| **期末处理** | `/api/v1/period/close` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **账簿** 科目表 | `/api/v1/ledger/subjects` | 200 | **200✅** | 200 | 403 | `isAccountant \|\| isFinanceManager \|\| isCompanyAdmin` |
| **账簿** 总账 | `/api/v1/ledger/general-ledger` | 200 | 200 | 200 | 403 | *(无直接导航)* |
| **账簿** 明细账 | `/api/v1/ledger/detail` | 200 | 200 | 200 | 403 | *(无直接导航)* |
| **税务** 税种管理 | `/api/v1/tax/types` | 200 | 403 | 403 | 403 | `isFinanceManager` |
| **工资** 工资项目 | `/api/v1/wage/projects` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **资产** 资产列表 | `/api/v1/fixed-assets/assets` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **资产** 折旧计算 | `/api/v1/fixed-assets/assets/{id}/depreciation` | 200 | 403 | 200 | 403 | 同上 |
| **档案** 档案列表 | `/api/v1/archive/documents` | 200 | 403 | 200 | 403 | `isAccountant` |
| **档案** 借阅/归还 | `/api/v1/archive/lendings*` | 200 | 403 | 200 | 403 | 同上 |
| **应收** 应收发票 | `/api/v1/receivable/invoices` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **应付** 应付发票 | `/api/v1/payable/invoices` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **预算** 预算项目 | `/api/v1/budget/projects` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **成本** 成本中心 | `/api/v1/cost/centers` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |
| **辅助核算** 客户/供应商 | `/api/v1/auxiliary/customers\|suppliers` | 200 | 403 | 200 | 403 | `isAccountant \|\| isFinanceManager` |

### 2.3 `require_roles` 隐式别名规则（`security.py`）

以下跨角色权限是代码内置的，不需要在后端显式声明：

| 条件 | 效果 |
|---|---|
| `role_type == 'finance'` + 请求 `cash` 或 `accounting` 资源 | ✅ 放行（finance 跨读 cash/accounting） |
| 有 `system` 权限码 | ✅ 可访问 `setup` API（公司管理员） |
| 有 `cash:*` 系统权限 | ✅ 可访问 `cash` API |
| 有 `accounting:*` 系统权限 | ✅ 可访问 `accounting` API |

---

## 三、已知问题与修复状态

| # | 问题 | 根因 | 修复状态 |
|---|---|---|---|
| **P1** | 固定资产、档案管理 API 路径 404 | `fixed_asset.py` 前端导航用 `/fixed_asset/list`，但 API 实际路径是 `/fixed-assets/assets`；`archive.py` 类似 | ✅ 已确认：前端 API 封装使用正确路径，功能正常，无需修改 |
| **P2** | 公司经理(`query`)无法访问业务单据（收款/付款/报销） | `business.py` 的 receipts/payments/reimbursements 只有 `accounting, cash, finance`，没有 `query` | ⚠️ 待修复 |
| **P3** | 公司经理(`query`)无法访问科目管理 | `ledger.py` 的 `subjects` 只有 `accounting, finance, setup`，没有 `query` | ⚠️ 待修复 |
| **P4** | 出纳(`cash`)无法访问报表 | `security.py` 中 finance 可跨读 accounting 和 cash，但 cash 角色没有映射到 report 权限 | ✅ 预期设计（出纳不负责报表） |
| **P5** | 会计无法访问税务管理 | `tax.py` 只有 `finance`，没有 `accounting` | ✅ 预期设计（税务由财务经理管理） |

---

## 四、导航菜单与权限条件

| 导航区块 | 可见条件 | 包含页面 |
|---|---|---|
| 系统管理 | `isSystemAdmin \|\| isCompanyAdmin` | 用户/部门/角色管理、审计日志、数据备份、科目映射 |
| 财务会计 | `isAccountant \|\| isFinanceManager \|\| isCompanyManager` | 凭证管理、新增凭证（仅会计）、期末处理 |
| **日常单据** | `isCashier \|\| isFinanceManager` | 费用报销单、收款单、付款单 |
| 业务管理 | `isAccountant \|\| isFinanceManager \|\| isCompanyAdmin` | 科目管理、应收/应付发票、固定资产、工资核算、客户/供应商 |
| 资金管理 | `isCashier \|\| isFinanceManager` | 银行账户、资金流水 |
| 报表中心 | `isAccountant \|\| isFinanceManager \|\| isCompanyManager` | 资产负债表、利润表、现金流量表 |
| 发票管理 | `isCashier \|\| isFinanceManager \|\| isAccountant` | 发票领购、发票开具、发票导入 |
| 税务管理 | `isFinanceManager` | 税种管理、税务申报、税金计提 |
| 预算管理 | `isAccountant \|\| isFinanceManager` | 预算项目 |
| 成本管理 | `isAccountant \|\| isFinanceManager` | 成本中心 |
| 档案管理 | `isAccountant` | 档案管理 |

### 4.1 出纳专属导航条件

| 导航区块 | 可见条件 | 包含页面 |
|---|---|---|
| 资金管理 | `isCashier \|\| isFinanceManager` | 银行账户、资金流水 |
| 日常单据 | `isCashier \|\| isFinanceManager` | 收款单、付款单、费用报销单 |
| 发票管理 | `isCashier \|\| isFinanceManager \|\| isAccountant` | 发票领购、发票开具 |

---

## 五、与原设计文档差异说明

本文档为工程实现层精确到每个 API 端点的权限记录。
原设计层定义见 `docs/需求分析书.md` §2 和 `docs/技术方案书.md` §6.4（角色权限矩阵，SaaS阶段）。

### 5.1 角色映射对照

| 原设计角色 | 原设计中的职责 | 工程实现 `role_type` | 对应关系 |
|---|---|---|---|
| 老板 | 看经营状况/报表 | `query` (company_manager) | ✅ 同名映射 |
| 会计 | 日常制单、发票采集 | `accounting` (accountant) | ✅ 同名映射 |
| 主管会计 | 审核凭证、过账、审批关键操作 | `finance` (finance_manager) | ✅ 同名映射 |
| 出纳 | 录入银行流水、管理收付款 | `cash` (cashier) | ✅ 同名映射 |
| 代账主管 | 看团队效率、分配客户 | *(MVP不支持)* | ❌ 未实现 |
| 系统管理员 | *(原设计无此角色)* | `isSystemAdmin=true` (system_admin) | ⚠️ 新增（系统级） |

### 5.2 权限差异明细

| 权限项 | 原设计 | 工程实现 | 差异说明 |
|---|---|---|---|
| 出纳录入业务单据（收款/付款/报销） | 未明确提及 | 出纳可录入（`cash` 在 business.py 的 `accounting, cash, finance` 中） | ⚠️ 合理扩展：出纳是收款/付款的实际操作人，业务单据是资金收付的载体 |
| 出纳查看报表 | ❌ 不应看 | ❌ 后端 403 | ✅ 与原设计一致 |
| 公司经理查看业务单据 | 原设计无明确说明 | 修复后公司经理可查看业务单据列表和详情（审批视角） | ⚠️ 合理扩展：公司经理需了解业务单据以便审批 |
| 公司经理查看科目表 | 原设计无明确说明 | 修复后公司经理可查看科目表 | ⚠️ 合理扩展：公司经理需了解科目结构 |
| 科目映射配置权限 | 原设计无明确说明 | `setup`（公司管理员）和 `finance` 可访问 | ⚠️ 合理扩展：科目映射是配置行为 |
| 凭证新建权限 | 会计+主管会计 | 仅 `accounting` | ✅ 一致 |
| 凭证审核/过账权限 | 主管会计 | `finance` | ✅ 一致 |
| 税务管理权限 | 主管会计 | `finance` | ✅ 一致 |

### 5.3 未实现的原设计项

| 项目 | 原设计描述 | 状态 |
|---|---|---|
| 代账主管角色 | 代账公司主管，看团队效率、分配客户 | ❌ MVP 阶段不支持 |
| 强制审核人≠制单人 | 主管会计不可审核自己制作的凭证 | ❌ 后端未强制校验（`reviewer_id != creator_id`） |
| 金额阈值触发额外复核 | 大金额触发四眼原则 | ❌ 未实现 |
| 多租户行级权限 | 会计仅见自己负责的客户 | ❌ MVP 单租户 |

---

## 六、开发说明

### 6.1 前端开发服务

前端基于 Vite 开发，端口 3001，代理 `/api` 请求到后端容器 `huihua-backend:8000`。
由于 docker 网络限制，Vite 与 backend 分属不同网络（host vs `huihua-financial-master_hf-network`），需将 backend 端口映射到 host：

```bash
# 后端容器需要暴露 8000 端口
docker run -d --name huihua-backend --network huihua-financial-master_hf-network -p 8000:8000 ...

# 前端 Vite 配置
# vite.config.js 中 proxy target 指向 http://localhost:8000（host 上的 backend）
```

### 七、修改记录

| 日期 | 修改内容 |
|---|---|
| 2026-05-23 | 初始文档建立；基于 `backend/app/core/security.py` + `MainLayout.vue` + API 实测结果 |
| 2026-05-23 | 确认资产/档案路径前端 API 已正确封装，功能正常；修复公司经理业务单据和科目表权限；补充与原设计文档差异说明 |
| 2026-05-23 | 开发说明补充（Vite dev server 端口映射配置）；修正 manager(query) 业务单据和科目表状态从 403❌ → 200✅（已修复） |