<template>
  <div class="review-page">
    <div class="page-header">
      <h3>银行流水审核工作台</h3>
      <div class="header-actions">
        <el-button
          type="primary"
          size="small"
          @click="$router.push('/bank/import')"
        >
          导入流水
        </el-button>
        <el-button
          size="small"
          @click="loadData"
        >
          刷新
        </el-button>
      </div>
    </div>

    <!-- 顶部统计卡片 -->
    <el-row
      :gutter="16"
      class="stat-row"
    >
      <el-col :span="6">
        <el-card
          shadow="hover"
          class="stat-card"
        >
          <p class="stat-num">
            {{ stats.monthly_txns }}
          </p>
          <p class="stat-label">
            本月流水
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card
          shadow="hover"
          class="stat-card pending"
        >
          <p class="stat-num">
            {{ stats.pending_count }}
          </p>
          <p class="stat-label">
            待审核
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card
          shadow="hover"
          class="stat-card ai-ok"
        >
          <p class="stat-num">
            {{ stats.ai_processed_count }}
          </p>
          <p class="stat-label">
            AI已处理
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card
          shadow="hover"
          class="stat-card manual"
        >
          <p class="stat-num danger">
            {{ stats.manual_pending_count }}
          </p>
          <p class="stat-label">
            待人工处理
          </p>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select
          v-model="filterStatus"
          placeholder="状态筛选"
          size="small"
          clearable
          style="width: 140px"
          @change="loadData"
        >
          <el-option
            label="全部"
            value=""
          />
          <el-option
            label="待审核"
            value="classified"
          />
          <el-option
            label="AI已处理"
            value="ai_processed"
          />
          <el-option
            label="待人工处理"
            value="manual_pending"
          />
          <el-option
            label="已完成"
            value="approved"
          />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="small"
          value-format="YYYY-MM-DD"
          style="width: 240px"
          @change="loadData"
        />
      </div>

      <!-- 批量操作栏 -->
      <div class="batch-bar">
        <el-checkbox
          v-model="selectAll"
          @change="onSelectAll"
        >
          全选
        </el-checkbox>
        <span class="selected-count">已选 {{ selectedIds.length }} 笔</span>
        <el-button
          size="small"
          type="primary"
          :disabled="selectedIds.length === 0"
          @click="batchSubmit"
        >
          批量提交审核
        </el-button>
        <el-button
          size="small"
          type="danger"
          :disabled="selectedIds.length === 0"
          @click="batchReject"
        >
          批量驳回
        </el-button>
      </div>

      <!-- 流水列表 -->
      <el-table
        ref="tableRef"
        :data="txnList"
        border
        stripe
        size="small"
        row-class-name="clickable-row"
        @selection-change="onSelectionChange"
        @row-click="openPreview"
      >
        <el-table-column
          type="selection"
          width="40"
          @click.stop
        />
        <el-table-column
          prop="txn_date"
          label="日期"
          width="90"
        />
        <el-table-column
          prop="description"
          label="摘要"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          prop="counterparty_name"
          label="对方"
          width="140"
          show-overflow-tooltip
        />
        <el-table-column
          label="金额"
          width="130"
          align="right"
        >
          <template #default="{ row }">
            <span :class="amountClass(row)">{{ amountDisplay(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="分类"
          width="110"
        >
          <template #default="{ row }">
            <el-tag
              :type="classificationTag(row.classification)"
              size="small"
            >
              {{ classificationLabel(row.classification) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="置信度"
          width="80"
        >
          <template #default="{ row }">
            <span v-if="row.ai_confidence">{{ row.ai_confidence }}%</span>
            <span
              v-else
              class="text-muted"
            >—</span>
          </template>
        </el-table-column>
        <el-table-column
          label="草稿状态"
          width="100"
        >
          <template #default="{ row }">
            <el-tag
              v-if="row.draft_voucher_id"
              type="info"
              size="small"
            >
              凭证草稿
            </el-tag>
            <el-tag
              v-else-if="row.draft_payment_id"
              type="info"
              size="small"
            >
              单据草稿
            </el-tag>
            <span
              v-else
              class="text-muted"
            >无</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="160"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              @click.stop="openPreview(row)"
            >
              预览
            </el-button>
            <el-button
              link
              type="success"
              size="small"
              @click.stop="submitOne(row)"
            >
              提交
            </el-button>
            <el-button
              link
              type="danger"
              size="small"
              @click.stop="rejectOne(row)"
            >
              驳回
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 12px; justify-content: flex-end"
        @current-change="loadData"
      />
    </el-card>

    <!-- 底部选中合计 -->
    <div
      v-if="selectedIds.length > 0"
      class="bottom-summary"
    >
      已选 <b>{{ selectedIds.length }}</b> 笔，合计 <b class="amount-income">{{ selectedTotal }}</b>
    </div>

    <!-- 草稿预览弹窗 -->
    <DraftPreviewDialog
      v-model="showPreview"
      :txn-id="currentTxnId"
      @submitted="onSubmitted"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getReviewList, getReviewStats, submitReview, rejectManual } from '@/api/modules/bank_txn_review'
import DraftPreviewDialog from './DraftPreviewDialog.vue'

interface TxnRow {
  id: string
  txn_date: string
  description: string
  counterparty_name: string
  debit: string
  credit: string
  direction: string
  classification: string
  ai_confidence?: number
  ai_business_scene?: string
  ai_suggested_action?: string
  status: string
  draft_voucher_id?: string
  draft_payment_id?: string
}

interface Stats {
  monthly_txns: number
  pending_count: number
  ai_processed_count: number
  manual_pending_count: number
}

const txnList = ref<TxnRow[]>([])
const stats = ref<Stats>({ monthly_txns: 0, pending_count: 0, ai_processed_count: 0, manual_pending_count: 0 })
const selectedIds = ref<string[]>([])
const selectAll = ref(false)
const tableRef = ref()
const filterStatus = ref('')
const dateRange = ref<string[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const showPreview = ref(false)
const currentTxnId = ref('')

onMounted(() => {
  loadStats()
  loadData()
})

async function loadStats() {
  try {
    const res: any = await getReviewStats()
    if (res?.data) {
      stats.value = res.data
    }
  } catch { /* silent */ }
}

async function loadData() {
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    if (dateRange.value?.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res: any = await getReviewList(params)
    const list = res?.data?.list || res?.data || []
    if (Array.isArray(list)) {
      txnList.value = list
      total.value = res?.data?.total || list.length
    }
  } catch (e: any) {
    ElMessage.error('加载流水列表失败: ' + (e?.message || ''))
  }
}

function onSelectionChange(rows: TxnRow[]) {
  selectedIds.value = rows.map(r => r.id)
  selectAll.value = rows.length === txnList.value.length && txnList.value.length > 0
}

function onSelectAll() {
  // handled internally by el-table
}

function openPreview(row: TxnRow) {
  currentTxnId.value = row.id
  showPreview.value = true
}

async function submitOne(row: TxnRow) {
  try {
    const res: any = await submitReview({ txn_ids: [row.id] })
    if (res?.data?.approved_count > 0) {
      ElMessage.success('提交审核成功')
      loadData()
      loadStats()
    }
  } catch (e: any) {
    ElMessage.error('提交失败: ' + (e?.message || ''))
  }
}

async function rejectOne(row: TxnRow) {
  try {
    await ElMessageBox.confirm('确认驳回该流水至待人工处理？', '驳回确认', { type: 'warning' })
    const res: any = await rejectManual({ txn_ids: [row.id] })
    if (res?.data?.rejected_count > 0) {
      ElMessage.success('已驳回至待人工处理')
      loadData()
      loadStats()
    }
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('驳回失败: ' + (e?.message || ''))
  }
}

async function batchSubmit() {
  if (selectedIds.value.length === 0) return
  try {
    const res: any = await submitReview({ txn_ids: selectedIds.value })
    if (res?.data?.approved_count > 0) {
      ElMessage.success(`成功提交 ${res.data.approved_count} 笔`)
      selectedIds.value = []
      loadData()
      loadStats()
    }
  } catch (e: any) {
    ElMessage.error('批量提交失败: ' + (e?.message || ''))
  }
}

async function batchReject() {
  if (selectedIds.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确认驳回选中的 ${selectedIds.value.length} 笔流水至待人工处理？`, '批量驳回确认', { type: 'warning' })
    const res: any = await rejectManual({ txn_ids: selectedIds.value })
    if (res?.data?.rejected_count > 0) {
      ElMessage.success(`已驳回 ${res.data.rejected_count} 笔`)
      selectedIds.value = []
      loadData()
      loadStats()
    }
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('批量驳回失败: ' + (e?.message || ''))
  }
}

function onSubmitted() {
  loadData()
  loadStats()
}

function amountClass(row: TxnRow) {
  if (row.classification === 'business_receipt') return 'amount-income'
  if (row.classification === 'business_payment') return 'amount-expense'
  return row.direction === 'in' ? 'amount-income' : 'amount-expense'
}

function amountDisplay(row: TxnRow) {
  const debit = Number(row.debit) || 0
  const credit = Number(row.credit) || 0
  const amt = debit > 0 ? debit : credit
  const sign = row.direction === 'in' || debit > 0 ? '+' : '-'
  return `${sign}¥${amt.toLocaleString('en', { minimumFractionDigits: 2 })}`
}

function classificationTag(val: string) {
  const map: Record<string, string> = {
    business_receipt: 'success',
    business_payment: 'danger',
    bank_fee: 'warning',
    interest_income: 'primary',
    internal_transfer: 'info',
    tax_payment: 'warning',
    social_security: 'info',
    insurance_fee: 'warning',
    pending: 'danger',
  }
  return map[val] || 'info'
}

function classificationLabel(val: string) {
  const map: Record<string, string> = {
    business_receipt: '业务收款',
    business_payment: '业务付款',
    bank_fee: '银行费用',
    interest_income: '利息收入',
    internal_transfer: '内部转账',
    tax_payment: '税务缴费',
    social_security: '社保缴费',
    insurance_fee: '保险费用',
    pending: '待处理',
  }
  return map[val] || val || '—'
}

const selectedTotal = computed(() => {
  const ids = new Set(selectedIds.value)
  let sum = 0
  for (const t of txnList.value) {
    if (ids.has(t.id)) {
      const debit = Number(t.debit) || 0
      const credit = Number(t.credit) || 0
      sum += debit > 0 ? debit : credit
    }
  }
  return `¥${sum.toLocaleString('en', { minimumFractionDigits: 2 })}`
})
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.header-actions { display: flex; gap: 8px; }
.stat-row { margin-bottom: 16px; }
.stat-card { text-align: center; }
.stat-num { font-size: 28px; font-weight: 700; line-height: 1.2; }
.stat-num.danger { color: #f56c6c; }
.stat-label { font-size: 13px; color: #666; margin-top: 4px; }
.stat-card.pending .stat-num { color: #e6a23c; }
.stat-card.ai-ok .stat-num { color: #67c23a; }
.stat-card.manual .stat-num { color: #f56c6c; }
.filter-bar { display: flex; gap: 8px; margin-bottom: 12px; }
.batch-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.selected-count { font-size: 13px; color: #666; }
.clickable-row { cursor: pointer; }
.text-muted { color: #999; font-size: 13px; }
.bottom-summary {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-top: none;
  padding: 10px 16px;
  font-size: 13px;
  color: #666;
  text-align: right;
}
.amount-income { color: #67c23a; }
.amount-expense { color: #f56c6c; }
</style>
