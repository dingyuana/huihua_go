<template>
  <div class="aging-page">
    <div class="page-header">
      <h3>账龄分析</h3>
      <p class="page-hint">按客户维度统计未结清应收/应付余额，识别逾期风险</p>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="分析方向">
          <el-radio-group v-model="direction">
            <el-radio-button label="ar">应收账龄</el-radio-button>
            <el-radio-button label="ap">应付账龄</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="截止日期">
          <el-date-picker v-model="asOf" type="date" value-format="YYYY-MM-DD" style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" class="stat-row" v-if="summary">
      <el-col :span="4"><el-card shadow="hover" class="stat-card"><p class="stat-num">¥{{ fmt(summary.total_current) }}</p><p class="stat-label">未到期</p></el-card></el-col>
      <el-col :span="5"><el-card shadow="hover" class="stat-card overdue"><p class="stat-num">¥{{ fmt(summary.total_0_30) }}</p><p class="stat-label">0-30 天</p></el-card></el-col>
      <el-col :span="5"><el-card shadow="hover" class="stat-card overdue"><p class="stat-num">¥{{ fmt(summary.total_30_60) }}</p><p class="stat-label">30-60 天</p></el-card></el-col>
      <el-col :span="5"><el-card shadow="hover" class="stat-card overdue"><p class="stat-num">¥{{ fmt(summary.total_60_90) }}</p><p class="stat-label">60-90 天</p></el-card></el-col>
      <el-col :span="5"><el-card shadow="hover" class="stat-card serious"><p class="stat-num">¥{{ fmt(summary.total_90_plus) }}</p><p class="stat-label">90+ 天</p></el-card></el-col>
    </el-row>

    <el-card>
      <el-table :data="buckets" border stripe size="small" v-loading="loading">
        <el-table-column prop="party_name" label="客户/供应商" min-width="200" show-overflow-tooltip />
        <el-table-column label="未到期" width="120" align="right">
          <template #default="{ row }"><span>¥{{ fmt(row.current) }}</span></template>
        </el-table-column>
        <el-table-column label="0-30 天" width="120" align="right">
          <template #default="{ row }"><span :class="numCls(row.b_0_30)">¥{{ fmt(row.b_0_30) }}</span></template>
        </el-table-column>
        <el-table-column label="30-60 天" width="120" align="right">
          <template #default="{ row }"><span :class="numCls(row.b_30_60)">¥{{ fmt(row.b_30_60) }}</span></template>
        </el-table-column>
        <el-table-column label="60-90 天" width="120" align="right">
          <template #default="{ row }"><span :class="numCls(row.b_60_90)">¥{{ fmt(row.b_60_90) }}</span></template>
        </el-table-column>
        <el-table-column label="90+ 天" width="120" align="right">
          <template #default="{ row }"><span :class="numCls(row.b_90_plus)">¥{{ fmt(row.b_90_plus) }}</span></template>
        </el-table-column>
        <el-table-column label="合计" width="140" align="right">
          <template #default="{ row }"><b>¥{{ fmt(row.total) }}</b></template>
        </el-table-column>
        <el-table-column prop="invoice_count" label="单据数" width="80" align="right" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchARAging, fetchAPAging, type AgingBucket, type AgingSummary } from '@/api/modules/aging'

const direction = ref<'ar' | 'ap'>('ar')
const asOf = ref('')
const loading = ref(false)
const buckets = ref<AgingBucket[]>([])
const summary = ref<AgingSummary | null>(null)

async function loadData() {
  loading.value = true
  try {
    const res: any = direction.value === 'ar' ? await fetchARAging(asOf.value || undefined) : await fetchAPAging(asOf.value || undefined)
    const data = res?.data || res
    buckets.value = data?.buckets || []
    summary.value = data?.summary || null
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

function fmt(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function numCls(val: any): string {
  const n = Number(val) || 0
  if (n === 0) return ''
  if (n > 10000) return 'amount-serious'
  return 'amount-overdue'
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.aging-page {
  .page-header { margin-bottom: 16px; h3 { font-size: 18px; margin: 0 0 4px; } .page-hint { font-size: 12px; color: #999; margin: 0; } }
  .filter-card { margin-bottom: 12px; }
  .stat-row { margin-bottom: 12px; }
  .stat-card { text-align: center; .stat-num { font-size: 20px; font-weight: 700; margin-bottom: 4px; color: #333; } .stat-label { font-size: 12px; color: #999; }
    &.overdue .stat-num { color: #d48806; }
    &.serious .stat-num { color: #cf1322; }
  }
  .amount-overdue { color: #d48806; }
  .amount-serious { color: #cf1322; font-weight: 600; }
}
</style>
