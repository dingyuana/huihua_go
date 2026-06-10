<template>
  <div class="manual-match">
    <div class="page-header"><h3>手工核销</h3></div>

    <!-- 选择区 -->
    <el-row :gutter="16" class="selection-row">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>选择收款单</template>
          <el-select
            v-model="selectedPaymentId"
            placeholder="搜索收款单"
            filterable
            style="width: 100%"
            :loading="loadingPayments"
            @change="onPaymentChange"
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
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>可选发票（未核销余额 &gt; 0）</template>
          <el-input
            v-model="invoiceFilter"
            placeholder="搜索发票号"
            size="small"
            clearable
            style="margin-bottom:8px;width:100%"
          />
          <el-table
            ref="invoiceTableRef"
            :data="filteredInvoices"
            size="small"
            border
            max-height="400"
            @row-click="addAllocation"
          >
            <el-table-column prop="invoice_no" label="发票号" width="130" />
            <el-table-column label="客户" width="120">
              <template #default="{ row }">{{ row.customer_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="未结清" width="110" align="right">
              <template #default="{ row }">¥{{ row.outstanding }}</template>
            </el-table-column>
            <el-table-column label="日期" width="90">
              <template #default="{ row }">{{ row.date || '-' }}</template>
            </el-table-column>
            <el-table-column label="可选" width="60">
              <template #default="{ row }"><el-checkbox :model-value="isSelected(row)" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <template #header>收款单详情</template>
          <div v-if="currentPayment" class="payment-detail">
            <p><b>对方：</b>{{ currentPayment.counterparty_name || '-' }}</p>
            <p><b>金额：</b>¥{{ paymentAmount(currentPayment) }}</p>
            <p><b>日期：</b>{{ currentPayment.txn_date || '-' }}</p>
            <p v-if="currentPayment.description"><b>摘要：</b>{{ currentPayment.description }}</p>
          </div>
          <div v-else class="payment-detail empty">请先选择收款单</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 预检结果 -->
    <el-card v-if="precheckItems.length" shadow="never" class="precheck-card">
      <template #header>预检结果</template>
      <el-table :data="precheckItems" size="small" border>
        <el-table-column prop="name" label="检查项" width="150" />
        <el-table-column prop="message" label="结果" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'passed'" size="small" type="success">通过</el-tag>
            <el-tag v-else-if="row.status === 'warning'" size="small" type="warning">警告</el-tag>
            <el-tag v-else size="small" type="danger">阻塞</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 已分配核销 -->
    <el-card shadow="never" class="allocation-card">
      <template #header>已分配核销</template>
      <el-table :data="allocations" size="small" border>
        <el-table-column prop="invoice_no" label="发票号" width="130" />
        <el-table-column prop="customer_name" label="客户" width="120" />
        <el-table-column label="未结清" width="100" align="right">
          <template #default="{ row }">¥{{ row.outstanding }}</template>
        </el-table-column>
        <el-table-column label="本次核销" width="150">
          <template #default="{ row }">
            <el-input-number
              v-model="row.amount"
              :min="0.01"
              :max="parseFloat(row.outstanding)"
              :step="0.01"
              :precision="2"
              controls-position="right"
              style="width:130px"
            />
          </template>
        </el-table-column>
        <el-table-column label="核销后余额" width="110" align="right">
          <template #default="{ row }">
            ¥{{ (parseFloat(row.outstanding) - (parseFloat(row.amount) || 0)).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60">
          <template #default="{ $index }">
            <el-button link type="danger" size="small" @click="allocations.splice($index, 1)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="allocation-summary">
        <span>合计核销: <b>¥{{ totalAmount }}</b></span>
        <span class="remaining">剩余可分配: <b>¥{{ remaining }}</b></span>
      </div>
      <el-button
        type="primary"
        class="exec-btn"
        :disabled="!selectedPaymentId || allocations.length === 0"
        :loading="executing"
        @click="execute"
      >
        执行核销
      </el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

// 收款单
const payments = ref<any[]>([])
const selectedPaymentId = ref('')
const currentPayment = computed(() => payments.value.find(p => p.id === selectedPaymentId.value))
const loadingPayments = ref(false)

// 发票
const invoices = ref<any[]>([])
const invoiceFilter = ref('')
const loadingInvoices = ref(false)

// 预检结果
const precheckItems = ref<any[]>([])

// 分配
interface Allocation {
  invoice_id: string
  invoice_no: string
  customer_name: string
  outstanding: string
  amount: string
}
const allocations = ref<Allocation[]>([])
const executing = ref(false)

function paymentLabel(p: any): string {
  const amt = p.credit && p.credit !== '0' ? p.credit : p.debit
  const party = p.counterparty_name || '无名'
  return `¥${amt} ${party}`
}

function paymentAmount(p: any): string {
  return p.credit && p.credit !== '0' ? p.credit : p.debit
}

const filteredInvoices = computed(() => {
  if (!invoiceFilter.value) return invoices.value
  const q = invoiceFilter.value.toLowerCase()
  return invoices.value.filter((inv: any) =>
    (inv.invoice_no || '').toLowerCase().includes(q)
  )
})

async function loadPayments() {
  loadingPayments.value = true
  try {
    const bankRes: any = await request.get('/bank-accounts')
    const accounts: any[] = bankRes?.data?.list || bankRes?.data || []
    const bankAccount = accounts.find((a: any) => !a.is_cash && a.is_active) || accounts[0]
    if (!bankAccount) { payments.value = []; return }

    const res: any = await request.get('/bank-transactions', {
      params: { bank_account_id: bankAccount.id, page: 1, page_size: 200 },
    })
    const list: any[] = res?.data ?? res ?? []
    payments.value = list.filter((t: any) => !t.matched)
  } catch { payments.value = [] }
  finally { loadingPayments.value = false }
}

async function loadInvoices() {
  loadingInvoices.value = true
  try {
    const res: any = await request.get('/invoices/unmatched', {
      params: { page_size: 500 },
    })
    const list: any[] = res?.data ?? res ?? []
    invoices.value = list.map((inv: any) => ({
      id: inv.id,
      invoice_no: inv.invoice_no ?? inv.invoice_number ?? '',
      customer_name: inv.customer_name ?? '',
      outstanding: inv.outstanding_amount ?? inv.outstanding ?? '0',
      date: inv.posting_date ?? inv.date ?? '',
      total_amount: inv.total_amount ?? '0',
    }))
  } catch { invoices.value = [] }
  finally { loadingInvoices.value = false }
}

onMounted(() => {
  loadPayments()
  loadInvoices()
})

function isSelected(row: any) {
  return allocations.value.some(a => a.invoice_id === row.id)
}

async function onPaymentChange() {
  // 清空分配和预检
  allocations.value = []
  precheckItems.value = []
}

async function addAllocation(row: any) {
  if (isSelected(row)) return
  allocations.value.push({
    invoice_id: row.id,
    invoice_no: row.invoice_no,
    customer_name: row.customer_name || '',
    outstanding: row.outstanding,
    amount: '',
  })
  // 自动预检
  await runPrecheck(row.id)
}

async function runPrecheck(invoiceId: string) {
  if (!selectedPaymentId.value || !invoiceId) return
  try {
    const res: any = await request.post('/reconciliation/precheck', {
      payment_id: selectedPaymentId.value,
      invoice_id: invoiceId,
    })
    const items = res?.data?.checks || res?.data
    if (Array.isArray(items)) {
      precheckItems.value = items
    }
  } catch { /* ignore */ }
}

const totalAmount = computed(() => {
  const sum = allocations.value.reduce((a, b) => a + (parseFloat(b.amount || '0') || 0), 0)
  return sum.toFixed(2)
})

const remaining = computed(() => {
  const amt = currentPayment.value ? parseFloat(paymentAmount(currentPayment.value) || '0') : 0
  return (amt - parseFloat(totalAmount.value)).toFixed(2)
})

async function execute() {
  if (!selectedPaymentId.value || allocations.value.length === 0) return
  executing.value = true
  try {
    await request.post('/reconciliation/manual', {
      bank_transaction_id: selectedPaymentId.value,
      allocations: allocations.value.map(a => ({
        invoice_id: a.invoice_id,
        amount: (parseFloat(a.amount || '0')).toFixed(2),
      })),
    })
    ElMessage.success('核销执行成功！')
    allocations.value = []
    precheckItems.value = []
    loadInvoices()
    loadPayments()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || e?.message || '核销失败')
  } finally {
    executing.value = false
  }
}
</script>

<style scoped>
.page-header h3 { font-size: 18px; margin-bottom: 16px; }
.selection-row { margin-bottom: 16px; }
.payment-detail { font-size: 13px; line-height: 1.8; }
.payment-detail.empty { color: #999; }
.precheck-card { margin-bottom: 16px; }
.allocation-card { }
.allocation-summary { margin-top: 12px; display: flex; gap: 24px; font-size: 14px; }
.remaining { color: #999; }
.exec-btn { margin-top: 12px; }
</style>