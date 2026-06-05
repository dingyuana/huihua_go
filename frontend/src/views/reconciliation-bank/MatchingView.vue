<template>
  <div class="bank-rec">
    <div class="page-header">
      <h3>银企对账</h3>
      <div>
        <el-select v-model="bankAccount" placeholder="选择银行账户" style="width: 240px; margin-right: 8px">
          <el-option v-for="acct in bankAccounts" :key="acct.id" :label="`${acct.bank_name} (${maskAccount(acct.account_number)})`" :value="acct.id" />
        </el-select>
        <el-select v-model="periodNo" placeholder="会计期间" style="width: 160px; margin-right: 8px">
          <el-option v-for="p in periods" :key="p.period_no" :label="p.period_label" :value="p.period_no" />
        </el-select>
        <el-button type="primary" :loading="loading" @click="runMatch">执行对账</el-button>
      </div>
    </div>

    <!-- 锁定状态提示 -->
    <el-alert
      v-if="locked"
      :title="`对账已锁定（锁定人：${lockedBy || '未知'}，锁定时间：${lockedAt || '未知'}）`"
      type="warning"
      show-icon
      closable
      class="mb-16"
    />

    <!-- 统计概览 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val">{{ stats.total }}</p>
          <p class="stat-lbl">总笔数</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val success">{{ stats.autoMatched }}</p>
          <p class="stat-lbl">自动勾兑</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val warning">{{ stats.needConfirm }}</p>
          <p class="stat-lbl">待确认</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="stat-val danger">{{ stats.unmatched }}</p>
          <p class="stat-lbl">未匹配</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 自动勾兑率 -->
    <el-card shadow="never" class="rate-card">
      <span>自动勾兑率：</span>
      <el-progress
        :percentage="stats.autoMatchRate"
        :color="stats.autoMatchRate > 90 ? '#52c41a' : stats.autoMatchRate > 80 ? '#faad14' : '#ff4d4f'"
        :stroke-width="16"
      />
    </el-card>

    <!-- 匹配列表 -->
    <el-card shadow="never">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="`自动匹配 (${stats.autoMatched})`" name="auto" />
        <el-tab-pane :label="`待确认 (${stats.needConfirm})`" name="confirm" />
        <el-tab-pane :label="`未匹配 (${stats.unmatched})`" name="unmatched" />
      </el-tabs>

      <el-table :data="filteredList" border stripe size="small" v-loading="loading">
        <el-table-column label="匹配得分" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.score >= 85 ? 'success' : row.score >= 60 ? 'warning' : 'danger'"
              size="small"
            >
              {{ row.score }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bank_txn" label="银行流水" min-width="200" />
        <el-table-column prop="gl_entry" label="GL 条目" min-width="200" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button
              v-if="row.needConfirm"
              size="small"
              type="primary"
              :loading="confirmingId === row.id"
              @click="handleConfirm(row)"
            >
              确认
            </el-button>
            <el-tag v-else size="small" type="success">已匹配</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="rec-actions">
      <el-button @click="$router.push('/bank-reconciliation/balance')">查看余额调节表</el-button>
      <el-button @click="$router.push('/bank-reconciliation/pending-confirm')">待确认列表</el-button>
      <el-button
        v-if="!locked"
        type="primary"
        :loading="locking"
        @click="handleLock"
      >
        锁定对账结果
      </el-button>
      <el-button
        v-else
        type="info"
        disabled
      >
        已锁定
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import {
  runBankReconciliation,
  confirmMatch,
  lockReconciliation,
  unlockReconciliation,
  fetchReconciliationStatus,
} from '@/api/modules/reconciliation'
import type { MatchItem } from '@/api/modules/reconciliation'

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
const activeTab = ref('auto')
const loading = ref(false)
const confirmingId = ref('')
const locking = ref(false)
const locked = ref(false)
const lockedBy = ref('')
const lockedAt = ref('')

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

const stats = ref({ total: 0, autoMatched: 0, needConfirm: 0, unmatched: 0, autoMatchRate: 0 })
const matchList = ref<MatchItem[]>([])

const filteredList = computed(() => {
  if (activeTab.value === 'auto') return matchList.value.filter(m => !m.needConfirm && m.score >= 60)
  if (activeTab.value === 'confirm') return matchList.value.filter(m => m.needConfirm)
  return matchList.value.filter(m => m.score < 60 && !m.needConfirm)
})

async function loadStatus() {
  if (!bankAccount.value) return
  try {
    const res: any = await fetchReconciliationStatus(bankAccount.value, periodNo.value)
    const data = res?.data || res
    if (data) {
      locked.value = data.locked ?? false
      lockedBy.value = data.locked_by || ''
      lockedAt.value = data.locked_at || ''
    }
  } catch {
    locked.value = false
  }
}

async function runMatch() {
  if (!bankAccount.value) {
    ElMessage.warning('请先选择银行账户')
    return
  }
  loading.value = true
  try {
    const res: any = await runBankReconciliation(bankAccount.value)
    const data = res?.data || res
    if (data) {
      stats.value = {
        total: data.total || data.matched_count || 0,
        autoMatched: data.autoMatched || data.matched_count || 0,
        needConfirm: data.needConfirm || 0,
        unmatched: data.unmatched || 0,
        autoMatchRate: data.autoMatchRate || 0,
      }
      if (data.matches) matchList.value = data.matches
    }
  } catch {
    // keep empty state
  }
  loading.value = false
}

async function handleConfirm(row: MatchItem) {
  confirmingId.value = row.id
  try {
    await confirmMatch(row.id)
    row.needConfirm = false
    stats.value.needConfirm = Math.max(0, stats.value.needConfirm - 1)
    stats.value.autoMatched += 1
    stats.value.autoMatchRate = Math.round(stats.value.autoMatched / stats.value.total * 100)
    ElMessage.success('匹配已确认')
  } catch {
    ElMessage.error('确认失败')
  }
  confirmingId.value = ''
}

async function handleLock() {
  if (!bankAccount.value) {
    ElMessage.warning('请先选择银行账户')
    return
  }
  locking.value = true
  try {
    await lockReconciliation(bankAccount.value, periodNo.value)
    ElMessage.success('对账结果已锁定')
    locked.value = true
    lockedBy.value = 'current_user'
    lockedAt.value = new Date().toLocaleString()
  } catch {
    ElMessage.error('锁定失败')
  }
  locking.value = false
}

onMounted(() => {
  loadBankAccounts().then(() => loadPeriods()).then(() => {
    loadStatus()
    runMatch()
  })
})
</script>

<style scoped lang="scss">
.mb-16 { margin-bottom: 16px; }
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
  &.success { color: #52c41a; }
  &.warning { color: #faad14; }
  &.danger { color: #ff4d4f; }
}
.stat-lbl { font-size: 13px; color: #999; }
.rate-card {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.rec-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>