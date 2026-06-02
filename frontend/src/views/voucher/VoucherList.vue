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
          <el-select v-model="filter.status" placeholder="全部" style="width: 100px" clearable>
            <el-option label="草稿" :value="0" />
            <el-option label="已审核" :value="1" />
            <el-option label="已作废" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="filter.dateRange" type="daterange" range-separator="~" style="width: 220px" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item><el-button type="primary" @click="fetchVouchers">查询</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <el-table :data="vouchers" border stripe size="small" @row-click="goDetail">
        <el-table-column prop="voucher_no" label="凭证号" width="150" />
        <el-table-column prop="posting_date" label="日期" width="110">
          <template #default="{ row }">{{ (row.posting_date || '').slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column label="对方名称" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.counterparty_name || '—' }}</template>
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
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'
import DocStatusTag from '@/components/business/DocStatusTag.vue'

const router = useRouter()

const filter = reactive({ type: '', status: null as number | null, dateRange: null as any })

const vouchers = ref<any[]>([])

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
</style>
