# SPEC: TASK04 — 前端：费用报销单页面

## 基本信息

- **任务 ID**: phase1-bill-004
- **类型**: feature
- **优先级**: high
- **依赖**: TASK01 + TASK02（后端接口可用）
- **执行者**: OpenCode

## 背景

会计日常第一个高频操作是"填报销单"。需要新增费用报销单页面，支持填写、提交审核、生成凭证。

## 目标

创建费用报销单列表页 + 新增/编辑表单页，包含文件上传和一键生成凭证功能。

## 技术约束

- 新建文件：`frontend/src/pages/business/ReimbursementList.vue` + `ReimbursementForm.vue`
- 在 `frontend/src/router/index.js` 中注册路由
- 新建 `frontend/src/api/business.js` 存放 API 函数
- 使用 Element Plus 现有组件
- 新建菜单入口（在 MainLayout.vue 的导航中追加）
- 复用现有的文件上传组件

## 详细设计

### 页面1：费用报销单列表 (`ReimbursementList.vue`)

```
┌──────────────────────────────────────────────────────────┐
│  费用报销单                              [+ 新增报销单]  │
├──────────────────────────────────────────────────────────┤
│  [日期筛选] [费用类型筛选] [状态筛选]     [搜索]          │
├──────────────────────────────────────────────────────────┤
│  编号    | 日期   | 费用类型 | 金额    | 状态   | 操作   │
│  FY...   | 5/23  | 差旅费  | ¥1,200 | 已生单 | [查看]   │
│  FY...   | 5/22  | 办公费  | ¥500   | 待审核 | [生成凭证]│
│  ...                                                     │
├──────────────────────────────────────────────────────────┤
│  第1页/共3页  [<] [1] [2] [3] [>]                      │
└──────────────────────────────────────────────────────────┘
```

**操作按钮说明**：
- **草稿状态**：显示 [编辑] [删除] [提交审核]
- **待审核状态**：显示 [审核通过] [审核驳回]（管理员）
- **已审核状态**：显示 [生成凭证]
- **已生成凭证状态**：显示 [查看凭证]（跳转到凭证详情页）

### 页面2：费用报销单表单 (`ReimbursementForm.vue`)

```
┌────────────────────────────────────────────┐
│  新增费用报销单                              │
├────────────────────────────────────────────┤
│  报销日期: [2026-05-23  ▼]                  │
│  费用类型: [差旅费 ▼]                       │
│  报销金额: [________] 元                    │
│  报销事由: [______________________________] │
│  所属部门: [财务部 ▼]                      │
│  收款方式: [银行转账 ▼]                     │
│  银行账户: [工商银行(尾号8888) ▼]            │
│  附件上传: [选择文件] 已上传2张              │
├────────────────────────────────────────────┤
│  [取消]  [保存草稿]  [保存并提交审核]        │
└────────────────────────────────────────────┘
```

### API 函数（`frontend/src/api/business.js`）

```js
// 费用报销单
export const getReimbursements = (params) => request.get('/v1/business/reimbursements', { params })
export const getReimbursementById = (id) => request.get(`/v1/business/reimbursements/${id}`)
export const createReimbursement = (data) => request.post('/v1/business/reimbursements', data)
export const updateReimbursement = (id, data) => request.put(`/v1/business/reimbursements/${id}`, data)
export const deleteReimbursement = (id) => request.delete(`/v1/business/reimbursements/${id}`)
export const submitReimbursement = (id) => request.post(`/v1/business/reimbursements/${id}/submit`)
export const approveReimbursement = (id) => request.post(`/v1/business/reimbursements/${id}/approve`)
export const rejectReimbursement = (id, reason) => request.post(`/v1/business/reimbursements/${id}/reject`, { reason })
export const generateVoucher = (id) => request.post(`/v1/business/reimbursements/${id}/generate-voucher`)
```

### 路由注册

```js
{
  path: 'business/reimbursements',
  name: 'ReimbursementList',
  component: () => import('../pages/business/ReimbursementList.vue'),
  meta: { title: '费用报销单', roles: ['accountant', 'finance_manager'] }
},
{
  path: 'business/reimbursements/create',
  name: 'ReimbursementCreate',
  component: () => import('../pages/business/ReimbursementForm.vue'),
  meta: { title: '新增报销单', roles: ['accountant'] }
}
```

### 菜单入口

在 MainLayout 的导航中，将"费用报销单"加到"凭证管理"附近。

## 验收标准

- [ ] 费用报销单列表可查看、筛选、分页
- [ ] 新增报销单表单可填写并保存草稿
- [ ] 报销单可提交审核、审核通过、驳回
- [ ] 审核通过后点击"生成凭证"成功，跳转到凭证详情
- [ ] 附件上传正常
- [ ] 浏览器控制台无报错

## OpenCode 指令

**目标**：在 `/root/huihua-financial-master/frontend` 中新建费用报销单列表+表单页面。

**约束**：
- 新建文件：`src/pages/business/ReimbursementList.vue` + `ReimbursementForm.vue`
- 新建 `src/api/business.js`
- 修改 `src/router/index.js` 加入路由
- 在 MainLayout.vue 的导航中增加菜单项
- 参考现有 `src/pages/accounting/VoucherList.vue` 的列表风格

**上下文**：
- repo: `/root/huihua-financial-master`
- 导航布局文件：`frontend/src/layouts/MainLayout.vue`
- 现有VoucherList样式参考：列表+筛选+分页的写法

**验收**：
- 列表页可加载数据、筛选、分页
- 表单页可新建和编辑
- 生成凭证功能可用
