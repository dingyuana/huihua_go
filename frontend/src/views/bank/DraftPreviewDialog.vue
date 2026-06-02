<template>
  <el-dialog
    v-model="visible"
    title="草稿预览"
    width="640px"
    :close-on-click-modal="false"
    @close="onClose"
  >
    <div
      v-if="loading"
      style="text-align: center; padding: 40px;"
    >
      <el-icon
        class="is-loading"
        :size="24"
      >
        <Loading />
      </el-icon>
      <p style="color: #999; margin-top: 8px;">
        加载中...
      </p>
    </div>

    <template v-else-if="previewData">
      <!-- 流水详情 -->
      <el-descriptions
        :column="2"
        border
        size="small"
        class="txn-info"
      >
        <el-descriptions-item label="日期">
          {{ previewData.bank_txn?.txn_date }}
        </el-descriptions-item>
        <el-descriptions-item label="方向">
          {{ previewData.bank_txn?.direction === 'in' ? '收入' : '支出' }}
        </el-descriptions-item>
        <el-descriptions-item label="对方">
          {{ previewData.bank_txn?.counterparty_name || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="金额">
          <span :class="amountClass">{{ amountDisplay }}</span>
        </el-descriptions-item>
        <el-descriptions-item
          label="摘要"
          :span="2"
        >
          {{ previewData.bank_txn?.description || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- AI 分析结果 -->
      <div class="section-title">
        AI 分析结果
      </div>
      <el-card
        shadow="never"
        class="ai-card"
      >
        <el-descriptions
          :column="3"
          size="small"
        >
          <el-descriptions-item label="业务场景">
            {{ previewData.ai_result?.business_scene || '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="置信度">
            <el-tag
              v-if="previewData.ai_result?.confidence"
              size="small"
              :type="confidenceType"
            >
              {{ previewData.ai_result.confidence }}%
            </el-tag>
            <span v-else>—</span>
          </el-descriptions-item>
          <el-descriptions-item label="建议动作">
            {{ suggestedActionLabel }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 凭证草稿预览 -->
      <template v-if="previewData.draft_voucher">
        <div class="section-title">
          凭证草稿预览
        </div>
        <el-table
          :data="previewData.draft_voucher.lines || []"
          size="small"
          border
        >
          <el-table-column
            prop="account_name"
            label="科目"
            min-width="160"
          />
          <el-table-column
            prop="debit"
            label="借方"
            width="120"
            align="right"
          >
            <template #default="{ row }">
              {{ row.debit && row.debit !== '0.00' ? '¥' + row.debit : '' }}
            </template>
          </el-table-column>
          <el-table-column
            prop="credit"
            label="贷方"
            width="120"
            align="right"
          >
            <template #default="{ row }">
              {{ row.credit && row.credit !== '0.00' ? '¥' + row.credit : '' }}
            </template>
          </el-table-column>
        </el-table>
        <p class="voucher-summary">
          摘要: {{ previewData.draft_voucher.summary || '—' }}
        </p>
      </template>

      <!-- 草稿单据预览 -->
      <template v-else-if="previewData.or_draft_payment">
        <div class="section-title">
          单据草稿预览
        </div>
        <el-card
          shadow="never"
          class="payment-card"
        >
          <p>单据类型: {{ previewData.or_draft_payment.payment_type || '收款单' }}</p>
          <p>对方单位: {{ previewData.or_draft_payment.party_name || '—' }}</p>
        </el-card>
      </template>

      <div
        v-else
        class="no-draft"
      >
        <el-empty
          description="暂无草稿信息"
          :image-size="48"
        />
      </div>
    </template>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="onClose">
          取消
        </el-button>
        <el-button
          type="danger"
          @click="handleReject"
        >
          驳回
        </el-button>
        <el-button
          type="primary"
          @click="handleSubmit"
        >
          确认并提交审核
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { previewDraft, submitReview, rejectManual } from '@/api/modules/bank_txn_review'

interface PreviewData {
  bank_txn: {
    txn_date?: string
    direction?: string
    counterparty_name?: string
    description?: string
    debit?: string
    credit?: string
  }
  ai_result?: {
    business_scene?: string
    suggested_action?: string
    confidence?: number
  }
  draft_voucher?: {
    id: string
    lines: Array<{ account_name: string; debit: string; credit: string }>
    summary: string
  }
  or_draft_payment?: Record<string, any>
}

const props = defineProps<{ modelValue: boolean; txnId: string }>()
const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  'submitted': []
}>()

const visible = ref(false)
const loading = ref(false)
const previewData = ref<PreviewData | null>(null)

watch(() => props.modelValue, async (val) => {
  visible.value = val
  if (val && props.txnId) {
    await loadPreview()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

async function loadPreview() {
  loading.value = true
  previewData.value = null
  try {
    const res: any = await previewDraft(props.txnId)
    previewData.value = res?.data || null
  } catch (e: any) {
    ElMessage.error('加载预览失败: ' + (e?.message || ''))
  } finally {
    loading.value = false
  }
}

function onClose() {
  visible.value = false
  previewData.value = null
}

async function handleSubmit() {
  try {
    const res: any = await submitReview({ txn_ids: [props.txnId] })
    if (res?.data?.approved_count > 0) {
      ElMessage.success('提交审核成功')
      emit('submitted')
      onClose()
    }
  } catch (e: any) {
    ElMessage.error('提交失败: ' + (e?.message || ''))
  }
}

async function handleReject() {
  try {
    const res: any = await rejectManual({ txn_ids: [props.txnId] })
    if (res?.data?.rejected_count > 0) {
      ElMessage.success('已驳回至待人工处理')
      emit('submitted')
      onClose()
    }
  } catch (e: any) {
    ElMessage.error('驳回失败: ' + (e?.message || ''))
  }
}

const amountDisplay = computed(() => {
  if (!previewData.value?.bank_txn) return '—'
  const t = previewData.value.bank_txn
  const debit = Number(t.debit) || 0
  const credit = Number(t.credit) || 0
  const amt = debit > 0 ? debit : credit
  const sign = t.direction === 'in' || debit > 0 ? '+' : '-'
  return `${sign}¥${amt.toLocaleString('en', { minimumFractionDigits: 2 })}`
})

function amountClass() {
  const t = previewData.value?.bank_txn
  if (!t) return ''
  const debit = Number(t.debit) || 0
  return debit > 0 ? 'amount-income' : 'amount-expense'
}

const confidenceType = computed(() => {
  const c = previewData.value?.ai_result?.confidence
  if (!c) return 'info'
  if (c >= 80) return 'success'
  if (c >= 60) return 'warning'
  return 'danger'
})

const suggestedActionLabel = computed(() => {
  const map: Record<string, string> = {
    auto_voucher: '生成凭证',
    payment_entry: '生成收款/付款单',
    manual_pending: '待人工处理',
  }
  return map[previewData.value?.ai_result?.suggested_action || ''] || previewData.value?.ai_result?.suggested_action || '—'
})
</script>

<style scoped>
.txn-info { margin-bottom: 16px; }
.section-title { font-size: 14px; font-weight: 600; margin: 12px 0 8px; color: #333; }
.ai-card { background: #f5f7fa; }
.voucher-summary { font-size: 13px; color: #666; margin-top: 8px; }
.payment-card { background: #f5f7fa; p { margin: 4px 0; font-size: 13px; } }
.no-draft { text-align: center; padding: 16px; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 12px; }
.amount-income { color: #67c23a; font-weight: 600; }
.amount-expense { color: #f56c6c; font-weight: 600; }
</style>
