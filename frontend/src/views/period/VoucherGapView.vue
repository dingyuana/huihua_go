<template>
  <div class="voucher-gap-page">
    <div class="page-header">
      <h3>凭证断号检测</h3>
      <div>
        <el-date-picker
          v-model="period"
          type="month"
          value-format="YYYY-MM"
          placeholder="选择期间"
          style="width: 160px; margin-right: 8px"
        />
        <el-button
          type="primary"
          :loading="loading"
          @click="runCheck"
        >
          执行检测
        </el-button>
      </div>
    </div>

    <!-- 整体状态 -->
    <el-card
      :class="['status-card', overallStatus]"
      shadow="never"
    >
      <div class="status-badge">
        <el-tag
          :type="overallTagType"
          size="large"
        >
          {{ statusText }}
        </el-tag>
      </div>
    </el-card>

    <!-- 统计概览 -->
    <CheckSummaryCard
      :summary="summary"
      :loading="loading"
    />

    <!-- 断号列表 -->
    <el-card
      shadow="never"
      class="gap-list-card"
    >
      <CheckResultPanel
        :checks="gapChecks"
        :loading="loading"
        title="断号明细"
        @action="handleGapAction"
      />
    </el-card>

    <!-- 快速修复提示 -->
    <el-card
      v-if="hasMissing && !loading"
      shadow="never"
      class="fix-card"
    >
      <el-alert
        title="存在缺失断号"
        type="error"
        :closable="false"
        show-icon
      >
        <template #default>
          <p>缺失的凭证号不可自动修复，请确认：</p>
          <ol style="margin: 8px 0; padding-left: 20px">
            <li>是否确实未录入该编号的凭证</li>
            <li>是否已被其他用户删除/作废</li>
            <li>是否需要补录缺失凭证</li>
          </ol>
        </template>
      </el-alert>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchVoucherGaps } from '@/api/modules/period'
import type { VoucherGap } from '@/api/modules/period'
import CheckResultPanel from '@/components/check/CheckResultPanel.vue'
import CheckSummaryCard from '@/components/check/CheckSummaryCard.vue'
import type { CheckItem, CheckSummary } from '@/types/check'

const now = new Date()
const period = ref(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)
const loading = ref(false)

const gaps = ref<VoucherGap[]>([])

const hasMissing = ref(false)

const gapChecks = computed<CheckItem[]>(() =>
  gaps.value.map(g => ({
    id: `gap-${g.gap_type}-${g.expected_no}`,
    name: g.gap_type === 'missing' ? '凭证缺失' : '凭证作废',
    status: g.gap_type === 'missing' ? 'blocked' as const : 'warning' as const,
    message: g.message,
    action: g.gap_type === 'voided' ? { label: '查看凭证' } : undefined,
  }))
)

const overallStatus = computed(() => {
  if (gaps.value.some(g => g.gap_type === 'missing')) return 'red'
  if (gaps.value.some(g => g.gap_type === 'voided')) return 'yellow'
  return 'green'
})

const overallTagType = computed(() => {
  if (overallStatus.value === 'red') return 'danger'
  if (overallStatus.value === 'yellow') return 'warning'
  return 'success'
})

const statusText = computed(() => {
  if (overallStatus.value === 'red') return '🔴 存在缺失断号，需人工处理'
  if (overallStatus.value === 'yellow') return '🟡 存在作废断号，请确认合理性'
  return '🟢 凭证编号连续'
})

const summary = computed<CheckSummary>(() => ({
  total: gaps.value.length,
  passed: gaps.value.filter(g => g.gap_type === 'voided').length,
  warning: gaps.value.filter(g => g.gap_type === 'voided').length,
  blocked: gaps.value.filter(g => g.gap_type === 'missing').length,
  pending: 0,
}))

function handleGapAction(_checkId: string) {
  // 预留：查看凭证详情
}

async function runCheck() {
  loading.value = true
  const [year, month] = period.value.split('-').map(Number)
  try {
    const res: any = await fetchVoucherGaps(year, month)
    const data = res?.data || res
    if (data?.voucher_gaps) {
      gaps.value = data.voucher_gaps
      hasMissing.value = data.has_missing || data.voucher_gaps.some((g: VoucherGap) => g.gap_type === 'missing')
      return
    }
  } catch {
    // keep empty state
  }
  loading.value = false
}

onMounted(runCheck)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.status-card {
  margin-bottom: 16px;
  text-align: center;
  padding: 8px;
  &.red { border-left: 4px solid #ff4d4f; }
  &.yellow { border-left: 4px solid #faad14; }
  &.green { border-left: 4px solid #52c41a; }
}
.status-badge {
  padding: 8px 0;
}
.gap-list-card {
  margin-bottom: 16px;
}
.fix-card {
  margin-bottom: 16px;
}
</style>
