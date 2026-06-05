<template>
  <div class="credit-page">
    <div class="page-header">
      <h3>信用管控</h3>
      <p class="page-hint">监控客户/供应商信用额度占用情况，识别超限风险</p>
    </div>

    <el-card class="stat-row">
      <el-row :gutter="12">
        <el-col :span="8">
          <el-statistic :value="overLimitCount" :value-style="{ color: '#cf1322' }">
            <template #title><span style="color:#999">超限客户数</span></template>
          </el-statistic>
        </el-col>
        <el-col :span="8">
          <el-statistic :value="totalOverLimit" :precision="2" :value-style="{ color: '#cf1322' }">
            <template #title><span style="color:#999">超限总额 (元)</span></template>
          </el-statistic>
        </el-col>
        <el-col :span="8">
          <el-statistic :value="totalCreditLimit" :precision="2" :value-style="{ color: '#389e0d' }">
            <template #title><span style="color:#999">总授信额度 (元)</span></template>
          </el-statistic>
        </el-col>
      </el-row>
    </el-card>

    <el-card>
      <el-table :data="list" border stripe size="small" v-loading="loading">
        <el-table-column prop="party_name" label="客户/供应商" min-width="200" show-overflow-tooltip />
        <el-table-column label="授信额度" width="140" align="right">
          <template #default="{ row }"><span>¥{{ fmt(row.credit_limit) }}</span></template>
        </el-table-column>
        <el-table-column label="已用" width="140" align="right">
          <template #default="{ row }"><span class="amount-used">¥{{ fmt(row.credit_used) }}</span></template>
        </el-table-column>
        <el-table-column label="可用" width="140" align="right">
          <template #default="{ row }"><span class="amount-avail">¥{{ fmt(row.available) }}</span></template>
        </el-table-column>
        <el-table-column label="使用率" width="120" align="right">
          <template #default="{ row }">
            <el-progress :percentage="Number(row.utilization_pct) || 0" :status="row.over_limit ? 'exception' : (Number(row.utilization_pct) > 80 ? 'warning' : 'success')" :stroke-width="14" />
          </template>
        </el-table-column>
        <el-table-column label="允许超期" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.overdraft_allowed ? 'success' : 'info'" size="small">{{ row.overdraft_allowed ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.over_limit" type="danger" size="small">超限</el-tag>
            <el-tag v-else-if="Number(row.utilization_pct) > 80" type="warning" size="small">高占用</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchOverLimitParties, type CreditStatus } from '@/api/modules/credit_control'

const loading = ref(false)
const list = ref<CreditStatus[]>([])

const overLimitCount = computed(() => list.value.length)
const totalOverLimit = computed(() => list.value.reduce((s, r) => s + Math.max(0, Number(r.credit_used) - Number(r.credit_limit)), 0))
const totalCreditLimit = computed(() => list.value.reduce((s, r) => s + Number(r.credit_limit), 0))

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchOverLimitParties()
    const data = res?.data || res
    list.value = data?.list || []
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

onMounted(loadData)
</script>

<style scoped lang="scss">
.credit-page {
  .page-header { margin-bottom: 16px; h3 { font-size: 18px; margin: 0 0 4px; } .page-hint { font-size: 12px; color: #999; margin: 0; } }
  .stat-row { margin-bottom: 12px; }
  .amount-used { color: #d4380d; font-weight: 600; }
  .amount-avail { color: #389e0d; }
}
</style>
