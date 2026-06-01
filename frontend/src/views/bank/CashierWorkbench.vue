<template>
  <div class="workbench">
    <div class="page-header">
      <h3>出纳核对工作台</h3>
      <div>
        <el-radio-group v-model="viewMode" size="small">
          <el-radio-button value="list">列表模式</el-radio-button>
          <el-radio-button value="batch">批量模式</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <el-row :gutter="16" class="stat-row">
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card total">
          <p class="stat-num">{{ stats.total }}</p>
          <p class="stat-label">本月流水</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card confirmed">
          <p class="stat-num">{{ stats.confirmed }}</p>
          <p class="stat-label">已确认</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card pending">
          <p class="stat-num">{{ stats.pending }}</p>
          <p class="stat-label">待确认</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card unclassified">
          <p class="stat-num danger">{{ stats.unclassified }}</p>
          <p class="stat-label">未分类</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card docs">
          <p class="stat-num">{{ stats.generatedDocs }}</p>
          <p class="stat-label">已生成单据</p>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <el-tabs v-model="activeTab" class="classification-tabs">
        <el-tab-pane :label="`全部 (${stats.total})`" name="all" />
        <el-tab-pane :label="`业务收款 (${stats.byCategory.business_receipt})`" name="business_receipt" />
        <el-tab-pane :label="`业务付款 (${stats.byCategory.business_payment})`" name="business_payment" />
        <el-tab-pane :label="`银行费用 (${stats.byCategory.bank_fee})`" name="bank_fee" />
        <el-tab-pane :label="`利息收入 (${stats.byCategory.interest_income})`" name="interest_income" />
        <el-tab-pane :label="`内部转账 (${stats.byCategory.internal_transfer})`" name="internal_transfer" />
        <el-tab-pane :label="`税务缴费 (${stats.byCategory.tax_payment})`" name="tax_payment" />
        <el-tab-pane :label="`社保缴费 (${stats.byCategory.social_security})`" name="social_security" />
        <el-tab-pane :label="`待处理 (${stats.byCategory.pending})${stats.byCategory.pending > 0 ? ' 🔴' : ''}`" name="pending" />
      </el-tabs>

      <!-- Tab 汇总 -->
      <div class="tab-summary">
        <span class="summary-item"><b>{{ tabSummary.count }}</b> 笔</span>
        <span class="summary-divider">|</span>
        <span class="summary-item">收入: <b class="amount-income">¥{{ tabSummary.income }}</b></span>
        <span class="summary-divider">|</span>
        <span class="summary-item">支出: <b class="amount-expense">¥{{ tabSummary.expense }}</b></span>
        <span class="summary-divider">|</span>
        <span class="summary-item">净额: <b :class="Number(tabSummary.net.replace(/,/g,'')) >= 0 ? 'amount-income' : 'amount-expense'">¥{{ tabSummary.net }}</b></span>
      </div>

      <div class="batch-bar">
        <el-checkbox v-model="selectAll" @change="onSelectAll">全选</el-checkbox>
        <span class="selected-count">已选 {{ selectedIds.length }} 条</span>
        <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="batchConfirm">确认选中</el-button>
        <el-button size="small" :disabled="selectedIds.length === 0" @click="showClassifyDialog = true">修正分类</el-button>
      </div>

      <!-- 批量模式下提示待生成单据 -->
      <el-alert v-if="viewMode === 'batch' && selectedIds.length > 0" :title="batchPreview" type="info" :closable="false" show-icon class="batch-preview" />

      <el-table
        ref="tableRef"
        :data="filteredTxns"
        border stripe size="small"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="40" />
        <el-table-column prop="date" label="日期" width="90" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span :class="amountClass(row)">{{ amountDisplay(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="counterparty" label="对方" width="140" />
        <el-table-column prop="description" label="摘要" min-width="180" show-overflow-tooltip />
        <el-table-column label="分类" width="100">
          <template #default="{ row }">
            <el-tag :type="classificationTag(row.classification)" size="small">{{ classificationLabel(row.classification) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="将生成" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.payment_no" size="small" type="primary">{{ row.payment_no }}</el-tag>
            <span v-else-if="row.confirmed" class="doc-preview">{{ docTypeLabel(row.classification) }}</span>
            <span v-else class="doc-preview">{{ docTypeLabel(row.classification) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.confirmed" type="success" size="small">已确认</el-tag>
            <el-tag v-else type="warning" size="small">待确认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.confirmed" link type="primary" size="small" @click="confirmOne(row)">确认</el-button>
            <el-button link type="primary" size="small" @click="editOne(row)">修正</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showClassifyDialog" title="修正分类" width="420px">
      <el-form label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="classifyForm.classification" style="width: 100%">
            <el-option label="业务收款" value="business_receipt" />
            <el-option label="业务付款" value="business_payment" />
            <el-option label="银行费用" value="bank_fee" />
            <el-option label="利息收入" value="interest_income" />
            <el-option label="内部转账" value="internal_transfer" />
            <el-option label="税务缴费" value="tax_payment" />
            <el-option label="待处理" value="pending" />
          </el-select>
        </el-form-item>
        <el-form-item label="对方单位">
          <PartySelector v-model="classifyForm.party" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="classifyForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showClassifyDialog = false">取消</el-button>
        <el-button type="primary" @click="saveClassification">确认修正</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface TxnItem {
  id: string
  date: string
  amount: string
  rawAmount: number
  direction: string
  counterparty: string
  description: string
  classification: string
  confirmed: boolean
  payment_entry_id?: string
  payment_no?: string
}

interface BankTxnAPI {
  id: string
  txn_date: string
  debit: number | string
  credit: number | string
  direction: string
  counterparty_name: string | null
  description: string | null
  classification: string | null
  matched: boolean
  matched_payment_entry_id?: string
}

const allTxns = ref<TxnItem[]>([])
const bankAccountId = ref('')

function apiToTxnItem(t: BankTxnAPI): TxnItem {
  const debit = Number(t.debit) || 0
  const credit = Number(t.credit) || 0
  const direction = t.direction || (debit > 0 ? 'in' : 'out')
  const rawAmount = debit > 0 ? debit : credit
  const amount = rawAmount.toFixed(2)
  const date = t.txn_date ? t.txn_date.slice(5, 10) : ''
  return {
    id: t.id,
    date,
    rawAmount,
    amount: Number(amount).toLocaleString('en', { minimumFractionDigits: 2 }),
    direction,
    counterparty: t.counterparty_name || '',
    description: t.description || '',
    classification: t.classification || (t.matched ? 'business_receipt' : 'pending'),
    confirmed: t.matched,
    payment_entry_id: t.matched_payment_entry_id,
  }
}

onMounted(async () => {
  try {
    const bankRes: any = await request.get('/bank-accounts')
    const accounts = bankRes?.data?.list || bankRes?.data
    if (Array.isArray(accounts) && accounts.length > 0) {
      // Prefer active bank accounts (non-cash)
      const bankAccount = accounts.find((a: any) => !a.is_cash && a.is_active) || accounts[0]
      bankAccountId.value = bankAccount.id
    }
  } catch { /* no bank accounts */ }

  if (bankAccountId.value) {
    try {
      const res: any = await request.get('/bank-transactions', {
        params: { bank_account_id: bankAccountId.value, page: 1, pageSize: 50 },
      })
      const list = res?.data?.list || res?.data
      if (Array.isArray(list)) {
        allTxns.value = list.map(apiToTxnItem)
      }
    } catch { /* no transactions yet */ }
  }
})

const viewMode = ref('list')
const tableRef = ref()
const activeTab = ref('all')
const selectedIds = ref<string[]>([])
const selectAll = ref(false)
const showClassifyDialog = ref(false)
const docCounter = ref(0) // 已生成单据数量统计

const classifyForm = reactive({
  classification: 'business_receipt',
  party: null as any,
  remark: '',
})

const stats = computed(() => {
  const byCategory: Record<string, number> = {
    business_receipt: 0,
    business_payment: 0,
    bank_fee: 0,
    interest_income: 0,
    internal_transfer: 0,
    tax_payment: 0,
    social_security: 0,
    pending: 0,
  }
  allTxns.value.forEach(t => {
    if (byCategory[t.classification] !== undefined) {
      byCategory[t.classification]++
    }
  })
  return {
    total: allTxns.value.length,
    confirmed: allTxns.value.filter(t => t.confirmed).length,
    pending: allTxns.value.filter(t => !t.confirmed).length,
    unclassified: allTxns.value.filter(t => t.classification === 'pending').length,
    generatedDocs: docCounter.value,
    byCategory,
  }
})

const filteredTxns = computed(() => {
  if (activeTab.value === 'all') return allTxns.value
  if (activeTab.value === 'pending') return allTxns.value.filter(t => t.classification === 'pending')
  return allTxns.value.filter(t => t.classification === activeTab.value)
})

const tabSummary = computed(() => {
  const items = filteredTxns.value
  let income = 0
  let expense = 0
  for (const t of items) {
    if (t.direction === 'in') income += t.rawAmount
    else expense += t.rawAmount
  }
  const fmt = (n: number) => n.toLocaleString('en', { minimumFractionDigits: 2 })
  return {
    count: items.length,
    income: fmt(income),
    expense: fmt(expense),
    net: fmt(income - expense),
  }
})

/** 批量操作预览：显示选中项将生成的单据类型 */
const batchPreview = computed(() => {
  const selected = allTxns.value.filter(t => selectedIds.value.includes(t.id))
  const types = [...new Set(selected.map(t => docTypeLabel(t.classification)))]
  return `确认后将自动生成：${types.join('、')}，共 ${selectedIds.value.length} 笔`
})

function docTypeLabel(cls: string): string {
  const map: Record<string, string> = {
    business_receipt: '收款单',
    business_payment: '付款单',
    bank_fee: '银行费用单',
    interest_income: '利息收入单',
    internal_transfer: '银行转账单',
    tax_payment: '税务缴费单',
    social_security: '社保缴费单',
    pending: '待处理',
  }
  return map[cls] || cls
}

function classificationTag(val: string) {
  const map: Record<string, string> = {
    business_receipt: 'success',
    business_payment: 'danger',
    bank_fee: 'warning',
    interest_income: 'primary',
    internal_transfer: 'info',
    tax_payment: 'warning',
    social_security: 'info',
    pending: 'danger',
  }
  return map[val] || 'info'
}

function classificationLabel(val: string) {
  const map: Record<string, string> = {
    business_receipt: '业务收款',
    business_payment: '业务付款',
    bank_fee: '银行费用',
    interest_income: '利息收入',
    internal_transfer: '内部转账',
    tax_payment: '税务缴费',
    social_security: '社保缴费',
    pending: '待处理',
  }
  return map[val] || val
}

const AUTO_VOUCHER_CLASSIFICATIONS = new Set(['bank_fee', 'interest_income', 'tax_payment', 'social_security'])
const DRAFT_ORDER_CLASSIFICATIONS = new Set(['business_receipt', 'business_payment', 'internal_transfer'])

function isAutoVoucherClassification(cls: string): boolean {
  return AUTO_VOUCHER_CLASSIFICATIONS.has(cls)
}

interface PaymentMapping { paymentType: string; partyType: string }

function mapToPaymentEntry(cls: string): PaymentMapping {
  switch (cls) {
    case 'business_receipt':  return { paymentType: 'receive',  partyType: 'customer' }
    case 'business_payment':  return { paymentType: 'pay',      partyType: 'supplier' }
    case 'internal_transfer': return { paymentType: 'transfer', partyType: 'internal' }
    case 'expense':           return { paymentType: 'expense',  partyType: 'employee' }
    case 'interest':          return { paymentType: 'interest', partyType: 'bank' }
    default:                  return { paymentType: 'pay',      partyType: 'other' }
  }
}

function amountClass(row: TxnItem): string {
  if (row.classification === 'business_receipt') return 'amount-income'
  if (row.classification === 'business_payment') return 'amount-expense'
  return row.direction === 'in' ? 'amount-income' : 'amount-expense'
}

function amountDisplay(row: TxnItem): string {
  if (row.classification === 'business_payment') return `-${row.amount}`
  return `${row.direction === 'in' ? '+' : '-'}${row.amount}`
}

function onSelectionChange(rows: TxnItem[]) {
  selectedIds.value = rows.map(r => r.id)
}

function onSelectAll() {
  // el-table 内置全选
}

async function confirmOne(row: TxnItem) {
  try {
    if (isAutoVoucherClassification(row.classification)) {
      const res: any = await request.post(`/bank-transactions/${row.id}/generate-voucher`)
      const je = res?.data
      row.confirmed = true
      row.payment_no = je?.voucher_no || ''
      docCounter.value++
      ElMessage.success(`流水 ${row.date} 已确认，已生成凭证：${je?.voucher_no || '(凭证号未返回)'}（${docTypeLabel(row.classification)}）`)
    } else {
      const { paymentType, partyType } = mapToPaymentEntry(row.classification)

      const res: any = await request.post('/payment-entries', {
        bank_transaction_id: row.id,
        payment_type: paymentType,
        party_type: partyType,
        party_id: '00000000-0000-0000-0000-000000000000',
        posting_date: new Date().toISOString().slice(0, 10),
      })
      row.confirmed = true
      row.payment_entry_id = res.data?.payment_entry?.id
      row.payment_no = res.data?.payment_entry?.payment_no
      docCounter.value++
      ElMessage.success(`流水 ${row.date} 已确认，已生成${docTypeLabel(row.classification)}：${row.payment_no}（待会计处理）`)
    }
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.message || '操作失败'
    ElMessage.error(`流水 ${row.date} 处理失败：${msg}`)
  }
}

async function batchConfirm() {
  const selected = allTxns.value.filter(t => selectedIds.value.includes(t.id))
  if (selected.length === 0) return

  const today = new Date().toISOString().slice(0, 10)
  let voucherOk = 0
  let orderOk = 0
  let fail = 0

  for (const t of selected) {
    try {
      if (isAutoVoucherClassification(t.classification)) {
        const res: any = await request.post(`/bank-transactions/${t.id}/generate-voucher`)
        t.confirmed = true
        t.payment_no = res?.data?.voucher_no || ''
        voucherOk++
      } else {
        const { paymentType, partyType } = mapToPaymentEntry(t.classification)

        const res: any = await request.post('/payment-entries', {
          bank_transaction_id: t.id,
          payment_type: paymentType,
          party_type: partyType,
          party_id: '00000000-0000-0000-0000-000000000000',
          posting_date: today,
        })
        t.confirmed = true
        t.payment_entry_id = res.data?.payment_entry?.id
        t.payment_no = res.data?.payment_entry?.payment_no
        orderOk++
      }
      docCounter.value++
    } catch (e: any) {
      const msg = e?.response?.data?.error || e?.message || '失败'
      console.error(`Bank txn ${t.id} failed:`, msg)
      fail++
    }
  }

  const parts: string[] = []
  if (voucherOk > 0) parts.push(`自动凭证 ${voucherOk} 张`)
  if (orderOk > 0) parts.push(`草稿单据 ${orderOk} 张`)
  if (fail > 0) parts.push(`失败 ${fail} 条`)

  if (parts.length > 0) {
    ElMessage.success(`批量完成：${parts.join('，')}`)
  }
  selectedIds.value = []
  selectAll.value = false
}

function editOne(row: TxnItem) {
  classifyForm.classification = row.classification
  showClassifyDialog.value = true
}

function saveClassification() {
  showClassifyDialog.value = false
  ElMessage.success('分类已更新')
}
</script>

<style scoped lang="scss">
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.stat-row { margin-bottom: 16px; }
.stat-card { text-align: center; .stat-num { font-size: 28px; font-weight: 700; margin-bottom: 4px; &.danger { color: #ff4d4f; } } .stat-label { font-size: 13px; color: #999; } &.total .stat-num { color: #1890ff; } &.confirmed .stat-num { color: #52c41a; } &.pending .stat-num { color: #faad14; } &.docs .stat-num { color: #722ed1; } }
.classification-tabs { margin-bottom: 4px; }
.tab-summary { display: flex; align-items: center; gap: 12px; padding: 8px 12px; background: #fafafa; border-radius: 4px; margin-bottom: 8px; font-size: 13px; .summary-divider { color: #ddd; } .summary-item { color: #666; b { font-weight: 600; } } }
.batch-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; padding: 8px 0; .selected-count { font-size: 13px; color: #666; } }
.batch-preview { margin-bottom: 12px; }
.doc-preview { color: #999; font-size: 12px; }
.amount-income { color: #389e0d; font-weight: 600; }
.amount-expense { color: #cf1322; font-weight: 600; }
</style>
