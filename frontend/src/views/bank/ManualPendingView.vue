<template>
  <div class="manual-pending-page">
    <div class="page-header">
      <h3>待人工处理工作台</h3>
      <el-button
        size="small"
        @click="loadData"
      >
        刷新
      </el-button>
    </div>

    <el-card>
      <el-table
        :data="txnList"
        border
        stripe
        size="small"
        @selection-change="onSelectionChange"
      >
        <el-table-column
          type="selection"
          width="40"
        />
        <el-table-column
          prop="txn_date"
          label="日期"
          width="90"
        />
        <el-table-column
          prop="description"
          label="摘要"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          prop="counterparty_name"
          label="对方"
          width="140"
          show-overflow-tooltip
        />
        <el-table-column
          label="金额"
          width="130"
          align="right"
        >
          <template #default="{ row }">
            <span :class="amountClass(row)">{{ amountDisplay(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="无匹配原因"
          min-width="160"
        >
          <template #default="{ row }">
            <span class="reason-text">{{ row.ai_business_scene || '无分类规则匹配' }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="260"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              @click="makeVoucher(row)"
            >
              做凭证
            </el-button>
            <el-button
              link
              type="success"
              size="small"
              @click="makeReceipt(row)"
            >
              做收款单
            </el-button>
            <el-button
              link
              type="danger"
              size="small"
              @click="makePayment(row)"
            >
              做付款单
            </el-button>
            <el-button
              link
              type="info"
              size="small"
              @click="markDone(row)"
            >
              标记完成
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 12px; justify-content: flex-end"
        @current-change="loadData"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getReviewList, submitReview } from '@/api/modules/bank_txn_review'
import request from '@/api/request'

interface TxnRow {
  id: string
  txn_date: string
  description: string
  counterparty_name: string
  debit: string
  credit: string
  direction: string
  classification: string
  ai_business_scene?: string
}

const txnList = ref<TxnRow[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const selectedIds = ref<string[]>([])

onMounted(() => {
  loadData()
})

async function loadData() {
  try {
    // Backend returns { data: [...], total, page, page_size }
    // Axios interceptor unwraps response.data, so res = { data: [...], total, page, page_size }
    const res: any = await getReviewList({ status: 'manual_pending', page: page.value, page_size: pageSize.value })
    const list = res?.data || []
    if (Array.isArray(list)) {
      txnList.value = list
      total.value = res?.total || list.length
    }
  } catch (e: any) {
    ElMessage.error('加载失败: ' + (e?.message || ''))
  }
}

function onSelectionChange(rows: TxnRow[]) {
  selectedIds.value = rows.map(r => r.id)
}

function amountClass(row: TxnRow) {
  return row.direction === 'in' ? 'amount-income' : 'amount-expense'
}

function amountDisplay(row: TxnRow) {
  const debit = Number(row.debit) || 0
  const credit = Number(row.credit) || 0
  const amt = debit > 0 ? debit : credit
  const sign = row.direction === 'in' || debit > 0 ? '+' : '-'
  return `${sign}¥${amt.toLocaleString('en', { minimumFractionDigits: 2 })}`
}

async function makeVoucher(row: TxnRow) {
  // 跳转凭证编辑页，带入流水信息
  ElMessage.info('跳转凭证编辑页（待实现）')
}

async function makeReceipt(row: TxnRow) {
  try {
    const data: any = await request.post('/payment-entries', {
      bank_transaction_id: row.id,
      payment_type: 'receive',
      party_type: 'customer',
      party_id: '00000000-0000-0000-0000-000000000000',
      posting_date: new Date().toISOString().slice(0, 10),
    })
    if (data?.payment_entry?.id) {
      ElMessage.success('收款单已生成')
      loadData()
    }
  } catch (e: any) {
    ElMessage.error('生成收款单失败: ' + (e?.message || ''))
  }
}

async function makePayment(row: TxnRow) {
  try {
    const data: any = await request.post('/payment-entries', {
      bank_transaction_id: row.id,
      payment_type: 'pay',
      party_type: 'supplier',
      party_id: '00000000-0000-0000-0000-000000000000',
      posting_date: new Date().toISOString().slice(0, 10),
    })
    if (data?.payment_entry?.id) {
      ElMessage.success('付款单已生成')
      loadData()
    }
  } catch (e: any) {
    ElMessage.error('生成付款单失败: ' + (e?.message || ''))
  }
}

async function markDone(row: TxnRow) {
  try {
    // submitReview returns { data: { approved_count, results } } after Axios unwrap
    const res: any = await submitReview({ txn_ids: [row.id] })
    if (res?.data?.approved_count > 0) {
      ElMessage.success('已标记为完成')
      loadData()
    }
  } catch (e: any) {
    ElMessage.error('操作失败: ' + (e?.message || ''))
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.reason-text { color: #999; font-size: 13px; }
.amount-income { color: #67c23a; }
.amount-expense { color: #f56c6c; }
</style>
