<template>
  <div class="review-page">
    <div class="page-header">
      <h3>凭证审核工作台</h3>
      <el-tag type="warning" size="large">待审核: {{ pendingCount }} 张</el-tag>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="日期">
          <el-date-picker v-model="filterDateRange" type="daterange" range-separator="~" style="width: 220px" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item><el-button type="primary" @click="fetchPending">查询</el-button></el-form-item>
      </el-form>
    </el-card>

    <!-- 统计汇总 -->
    <el-row :gutter="8" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card amount">
          <div class="stat-summary">
            <span class="stat-label">待审金额</span>
            <span class="stat-amount">{{ formatAmount(stats.totalAmount) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card debit">
          <div class="stat-summary">
            <span class="stat-label">借方合计</span>
            <span class="stat-amount">{{ formatAmount(stats.debitTotal) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card credit">
          <div class="stat-summary">
            <span class="stat-label">贷方合计</span>
            <span class="stat-amount">{{ formatAmount(stats.creditTotal) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card total-count">
          <div class="stat-summary">
            <span class="stat-label">待审数量</span>
            <span class="stat-amount">{{ stats.count }} 张</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <!-- 批量操作 -->
      <div class="batch-bar">
        <el-checkbox v-model="selectAll" @change="toggleAll">全选</el-checkbox>
        <span class="selected-count">已选 {{ selectedIds.length }} 张</span>
        <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="batchApprove">审核通过</el-button>
        <el-button size="small" :disabled="selectedIds.length === 0" @click="showRejectDialog = true">驳回</el-button>
      </div>

      <el-table :data="pendingVouchers" border stripe size="small" @selection-change="onSelection" @row-click="openDetail">
        <el-table-column type="selection" width="40" @click.stop />
        <el-table-column prop="voucher_no" label="凭证号" width="150" />
        <el-table-column prop="date" label="日期" width="110">
          <template #default="{ row }">{{ (row.date || '').slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column label="对方名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.counterparty_name || '—' }}</template>
        </el-table-column>
        <el-table-column label="科目" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.first_account_code">
              <el-tag size="small" type="info">{{ row.first_account_code }}</el-tag>
              {{ row.first_account_name }}
            </span>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="来源单据" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.source_doc_no" size="small" type="success">{{ row.source_doc_no }}</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="摘要" min-width="200" show-overflow-tooltip />
        <el-table-column label="借方合计" width="120" align="right"><template #default="{ row }">{{ row.debit_total || '0.00' }}</template></el-table-column>
        <el-table-column label="贷方合计" width="120" align="right"><template #default="{ row }">{{ row.credit_total || '0.00' }}</template></el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }"><DocStatusTag :docstatus="row.docstatus" /></template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 详情抽屉：分屏审核 -->
    <el-drawer v-model="showDetail" :title="`审核凭证: ${currentVoucher?.voucher_no}`" size="600px">
      <template v-if="currentVoucher">
        <el-tabs v-model="detailTab">
          <el-tab-pane label="凭证信息" name="voucher">
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="凭证号">{{ currentVoucher.voucher_no }}</el-descriptions-item>
              <el-descriptions-item label="日期">{{ currentVoucher.date }}</el-descriptions-item>
              <el-descriptions-item label="摘要" :span="2">{{ currentVoucher.remark }}</el-descriptions-item>
              <el-descriptions-item label="制单人">{{ currentVoucher.creator }}</el-descriptions-item>
              <el-descriptions-item label="金额">{{ currentVoucher.amount }}</el-descriptions-item>
            </el-descriptions>

            <h4 class="section-title">分录明细</h4>
            <el-table :data="currentVoucher.lines || []" size="small" border>
              <el-table-column prop="account" label="科目" min-width="160" />
              <el-table-column prop="debit" label="借方" width="120" align="right" />
              <el-table-column prop="credit" label="贷方" width="120" align="right" />
            </el-table>

            <!-- AI 风险详情 -->
            <h4 class="section-title">AI 风控分析</h4>
            <el-card v-if="currentVoucher.risk.items.length" shadow="never" class="risk-detail-card">
              <div v-for="(item, i) in currentVoucher.risk.items" :key="i" class="risk-item">
                <el-tag :type="item.severity === 'error' ? 'danger' : 'warning'" size="small">
                  {{ item.severity === 'error' ? '⚠️ 风险' : '💡 提示' }}
                </el-tag>
                <span class="risk-msg">{{ item.message }}</span>
                <p v-if="item.suggestion" class="risk-suggestion">{{ item.suggestion }}</p>
              </div>
            </el-card>
            <el-empty v-else description="AI 风控未发现异常" :image-size="60" />
          </el-tab-pane>

          <el-tab-pane label="原始单据" name="source">
            <el-empty description="关联原始单据（待对接）" :image-size="60" />
            <p class="source-hint">审核时可联查银行流水/发票等原始单据</p>
          </el-tab-pane>
        </el-tabs>
      </template>

      <template #footer>
        <div class="drawer-actions">
          <el-button @click="showDetail = false">关闭</el-button>
          <el-button type="danger" @click="openRejectFromDrawer">驳回</el-button>
          <el-button type="primary" @click="approveCurrent">审核通过</el-button>
        </div>
      </template>
    </el-drawer>

    <!-- 驳回弹窗 -->
    <el-dialog v-model="showRejectDialog" title="驳回凭证" width="420px">
      <el-form>
        <el-form-item label="驳回原因" required>
          <el-input v-model="rejectReason" type="textarea" :rows="3" placeholder="请填写驳回原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectDialog = false">取消</el-button>
        <el-button type="primary" @click="reject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import DocStatusTag from '@/components/business/DocStatusTag.vue'

interface RiskItem {
  severity: 'error' | 'warning'
  message: string
  suggestion?: string
}

interface VoucherItem {
  id: string
  voucher_no: string
  date: string
  remark: string
  amount: string
  creator: string
  risk: { level: string; items: RiskItem[] }
  lines?: { account: string; debit: string; credit: string }[]
  debit_total?: number | string
  credit_total?: number | string
  docstatus?: number
  counterparty_name?: string
  first_account_code?: string
  first_account_name?: string
  source_doc_no?: string
}

const filterDateRange = ref<any>(null)
const pendingCount = ref(0)
const pendingVouchers = ref<VoucherItem[]>([])

const stats = computed(() => {
  const list = pendingVouchers.value
  let debitTotal = 0
  let creditTotal = 0
  for (const v of list) {
    debitTotal += Number(v.debit_total || 0)
    creditTotal += Number(v.credit_total || 0)
  }
  return {
    totalAmount: debitTotal,
    debitTotal,
    creditTotal,
    count: list.length,
  }
})

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function fetchPending() {
  try {
    const params: any = {}
    if (filterDateRange.value && filterDateRange.value.length === 2) {
      params.start_date = filterDateRange.value[0]
      params.end_date = filterDateRange.value[1]
    }
    const res: any = await request.get('/vouchers/pending-review', { params })
    const list = res?.list || res?.data?.list || res?.data
    if (Array.isArray(list)) { pendingVouchers.value = list; pendingCount.value = list.length }
  } catch { /* silent fail - show empty */ }
}

onMounted(() => fetchPending())

const selectAll = ref(false)
const selectedIds = ref<string[]>([])
const showRejectDialog = ref(false)
const rejectReason = ref('')

// 详情抽屉
const showDetail = ref(false)
const detailTab = ref('voucher')
const currentVoucher = ref<VoucherItem | null>(null)

function onSelection(rows: any[]) {
  selectedIds.value = rows.map((r: any) => r.id)
}

function toggleAll() { /* handled by el-table */ }

async function openDetail(row: VoucherItem) {
  currentVoucher.value = { ...row }
  detailTab.value = 'voucher'
  showDetail.value = true
  // Fetch full voucher detail including journal entry lines
  try {
    const res: any = await request.get(`/vouchers/${row.id}`)
    const data = res?.data || res
    if (data?.journal_entry_lines) {
      currentVoucher.value = {
        ...currentVoucher.value,
        lines: data.journal_entry_lines.map((l: any) => ({
          account: [l.account_code, l.account_name].filter(Boolean).join(' '),
          debit: l.debit || '0.00',
          credit: l.credit || '0.00',
        })),
      }
    }
    // Update basic fields from journal_entry if available
    if (data?.journal_entry) {
      const je = data.journal_entry
      currentVoucher.value = {
        ...currentVoucher.value,
        date: je.posting_date || currentVoucher.value.date,
        remark: je.remark || currentVoucher.value.remark,
        amount: je.debit_total || je.credit_total || currentVoucher.value.amount,
      }
    }
  } catch { /* silently use list data */ }
}

function approveCurrent() {
  if (!currentVoucher.value) return
  ElMessage.success(`凭证 ${currentVoucher.value.voucher_no} 已审核通过`)
  pendingVouchers.value = pendingVouchers.value.filter(v => v.id !== currentVoucher.value!.id)
  pendingCount.value = pendingVouchers.value.length
  showDetail.value = false
}

function openRejectFromDrawer() {
  showDetail.value = false
  showRejectDialog.value = true
  selectedIds.value = currentVoucher.value ? [currentVoucher.value.id] : []
}

async function batchApprove() {
  try {
    for (const id of selectedIds.value) {
      await request.post(`/approvals/${id}/approve`, { comment: '审核通过' })
    }
    ElMessage.success(`已审核通过 ${selectedIds.value.length} 张凭证`)
    pendingVouchers.value = pendingVouchers.value.filter(v => !selectedIds.value.includes(v.id))
    pendingCount.value = pendingVouchers.value.length
    selectedIds.value = []
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '审核失败')
  }
}

async function reject() {
  if (!rejectReason.value) { ElMessage.warning('请填写驳回原因'); return }
  try {
    for (const id of selectedIds.value) {
      await request.post(`/approvals/${id}/reject`, { reason: rejectReason.value })
    }
    ElMessage.success(`已驳回 ${selectedIds.value.length} 张凭证`)
    showRejectDialog.value = false
    rejectReason.value = ''
    pendingVouchers.value = pendingVouchers.value.filter(v => !selectedIds.value.includes(v.id))
    pendingCount.value = pendingVouchers.value.length
    selectedIds.value = []
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '驳回失败')
  }
}
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.filter-card { margin-bottom: 16px; }
.stat-row { margin-bottom: 12px; }
.stat-card {
  text-align: center;
  .stat-summary {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    flex-wrap: wrap;
    line-height: 1;
  }
  .stat-label { font-size: 14px; font-weight: 500; color: #666; }
  .stat-amount { font-size: 20px; font-weight: 700; }
  &.amount .stat-amount { color: #096dd9; }
  &.debit .stat-amount { color: #389e0d; }
  &.credit .stat-amount { color: #cf1322; }
  &.total-count .stat-amount { color: #333; }
}
.batch-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.selected-count { font-size: 13px; color: #666; }
.section-title { font-size: 14px; font-weight: 600; margin: 16px 0 8px; }
.risk-detail-card { background: #fffbe6; border: 1px solid #ffe58f; }
.risk-item { margin-bottom: 12px; padding: 8px; background: #fff; border-radius: 4px; }
.risk-msg { margin-left: 4px; font-size: 13px; }
.risk-suggestion { color: #999; font-size: 12px; margin-top: 4px; padding-left: 4px; }
.source-hint { color: #999; font-size: 13px; text-align: center; margin-top: 8px; }
.drawer-actions { display: flex; justify-content: flex-end; gap: 12px; }
</style>
