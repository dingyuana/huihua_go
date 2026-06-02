<template>
  <div class="dashboard">
    <h2>欢迎使用慧财智能财务平台</h2>
    <p class="company-name">{{ companyName }}</p>
    <el-row :gutter="16" class="stat-cards">
      <el-col :span="6">
        <el-card shadow="hover">
          <p class="stat-label">本月流水</p>
          <p class="stat-value">{{ stats.monthlyTxns }} 笔</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <p class="stat-label">待审核凭证</p>
          <p class="stat-value warning">{{ stats.pendingVouchers }} 张</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <p class="stat-label">本月净利润</p>
          <p class="stat-value positive">¥{{ stats.monthlyProfit }}</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <p class="stat-label">待处理流水</p>
          <p class="stat-value danger">{{ stats.pendingTxns }} 条</p>
        </el-card>
      </el-col>
    </el-row>
    <el-card class="quick-actions">
      <template #header>
        <span>快速入口</span>
      </template>
      <el-space wrap>
        <el-button type="primary" @click="$router.push('/bank/import')">导入银行流水</el-button>
        <el-button @click="$router.push('/vouchers/create')">新增凭证</el-button>
        <el-button @click="$router.push('/period/health-check')">结账体检</el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import request from '@/api/request'
import { fetchDashboardStats } from '@/api/modules/dashboard'

const companyName = ref('加载中...')

const stats = reactive({
  monthlyTxns: 0,
  pendingVouchers: 0,
  monthlyProfit: '0',
  pendingTxns: 0,
})

onMounted(async () => {
  // Load company name
  try {
    const res: any = await request.get('/account-setup/status')
    const data = res?.data || res
    if (data?.company?.company_name) {
      companyName.value = data.company.company_name
    } else {
      companyName.value = ''
    }
  } catch {
    companyName.value = ''
  }

  // Try loading real dashboard stats (falls back to 0 on error)
  try {
    const res: any = await fetchDashboardStats()
    const data = res?.data || res
    if (data) {
      stats.monthlyTxns = data.monthly_txns ?? 0
      stats.pendingVouchers = data.pending_vouchers ?? 0
      stats.pendingTxns = data.pending_txns ?? 0
      if (data.monthly_profit != null) {
        stats.monthlyProfit = (Number(data.monthly_profit) / 100).toLocaleString()
      }
    }
  } catch {
    // Keep defaults (all zeros)
  }
})
</script>

<style scoped lang="scss">
.dashboard {
  padding: 24px;
  h2 { margin-bottom: 4px; }
  .company-name {
    color: #666;
    margin-bottom: 24px;
  }
  .stat-cards {
    margin-bottom: 24px;
  }
  .stat-label {
    font-size: 13px;
    color: #999;
    margin-bottom: 8px;
  }
  .stat-value {
    font-size: 24px;
    font-weight: 600;
    color: #333;
    &.warning { color: #faad14; }
    &.positive { color: #cf1322; }
    &.danger { color: #ff4d4f; }
  }
}
</style>
