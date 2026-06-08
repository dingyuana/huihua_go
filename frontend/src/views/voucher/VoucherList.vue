<template>
  <div class="voucher-page">
    <div class="page-header">
      <h3>凭证列表</h3>
      <el-button type="primary" @click="$router.push('/vouchers/create')">+ 新增凭证</el-button>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="类型">
          <el-select v-model="filter.type" placeholder="全部" style="width: 100px" clearable>
            <el-option v-for="t in ['记','银','现','转','折旧','结转']" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 120px" clearable>
            <el-option label="草稿" :value="0" />
            <el-option label="已提交" :value="1" />
            <el-option label="已审核" :value="2" />
            <el-option label="已作废" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="filter.dateRange" type="daterange" range-separator="~" style="width: 220px" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item><el-button type="primary" @click="fetchVouchers">查询</el-button></el-form-item>
      </el-form>
    </el-card>

    <!-- 统计汇总 -->
    <el-row :gutter="8" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card debit">
          <div class="stat-summary">
            <span class="stat-label">借方总额</span>
            <span class="stat-amount">{{ formatAmount(stats.debitTotal) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card credit">
          <div class="stat-summary">
            <span class="stat-label">贷方总额</span>
            <span class="stat-amount">{{ formatAmount(stats.creditTotal) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card balance">
          <div class="stat-summary">
            <span class="stat-label">借贷差额</span>
            <span class="stat-amount" :class="stats.balance === 0 ? 'zero' : stats.balance > 0 ? 'debit' : 'credit'">
              {{ formatAmount(stats.balance) }} 元
            </span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card total">
          <div class="stat-summary">
            <span class="stat-label">凭证总数</span>
            <span class="stat-amount">{{ stats.count }} 张</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <el-table :data="vouchers" border stripe size="small" @row-click="goDetail">
        <el-table-column prop="voucher_no" label="凭证号" width="150" />
        <el-table-column prop="posting_date" label="日期" width="110">
          <template #default="{ row }">{{ (row.posting_date || '').slice(0, 10) }}</template>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'
import DocStatusTag from '@/components/business/DocStatusTag.vue'

const router = useRouter()

const filter = reactive({ type: '', status: null as number | null, dateRange: null as any })

const vouchers = ref<any[]>([])

const stats = computed(() => {
  const list = vouchers.value
  let debitTotal = 0
  let creditTotal = 0
  for (const v of list) {
    debitTotal += Number(v.debit_total || 0)
    creditTotal += Number(v.credit_total || 0)
  }
  return {
    debitTotal,
    creditTotal,
    balance: debitTotal - creditTotal,
    count: list.length,
  }
})

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function fetchVouchers() {
  try {
    const params: any = { limit: 100, offset: 0 }
    if (filter.type) params.voucher_type = filter.type
    if (filter.status !== null && filter.status !== undefined) params.doc_status = filter.status
    if (filter.dateRange && filter.dateRange.length === 2) {
      params.start_date = filter.dateRange[0]
      params.end_date = filter.dateRange[1]
    }
    const res: any = await request.get('/vouchers', { params })
    const list = res?.vouchers || res?.data?.list || res?.data
    if (Array.isArray(list)) { vouchers.value = list }
  } catch { /* no data */ }
}

onMounted(() => fetchVouchers())

function goDetail(row: any) {
  router.push(`/vouchers/${row.id}`)
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
  &.debit .stat-amount { color: #096dd9; }
  &.credit .stat-amount { color: #cf1322; }
  &.balance {
    .stat-amount.debit { color: #096dd9; }
    .stat-amount.credit { color: #cf1322; }
    .stat-amount.zero { color: #999; }
  }
  &.total .stat-amount { color: #333; }
}
</style>
