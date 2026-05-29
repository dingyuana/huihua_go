# SPEC: TASK05 — 前端：收款单 & 付款单页面

## 基本信息

- **任务 ID**: phase1-bill-005
- **类型**: feature
- **优先级**: high
- **依赖**: TASK01 + TASK02（后端接口可用）
- **执行者**: OpenCode

## 背景

费用报销单之后，会计第二类和第三类高频操作是收款单和付款单。

## 目标

创建收款单列表+表单页、付款单列表+表单页。收付款单结构相似，可复用部分代码。

## 技术约束

- 新建4个文件：`ReceiptList.vue`、`ReceiptForm.vue`、`PaymentList.vue`、`PaymentForm.vue`
- 放入 `src/pages/business/` 目录
- 路由注册到 `src/router/index.js`
- 在 MainLayout.vue 导航中添加菜单项
- API 写在 `src/api/business.js`（TASK04 已创建）

## 详细设计

### 页面结构（收款单和付款单类似）

**列表页**：
```
┌────────────────────────────────────────────────────────┐
│  收款单                                 [+ 新增收款]  │
├────────────────────────────────────────────────────────┤
│  [日期筛选] [客户筛选] [状态筛选]        [搜索]        │
├────────────────────────────────────────────────────────┤
│  编号     | 日期   | 客户名称  | 金额    | 状态  | 操作│
│  SK...    | 5/23  | XX公司   | ¥50,000 | 已生单| [查看]│
│  SK...    | 5/22  | YY公司   | ¥12,000 | 待审核| [生成]│
│  ...                                                   │
└────────────────────────────────────────────────────────┘
```

**表单页**：
```
┌────────────────────────────────────────────┐
│  新增收款单                                  │
├────────────────────────────────────────────┤
│  收款日期: [2026-05-23  ▼]                  │
│  客户名称: [选择客户 ▼]                     │
│  收款金额: [________] 元                    │
│  收款方式: [银行转账 ▼]                     │
│  收款账户: [工商银行(尾号8888) ▼]              │
│  摘要: [________________________________]   │
├────────────────────────────────────────────┤
│  [取消]  [保存草稿]  [保存并提交审核]        │
└────────────────────────────────────────────┘
```

付款单表单结构和收款单基本相同，将"客户名称"替换为"供应商名称"，客户选择器改为供应商选择器。

### API 函数（追加到 `src/api/business.js`）

```js
// 收款单
export const getReceipts = (params) => request.get('/v1/business/receipts', { params })
export const getReceiptById = (id) => request.get(`/v1/business/receipts/${id}`)
export const createReceipt = (data) => request.post('/v1/business/receipts', data)
export const updateReceipt = (id, data) => request.put(`/v1/business/receipts/${id}`, data)
export const deleteReceipt = (id) => request.delete(`/v1/business/receipts/${id}`)
export const submitReceipt = (id) => request.post(`/v1/business/receipts/${id}/submit`)
export const approveReceipt = (id) => request.post(`/v1/business/receipts/${id}/approve`)
export const rejectReceipt = (id, reason) => request.post(`/v1/business/receipts/${id}/reject`, { reason })
export const generateVoucherForReceipt = (id) => request.post(`/v1/business/receipts/${id}/generate-voucher`)

// 付款单（同上，路径用 payments）
export const getPayments = (params) => request.get('/v1/business/payments', { params })
// ... 同理
export const generateVoucherForPayment = (id) => request.post(`/v1/business/payments/${id}/generate-voucher`)
```

### 路由注册

```js
{ path: 'business/receipts', name: 'ReceiptList', ... },
{ path: 'business/receipts/create', name: 'ReceiptCreate', ... },
{ path: 'business/payments', name: 'PaymentList', ... },
{ path: 'business/payments/create', name: 'PaymentCreate', ... },
```

## 验收标准

- [ ] 收款单列表可查看、筛选客户、按状态过滤
- [ ] 新增收款单选择客户后自动带入客户名
- [ ] 付款单列表可查看、筛选供应商
- [ ] 生成凭证成功后，应收账款/应付账款余额更新
- [ ] 所有状态流转正常（草稿→待审核→已审核→已生成凭证）

## OpenCode 指令

**目标**：在 `/root/huihua-financial-master/frontend` 中新建收款单+付款单的列表页和表单页。

**约束**：
- 新建4个文件在 `src/pages/business/`
- 注册路由 + 导航菜单
- API 追加到 `src/api/business.js`
- 客户选择器复用现有的 CustomerList 中的客户数据接口
- 参考 TASK04 费用报销单的页面风格

**上下文**：
- repo: `/root/huihua-financial-master`
- 客户API：`src/api/auxiliary.js` 中有 `getCustomerList`
- 供应商API：`src/api/auxiliary.js` 中有 `getSupplierList`
- 银行账户API：`src/api/cash.js` 中有 `getBankAccounts`

**验收**：
- 收付款单列表页可查看数据、筛选、分页
- 表单页可选择客户/供应商和银行账户
- 生成凭证功能正常
