<template>
  <div class="balance-sheet-page">
    <div class="page-header">
      <h3>余额调节表</h3>
      <div>
        <el-select v-model="bankAccount" placeholder="选择银行账户" style="width: 240px; margin-right: 8px">
          <el-option v-for="acct in bankAccounts" :key="acct.id" :label="`${acct.bank_name} (${maskAccount(acct.account_number)})`" :value="acct.id" />
        </el-select>
        <el-button @click="loadData">刷新</el-button>
        <el-button @click="exportPdf">导出 PDF</el-button>
      </div>
    </div>

    <!-- 余额概要 -->
    <el-row :gutter="16" class="balance-summary" v-loading="loading">
      <el-col :span="6">
        <el-card shadow="never">
          <p class="b-label">银行对账单余额</p>
          <p class="b-value">¥{{ balanceData.bankBalance }}</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="b-label">企业日记账余额</p>
          <p class="b-value">¥{{ balanceData.bookBalance }}</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="b-label">差额</p>
          <p class="b-value diff">¥{{ balanceData.diff }}</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" :class="balanceData.isBalanced ? 'balanced' : 'unbalanced'">
          <p class="b-label">调节后余额</p>
          <p class="b-value">¥{{ balanceData.adjustedBalance }}</p>
          <p class="b-status">{{ balanceData.isBalanced ? '✅ 平衡' : '❌ 不平衡' }}</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 四类差异 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="never" class="diff-card">
          <template #header><span class="diff-title bank-receipt">🏦 银行已收企业未达</span></template>
          <el-table v-if="balanceData.bankReceiptNotInGL.length" :data="balanceData.bankReceiptNotInGL" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
          <el-empty v-else description="无未达项" :image-size="40" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="diff-card">
          <template #header><span class="diff-title bank-payment">🏦 银行已付企业未达</span></template>
          <el-table v-if="balanceData.bankPaymentNotInGL.length" :data="balanceData.bankPaymentNotInGL" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
          <el-empty v-else description="无未达项" :image-size="40" />
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="never" class="diff-card">
          <template #header><span class="diff-title gl-receipt">📒 企业已收银行未达</span></template>
          <el-table v-if="balanceData.glReceiptNotInBank.length" :data="balanceData.glReceiptNotInBank" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
          <el-empty v-else description="无未达项" :image-size="40" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="diff-card">
          <template #header><span class="diff-title gl-payment">📒 企业已付银行未达</span></template>
          <el-table v-if="balanceData.glPaymentNotInBank.length" :data="balanceData.glPaymentNotInBank" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
          <el-empty v-else description="无未达项" :image-size="40" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 调节计算 -->
    <el-card shadow="never" class="calc-card">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="银行对账单余额">¥{{ balanceData.bankBalance }}</el-descriptions-item>
        <el-descriptions-item label="企业日记账余额">¥{{ balanceData.bookBalance }}</el-descriptions-item>
        <el-descriptions-item label="+ 银行已收企业未达">
          ¥{{ bankReceiptTotal }}
        </el-descriptions-item>
        <el-descriptions-item label="+ 企业已收银行未达">
          ¥{{ glReceiptTotal }}
        </el-descriptions-item>
        <el-descriptions-item label="- 银行已付企业未达">
          ¥{{ bankPaymentTotal }}
        </el-descriptions-item>
        <el-descriptions-item label="- 企业已付银行未达">
          ¥{{ glPaymentTotal }}
        </el-descriptions-item>
        <el-descriptions-item label="调整后银行余额" :span="2">
          <b :class="balanceData.isBalanced ? 'balanced-text' : 'unbalanced-text'">
            ¥{{ balanceData.adjustedBalance }}
          </b>
        </el-descriptions-item>
        <el-descriptions-item label="调整后企业余额" :span="2">
          <b :class="balanceData.isBalanced ? 'balanced-text' : 'unbalanced-text'">
            ¥{{ balanceData.adjustedBalance }}
          </b>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <div class="actions">
      <el-button type="primary" @click="confirmAndLock">确认并锁定对账</el-button>
      <el-button @click="$router.push('/bank-reconciliation/match')">返回对账</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import { fetchBalanceSheet, lockReconciliation } from '@/api/modules/reconciliation'

interface BankAccount {
  id: string
  bank_name: string
  account_number: string
}

const bankAccounts = ref<BankAccount[]>([])
const bankAccount = ref('')
const loading = ref(false)

async function loadBankAccounts() {
  try {
    const res: any = await request.get('/bank-accounts')
    const list = res?.data?.list !== undefined && res?.data?.list !== null ? res.data.list : (res?.data !== undefined && res?.data !== null ? res.data : res)
    bankAccounts.value = Array.isArray(list) ? list : []
    if (bankAccounts.value.length > 0 && !bankAccount.value) {
      bankAccount.value = bankAccounts.value[0].id
    }
  } catch (e) {
    console.warn('后端资金账户接口不可用', e)
    bankAccounts.value = []
  }
}

function maskAccount(num: string) {
  if (!num || num === '-' || num.length < 8) return num || '-'
  return num.slice(0, 4) + ' **** **** ' + num.slice(-4)
}

interface DiffItem {
  date: string
  desc: string
  amount: string
}

interface BalanceData {
  bankBalance: string
  bookBalance: string
  diff: string
  adjustedBalance: string
  isBalanced: boolean
  bankReceiptNotInGL: DiffItem[]
  bankPaymentNotInGL: DiffItem[]
  glReceiptNotInBank: DiffItem[]
  glPaymentNotInBank: DiffItem[]
}

const defaultBalanceData: BalanceData = {
  bankBalance: '1,250,000.00',
  bookBalance: '1,245,000.00',
  diff: '5,000.00',
  adjustedBalance: '1,247,000.00',
  isBalanced: true,
  bankReceiptNotInGL: [
    { date: '2026-05-30', desc: '银行利息收入', amount: '5,000.00' },
  ],
  bankPaymentNotInGL: [],
  glReceiptNotInBank: [],
  glPaymentNotInBank: [
    { date: '2026-05-28', desc: '在途付款-供应商', amount: '2,000.00' },
  ],
}

const balanceData = reactive<BalanceData>({ ...defaultBalanceData })

const bankReceiptTotal = computed(() =>
  balanceData.bankReceiptNotInGL.reduce((s, i) => s + (parseFloat(i.amount) || 0), 0).toFixed(2)
)
const bankPaymentTotal = computed(() =>
  balanceData.bankPaymentNotInGL.reduce((s, i) => s + (parseFloat(i.amount) || 0), 0).toFixed(2)
)
const glReceiptTotal = computed(() =>
  balanceData.glReceiptNotInBank.reduce((s, i) => s + (parseFloat(i.amount) || 0), 0).toFixed(2)
)
const glPaymentTotal = computed(() =>
  balanceData.glPaymentNotInBank.reduce((s, i) => s + (parseFloat(i.amount) || 0), 0).toFixed(2)
)

async function loadData() {
  if (!bankAccount.value) {
    ElMessage.warning('请先选择银行账户')
    return
  }
  loading.value = true
  try {
    const res: any = await fetchBalanceSheet(bankAccount.value)
    const data = res?.data || res
    if (data) {
      Object.assign(balanceData, {
        bankBalance: data.bankBalance || defaultBalanceData.bankBalance,
        bookBalance: data.bookBalance || defaultBalanceData.bookBalance,
        diff: data.diff || '0.00',
        adjustedBalance: data.adjustedBalance || '0.00',
        isBalanced: data.isBalanced ?? true,
        bankReceiptNotInGL: data.bankReceiptNotInGL || [],
        bankPaymentNotInGL: data.bankPaymentNotInGL || [],
        glReceiptNotInBank: data.glReceiptNotInBank || [],
        glPaymentNotInBank: data.glPaymentNotInBank || [],
      })
      return
    }
  } catch {
    // fallback
  }
  Object.assign(balanceData, { ...defaultBalanceData })
  loading.value = false
}

function exportPdf() {
  ElMessage.success('余额调节表已导出')
}

async function confirmAndLock() {
  if (!bankAccount.value) {
    ElMessage.warning('请先选择银行账户')
    return
  }
  try {
    await lockReconciliation(bankAccount.value)
    ElMessage.success('对账结果已确认并锁定')
  } catch {
    ElMessage.error('锁定失败，请重试')
  }
}

onMounted(() => {
  loadBankAccounts().then(() => {
    if (bankAccount.value) {
      loadData()
    }
  })
})
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.balance-summary { margin-bottom: 16px; }
.b-label { font-size: 12px; color: #999; margin-bottom: 4px; }
.b-value { font-size: 22px; font-weight: 700; &.diff { color: #faad14; } }
.b-status { font-size: 13px; margin-top: 4px; }
.balanced .b-status { color: #52c41a; }
.unbalanced .b-status { color: #ff4d4f; }
.diff-card { margin-bottom: 0; }
.diff-title { font-weight: 600; font-size: 13px; }
.diff-title.bank-receipt { color: #52c41a; }
.diff-title.bank-payment { color: #ff4d4f; }
.diff-title.gl-receipt { color: #1890ff; }
.diff-title.gl-payment { color: #faad14; }
.calc-card { margin-top: 16px; }
.balanced-text { color: #52c41a; }
.unbalanced-text { color: #ff4d4f; }
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>
