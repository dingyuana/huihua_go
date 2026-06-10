<template>
  <div class="precheck">
    <div class="page-header"><h3>核销预检</h3></div>

    <!-- 选择区 -->
    <el-row :gutter="16" class="selection-row">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>选择收款单</template>
          <el-select
            v-model="selectedPayment"
            placeholder="搜索收款单"
            filterable
            style="width: 100%"
            :loading="loadingPayments"
          >
            <el-option
              v-for="p in payments"
              :key="p.id"
              :label="paymentLabel(p)"
              :value="p.id"
            />
          </el-select>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>选择发票</template>
          <el-select
            v-model="selectedInvoice"
            placeholder="搜索发票"
            filterable
            style="width: 100%"
            :loading="loadingInvoices"
          >
            <el-option
              v-for="inv in invoices"
              :key="inv.id"
              :label="invoiceLabel(inv)"
              :value="inv.id"
            />
          </el-select>
        </el-card>
      </el-col>
      <el-col :span="8" class="action-col">
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
    <el-card v-if="precheckDone" shadow="never" class="result-card">
      <template #header>核销预检结果</template>

      <CheckSummaryCard :summary="summary" />

      <CheckResultPanel
        :checks="checkResults"
        :loading="loading"
        @action="handleCheckAction"
      />

      <div class="precheck-actions">
        <BlockingGuard :blocked="blockerCount > 0" :blocked-count="blockerCount">
          <el-button
            type="primary"
            :loading="executing"
            :disabled="blockerCount > 0"
            @click="executeReconciliation"
          >
            执行核销
          </el-button>
          <el-button
            v-if="blockerCount > 0"
            type="warning"
            :loading="executing"
            @click="showForcePassDialog = true"
          >
            强制通过并核销
          </el-button>
        </BlockingGuard>
      </div>
    </el-card>

    <!-- 强制通过弹窗 -->
    <el-dialog v-model="showForcePassDialog" title="强制通过核销" width="450px">
      <el-alert type="warning" :closable="false" show-icon>
        <p>以下检查项未通过，强制通过需备注原因：</p>
        <ul>
          <li v-for="c in blockedChecks" :key="c.id" style="margin:4px 0">
            {{ c.name }}: {{ c.message }}
          </li>
        </ul>
      </el-alert>
      <el-form class="force-form">
        <el-form-item label="备注原因" required>
          <el-input
            v-model="forcePassReason"
            type="textarea"
            :rows="3"
            placeholder="请说明强制通过的原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showForcePassDialog = false">取消</el-button>
        <el-button type="primary" :loading="executing" :disabled="!forcePassReason" @click="executeForcePass">
          确认强制通过
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import CheckResultPanel from '@/components/check/CheckResultPanel.vue'
import CheckSummaryCard from '@/components/check/CheckSummaryCard.vue'
import BlockingGuard from '@/components/check/BlockingGuard.vue'
import type { CheckItem, CheckSummary } from '@/types/check'

const selectedPayment = ref('')
const selectedInvoice = ref('')
const precheckDone = ref(false)
const loading = ref(false)
const executing = ref(false)
const showForcePassDialog = ref(false)
const forcePassReason = ref('')

const checks = ref<CheckItem[]>([])

// 下拉列表数据
interface PaymentOption {
  id: string
  counterparty_name?: string
  debit: string
  credit: string
  description?: string
  txn_date: string
}

interface InvoiceOption {
  id: string
  invoice_no: string
  customer_name?: string
  total_amount: string
  outstanding_amount: string
  posting_date: string
}

const payments = ref<PaymentOption[]>([])
const invoices = ref<InvoiceOption[]>([])
const loadingPayments = ref(false)
const loadingInvoices = ref(false)

function paymentLabel(p: PaymentOption): string {
  const amt = p.credit && p.credit !== '0' ? p.credit : p.debit
  const party = p.counterparty_name ? p.counterparty_name : '无名'
  const desc = p.description ? `(${p.description})` : ''
  return `¥${amt} ${party} ${desc}`
}

function invoiceLabel(inv: InvoiceOption): string {
  const cust = inv.customer_name ? ` - ${inv.customer_name}` : ''
  return `${inv.invoice_no} ¥${inv.outstanding_amount}${cust}`
}

async function loadPayments() {
  loadingPayments.value = true
  try {
    const bankRes: any = await request.get('/bank-accounts')
    const accounts: any[] = bankRes?.data?.list || bankRes?.data || []
    const bankAccount = accounts.find((a: any) => !a.is_cash && a.is_active) || accounts[0]
    if (!bankAccount) {
      payments.value = []
      return
    }

    const res: any = await request.get('/bank-transactions', {
      params: { bank_account_id: bankAccount.id, page: 1, page_size: 200 },
    })
    const list: any[] = res?.data ?? res ?? []
    payments.value = list
      .filter((t: any) => !t.matched)
      .map(txn => ({
        id: txn.id,
      counterparty_name: txn.counterparty_name ?? '',
      debit: txn.debit ?? '0',
      credit: txn.credit ?? '0',
      description: txn.description ?? '',
      txn_date: txn.txn_date ?? '',
    }))
  } catch {
    payments.value = []
  } finally {
    loadingPayments.value = false
  }
}

async function loadInvoices() {
  loadingInvoices.value = true
  try {
    const res: any = await request.get('/invoices/unmatched', {
      params: { page_size: 200 },
    })
    const list: any[] = res?.data ?? res ?? []
    invoices.value = list.map(inv => ({
      id: inv.id,
      invoice_no: inv.invoice_no ?? inv.invoice_number ?? '',
      customer_name: inv.customer_name ?? '',
      total_amount: inv.total_amount ?? '0',
      outstanding_amount: inv.outstanding_amount ?? inv.outstanding ?? '0',
      posting_date: inv.posting_date ?? '',
    }))
  } catch {
    invoices.value = []
  } finally {
    loadingInvoices.value = false
  }
}

onMounted(() => {
  loadPayments()
  loadInvoices()
})

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
  } catch (e: any) {
    checks.value = []
    loading.value = false
    ElMessage.error(e?.message || '预检失败')
    return
  }
  checks.value = []
  loading.value = false
}

async function executeReconciliation() {
  if (!selectedPayment.value || !selectedInvoice.value) return
  executing.value = true
  try {
    // 先预检检查是否仍有blocked
    const preRes: any = await request.post('/reconciliation/precheck', {
      payment_id: selectedPayment.value,
      invoice_id: selectedInvoice.value,
    })
    const items = preRes?.data?.checks || preRes?.data
    if (Array.isArray(items) && items.some((c: any) => c.status === 'blocked')) {
      ElMessage.warning('存在阻塞项，请使用强制通过')
      executing.value = false
      return
    }

    // force-pass (即使全部passed也需要创建pair)
    const fpRes: any = await request.post('/reconciliation/precheck/force-pass', {
      payment_id: selectedPayment.value,
      invoice_id: selectedInvoice.value,
      reason: '预检通过，正常核销',
    })
    const pairId = fpRes?.data?.id || fpRes?.id
    if (!pairId) {
      ElMessage.error('创建核销对失败')
      executing.value = false
      return
    }

    // 执行核销
    await request.post('/reconciliation/execute', {
      pair_ids: [pairId],
    })
    ElMessage.success('已提交审核，等待审批')
    precheckDone.value = false
    selectedPayment.value = ''
    selectedInvoice.value = ''
    checks.value = []
    loadPayments()
    loadInvoices()
  } catch (e: any) {
    ElMessage.error(e?.message || '核销失败')
  } finally {
    executing.value = false
  }
}

async function executeForcePass() {
  executing.value = true
  try {
    const fpRes: any = await request.post('/reconciliation/precheck/force-pass', {
      payment_id: selectedPayment.value,
      invoice_id: selectedInvoice.value,
      reason: forcePassReason.value,
    })
    const pairId = fpRes?.data?.id || fpRes?.id
    if (!pairId) {
      ElMessage.error('强制通过失败')
      executing.value = false
      return
    }

    await request.post('/reconciliation/execute', {
      pair_ids: [pairId],
    })
    ElMessage.success('已提交审核，等待审批')
    showForcePassDialog.value = false
    forcePassReason.value = ''
    precheckDone.value = false
    selectedPayment.value = ''
    selectedInvoice.value = ''
    checks.value = []
    loadPayments()
    loadInvoices()
  } catch (e: any) {
    ElMessage.error(e?.message || '强制通过核销失败')
  } finally {
    executing.value = false
  }
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
