<template>
  <Layout>
    <div class="dashboard">
      <!-- 统计卡片 -->
      <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin-bottom: 24px;">
        <n-gi>
          <n-card>
            <div class="stat-card">
              <div class="stat-label">本月凭证数</div>
              <div class="stat-value">{{ stats.voucherCount }}</div>
            </div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card>
            <div class="stat-card">
              <div class="stat-label">未对账银行流水</div>
              <div class="stat-value">{{ stats.unmatchedCount }}</div>
            </div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card>
            <div class="stat-card">
              <div class="stat-label">本月发票数</div>
              <div class="stat-value">{{ stats.invoiceCount }}</div>
            </div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card>
            <div class="stat-card">
              <div class="stat-label">本月发票金额</div>
              <div class="stat-value">{{ formatMoney(stats.invoiceAmount) }}</div>
            </div>
          </n-card>
        </n-gi>
      </n-grid>

      <!-- 试算平衡表 -->
      <n-card title="试算平衡（本月汇总）" style="margin-bottom: 16px;">
        <n-table :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>科目类别</th>
              <th style="text-align: right;">借方余额</th>
              <th style="text-align: right;">贷方余额</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>资产</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.assetDebit) }}</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.assetCredit) }}</td>
            </tr>
            <tr>
              <td>负债</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.liabilityDebit) }}</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.liabilityCredit) }}</td>
            </tr>
            <tr>
              <td>权益</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.equityDebit) }}</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.equityCredit) }}</td>
            </tr>
            <tr>
              <td>成本/损益</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.expenseDebit) }}</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.expenseCredit) }}</td>
            </tr>
            <tr style="font-weight: 600;">
              <td>合计</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.totalDebit) }}</td>
              <td style="text-align: right;">{{ formatMoney(trialBalance.totalCredit) }}</td>
            </tr>
          </tbody>
        </n-table>
      </n-card>

      <!-- 最近凭证 -->
      <n-card title="最近凭证">
        <template #header-extra>
          <n-button text type="primary" @click="$router.push('/vouchers')">查看全部</n-button>
        </template>
        <n-table :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>日期</th>
              <th>凭证号</th>
              <th>摘要</th>
              <th>金额</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="v in recentVouchers" :key="v.id">
              <td>{{ v.date }}</td>
              <td>{{ v.number }}</td>
              <td>{{ getSummary(v) }}</td>
              <td>{{ formatMoney(getAmount(v)) }}</td>
              <td><StatusTag :status="v.status" /></td>
            </tr>
            <tr v-if="recentVouchers.length === 0">
              <td colspan="5" style="text-align: center; color: #999;">暂无数据</td>
            </tr>
          </tbody>
        </n-table>
      </n-card>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NGrid, NGi, NCard, NTable, NButton, useMessage } from 'naive-ui'
import Layout from '@/components/Layout.vue'
import StatusTag from '@/components/StatusTag.vue'
import { reportApi, voucherApi, bankTxnApi, invoiceApi } from '@/api/adapter/client'
import type { Voucher, TrialBalanceReport } from '@/types'

const message = useMessage()

const stats = ref({
  voucherCount: 0,
  unmatchedCount: 0,
  invoiceCount: 0,
  invoiceAmount: 0
})

const trialBalance = ref({
  assetDebit: 0, assetCredit: 0,
  liabilityDebit: 0, liabilityCredit: 0,
  equityDebit: 0, equityCredit: 0,
  expenseDebit: 0, expenseCredit: 0,
  totalDebit: 0, totalCredit: 0
})

const recentVouchers = ref<Voucher[]>([])

const formatMoney = (amount: number) => {
  return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(amount)
}

const getSummary = (v: Voucher) => {
  return v.lines?.[0]?.summary || '-'
}

const getAmount = (v: Voucher) => {
  if (!v.lines?.length) return 0
  return v.lines.reduce((sum, line) => sum + line.debit, 0)
}

onMounted(async () => {
  try {
    // 加载试算平衡表
    const tb = await reportApi.trialBalance()
    if (tb?.items?.length) {
      // 按类型汇总
      let assetDebit = 0, assetCredit = 0
      let liabilityDebit = 0, liabilityCredit = 0
      let equityDebit = 0, equityCredit = 0
      let expenseDebit = 0, expenseCredit = 0
      
      tb.items.forEach(item => {
        const code = item.account_code || ''
        const isExpense = code.startsWith('5') || code.startsWith('6') || code.startsWith('7')
        const isLiability = code.startsWith('2')
        const isEquity = code.startsWith('3') || code.startsWith('4')
        
        if (isExpense) {
          expenseDebit += item.debit_balance
          expenseCredit += item.credit_balance
        } else if (isLiability) {
          liabilityDebit += item.debit_balance
          liabilityCredit += item.credit_balance
        } else if (isEquity) {
          equityDebit += item.debit_balance
          equityCredit += item.credit_balance
        } else {
          assetDebit += item.debit_balance
          assetCredit += item.credit_balance
        }
      })
      
      trialBalance.value = {
        assetDebit, assetCredit,
        liabilityDebit, liabilityCredit,
        equityDebit, equityCredit,
        expenseDebit, expenseCredit,
        totalDebit: assetDebit + liabilityDebit + equityDebit + expenseDebit,
        totalCredit: assetCredit + liabilityCredit + equityCredit + expenseCredit
      }
    }
  } catch (e) {
    console.error('Failed to load trial balance:', e)
  }

  try {
    // 加载最近凭证
    const result = await voucherApi.list({ page: 1, page_size: 5 })
    recentVouchers.value = result.vouchers || []
  } catch (e) {
    console.error('Failed to load vouchers:', e)
  }

  try {
    // 加载未对账流水数
    const count = await bankTxnApi.unmatched()
    stats.value.unmatchedCount = count
  } catch (e) {
    console.error('Failed to load unmatched count:', e)
  }

  try {
    // 加载发票统计
    const invStats = await invoiceApi.stats()
    stats.value.invoiceCount = invStats.this_month_count
    stats.value.invoiceAmount = invStats.this_month_amount
  } catch (e) {
    console.error('Failed to load invoice stats:', e)
  }
})
</script>

<style scoped>
.dashboard {
  padding: 16px;
}
.stat-card {
  text-align: center;
}
.stat-label {
  color: #666;
  font-size: 14px;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #333;
}
</style>