<template>
  <div class="pending-confirm-page">
    <div class="page-header">
      <h3>银企对账待确认</h3>
      <div>
        <el-select v-model="bankAccount" placeholder="选择银行账户" style="width: 240px; margin-right: 8px">
          <el-option v-for="acct in bankAccounts" :key="acct.id" :label="`${acct.bank_name} (${maskAccount(acct.account_number)})`" :value="acct.id" />
        </el-select>
        <el-select v-model="periodNo" placeholder="选择会计期间" style="width: 160px; margin-right: 8px">
          <el-option v-for="p in periods" :key="p.period_no" :label="p.period_label" :value="p.period_no" />
        </el-select>
        <el-button type="primary" :loading="loading" @click="loadPending">查询</el-button>
      </div>
    </div>

    <!-- 统计 -->
    <el-row :gutter="16" class="stat-row" v-loading="loading">
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val">{{ stats.total }}</p>
          <p class="stat-lbl">候选对总数</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val warning">{{ stats.highScore }}</p>
          <p class="stat-lbl">高分候选对 (80-85)</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val info">{{ stats.lowScore }}</p>
          <p class="stat-lbl">低分候选对 (60-80)</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val">{{ stats.confirmed }}</p>
          <p class="stat-lbl">已确认</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 候选对列表 -->
    <el-card shadow="never">
      <el-table :data="pendingItems" border stripe size="small" v-loading="loading" empty-text="暂无待确认候选对">
        <el-table-column label="银行流水" prop="bankTxnDesc" min-width="200" show-overflow-tooltip />
        <el-table-column label="银行金额" prop="bankTxnAmt" width="120" align="right">
          <template #default="{ row }">¥{{ row.bankTxnAmt }}</template>
        </el-table-column>
        <el-table-column label="银行日期" prop="bankTxnDate" width="100" />
        <el-table-column label="GL分录" prop="glEntryDesc" min-width="200" show-overflow-tooltip />
        <el-table-column label="得分" prop="score.total_score" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.score.total_score >= 80 ? 'success' : 'warning'" size="small">
              {{ row.score.total_score.toFixed(1) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              :loading="confirmingId === row.bank_txn_id + row.gl_entry_id"
              @click="handleConfirm(row)"
            >
              确认勾兑
            </el-button>
            <el-button
              type="danger"
              size="small"
              :loading="rejectingId === row.bank_txn_id + row.gl_entry_id"
              @click="handleReject(row)"
            >
              拒绝
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 16px; justify-content: flex-end"
        @current-change="loadPending"
      />
    </el-card>

    <div class="page-actions">
      <el-button @click="$router.push('/bank-reconciliation/match')">返回对账</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import {
  fetchPendingConfirm,
  confirmMatchCandidate,
  rejectMatchCandidate,
} from '@/api/modules/reconciliation'
import type { PendingConfirmItem } from '@/api/modules/reconciliation'

interface BankAccount {
  id: string
  bank_name: string
  account_number: string
}

interface Period {
  period_no: number
  period_label: string
}

const bankAccounts = ref<BankAccount[]>([])
const bankAccount = ref('')
const periods = ref<Period[]>([])
const periodNo = ref<number | undefined>(undefined)
const loading = ref(false)
const pendingItems = ref<PendingConfirmItem[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const stats = ref({ total: 0, highScore: 0, lowScore: 0, confirmed: 0 })
const confirmingId = ref('')
const rejectingId = ref('')

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

async function loadPeriods() {
  try {
    const res: any = await request.get('/periods')
    const list = res?.data?.list !== undefined && res?.data?.list !== null ? res.data.list : (res?.data !== undefined && res?.data !== null ? res.data : res)
    periods.value = Array.isArray(list) ? list : []
    if (periods.value.length > 0) {
      const current = periods.value.find((p: any) => p.is_current) || periods.value[0]
      periodNo.value = current.period_no
    }
  } catch (e) {
    console.warn('后端期间接口不可用', e)
    periods.value = []
  }
}

function maskAccount(num: string) {
  if (!num || num === '-' || num.length < 8) return num || '-'
  return num.slice(0, 4) + ' **** **** ' + num.slice(-4)
}

async function loadPending() {
  if (!bankAccount.value) {
    ElMessage.warning('请先选择银行账户')
    return
  }
  loading.value = true
  try {
    const res: any = await fetchPendingConfirm(bankAccount.value, periodNo.value)
    const items: PendingConfirmItem[] = res?.data || res || []
    pendingItems.value = items
    total.value = items.length
    stats.value = {
      total: items.length,
      highScore: items.filter(i => i.score.total_score >= 80).length,
      lowScore: items.filter(i => i.score.total_score >= 60 && i.score.total_score < 80).length,
      confirmed: 0,
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载待确认数据失败')
  } finally {
    loading.value = false
  }
}

async function handleConfirm(row: PendingConfirmItem) {
  const key = row.bank_txn_id + row.gl_entry_id
  confirmingId.value = key
  try {
    await confirmMatchCandidate(row.bank_txn_id, row.gl_entry_id)
    ElMessage.success('勾兑已确认')
    pendingItems.value = pendingItems.value.filter(
      i => !(i.bank_txn_id === row.bank_txn_id && i.gl_entry_id === row.gl_entry_id)
    )
    total.value = Math.max(0, total.value - 1)
    stats.value.confirmed++
    stats.value.total = Math.max(0, stats.value.total - 1)
  } catch (e: any) {
    ElMessage.error(e?.message || '确认失败')
  } finally {
    confirmingId.value = ''
  }
}

async function handleReject(row: PendingConfirmItem) {
  const key = row.bank_txn_id + row.gl_entry_id
  rejectingId.value = key
  try {
    await rejectMatchCandidate(row.bank_txn_id, row.gl_entry_id)
    ElMessage.success('已拒绝')
    pendingItems.value = pendingItems.value.filter(
      i => !(i.bank_txn_id === row.bank_txn_id && i.gl_entry_id === row.gl_entry_id)
    )
    total.value = Math.max(0, total.value - 1)
    stats.value.total = Math.max(0, stats.value.total - 1)
  } catch (e: any) {
    ElMessage.error(e?.message || '拒绝失败')
  } finally {
    rejectingId.value = ''
  }
}

onMounted(() => {
  loadBankAccounts().then(() => loadPeriods()).then(() => loadPending())
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
.stat-row { margin-bottom: 16px; }
.stat-val {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 4px;
  &.warning { color: #faad14; }
  &.info { color: #1890ff; }
}
.stat-lbl { font-size: 13px; color: #999; }
.page-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>