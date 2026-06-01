<template>
  <div class="payment-page">
    <div class="page-header">
      <h3>收付款单</h3>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="单据类型">
          <el-select v-model="filter.paymentType" placeholder="全部" style="width: 120px" clearable>
            <el-option label="收款单" value="receive" />
            <el-option label="付款单" value="pay" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="filter.dateRange" type="daterange" range-separator="~"
            start-placeholder="开始" end-placeholder="结束" style="width: 240px" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="单据号/对方单位" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="12" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ stats.total }}</p>
          <p class="stat-label">单据总数</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card receive">
          <p class="stat-num">{{ stats.receiveCount }}</p>
          <p class="stat-label">收款单</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card pay">
          <p class="stat-num">{{ stats.payCount }}</p>
          <p class="stat-label">付款单</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ stats.totalAmount }}</p>
          <p class="stat-label">收付总额(元)</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 列表 -->
    <el-card>
      <el-table :data="payments" border stripe size="small" v-loading="loading">
        <el-table-column prop="payment_no" label="单据号" width="160" />
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.payment_type === 'receive' ? 'success' : 'danger'" size="small">
              {{ row.payment_type === 'receive' ? '收款' : '付款' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="party_name" label="对方单位" min-width="140" />
        <el-table-column prop="paid_amount" label="金额" width="120" align="right">
          <template #default="{ row }">
            <span :class="row.payment_type === 'receive' ? 'amount-income' : 'amount-expense'">
              {{ row.payment_type === 'receive' ? '+' : '-' }}{{ formatAmount(row.paid_amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="posting_date" label="日期" width="100" />
        <el-table-column prop="reference_no" label="参考号" width="130" />
        <el-table-column label="单据状态" width="90">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.docstatus" />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > 0"
        style="margin-top: 12px; justify-content: flex-end"
        background
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        layout="total, prev, pager, next"
        @current-change="onPageChange"
      />
    </el-card>

    <!-- 详情抽屉 -->
    <el-drawer v-model="showDrawer" :title="`收付款单 ${currentPayment?.payment_no || ''}`" size="480px">
      <template v-if="currentPayment">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="单据号">{{ currentPayment.payment_no }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="currentPayment.payment_type === 'receive' ? 'success' : 'danger'" size="small">
              {{ currentPayment.payment_type === 'receive' ? '收款' : '付款' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="对方单位">{{ currentPayment.party_name }}</el-descriptions-item>
          <el-descriptions-item label="金额">
            <b :class="currentPayment.payment_type === 'receive' ? 'amount-income' : 'amount-expense'">
              {{ currentPayment.payment_type === 'receive' ? '+' : '-' }}{{ formatAmount(currentPayment.paid_amount) }}
            </b>
          </el-descriptions-item>
          <el-descriptions-item label="参考号">{{ currentPayment.reference_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="日期">{{ currentPayment.posting_date }}</el-descriptions-item>
          <el-descriptions-item label="单据状态">
            <DocStatusTag :docstatus="currentPayment.docstatus" />
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentPayment.created_at }}</el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 20px; text-align: center">
          <el-button v-if="currentPayment.docstatus === 0" type="primary">提交审核</el-button>
          <el-button v-if="currentPayment.docstatus === 1" type="danger">作废</el-button>
          <el-button @click="showDrawer = false">关闭</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchPayments, fetchPaymentDetail } from '@/api/modules/payment'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { PaymentEntry } from '@/types/models/payment'

const loading = ref(false)
const payments = ref<PaymentEntry[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive({
  paymentType: '',
  dateRange: null as [string, string] | null,
  keyword: '',
})

const stats = computed(() => {
  const all = payments.value
  const receiveList = all.filter(p => p.payment_type === 'receive')
  const payList = all.filter(p => p.payment_type === 'pay')
  const totalReceive = receiveList.reduce((s, p) => s + Number(p.paid_amount || 0), 0)
  const totalPay = payList.reduce((s, p) => s + Number(p.paid_amount || 0), 0)
  return {
    total: all.length,
    receiveCount: receiveList.length,
    payCount: payList.length,
    totalAmount: (totalReceive - totalPay).toLocaleString('en', { minimumFractionDigits: 2 }),
  }
})

const showDrawer = ref(false)
const currentPayment = ref<PaymentEntry | null>(null)

async function loadData() {
  loading.value = true
  try {
    const [startDate, endDate] = filter.dateRange || [undefined, undefined]
    const res: any = await fetchPayments({
      page: page.value,
      pageSize: pageSize.value,
      payment_type: filter.paymentType || undefined,
      start_date: startDate,
      end_date: endDate,
    })
    payments.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total || (Array.isArray(payments.value) ? payments.value.length : 0)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    payments.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.paymentType = ''
  filter.dateRange = null
  filter.keyword = ''
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}

async function showDetail(row: PaymentEntry) {
  try {
    const res: any = await fetchPaymentDetail(row.id)
    currentPayment.value = res?.data || row
    showDrawer.value = true
  } catch {
    currentPayment.value = row
    showDrawer.value = true
  }
}

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.filter-card { margin-bottom: 12px; }
.stat-row { margin-bottom: 12px; }
.stat-card {
  text-align: center;
  .stat-num { font-size: 24px; font-weight: 700; margin-bottom: 4px; color: #333; }
  .stat-label { font-size: 12px; color: #999; }
  &.receive .stat-num { color: #389e0d; }
  &.pay .stat-num { color: #cf1322; }
}
.amount-income { color: #389e0d; font-weight: 600; }
.amount-expense { color: #cf1322; font-weight: 600; }
</style>
