# 凭证记账与审核工作流 — 设计文档

> 状态: 待开发
> 日期: 2026-05-19
> 优先级: P0

---

## 1. 现状问题

### 1.1 凭证工作流断裂

当前凭证生命周期：`draft → pending → approved → posted`

| 状态 | 前端操作 | 后端API | 前端API函数 |
|------|---------|---------|------------|
| draft | 提交 | `POST /vouchers/{id}/submit` | ✅ `submitVoucher` |
| pending | 审核 | `POST /vouchers/{id}/approve` | ✅ `approveVoucher` |
| pending | **驳回** | `POST /vouchers/{id}/reject` | ⚠️ 有函数无按钮 |
| approved | **记账** | `POST /vouchers/{id}/post` | ❌ 函数不存在 |
| approved | **反记账** | `POST /vouchers/{id}/unpost` | ✅ 直接用`request.post` |
| any | 编辑/详情 | `GET/PUT /vouchers/{id}` | ❌ 函数不存在 |

**具体缺口**：

1. **`approveVoucher` 调用错误** — 传入空`{}`，但后端期望 `{ comment?: string }`
2. **记账(post) API** — `POST /v1/accounting/vouchers/{id}/post` 后端已实现，但：
   - `accounting.js` 中无 `postVoucher` 函数
   - `VoucherList.vue` 的 `handlePost` 用 `request.post` 直接写死路径
   - `VoucherDetail.vue` 的 `handlePost` 没有实现
3. **驳回(reject) API** — `accounting.js` 有 `rejectVoucher`，但两个页面都没有"驳回"按钮
4. **凭证详情/编辑** — `getVoucherById`、`updateVoucher` 在 `accounting.js` 已定义，但页面没调用

### 1.2 凭证详情页与编辑页未对接

- `VoucherDetail.vue` — 纯静态演示数据，`handleSubmit/handleApprove/handlePost/handleDelete` 均未调用 API
- `VoucherForm.vue` — 同上
- `VoucherList.vue` — `fetchData()` 调用了 `getVoucherList`，但 catch 后静默使用静态演示数据

---

## 2. 凭证状态机

```
  draft ──submit──→ pending ──approve──→ approved ──post──→ posted
    ↑                   │                                    │
    └──reject───────────┘               unpost←──────────────┘
    ↑
    └──delete
```

**各状态允许的操作**：

| 当前状态 | 可执行动作 | 目标状态 |
|---------|-----------|---------|
| draft | 提交/编辑/删除 | pending / - / - |
| pending | 审核通过/驳回 | approved / draft |
| approved | 记账/反审核 | posted / pending |
| posted | 反记账 | approved |

---

## 3. API 缺口清单

### 3.1 后端已有、前端缺失的函数

| 后端路径 | 说明 | `accounting.js` 现状 |
|---------|------|---------------------|
| `POST /v1/accounting/vouchers/{id}/post` | 记账 | ❌ 不存在 |
| `GET /v1/accounting/vouchers/{id}` | 凭证详情 | ✅ `getVoucherById` |
| `PUT /v1/accounting/vouchers/{id}` | 编辑凭证 | ✅ `updateVoucher` |
| `DELETE /v1/accounting/vouchers/{id}` | 删除凭证 | ✅ `deleteVoucher` |
| `GET /v1/accounting/vouchers/summary/monthly` | 月度汇总 | ❌ 不存在 |

### 3.2 需要新增的前端 API 函数 (accounting.js)

```javascript
// 新增1: 记账
export const postVoucher = (id, data = {}) => {
  return request.post(`/v1/accounting/vouchers/${id}/post`, data)
}

// 新增2: 凭证详情(已有getVoucherById,检查是否已导入)
// 新增3: 月度汇总
export const getVoucherMonthlySummary = (params) => {
  return request.get('/v1/accounting/vouchers/summary/monthly', { params })
}
```

---

## 4. 页面改造设计

### 4.1 VoucherList.vue 改造

**按钮逻辑调整**：

```vue
<!-- 当前(有问题): -->
<el-button v-if="row.status === 'approved'" @click="handlePost">记账</el-button>
<!-- 驳回按钮缺失 -->

<!-- 改造后: -->
<!-- draft -->
<el-button @click="handleEdit">编辑</el-button>
<el-button type="success" @click="handleSubmit">提交</el-button>
<el-button type="danger" link @click="handleDelete">删除</el-button>

<!-- pending -->
<el-button type="warning" @click="handleApprove">审核通过</el-button>
<el-button type="danger" link @click="handleReject">驳回</el-button>

<!-- approved -->
<el-button type="primary" @click="handlePost">记账</el-button>

<!-- posted -->
<el-button type="info" link @click="handleUnpost">反记账</el-button>
```

**handlePost 实现**：
```javascript
const handlePost = async (row) => {
  try {
    await ElMessageBox.confirm(`确定记账「${row.voucherWord}第${row.voucherNo}号」吗？`, '记账确认')
    await postVoucher(row.id)
    ElMessage.success('记账成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e?.detail || '记账失败')
  }
}
```

**handleReject 实现**（新增）：
```javascript
const handleReject = async (row) => {
  try {
    await ElMessageBox.confirm(`确定驳回该凭证吗？`, '驳回确认')
    await rejectVoucher(row.id, {})
    ElMessage.success('已驳回')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e?.detail || '驳回失败')
  }
}
```

### 4.2 VoucherDetail.vue 改造

| 操作 | 当前 | 改造后 |
|------|------|--------|
| 打开页面 | 静态演示数据 | `getVoucherById(route.params.id)` |
| 提交 | `request.post` 演示 | `submitVoucher(id)` |
| 审核 | `request.post` 演示 | `approveVoucher(id, {})` |
| 记账 | 无实现 | `postVoucher(id)` |
| 驳回 | 无按钮 | `rejectVoucher(id, {})` |
| 反记账 | 无实现 | `request.post('/v1/accounting/vouchers/${id}/unpost')` |
| 删除 | `request.delete` 演示 | `deleteVoucher(id)` |

### 4.3 VoucherForm.vue 改造

| 操作 | 当前 | 改造后 |
|------|------|--------|
| 新增保存 | 静态演示 | `createVoucher(formData)` |
| 编辑保存 | 静态演示 | `updateVoucher(id, formData)` |

---

## 5. 实施步骤

### Step 1: 补充 accounting.js (新增1个函数)

```javascript
// account.js 新增 postVoucher
export const postVoucher = (id, data = {}) => {
  return request.post(`/v1/accounting/vouchers/${id}/post`, data)
}
```

### Step 2: 改造 VoucherList.vue

- 添加 `rejectVoucher` 导入（已有）
- 添加 `postVoucher` 导入（新增）
- 添加 `handleReject` 方法
- 修正 `handleApprove` 参数
- 修正 `handlePost` 使用 `postVoucher`
- 添加驳回按钮（pending 状态）

### Step 3: 改造 VoucherDetail.vue

- `onMounted` 调用 `getVoucherById`
- 各操作按钮绑定真实 API
- `handlePost` 调用 `postVoucher`

### Step 4: 改造 VoucherForm.vue

- `onMounted` 编辑模式调用 `getVoucherById`
- 保存时判断新建/编辑，调用对应 API

### Step 5: 真实数据联调

- 移除各 catch 块中的静态演示数据 fallback
- `fetchData()` 失败时显示错误而非静默

---

## 6. 验收标准

- [ ] 凭证列表展示真实数据（非静态演示）
- [ ] 提交/审核/驳回/记账/反记账 五步工作流完整走通
- [ ] 凭证详情页加载真实数据
- [ ] 凭证编辑页新建/编辑保存正常
- [ ] 错误提示正确显示
