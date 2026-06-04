<template>
  <div class="asset-detail-page">
    <div class="page-header">
      <h3>资产详情</h3>
      <div class="header-actions">
        <el-button @click="router.back()">返回</el-button>
      </div>
    </div>

    <div v-loading="loading">
      <!-- 基本信息 -->
      <el-card class="info-card" style="margin-bottom: 12px">
        <template #header>基本信息</template>
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="资产编号">{{ asset.asset_code }}</el-descriptions-item>
          <el-descriptions-item label="资产名称">{{ asset.asset_name }}</el-descriptions-item>
          <el-descriptions-item label="资产类别">{{ asset.category }}</el-descriptions-item>
          <el-descriptions-item label="购置日期">{{ asset.purchase_date }}</el-descriptions-item>
          <el-descriptions-item label="折旧方法">{{ asset.depreciation_method || '直线法' }}</el-descriptions-item>
          <el-descriptions-item label="折旧开始日期">{{ asset.depreciation_start_date }}</el-descriptions-item>
          <el-descriptions-item label="预计使用月数">{{ asset.useful_life_months }}</el-descriptions-item>
          <el-descriptions-item label="原值">{{ formatAmount(asset.original_value) }}</el-descriptions-item>
          <el-descriptions-item label="残值">{{ formatAmount(asset.residual_value) }}</el-descriptions-item>
          <el-descriptions-item label="累计折旧">{{ formatAmount(asset.accumulated_depreciation) }}</el-descriptions-item>
          <el-descriptions-item label="当前净值">{{ formatAmount(asset.net_value) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag v-if="asset.depreciation_status === 'normal'" type="success" size="small">正常</el-tag>
            <el-tag v-else-if="asset.depreciation_status === 'fully_depreciated'" type="info" size="small">已提足</el-tag>
            <el-tag v-else type="danger" size="small">已处置</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 折旧计划 -->
      <el-card class="schedule-card">
        <template #header>
          <span>折旧计划</span>
          <el-button v-if="!schedule.length" type="primary" size="small" style="float:right" @click="handleCreateSchedule">
            生成折旧计划
          </el-button>
        </template>

        <el-table v-if="schedule.length" :data="schedule" border stripe size="small">
          <el-table-column prop="period" label="期间" width="100" />
          <el-table-column prop="depreciation_amount" label="本期折旧" width="130" align="right">
            <template #default="{ row }">{{ formatAmount(row.depreciation_amount) }}</template>
          </el-table-column>
          <el-table-column prop="accumulated_amount" label="累计折旧" width="130" align="right">
            <template #default="{ row }">{{ formatAmount(row.accumulated_amount) }}</template>
          </el-table-column>
          <el-table-column prop="net_value" label="账面净值" width="130" align="right">
            <template #default="{ row }">{{ formatAmount(row.net_value) }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无折旧计划，点击按钮生成" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchAssetDetail, fetchDepreciationSchedule, createDepreciationSchedule } from '@/api/modules/asset'
import type { Asset, DepreciationScheduleRow } from '@/api/modules/asset'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const asset = ref<Asset>({} as Asset)
const schedule = ref<DepreciationScheduleRow[]>([])

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  const id = route.params.id as string
  if (!id) return
  loading.value = true
  try {
    const [assetRes, schedRes]: any = await Promise.all([
      fetchAssetDetail(id),
      fetchDepreciationSchedule(id).catch(() => ({ data: null })),
    ])
    asset.value = assetRes?.data || assetRes || {}
    const sched = schedRes?.data || schedRes
    schedule.value = sched?.schedule_rows || []
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleCreateSchedule() {
  try {
    await createDepreciationSchedule({ asset_id: route.params.id as string })
    ElMessage.success('折旧计划已生成')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '生成失败')
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.info-card, .schedule-card { margin-bottom: 12px; }
</style>