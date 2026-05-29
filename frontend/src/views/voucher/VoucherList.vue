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
          <el-date-picker v-model="filter.dateRange" type="daterange" range-separator="~" style="width: 220px" />
        </el-form-item>
        <el-form-item><el-button type="primary">查询</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <el-table :data="vouchers" border stripe size="small" @row-click="goDetail">
        <el-table-column prop="voucher_no" label="凭证号" width="150" />
        <el-table-column prop="posting_date" label="日期" width="90" />
        <el-table-column prop="remark" label="摘要" min-width="200" show-overflow-tooltip />
        <el-table-column label="借方合计" width="120" align="right"><template #default="{ row }">{{ row.debit_total }}</template></el-table-column>
        <el-table-column label="贷方合计" width="120" align="right"><template #default="{ row }">{{ row.credit_total }}</template></el-table-column>
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

const router = useRouter()

const filter = reactive({ type: '', status: null as number | null, dateRange: null as any })

const localVouchers = [
  { id: 'v1', voucher_no: '记-2026-05-0001', posting_date: '2026-05-06', remark: '收款-上海XX贸易公司', debit_total: '12,000.00', credit_total: '12,000.00', docstatus: 1 },
  { id: 'v2', voucher_no: '记-2026-05-0002', posting_date: '2026-05-07', remark: '付款-北京YY科技有限公司', debit_total: '5,000.00', credit_total: '5,000.00', docstatus: 0 },
  { id: 'v3', voucher_no: '记-2026-05-0003', posting_date: '2026-05-08', remark: '计提折旧', debit_total: '3,000.00', credit_total: '3,000.00', docstatus: 2 },
  { id: 'v4', voucher_no: '记-2026-05-0004', posting_date: '2026-05-09', remark: '银行手续费', debit_total: '50.00', credit_total: '50.00', docstatus: 1 },
]
const vouchers = ref(localVouchers)

onMounted(async () => {
  try {
    const res: any = await request.get('/vouchers')
    const list = res?.data?.list || res?.data
    if (Array.isArray(list)) { vouchers.value = list; return }
  } catch { /* fallback */ }
})

function goDetail(row: any) {
  router.push(`/vouchers/${row.id}`)
}
</script>
<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.filter-card { margin-bottom: 16px; }
</style>
