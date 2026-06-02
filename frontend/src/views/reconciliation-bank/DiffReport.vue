<template>
  <div class="diff-report-page">
    <div class="page-header">
      <h3>对账差异报告</h3>
      <div class="filter-bar">
        <el-select
          v-model="bankAccountId"
          placeholder="选择银行账户"
          style="width: 240px"
          filterable
        >
          <el-option
            v-for="acct in bankAccounts"
            :key="acct.id"
            :label="`${acct.bank_name} (${maskAccount(acct.account_number)})`"
            :value="acct.id"
          />
        </el-select>
        <el-input
          v-model="periodNo"
          placeholder="期间 YYYYMM"
          style="width: 140px; margin-left: 8px"
        />
        <el-button
          type="primary"
          :loading="loading"
          style="margin-left: 8px"
          @click="loadReport"
        >
          生成报告
        </el-button>
        <el-button
          :disabled="!report"
          @click="exportCsv"
        >
          导出 CSV
        </el-button>
      </div>
    </div>

    <el-row
      v-if="report"
      :gutter="16"
      class="summary"
    >
      <el-col :span="6">
        <el-card shadow="never">
          <p class="lbl">
            银行对账单余额
          </p>
          <p class="val">
            ¥{{ report.bank_balance }}
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="lbl">
            企业日记账余额
          </p>
          <p class="val">
            ¥{{ report.book_balance }}
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <p class="lbl">
            差额
          </p>
          <p
            class="val"
            :class="Number(report.difference) === 0 ? 'ok' : 'err'"
          >
            ¥{{ report.difference }}
          </p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card
          shadow="never"
          :class="report.adjusted_reconciled ? 'ok-bg' : 'err-bg'"
        >
          <p class="lbl">
            调节后状态
          </p>
          <p class="val">
            {{ report.adjusted_reconciled ? '✅ 已平衡' : '❌ 不平衡' }}
          </p>
          <p class="ts">
            生成时间: {{ report.generated_at }}
          </p>
        </el-card>
      </el-col>
    </el-row>

    <el-row
      v-if="report"
      :gutter="16"
      style="margin-top: 16px"
    >
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <span class="card-title bank">🏦 银行有，企业无（{{ report.bank_only_count }} 笔，¥{{ report.bank_only_total }}）</span>
          </template>
          <el-table
            v-if="report.bank_only_items.length"
            :data="report.bank_only_items"
            size="small"
            border
            stripe
          >
            <el-table-column
              prop="txn_date"
              label="日期"
              width="100"
            />
            <el-table-column
              prop="description"
              label="摘要"
              min-width="180"
              show-overflow-tooltip
            />
            <el-table-column
              prop="amount"
              label="金额"
              width="100"
              align="right"
            />
            <el-table-column
              prop="direction"
              label="方向"
              width="60"
            >
              <template #default="{ row }">
                <el-tag
                  :type="row.direction === 'in' ? 'success' : 'danger'"
                  size="small"
                >
                  {{ row.direction === 'in' ? '收' : '付' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="source_type"
              label="来源"
              width="80"
            />
          </el-table>
          <el-empty
            v-else
            description="无未达项"
            :image-size="48"
          />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <span class="card-title book">📒 企业有，银行无（{{ report.book_only_count }} 笔，¥{{ report.book_only_total }}）</span>
          </template>
          <el-table
            v-if="report.book_only_items.length"
            :data="report.book_only_items"
            size="small"
            border
            stripe
          >
            <el-table-column
              prop="txn_date"
              label="日期"
              width="100"
            />
            <el-table-column
              prop="description"
              label="摘要"
              min-width="180"
              show-overflow-tooltip
            />
            <el-table-column
              prop="amount"
              label="金额"
              width="100"
              align="right"
            />
            <el-table-column
              prop="direction"
              label="方向"
              width="60"
            >
              <template #default="{ row }">
                <el-tag
                  :type="row.direction === 'in' ? 'success' : 'danger'"
                  size="small"
                >
                  {{ row.direction === 'in' ? '收' : '付' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="source_type"
              label="来源"
              width="80"
            />
          </el-table>
          <el-empty
            v-else
            description="无未达项"
            :image-size="48"
          />
        </el-card>
      </el-col>
    </el-row>

    <el-empty
      v-if="!report && !loading"
      description="请选择银行账户和期间后点击「生成报告」"
      :image-size="80"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface BankAccountItem {
  id: string
  bank_name: string
  account_number: string
}

interface DiffItem {
  source_type: string
  txn_date: string
  description: string
  amount: string
  direction: string
}

interface DiffReport {
  bank_account_id: string
  bank_name: string
  period_no: number
  bank_balance: string
  book_balance: string
  difference: string
  bank_only_items: DiffItem[]
  book_only_items: DiffItem[]
  bank_only_total: string
  book_only_total: string
  bank_only_count: number
  book_only_count: number
  adjusted_reconciled: boolean
  generated_at: string
}

const bankAccounts = ref<BankAccountItem[]>([])
const bankAccountId = ref('')
const periodNo = ref(new Date().toISOString().slice(0, 7).replace('-', ''))
const report = ref<DiffReport | null>(null)
const loading = ref(false)

function maskAccount(num: string) {
  if (!num || num.length < 8) return num
  return num.slice(0, 4) + ' **** **** ' + num.slice(-4)
}

async function loadBankAccounts() {
  try {
    const res: any = await request.get('/bank-accounts')
    const list = res?.data?.list !== undefined ? res.data.list : (res?.data ?? [])
    bankAccounts.value = Array.isArray(list) ? list : []
    if (bankAccounts.value.length > 0 && !bankAccountId.value) {
      bankAccountId.value = bankAccounts.value[0].id
    }
  } catch (e) {
    ElMessage.error('加载银行账户失败')
  }
}

async function loadReport() {
  if (!bankAccountId.value) {
    ElMessage.warning('请选择银行账户')
    return
  }
  if (!periodNo.value || !/^\d{6}$/.test(periodNo.value)) {
    ElMessage.warning('请输入 6 位期间 (YYYYMM)')
    return
  }
  loading.value = true
  try {
    const res: any = await request.get('/bank-reconciliation/diff-report', {
      params: { bank_account_id: bankAccountId.value, period_no: periodNo.value },
    })
    report.value = res?.data ?? null
    if (!report.value) {
      ElMessage.warning('未能生成报告')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '生成报告失败')
    report.value = null
  } finally {
    loading.value = false
  }
}

function exportCsv() {
  if (!report.value) return
  const rows = [
    ['类型', '日期', '摘要', '方向', '金额'],
    ...report.value.bank_only_items.map(it => ['银行有企业无', it.txn_date, it.description, it.direction, it.amount]),
    ...report.value.book_only_items.map(it => ['企业有银行无', it.txn_date, it.description, it.direction, it.amount]),
  ]
  const csv = '\uFEFF' + rows.map(r => r.map(c => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `对账差异报告_${report.value.bank_name}_${periodNo.value}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  loadBankAccounts()
})
</script>

<style scoped lang="scss">
.diff-report-page {
  padding: 24px;
  .page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; margin: 0; } }
  .filter-bar { display: flex; align-items: center; }
  .summary .lbl { color: #999; font-size: 13px; margin-bottom: 4px; }
  .summary .val { font-size: 22px; font-weight: 600; }
  .summary .val.ok { color: #52c41a; }
  .summary .val.err { color: #ff4d4f; }
  .summary .ts { color: #999; font-size: 12px; margin-top: 4px; }
  .ok-bg { background: #f6ffed; }
  .err-bg { background: #fff1f0; }
  .card-title { font-weight: 600; }
  .card-title.bank { color: #1890ff; }
  .card-title.book { color: #fa8c16; }
}
</style>
