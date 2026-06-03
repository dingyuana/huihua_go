# 报表导出与审批 — 设计文档

> 状态: 待开发
> 日期: 2026-05-19
> 优先级: P1

---

## 1. 现状问题

### 1.1 资产负债表(BalanceSheet.vue)

| 功能 | 当前状态 | 后端API | 前端API |
|------|---------|---------|---------|
| 加载报表 | ✅ 调用 `getBalanceSheet` | `GET /reports/balance/calculate` | ✅ 已对接 |
| 导出 | ❌ `window.print()` 仅打印 | `GET /reports/balance/export` | ❌ 不存在 |
| 审批 | ❌ 无 | `POST /reports/{report_type}/approve` | ❌ 不存在 |
| 审批状态 | ❌ 无 | `GET /reports/{report_type}/approval-status` | ❌ 不存在 |

### 1.2 利润表(ProfitSheet.vue)

| 功能 | 当前状态 | 后端API | 前端API |
|------|---------|---------|---------|
| 加载报表 | ✅ 调用 `getProfitSheet` | `GET /reports/profit` | ✅ 已对接 |
| 导出 | ❌ `window.print()` 仅打印 | `GET /reports/profit/export` | ❌ 不存在 |
| 审批 | ❌ 无 | `POST /reports/{report_type}/approve` | ❌ 不存在 |
| 审批状态 | ❌ 无 | `GET /reports/{report_type}/approval-status` | ❌ 不存在 |

### 1.3 现金流量表(CashflowSheet.vue)

| 功能 | 当前状态 | 后端API | 前端API |
|------|---------|---------|---------|
| 加载报表 | ✅ 调用 `getCashflowSheet` | `GET /reports/cashflow` | ✅ 已对接 |
| 导出 | ❌ `window.print()` 仅打印 | 无专用导出端点 | ❌ 不存在 |

### 1.4 其他缺失报表

| 报表 | 后端 | 前端 |
|------|------|------|
| 资产负债表列表 `GET /reports/balance/list` | ✅ | ❌ |
| 利润表列表 `GET /reports/profit/list` | ✅ | ❌ |
| 现金流量表列表 `GET /reports/cashflow/list` | ✅ | ❌ |
| 费用明细 `GET /reports/expense_detail` | ✅ | ❌ |
| 部门费用 `GET /reports/expense_by_dept` | ✅ | ❌ |
| 收入明细 `GET /reports/revenue_detail` | ✅ | ❌ |
| 辅助余额表 `GET /reports/auxiliary-balance` | ✅ | ❌ |
| 科目余额表 `GET /reports/subject-balance` | ✅ | ❌ |

---

## 2. 设计方案

### 2.1 导出功能

**后端接口**: `GET /api/v1/reports/{report_type}/export`

| report_type | 含义 |
|------------|------|
| `balance` | 资产负债表 |
| `profit` | 利润表 |
| `cashflow` | 现金流量表 |

**前端实现**：
```javascript
// report.js 新增
export const exportReport = (reportType, params) => {
  return request.get(`/v1/reports/${reportType}/export`, {
    params,
    responseType: 'blob'  // 关键: 需要 blob 响应
  })
}
```

**UI 按钮改造**：
```vue
<el-button :icon="Download" @click="handleExport" :loading="exporting">
  {{ exporting ? '导出中...' : '导出Excel' }}
</el-button>
```

**handleExport 实现**：
```javascript
const handleExport = async () => {
  exporting.value = true
  try {
    const [y, m] = queryMonth.value.split('-')
    const blob = await exportReport('balance', { year: y, month: m })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `资产负债表_${queryMonth.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}
```

### 2.2 报表审批功能

**后端接口**：
- `POST /api/v1/reports/{report_type}/approve` — 审批
- `GET /api/v1/reports/{report_type}/approval-status` — 审批状态

**审批状态**：`draft → pending → approved`

**UI 改造 — 新增审批状态栏**：

```vue
<div class="report-header-bar">
  <el-tag v-if="approvalStatus === 'draft'" type="info">草稿</el-tag>
  <el-tag v-if="approvalStatus === 'pending'" type="warning">待审批</el-tag>
  <el-tag v-if="approvalStatus === 'approved'" type="success">已审批</el-tag>
  <el-button
    v-if="approvalStatus === 'draft'"
    type="primary"
    size="small"
    @click="handleApprove"
  >
    提交审批
  </el-button>
  <el-button
    v-if="approvalStatus === 'pending'"
    type="success"
    size="small"
    @click="handleApproveConfirm"
  >
    审批通过
  </el-button>
</div>
```

### 2.3 三张报表导出接口差异

| 报表 | 后端导出端点 | 返回格式 |
|------|------------|---------|
| 资产负债表 | `GET /reports/balance/export` | Excel |
| 利润表 | `GET /reports/profit/export` | Excel |
| 现金流量表 | 无专用端点 | 需复用 print() 或新增 |

**现金流量表导出现在缺失**，有两个选择：
1. 新增 `GET /reports/cashflow/export` 后端端点
2. 复用 `window.print()` 打印（不推荐，不够正式）

---

## 3. 实施步骤

### Step 1: report.js 新增导出函数

```javascript
export const exportBalanceReport = (params) => {
  return request.get('/v1/reports/balance/export', { params, responseType: 'blob' })
}

export const exportProfitReport = (params) => {
  return request.get('/v1/reports/profit/export', { params, responseType: 'blob' })
}

export const getReportApprovalStatus = (reportType, params) => {
  return request.get(`/v1/reports/${reportType}/approval-status`, { params })
}

export const approveReport = (reportType, params) => {
  return request.post(`/v1/reports/${reportType}/approve`, null, { params })
}
```

### Step 2: BalanceSheet.vue 改造

- 新增 `approvalStatus` ref
- `fetchData()` 后调用 `getReportApprovalStatus` 获取状态
- `handleExport` 使用 `exportBalanceReport`
- 新增审批状态栏 UI

### Step 3: ProfitSheet.vue 改造

同 BalanceSheet.vue 模式

### Step 4: CashflowSheet.vue 改造

- `handleExport` 暂时保持 `window.print()`（或等后端新增导出端点）
- 后续可补充审批功能

### Step 5: 后端补充（如果需要）

如果 `/reports/cashflow/export` 不存在，需要在 `report.py` 中新增。

---

## 4. 验收标准

- [ ] 资产负债表点击"导出Excel"能下载真实xlsx文件
- [ ] 利润表点击"导出Excel"能下载真实xlsx文件
- [ ] 报表页面正确显示当前审批状态
- [ ] 提交审批/审批通过按钮功能正常
- [ ] 现金流量表导出功能补充（如果后端支持）
