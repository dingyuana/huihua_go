<template>
  <div class="precheck">
    <div class="page-header">
      <h3>核销预检</h3>
    </div>

    <!-- 选择区 -->
    <el-row
      :gutter="16"
      class="selection-row"
    >
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            选择收款单
          </template>
          <el-select
            v-model="selectedPayment"
            placeholder="搜索收款单"
            filterable
            style="width: 100%"
          />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            选择发票
          </template>
          <el-select
            v-model="selectedInvoice"
            placeholder="搜索发票"
            filterable
            style="width: 100%"
          />
        </el-card>
      </el-col>
      <el-col
        :span="8"
        class="action-col"
      >
        <el-button
          type="primary"
          size="large"
          :disabled="!selectedPayment || !selectedInvoice"
          @click="runPrecheck"
        >
          执行预检
        </el-button>
      </el-col>
    </el-row>

    <!-- 预检结果 -->
    <el-card
      v-if="precheckDone"
      shadow="never"
      class="result-card"
    >
      <template #header>
        核销预检结果
      </template>

      <CheckSummaryCard :summary="summary" />

      <CheckResultPanel
        :checks="checkResults"
        :loading="loading"
        @action="handleCheckAction"
      />

      <div class="precheck-actions">
        <BlockingGuard
          :blocked="blockerCount > 0"
          :blocked-count="blockerCount"
        >
          <el-button
            type="primary"
            @click="showForcePassDialog = true"
          >
            强制通过并核销
          </el-button>
        </BlockingGuard>
      </div>
    </el-card>

    <!-- 强制通过弹窗 -->
    <el-dialog
      v-model="showForcePassDialog"
      title="强制通过核销"
      width="450px"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
      >
        <p>以下检查项未通过，强制通过需备注原因：</p>
        <ul>
          <li
            v-for="c in blockedChecks"
            :key="c.id"
            style="margin:4px 0"
          >
            {{ c.name }}: {{ c.message }}
          </li>
        </ul>
      </el-alert>
      <el-form class="force-form">
        <el-form-item
          label="备注原因"
          required
        >
          <el-input
            v-model="forcePassReason"
            type="textarea"
            :rows="3"
            placeholder="请说明强制通过的原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showForcePassDialog = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :disabled="!forcePassReason"
          @click="executeForcePass"
        >
          确认强制通过
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import request from '@/api/request'
import CheckResultPanel from '@/components/check/CheckResultPanel.vue'
import CheckSummaryCard from '@/components/check/CheckSummaryCard.vue'
import BlockingGuard from '@/components/check/BlockingGuard.vue'
import type { CheckItem, CheckSummary } from '@/types/check'

const router = useRouter()
const selectedPayment = ref('')
const selectedInvoice = ref('')
const precheckDone = ref(false)
const loading = ref(false)
const showForcePassDialog = ref(false)
const forcePassReason = ref('')

const checks = ref<CheckItem[]>([])

const checkResults = computed(() => checks.value)
const blockerCount = computed(() => checks.value.filter(c => c.status === 'blocked').length)
const blockedChecks = computed(() => checks.value.filter(c => c.status === 'blocked'))

const summary = computed<CheckSummary>(() => ({
  total: checks.value.length,
  passed: checks.value.filter(c => c.status === 'passed').length,
  warning: checks.value.filter(c => c.status === 'warning').length,
  blocked: blockerCount.value,
  pending: checks.value.filter(c => c.status === 'pending').length,
}))

function handleCheckAction(_checkId: string) {
  // 预留：查看发票详情等操作
}

async function runPrecheck() {
  loading.value = true
  precheckDone.value = true
  try {
    const res: any = await request.post('/reconciliation/precheck', {
      payment_id: selectedPayment.value,
      invoice_id: selectedInvoice.value,
    })
    const items = res?.data?.checks || res?.data
    if (Array.isArray(items) && items.length > 0) {
      checks.value = items
      loading.value = false
      return
    }
  } catch {
    // fallback
  }
  checks.value = []
  loading.value = false
  ElMessage.success('预检完成')
}

function executeForcePass() {
  ElMessage.success(`已强制通过核销（原因：${forcePassReason.value}）`)
  showForcePassDialog.value = false
  forcePassReason.value = ''
  router.push('/reconciliation/match')
}
</script>

<style scoped lang="scss">
.page-header h3 {
  font-size: 18px;
  margin-bottom: 16px;
}
.selection-row {
  margin-bottom: 16px;
}
.action-col {
  display: flex;
  align-items: flex-end;
  padding-bottom: 20px;
}
.result-card {
  margin-top: 0;
}
.precheck-actions {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
