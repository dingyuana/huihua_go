# 慧话财务 API 接口文档

> 生成时间: 2026-05-19 08:00
> 源码路径: `frontend/backend/app/api/v1/`
> 接口总数: 234
> 调用状态: ✅ 已调用 80 | ❌ 未调用 149 | ⚠️ 方法不匹配 5

---

## 调用状态说明

- ✅ **已调用**: 前端有调用后端接口
- ❌ **未调用**: 后端有接口但前端未调用
- ⚠️ **方法不匹配**: 路径存在但 HTTP 方法不同

---

## 目录

- [会计科目 (`accounting`)](#accounting) — 15 接口
- [档案管理 (`archive`)](#archive) — 11 接口
- [审计日志 (`audit`)](#audit) — 3 接口 ⚠️
- [认证授权 (`auth`)](#auth) — 3 接口 ⚠️
- [辅助核算 (`auxiliary`)](#auxiliary) — 10 接口 ⚠️
- [预算管理 (`budget`)](#budget) — 16 接口
- [现金银行 (`cash`)](#cash) — 12 接口
- [公司设置 (`company`)](#company) — 5 接口
- [成本管理 (`cost`)](#cost) — 15 接口
- [部门管理 (`depts`)](#depts) — 6 接口
- [固定资产 (`fixed_asset`)](#fixed_asset) — 10 接口
- [发票管理 (`invoice`)](#invoice) — 14 接口
- [账簿管理 (`ledger`)](#ledger) — 10 接口
- [应付账款 (`payable`)](#payable) — 13 接口
- [期间管理 (`period`)](#period) — 4 接口
- [权限管理 (`rbac`)](#rbac) — 7 接口
- [应收账款 (`receivable`)](#receivable) — 14 接口
- [财务报表 (`report`)](#report) — 20 接口
- [角色管理 (`roles`)](#roles) — 9 接口
- [税务管理 (`tax`)](#tax) — 13 接口
- [用户管理 (`users`)](#users) — 6 接口
- [工资管理 (`wage`)](#wage) — 18 接口

---

## 会计科目 (`accounting`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/vouchers` | ✅ 已调用 | accounting | 查询凭证列表，支持分页、筛选条件 |
| 🔵 GET | `/vouchers/generate-no` | ✅ 已调用 | accounting | 生成凭证编号 |
| 🟢 POST | `/vouchers` | ✅ 已调用 | accounting | 创建新凭证 |
| 🔵 GET | `/vouchers/{voucher_id}` | ❌ 未调用 | — | 查询单个凭证详情 |
| 🟡 PUT | `/vouchers/{voucher_id}` | ❌ 未调用 | — | 更新凭证信息 |
| 🔴 DELETE | `/vouchers/{voucher_id}` | ❌ 未调用 | — | 删除凭证 |
| 🟢 POST | `/vouchers/{voucher_id}/submit` | ❌ 未调用 | — | 提交凭证审核 |
| 🟢 POST | `/vouchers/{voucher_id}/reject` | ❌ 未调用 | — | 驳回凭证 |
| 🟢 POST | `/vouchers/{voucher_id}/approve` | ❌ 未调用 | — | 批准凭证 |
| 🟢 POST | `/vouchers/{voucher_id}/unpost` | ❌ 未调用 | — | 取消记账 |
| 🟢 POST | `/period/{year}/{month}/close` | ❌ 未调用 | — | 期末结账 |
| 🟢 POST | `/period/{year}/{month}/unclose` | ❌ 未调用 | — | 取消结账 |
| 🔵 GET | `/periods` | ❌ 未调用 | — | 查询期间列表 |
| 🔵 GET | `/vouchers/summary/monthly` | ❌ 未调用 | — | 月度凭证汇总 |
| 🔵 GET | `/period/{year}/{month}/pre-close-check` | ❌ 未调用 | — | 结账前检查 |

**小结**: ✅ 3 / ❌ 12 / ⚠️ 0

## 档案管理 (`archive`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/categories` | ✅ 已调用 | archive | 查询档案分类列表 |
| 🟢 POST | `/categories` | ✅ 已调用 | archive | 创建档案分类 |
| 🟡 PUT | `/categories/{category_id}` | ❌ 未调用 | — | 更新档案分类 |
| 🟢 POST | `/categories/init` | ✅ 已调用 | archive | 初始化档案分类 |
| 🔵 GET | `/documents` | ✅ 已调用 | archive | 查询档案文件列表 |
| 🟢 POST | `/documents` | ✅ 已调用 | archive | 上传档案文件 |
| 🟡 PUT | `/documents/{doc_id}` | ❌ 未调用 | — | 更新档案文件 |
| 🔴 DELETE | `/documents/{doc_id}` | ❌ 未调用 | — | 删除档案文件 |
| 🔵 GET | `/lendings` | ✅ 已调用 | archive | 查询借阅记录 |
| 🟢 POST | `/lendings` | ✅ 已调用 | archive | 创建借阅记录 |
| 🟢 POST | `/lendings/{lending_id}/return` | ❌ 未调用 | — | 归还档案 |

**小结**: ✅ 7 / ❌ 4 / ⚠️ 0

## 审计日志 (`audit`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/logs` | ❌ 未调用 | — | 查询审计日志列表 |
| 🔵 GET | `/logs/statistics` | ❌ 未调用 | — | 审计日志统计 |
| 🔵 GET | `/logs/export` | ❌ 未调用 | — | 导出审计日志 |

**小结**: ✅ 0 / ❌ 3 / ⚠️ 0

## 认证授权 (`auth`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🟢 POST | `/login` | ❌ 未调用 | — | 用户登录 |
| 🟢 POST | `/refresh` | ❌ 未调用 | — | 刷新访问令牌 |
| 🟢 POST | `/logout` | ❌ 未调用 | — | 用户登出 |

**小结**: ✅ 0 / ❌ 3 / ⚠️ 0

## 辅助核算 (`auxiliary`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/customers` | ❌ 未调用 | — | 查询客户列表 |
| 🟢 POST | `/customers` | ❌ 未调用 | — | 创建客户 |
| 🟡 PUT | `/customers/{customer_id}` | ❌ 未调用 | — | 更新客户信息 |
| 🔴 DELETE | `/customers/{customer_id}` | ❌ 未调用 | — | 删除客户 |
| 🔵 GET | `/suppliers` | ❌ 未调用 | — | 查询供应商列表 |
| 🟢 POST | `/suppliers` | ❌ 未调用 | — | 创建供应商 |
| 🟡 PUT | `/suppliers/{supplier_id}` | ❌ 未调用 | — | 更新供应商信息 |
| 🔴 DELETE | `/suppliers/{supplier_id}` | ❌ 未调用 | — | 删除供应商 |
| 🔵 GET | `/projects` | ❌ 未调用 | — | 查询项目列表 |
| 🟢 POST | `/projects` | ❌ 未调用 | — | 创建项目 |

**小结**: ✅ 0 / ❌ 10 / ⚠️ 0

## 预算管理 (`budget`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/projects` | ✅ 已调用 | budget | 查询预算项目列表 |
| 🟢 POST | `/projects` | ✅ 已调用 | budget | 创建预算项目 |
| 🟡 PUT | `/projects/{project_id}` | ❌ 未调用 | — | 更新预算项目 |
| 🔴 DELETE | `/projects/{project_id}` | ❌ 未调用 | — | 删除预算项目 |
| 🟢 POST | `/projects/init` | ✅ 已调用 | budget | 初始化预算项目 |
| 🔵 GET | `/plans` | ✅ 已调用 | budget | 查询预算计划列表 |
| 🔵 GET | `/plans/{plan_id}` | ❌ 未调用 | — | 查询预算计划详情 |
| 🟢 POST | `/plans` | ✅ 已调用 | budget | 创建预算计划 |
| 🟡 PUT | `/plans/{plan_id}` | ❌ 未调用 | — | 更新预算计划 |
| 🟢 POST | `/plans/{plan_id}/submit` | ❌ 未调用 | — | 提交预算计划 |
| 🟢 POST | `/plans/{plan_id}/approve` | ❌ 未调用 | — | 批准预算计划 |
| 🔵 GET | `/adjustments` | ✅ 已调用 | budget | 查询预算调整列表 |
| 🟢 POST | `/adjustments` | ✅ 已调用 | budget | 创建预算调整 |
| 🔵 GET | `/executions` | ✅ 已调用 | budget | 查询预算执行列表 |
| 🟢 POST | `/executions` | ✅ 已调用 | budget | 记录预算执行 |
| 🔵 GET | `/analysis` | ✅ 已调用 | budget | 预算分析报告 |

**小结**: ✅ 10 / ❌ 6 / ⚠️ 0

## 现金银行 (`cash`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/banks` | ✅ 已调用 | cash | 查询银行账户列表 |
| 🟢 POST | `/banks` | ✅ 已调用 | cash | 创建银行账户 |
| 🔵 GET | `/banks/{bank_id}` | ❌ 未调用 | — | 查询银行账户详情 |
| 🟡 PUT | `/banks/{bank_id}` | ❌ 未调用 | — | 更新银行账户 |
| 🔴 DELETE | `/banks/{bank_id}` | ❌ 未调用 | — | 删除银行账户 |
| 🔵 GET | `/transactions` | ✅ 已调用 | cash | 查询现金交易列表 |
| 🟢 POST | `/transactions` | ⚠️ 方法不匹配(前端:GET) | cash | 创建现金交易 |
| 🔵 GET | `/transactions/{txn_id}` | ❌ 未调用 | — | 查询交易详情 |
| 🟡 PUT | `/transactions/{txn_id}` | ❌ 未调用 | — | 更新交易记录 |
| 🔴 DELETE | `/transactions/{txn_id}` | ❌ 未调用 | — | 删除交易记录 |
| 🔵 GET | `/transactions/reconcile` | ❌ 未调用 | — | 查询对账记录 |
| 🟢 POST | `/transactions/reconcile` | ❌ 未调用 | — | 执行银行对账 |

**小结**: ✅ 3 / ❌ 8 / ⚠️ 1

## 公司设置 (`company`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `` | ✅ 已调用 | system | 查询公司信息 |
| 🟡 PUT | `` | ✅ 已调用 | system | 更新公司信息 |
| 🟢 POST | `/init` | ✅ 已调用 | system | 初始化公司数据 |
| 🟢 POST | `/account-books` | ❌ 未调用 | — | 创建账套 |
| 🟢 POST | `/reset` | ❌ 未调用 | — | 重置公司数据 |

**小结**: ✅ 3 / ❌ 2 / ⚠️ 0

## 成本管理 (`cost`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/centers` | ✅ 已调用 | cost | 查询成本中心列表 |
| 🟢 POST | `/centers` | ✅ 已调用 | cost | 创建成本中心 |
| 🟡 PUT | `/centers/{center_id}` | ❌ 未调用 | — | 更新成本中心 |
| 🔴 DELETE | `/centers/{center_id}` | ❌ 未调用 | — | 删除成本中心 |
| 🟢 POST | `/centers/init` | ✅ 已调用 | cost | 初始化成本中心 |
| 🔵 GET | `/elements` | ✅ 已调用 | cost | 查询成本要素列表 |
| 🟢 POST | `/elements` | ✅ 已调用 | cost | 创建成本要素 |
| 🟡 PUT | `/elements/{element_id}` | ❌ 未调用 | — | 更新成本要素 |
| 🟢 POST | `/elements/init` | ✅ 已调用 | cost | 初始化成本要素 |
| 🔵 GET | `/records` | ✅ 已调用 | cost | 查询成本记录列表 |
| 🟢 POST | `/records` | ✅ 已调用 | cost | 创建成本记录 |
| 🔵 GET | `/rules` | ✅ 已调用 | cost | 查询成本分配规则 |
| 🟢 POST | `/rules` | ✅ 已调用 | cost | 创建成本分配规则 |
| 🔴 DELETE | `/rules/{rule_id}` | ❌ 未调用 | — | 删除成本分配规则 |
| 🔵 GET | `/report` | ✅ 已调用 | cost | 成本报表 |

**小结**: ✅ 11 / ❌ 4 / ⚠️ 0

## 部门管理 (`depts`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/tree` | ✅ 已调用 | system | 获取部门树形结构 |
| 🔵 GET | `` | ✅ 已调用 | system | 查询部门列表 |
| 🔵 GET | `/{dept_id}` | ❌ 未调用 | — | 查询部门详情 |
| 🟢 POST | `` | ✅ 已调用 | system | 创建部门 |
| 🟡 PUT | `/{dept_id}` | ❌ 未调用 | — | 更新部门信息 |
| 🔴 DELETE | `/{dept_id}` | ❌ 未调用 | — | 删除部门 |

**小结**: ✅ 3 / ❌ 3 / ⚠️ 0

## 固定资产 (`fixed_asset`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/assets` | ✅ 已调用 | fixed_asset | 查询固定资产列表 |
| 🔵 GET | `/assets/{asset_id}` | ❌ 未调用 | — | 查询固定资产详情 |
| 🟢 POST | `/assets` | ✅ 已调用 | fixed_asset | 新增固定资产 |
| 🟡 PUT | `/assets/{asset_id}` | ❌ 未调用 | — | 更新固定资产 |
| 🔴 DELETE | `/assets/{asset_id}` | ❌ 未调用 | — | 删除固定资产 |
| 🟢 POST | `/assets/{asset_id}/depreciation` | ❌ 未调用 | — | 计提折旧 |
| 🔵 GET | `/assets/{asset_id}/depreciations` | ❌ 未调用 | — | 查询折旧记录 |
| 🔵 GET | `/statistics` | ✅ 已调用 | fixed_asset | 固定资产统计 |
| 🔵 GET | `/categories` | ✅ 已调用 | fixed_asset | 查询资产分类 |
| 🟢 POST | `/categories` | ⚠️ 方法不匹配(前端:GET) | fixed_asset | 创建资产分类 |

**小结**: ✅ 4 / ❌ 5 / ⚠️ 1

## 发票管理 (`invoice`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/purchases` | ✅ 已调用 | invoice | 查询采购发票列表 |
| 🔵 GET | `/purchases/{purchase_id}` | ❌ 未调用 | — | 查询采购发票详情 |
| 🟢 POST | `/purchases` | ✅ 已调用 | invoice | 创建采购发票 |
| 🟡 PUT | `/purchases/{purchase_id}` | ❌ 未调用 | — | 更新采购发票 |
| 🔴 DELETE | `/purchases/{purchase_id}` | ❌ 未调用 | — | 删除采购发票 |
| 🔵 GET | `/issues` | ✅ 已调用 | invoice | 查询销项发票列表 |
| 🔵 GET | `/issues/{issue_id}` | ❌ 未调用 | — | 查询销项发票详情 |
| 🟢 POST | `/issues` | ✅ 已调用 | invoice | 创建销项发票 |
| 🟡 PUT | `/issues/{issue_id}` | ❌ 未调用 | — | 更新销项发票 |
| 🔴 DELETE | `/issues/{issue_id}` | ❌ 未调用 | — | 删除销项发票 |
| 🟢 POST | `/issues/{issue_id}/void` | ❌ 未调用 | — | 作废发票 |
| 🟢 POST | `/issues/{issue_id}/redflush` | ❌ 未调用 | — | 红冲发票 |
| 🔵 GET | `/statistics` | ✅ 已调用 | invoice | 发票统计 |
| 🔵 GET | `/available-numbers` | ✅ 已调用 | invoice | 查询可用发票号 |

**小结**: ✅ 6 / ❌ 8 / ⚠️ 0

## 账簿管理 (`ledger`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/subjects` | ✅ 已调用 | ledger | 查询会计科目列表 |
| 🔵 GET | `/subjects/{subject_id}` | ❌ 未调用 | — | 查询科目详情 |
| 🟢 POST | `/subjects` | ✅ 已调用 | ledger | 创建会计科目 |
| 🟡 PUT | `/subjects/{subject_id}` | ❌ 未调用 | — | 更新会计科目 |
| 🔴 DELETE | `/subjects/{subject_id}` | ❌ 未调用 | — | 删除会计科目 |
| 🔵 GET | `/accounts/{year}/{month}` | ❌ 未调用 | — | 查询月度账簿 |
| 🟢 POST | `/accounts` | ❌ 未调用 | — | 创建账簿记录 |
| 🟡 PUT | `/accounts/{account_id}` | ❌ 未调用 | — | 更新账簿记录 |
| 🔵 GET | `/general-ledger` | ✅ 已调用 | report | 查询总账 |
| 🔵 GET | `/detail-ledger` | ✅ 已调用 | report | 查询明细账 |

**小结**: ✅ 4 / ❌ 6 / ⚠️ 0

## 应付账款 (`payable`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/invoices` | ✅ 已调用 | payable | 查询应付发票列表 |
| 🔵 GET | `/invoices/{invoice_id}` | ❌ 未调用 | — | 查询应付发票详情 |
| 🟢 POST | `/invoices` | ✅ 已调用 | payable | 创建应付发票 |
| 🟡 PUT | `/invoices/{invoice_id}` | ❌ 未调用 | — | 更新应付发票 |
| 🔴 DELETE | `/invoices/{invoice_id}` | ❌ 未调用 | — | 删除应付发票 |
| 🟡 PUT | `/invoices/{invoice_id}/void` | ❌ 未调用 | — | 作废应付发票 |
| 🔵 GET | `` | ❌ 未调用 | — | 查询应付账款列表 |
| 🔵 GET | `/{payable_id}` | ❌ 未调用 | — | 查询应付账款详情 |
| 🟢 POST | `` | ❌ 未调用 | — | 创建应付账款 |
| 🟡 PUT | `/{payable_id}` | ❌ 未调用 | — | 更新应付账款 |
| 🔴 DELETE | `/{payable_id}` | ❌ 未调用 | — | 删除应付账款 |
| 🟡 PUT | `/{payable_id}/writeoff` | ❌ 未调用 | — | 核销应付账款 |
| 🔵 GET | `/invoices/aging` | ❌ 未调用 | — | 应付账龄分析 |

**小结**: ✅ 2 / ❌ 11 / ⚠️ 0

## 期间管理 (`period`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🟢 POST | `/profit-loss-entry` | ✅ 已调用 | report | 结转损益 |
| 🔵 GET | `/profit-loss-balance` | ❌ 未调用 | — | 查询损益余额 |
| 🔵 GET | `/profit-loss-status` | ❌ 未调用 | — | 查询损益状态 |
| 🔵 GET | `/profit-loss-preview` | ✅ 已调用 | report | 预览损益结转 |

**小结**: ✅ 2 / ❌ 2 / ⚠️ 0

## 权限管理 (`rbac`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/permissions` | ✅ 已调用 | system | 查询权限列表 |
| 🔵 GET | `/roles` | ❌ 未调用 | — | 查询角色列表 |
| 🔵 GET | `/roles/{role_id}` | ❌ 未调用 | — | 查询角色详情 |
| 🟢 POST | `/roles` | ❌ 未调用 | — | 创建角色 |
| 🟡 PUT | `/roles/{role_id}` | ❌ 未调用 | — | 更新角色 |
| 🔴 DELETE | `/roles/{role_id}` | ❌ 未调用 | — | 删除角色 |
| 🔵 GET | `/users/{user_id}/permissions` | ❌ 未调用 | — | 查询用户权限 |

**小结**: ✅ 1 / ❌ 6 / ⚠️ 0

## 应收账款 (`receivable`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/invoices` | ✅ 已调用 | receivable | 查询应收发票列表 |
| 🔵 GET | `/invoices/{invoice_id}` | ❌ 未调用 | — | 查询应收发票详情 |
| 🟢 POST | `/invoices` | ✅ 已调用 | receivable | 创建应收发票 |
| 🟡 PUT | `/invoices/{invoice_id}` | ❌ 未调用 | — | 更新应收发票 |
| 🔴 DELETE | `/invoices/{invoice_id}` | ❌ 未调用 | — | 删除应收发票 |
| 🟡 PUT | `/invoices/{invoice_id}/void` | ❌ 未调用 | — | 作废应收发票 |
| 🔵 GET | `` | ❌ 未调用 | — | 查询应收账款列表 |
| 🔵 GET | `/{receivable_id}` | ❌ 未调用 | — | 查询应收账款详情 |
| 🟢 POST | `` | ❌ 未调用 | — | 创建应收账款 |
| 🟡 PUT | `/{receivable_id}` | ❌ 未调用 | — | 更新应收账款 |
| 🔴 DELETE | `/{receivable_id}` | ❌ 未调用 | — | 删除应收账款 |
| 🟡 PUT | `/{receivable_id}/writeoff` | ❌ 未调用 | — | 核销应收账款 |
| 🔵 GET | `/invoices/aging` | ❌ 未调用 | — | 应收账龄分析 |
| 🔵 GET | `/invoices/warnings` | ❌ 未调用 | — | 逾期预警 |

**小结**: ✅ 2 / ❌ 12 / ⚠️ 0

## 财务报表 (`report`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/balance` | ❌ 未调用 | — | 查询资产负债表 |
| 🟢 POST | `/balance` | ❌ 未调用 | — | 生成资产负债表 |
| 🔵 GET | `/balance/list` | ❌ 未调用 | — | 查询资产负债表历史 |
| 🔵 GET | `/profit` | ✅ 已调用 | report | 查询利润表 |
| 🟢 POST | `/profit` | ⚠️ 方法不匹配(前端:GET) | report | 生成利润表 |
| 🔵 GET | `/profit/list` | ❌ 未调用 | — | 查询利润表历史 |
| 🔵 GET | `/cashflow` | ✅ 已调用 | report | 查询现金流量表 |
| 🟢 POST | `/cashflow` | ⚠️ 方法不匹配(前端:GET) | report | 生成现金流量表 |
| 🔵 GET | `/cashflow/list` | ❌ 未调用 | — | 查询现金流量表历史 |
| 🔵 GET | `/balance/calculate` | ✅ 已调用 | report | 计算资产负债表 |
| 🔵 GET | `/profit/calculate` | ❌ 未调用 | — | 计算利润表 |
| 🔵 GET | `/cashflow/calculate` | ❌ 未调用 | — | 计算现金流量表 |
| 🔵 GET | `/balance/export` | ❌ 未调用 | — | 导出资产负债表 |
| 🔵 GET | `/profit/export` | ❌ 未调用 | — | 导出利润表 |
| 🟢 POST | `/{report_type}/approve` | ❌ 未调用 | — | 报表审批 |
| 🔵 GET | `/{report_type}/approval-status` | ❌ 未调用 | — | 查询报表审批状态 |
| 🟢 POST | `/{report_type}/record-export` | ❌ 未调用 | — | 导出报表记录 |
| 🔵 GET | `/expense_detail` | ❌ 未调用 | — | 费用明细 |
| 🔵 GET | `/expense_by_dept` | ❌ 未调用 | — | 部门费用分析 |
| 🔵 GET | `/revenue_detail` | ❌ 未调用 | — | 收入明细 |

**小结**: ✅ 3 / ❌ 15 / ⚠️ 2

## 角色管理 (`roles`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `` | ✅ 已调用 | system | 查询角色列表 |
| 🔵 GET | `/permissions` | ✅ 已调用 | system | 查询角色权限列表 |
| 🔵 GET | `/{role_id}` | ❌ 未调用 | — | 查询角色详情 |
| 🟢 POST | `` | ✅ 已调用 | system | 创建角色 |
| 🟡 PUT | `/{role_id}` | ❌ 未调用 | — | 更新角色 |
| 🔴 DELETE | `/{role_id}` | ❌ 未调用 | — | 删除角色 |
| 🟢 POST | `/permissions` | ⚠️ 方法不匹配(前端:GET) | system | 批量设置角色权限 |
| 🟡 PUT | `/permissions/{permission_id}` | ❌ 未调用 | — | 更新角色权限 |
| 🔴 DELETE | `/permissions/{permission_id}` | ❌ 未调用 | — | 删除角色权限 |

**小结**: ✅ 3 / ❌ 5 / ⚠️ 1

## 税务管理 (`tax`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/types` | ✅ 已调用 | tax | 查询税种列表 |
| 🟢 POST | `/types` | ✅ 已调用 | tax | 创建税种 |
| 🔵 GET | `/types/{tax_id}` | ❌ 未调用 | — | 查询税种详情 |
| 🟡 PUT | `/types/{tax_id}` | ❌ 未调用 | — | 更新税种 |
| 🔴 DELETE | `/types/{tax_id}` | ❌ 未调用 | — | 删除税种 |
| 🔵 GET | `/declarations` | ✅ 已调用 | tax | 查询纳税申报列表 |
| 🟢 POST | `/declarations` | ✅ 已调用 | tax | 创建纳税申报 |
| 🔵 GET | `/declarations/{decl_id}` | ❌ 未调用 | — | 查询纳税申报详情 |
| 🟡 PUT | `/declarations/{decl_id}` | ❌ 未调用 | — | 更新纳税申报 |
| 🔴 DELETE | `/declarations/{decl_id}` | ❌ 未调用 | — | 删除纳税申报 |
| 🟡 PUT | `/declarations/{decl_id}/submit` | ❌ 未调用 | — | 提交纳税申报 |
| 🟡 PUT | `/declarations/{decl_id}/approve` | ❌ 未调用 | — | 审批纳税申报 |
| 🟡 PUT | `/declarations/{decl_id}/file` | ❌ 未调用 | — | 归档纳税申报 |

**小结**: ✅ 4 / ❌ 9 / ⚠️ 0

## 用户管理 (`users`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `` | ✅ 已调用 | system | 查询用户列表 |
| 🔵 GET | `/{user_id}` | ❌ 未调用 | — | 查询用户详情 |
| 🟢 POST | `` | ✅ 已调用 | system | 创建用户 |
| 🟡 PUT | `/{user_id}` | ❌ 未调用 | — | 更新用户信息 |
| 🔴 DELETE | `/{user_id}` | ❌ 未调用 | — | 删除用户 |
| 🟡 PUT | `/{user_id}/password` | ❌ 未调用 | — | 修改用户密码 |

**小结**: ✅ 2 / ❌ 4 / ⚠️ 0

## 工资管理 (`wage`)

| 方法 | 路径 | 状态 | 前端模块 | 说明 |
|------|------|------|----------|------|
| 🔵 GET | `/projects` | ✅ 已调用 | wage | 查询工资项目列表 |
| 🟢 POST | `/projects` | ✅ 已调用 | wage | 创建工资项目 |
| 🟡 PUT | `/projects/{project_id}` | ❌ 未调用 | — | 更新工资项目 |
| 🔴 DELETE | `/projects/{project_id}` | ❌ 未调用 | — | 删除工资项目 |
| 🟢 POST | `/projects/init` | ✅ 已调用 | wage | 初始化工资项目 |
| 🔵 GET | `/employees` | ✅ 已调用 | wage | 查询员工列表 |
| 🔵 GET | `/employees/{employee_id}` | ❌ 未调用 | — | 查询员工详情 |
| 🟢 POST | `/employees` | ✅ 已调用 | wage | 创建员工 |
| 🟡 PUT | `/employees/{employee_id}` | ❌ 未调用 | — | 更新员工信息 |
| 🔵 GET | `/registers` | ✅ 已调用 | wage | 查询工资登记表 |
| 🔵 GET | `/registers/{register_id}` | ❌ 未调用 | — | 查询工资登记详情 |
| 🟢 POST | `/registers` | ✅ 已调用 | wage | 创建工资登记 |
| 🟢 POST | `/registers/{register_id}/calculate` | ❌ 未调用 | — | 计算工资 |
| 🔵 GET | `/registers/{register_id}/details` | ❌ 未调用 | — | 查询工资明细 |
| 🟡 PUT | `/registers/{register_id}/details/{detail_id}` | ❌ 未调用 | — | 更新工资明细 |
| 🟢 POST | `/registers/{register_id}/confirm` | ❌ 未调用 | — | 确认工资 |
| 🟢 POST | `/registers/{register_id}/generate-voucher` | ❌ 未调用 | — | 生成工资凭证 |
| 🔵 GET | `/payslip/{employee_id}` | ❌ 未调用 | — | 查询工资条 |

**小结**: ✅ 7 / ❌ 11 / ⚠️ 0