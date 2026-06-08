<template>
  <div class="payment-page">
    <div class="page-header">
      <h3>收付款单</h3>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
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

    <!-- 统计汇总 -->
    <el-row :gutter="8" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card receive">
          <div class="stat-summary">
            <span class="stat-label">收款总额</span>
            <span class="stat-amount">{{ formatAmount(stats.receiveAmount) }} 元</span>
            <span class="stat-count">共 {{ stats.receiveCount }} 笔</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card pay">
          <div class="stat-summary">
            <span class="stat-label">付款总额</span>
            <span class="stat-amount">{{ formatAmount(stats.payAmount) }} 元</span>
            <span class="stat-count">共 {{ stats.payCount }} 笔</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card net">
          <div class="stat-summary">
            <span class="stat-label">总金额</span>
            <span class="stat-amount" :class="stats.netAmount >= 0 ? 'receive' : 'pay'">{{ stats.netAmount >= 0 ? '+' : '' }}{{ formatAmount(stats.netAmount) }} 元</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card total">
          <div class="stat-summary">
            <span class="stat-label">总笔数</span>
            <span class="stat-amount">{{ stats.total }} 笔</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Tab 切换 + 列表 -->
    <el-card>
      <el-tabs v-model="activeTab" class="payment-tabs">
        <el-tab-pane label="收款" name="receive">
          <el-table :data="payments" border stripe size="small" v-loading="loading">
            <el-table-column prop="payment_no" label="单据号" width="160" />
            <el-table-column prop="counterparty_name" label="对方单位" min-width="140" />
            <el-table-column prop="paid_amount" label="收款金额" width="120" align="right">
              <template #default="{ row }">
                <span class="amount-income">+{{ formatAmount(row.paid_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="posting_date" label="日期" width="100" />
            <el-table-column prop="reference_no" label="参考号" width="130" />
            <el-table-column label="单据状态" width="90">
              <template #default="{ row }"><DocStatusTag :docstatus="row.docstatus" /></template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="付款" name="pay">
          <el-table :data="payments" border stripe size="small" v-loading="loading">
            <el-table-column prop="payment_no" label="单据号" width="160" />
            <el-table-column prop="counterparty_name" label="对方单位" min-width="140" />
            <el-table-column prop="paid_amount" label="付款金额" width="120" align="right">
              <template #default="{ row }">
                <span class="amount-expense">-{{ formatAmount(row.paid_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="posting_date" label="日期" width="100" />
            <el-table-column prop="reference_no" label="参考号" width="130" />
            <el-table-column label="单据状态" width="90">
              <template #default="{ row }"><DocStatusTag :docstatus="row.docstatus" /></template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

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
          <el-descriptions-item label="对方单位">{{ currentPayment.counterparty_name }}</el-descriptions-item>
          <el-descriptions-item label="收款方式">
            <el-tag size="small" v-if="currentPayment.payment_method">
              {{ { bank: '银行转账', cash: '现金', wechat: '微信', alipay: '支付宝', other: '其他' }[currentPayment.payment_method] || currentPayment.payment_method }}
            </el-tag>
            <span v-else>银行转账</span>
          </el-descriptions-item>
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

        <!-- 发票核销 -->
        <template v-if="currentPayment.party_id">
          <el-divider />
          <div class="allocation-section">
            <div class="section-header">
              <h4>发票核销</h4>
              <el-button size="small" :loading="invoiceLoading" @click="loadUnmatchedInvoices">
                {{ invoiceLoaded ? '刷新未核销发票' : '加载未核销发票' }}
              </el-button>
            </div>
            <el-table
              v-if="unmatchedInvoices.length > 0"
              :data="unmatchedInvoices"
              border size="small" max-height="300"
              @selection-change="(rows: any[]) => { selectedInvoices = rows }"
            >
              <el-table-column type="selection" width="40" />
              <el-table-column prop="invoice_no" label="发票号" width="140" />
              <el-table-column label="原金额" width="100" align="right">
                <template #default="{ row }">{{ formatInvoiceAmount(row.total_amount) }}</template>
              </el-table-column>
              <el-table-column label="未核销" width="110" align="right">
                <template #default="{ row }">{{ formatInvoiceAmount(row.outstanding_amount) }}</template>
              </el-table-column>
              <el-table-column label="本次核销" width="140" align="right">
                <template #default="{ row }">
                  <el-input-number
                    v-model="allocationMap[row.id]"
                    :min="0"
                    :max="Number(row.outstanding_amount)"
                    :precision="2"
                    :step="100"
                    size="small"
                    controls-position="right"
                    style="width: 130px"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="posting_date" label="日期" width="90" />
            </el-table>
            <el-empty v-else-if="invoiceLoaded" :description="'暂无未核销发票'" />
            <div v-if="unmatchedInvoices.length > 0" style="margin-top: 12px; text-align: right">
              <el-button type="primary" :loading="allocating" @click="handleAllocate">
                确认核销
              </el-button>
            </div>
          </div>
        </template>

        <div style="margin-top: 20px; text-align: center">
          <el-button v-if="currentPayment.docstatus === 0" type="primary" :loading="voucherLoading" @click="handleGenerateVoucher">生成凭证</el-button>
          <el-button v-if="currentPayment.docstatus === 1" type="danger">作废</el-button>
          <el-button @click="showDrawer = false">关闭</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchPayments, fetchPaymentDetail, generateVoucherFromPayment, fetchUnmatchedInvoices, allocateInvoices } from '@/api/modules/payment'
import { deleteVoucher } from '@/api/modules/voucher'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { PaymentEntry } from '@/types/models/payment'

const loading = ref(false)
const payments = ref<PaymentEntry[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const activeTab = ref('receive')

const filter = reactive({
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
    receiveAmount: totalReceive,
    payAmount: totalPay,
    netAmount: totalReceive - totalPay,
  }
})

watch(activeTab, () => {
  page.value = 1
  loadData()
})

const showDrawer = ref(false)
const currentPayment = ref<PaymentEntry | null>(null)
const voucherLoading = ref(false)
const voidLoading = ref(false)

// Invoice matching
const unmatchedInvoices = ref<any[]>([])
const selectedInvoices = ref<any[]>([])
const invoiceLoading = ref(false)
const invoiceLoaded = ref(false)
const allocating = ref(false)
const allocationMap = reactive<Record<string, number>>({})

async function handleGenerateVoucher() {
  if (!currentPayment.value) return
  voucherLoading.value = true
  try {
    const res: any = await generateVoucherFromPayment(currentPayment.value.id)
    const voucherNo = res?.data?.voucher_no || ''
    currentPayment.value.docstatus = 1 // posted
    ElMessage.success(`凭证已生成: ${voucherNo}`)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '凭证生成失败')
  } finally {
    voucherLoading.value = false
  }
}

async function loadUnmatchedInvoices() {
  if (!currentPayment.value?.party_id) return
  invoiceLoading.value = true
  try {
    const res: any = await fetchUnmatchedInvoices(currentPayment.value.party_id)
    const list = res?.data || []
    unmatchedInvoices.value = Array.isArray(list) ? list : []
    invoiceLoaded.value = true
    for (const key of Object.keys(allocationMap)) {
      delete allocationMap[key]
    }
    unmatchedInvoices.value.forEach((inv: any) => {
      allocationMap[inv.id] = Number(inv.outstanding_amount) || 0
    })
  } catch (e: any) {
    const errMsg = e?.response?.data?.error || e?.response?.data?.message || e?.message || ''
    ElMessage.error('加载未核销发票失败: ' + errMsg)
  } finally {
    invoiceLoading.value = false
  }
}

async function handleAllocate() {
  if (!currentPayment.value?.id || selectedInvoices.value.length === 0) {
    ElMessage.warning('请选择要核销的发票')
    return
  }
  allocating.value = true
  try {
    const allocations = selectedInvoices.value.map((inv: any) => ({
      invoice_id: inv.id,
      allocated_amount: allocationMap[inv.id] || Number(inv.outstanding_amount) || 0,
    }))
    await allocateInvoices(currentPayment.value.id, allocations)
    ElMessage.success('核销成功')
    selectedInvoices.value = []
    loadUnmatchedInvoices()
  } catch (e: any) {
    ElMessage.error('核销失败: ' + (e?.response?.data?.error || e?.message || ''))
  } finally {
    allocating.value = false
  }
}

function formatInvoiceAmount(val: any): string {
  const n = Number(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const [startDate, endDate] = filter.dateRange || [undefined, undefined]
    const res: any = await fetchPayments({
      page: page.value,
      pageSize: pageSize.value,
      payment_type: activeTab.value,
      start_date: startDate,
      end_date: endDate,
      keyword: filter.keyword || undefined,
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
  .stat-summary {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    flex-wrap: wrap;
    line-height: 1;
  }
  .stat-label { font-size: 14px; font-weight: 500; color: #666; }
  .stat-amount { font-size: 20px; font-weight: 700; }
  .stat-count { font-size: 20px; font-weight: 700; }
  &.receive {
    .stat-amount, .stat-count { color: #389e0d; }
  }
  &.pay {
    .stat-amount, .stat-count { color: #cf1322; }
  }
  &.net {
    .stat-amount { color: #096dd9; }
  }
  &.total {
    .stat-amount { color: #333; }
  }
}
.payment-tabs { margin-top: -16px; }
.amount-income { color: #389e0d; font-weight: 600; }
.amount-expense { color: #cf1322; font-weight: 600; }
.allocation-section {
  margin-top: 4px;
  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    h4 { margin: 0; font-size: 14px; }
  }
}
</style>
