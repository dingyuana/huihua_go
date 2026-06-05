<template>
  <div class="ar-invoice-page">
    <div class="page-header">
      <h3>应收款单</h3>
      <p class="page-hint">销售发票确认后自动生成</p>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 140px" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已确认" value="confirmed" />
            <el-option label="已冲销" value="reversed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="关联发票号" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ stats.total }}</p>
          <p class="stat-label">单据总数</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card draft">
          <p class="stat-num">{{ stats.draftCount }}</p>
          <p class="stat-label">草稿</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card confirmed">
          <p class="stat-num">{{ stats.confirmedCount }}</p>
          <p class="stat-label">已确认</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">¥{{ stats.totalAmount }}</p>
          <p class="stat-label">应收总额</p>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <el-table :data="filteredList" border stripe size="small" v-loading="loading">
        <el-table-column prop="invoice_no" label="关联发票号" min-width="180" show-overflow-tooltip />
        <el-table-column label="金额" width="140" align="right">
          <template #default="{ row }">
            <span class="amount-amount">¥{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="due_date" label="到期日" width="110">
          <template #default="{ row }">
            {{ row.due_date || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source_type" label="来源" width="100">
          <template #default="{ row }">
            {{ row.source_type || 'sales_invoice' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="生成时间" width="160" />
        <el-table-column prop="confirmed_at" label="确认时间" width="160">
          <template #default="{ row }">
            {{ row.confirmed_at || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="showDrawer" :title="`应收款单 ${currentItem?.invoice_no || ''}`" size="480px">
      <template v-if="currentItem">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="应收单ID">{{ currentItem.id }}</el-descriptions-item>
          <el-descriptions-item label="关联发票号">{{ currentItem.invoice_no }}</el-descriptions-item>
          <el-descriptions-item label="客户ID">{{ currentItem.customer_id }}</el-descriptions-item>
          <el-descriptions-item label="应收金额">¥{{ formatAmount(currentItem.amount) }}</el-descriptions-item>
          <el-descriptions-item label="到期日">{{ currentItem.due_date || '—' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="来源类型">{{ currentItem.source_type }}</el-descriptions-item>
          <el-descriptions-item label="生成时间">{{ currentItem.created_at }}</el-descriptions-item>
          <el-descriptions-item label="确认时间">{{ currentItem.confirmed_at || '—' }}</el-descriptions-item>
          <el-descriptions-item label="批准时间">{{ currentItem.approved_at || '—' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchArInvoices, type ArInvoice } from '@/api/modules/ar_invoice'

const loading = ref(false)
const list = ref<ArInvoice[]>([])
const total = ref(0)

const filter = reactive({
  status: '',
  keyword: '',
})

const filteredList = computed(() => {
  let result = list.value
  if (filter.status) {
    result = result.filter(r => r.status === filter.status)
  }
  if (filter.keyword) {
    const kw = filter.keyword.toLowerCase()
    result = result.filter(r => (r.invoice_no || '').toLowerCase().includes(kw))
  }
  return result
})

const stats = computed(() => {
  const all = filteredList.value
  const draft = all.filter(r => r.status === 'draft')
  const confirmed = all.filter(r => r.status === 'confirmed')
  const totalAmount = all.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  return {
    total: all.length,
    draftCount: draft.length,
    confirmedCount: confirmed.length,
    totalAmount: totalAmount.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
  }
})

const showDrawer = ref(false)
const currentItem = ref<ArInvoice | null>(null)

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchArInvoices()
    const data = res?.data || res
    list.value = data?.list || []
    total.value = data?.total || list.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    list.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.status = ''
  filter.keyword = ''
}

function showDetail(row: ArInvoice) {
  currentItem.value = row
  showDrawer.value = true
}

function statusLabel(s: string): string {
  switch (s) {
    case 'draft': return '草稿'
    case 'confirmed': return '已确认'
    case 'reversed': return '已冲销'
    default: return s || '—'
  }
}

function statusTag(s: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  switch (s) {
    case 'draft': return 'warning'
    case 'confirmed': return 'success'
    case 'reversed': return 'info'
    default: return 'info'
  }
}

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.ar-invoice-page {
  .page-header {
    margin-bottom: 16px;
    h3 { font-size: 18px; margin: 0 0 4px; }
    .page-hint { font-size: 12px; color: #999; margin: 0; }
  }
  .filter-card { margin-bottom: 12px; }
  .stat-row { margin-bottom: 12px; }
  .stat-card {
    text-align: center;
    .stat-num { font-size: 22px; font-weight: 700; margin-bottom: 4px; color: #333; }
    .stat-label { font-size: 12px; color: #999; }
    &.draft .stat-num { color: #d48806; }
    &.confirmed .stat-num { color: #389e0d; }
  }
  .amount-amount { color: #d4380d; font-weight: 600; }
}
</style>
