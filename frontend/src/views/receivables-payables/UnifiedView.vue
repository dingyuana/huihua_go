<template>
  <div class="unified-page">
    <div class="page-header">
      <h3>应收应付汇总</h3>
      <p class="page-hint">集中查看应收、应付、预收、预付单据</p>
    </div>

    <!-- 全局统计 -->
    <el-row :gutter="12" class="global-stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card ar"><p class="stat-num">{{ globals.arCount }}</p><p class="stat-label">应收单</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card ap"><p class="stat-num">{{ globals.apCount }}</p><p class="stat-label">应付单</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card adv-rec"><p class="stat-num">{{ globals.advRecCount }}</p><p class="stat-label">预收单</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card adv-pay"><p class="stat-num">{{ globals.advPayCount }}</p><p class="stat-label">预付单</p></el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" class="global-stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card"><p class="stat-num">¥{{ globals.arTotal }}</p><p class="stat-label">应收总额</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card"><p class="stat-num">¥{{ globals.apTotal }}</p><p class="stat-label">应付总额</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card"><p class="stat-num">¥{{ globals.advRecAmt }}</p><p class="stat-label">预收总额</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card"><p class="stat-num">¥{{ globals.advPayAmt }}</p><p class="stat-label">预付总额</p></el-card>
      </el-col>
    </el-row>

    <!-- 折叠面板 -->
    <el-collapse v-model="activeNames">
      <!-- 应收款单 -->
      <el-collapse-item title="应收款单" name="ar">
        <template #title>
          <span class="collapse-header ar-header">应收款单 <el-tag size="small" type="warning">{{ arList.length }}</el-tag></span>
        </template>
        <el-table :data="arList" border stripe size="small" v-loading="loadingAR" max-height="400">
          <el-table-column label="发票号" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">{{ row.invoice_no }}</template>
          </el-table-column>
          <el-table-column prop="customer_name" label="客户" min-width="140" show-overflow-tooltip />
          <el-table-column label="金额" width="130" align="right">
            <template #default="{ row }"><span class="amt-amount">¥{{ fmt(row.amount) }}</span></template>
          </el-table-column>
          <el-table-column label="已收" width="110" align="right">
            <template #default="{ row }"><span class="amt-paid">¥{{ fmt(row.paid_amount) }}</span></template>
          </el-table-column>
          <el-table-column label="未收" width="120" align="right">
            <template #default="{ row }"><span :class="outstandingCls(row.outstanding_amount)">¥{{ fmt(row.outstanding_amount) }}</span></template>
          </el-table-column>
          <el-table-column prop="due_date" label="到期日" width="100" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }"><el-tag :type="arStatusTag(row.status)" size="small">{{ arStatusLabel(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="70" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="goToArDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <!-- 应付款单 -->
      <el-collapse-item title="应付款单" name="ap">
        <template #title>
          <span class="collapse-header ap-header">应付款单 <el-tag size="small" type="warning">{{ apList.length }}</el-tag></span>
        </template>
        <el-table :data="apList" border stripe size="small" v-loading="loadingAP" max-height="400">
          <el-table-column label="发票号" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">{{ row.invoice_no }}</template>
          </el-table-column>
          <el-table-column prop="supplier_name" label="供应商" min-width="140" show-overflow-tooltip />
          <el-table-column label="金额" width="130" align="right">
            <template #default="{ row }"><span class="amt-amount">¥{{ fmt(row.amount) }}</span></template>
          </el-table-column>
          <el-table-column label="已付" width="110" align="right">
            <template #default="{ row }"><span class="amt-paid">¥{{ fmt(row.paid_amount) }}</span></template>
          </el-table-column>
          <el-table-column label="未付" width="120" align="right">
            <template #default="{ row }"><span :class="outstandingCls(row.outstanding_amount)">¥{{ fmt(row.outstanding_amount) }}</span></template>
          </el-table-column>
          <el-table-column prop="due_date" label="到期日" width="100" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }"><el-tag :type="apStatusTag(row.status)" size="small">{{ apStatusLabel(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="70" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="goToApDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <!-- 预收款单 -->
      <el-collapse-item title="预收款单" name="adv-rec">
        <template #title>
          <span class="collapse-header adv-rec-header">预收款单 <el-tag size="small" type="warning">{{ advRecList.length }}</el-tag></span>
        </template>
        <el-table :data="advRecList" border stripe size="small" v-loading="loadingAdvRec" max-height="400">
          <el-table-column prop="advance_no" label="预收单号" min-width="160" show-overflow-tooltip />
          <el-table-column prop="customer_id" label="客户" width="240" show-overflow-tooltip />
          <el-table-column label="金额" width="130" align="right">
            <template #default="{ row }">¥{{ fmt(row.amount) }}</template>
          </el-table-column>
          <el-table-column label="已分配" width="110" align="right">
            <template #default="{ row }">¥{{ fmt(row.allocated_amount) }}</template>
          </el-table-column>
          <el-table-column label="未分配" width="120" align="right">
            <template #default="{ row }"><span class="amt-outstanding">¥{{ fmt(row.outstanding_amount) }}</span></template>
          </el-table-column>
          <el-table-column prop="received_date" label="收款日" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }"><el-tag :type="advStatusTag(row.status)" size="small">{{ advStatusLabel(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="70" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="goToAdvRecDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <!-- 预付款单 -->
      <el-collapse-item title="预付款单" name="adv-pay">
        <template #title>
          <span class="collapse-header adv-pay-header">预付款单 <el-tag size="small" type="warning">{{ advPayList.length }}</el-tag></span>
        </template>
        <el-table :data="advPayList" border stripe size="small" v-loading="loadingAdvPay" max-height="400">
          <el-table-column prop="advance_no" label="预付单号" min-width="160" show-overflow-tooltip />
          <el-table-column prop="supplier_id" label="供应商" width="240" show-overflow-tooltip />
          <el-table-column label="金额" width="130" align="right">
            <template #default="{ row }">¥{{ fmt(row.amount) }}</template>
          </el-table-column>
          <el-table-column label="已分配" width="110" align="right">
            <template #default="{ row }">¥{{ fmt(row.allocated_amount) }}</template>
          </el-table-column>
          <el-table-column label="未分配" width="120" align="right">
            <template #default="{ row }"><span class="amt-outstanding">¥{{ fmt(row.outstanding_amount) }}</span></template>
          </el-table-column>
          <el-table-column prop="paid_date" label="付款日" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }"><el-tag :type="advStatusTag(row.status)" size="small">{{ advStatusLabel(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="70" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="goToAdvPayDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchArInvoices, type ArInvoice } from '@/api/modules/ar_invoice'
import { fetchApInvoices, type ApInvoice } from '@/api/modules/ap_invoice'
import { fetchAdvanceReceipts, type AdvanceReceipt } from '@/api/modules/advance_receipt'
import { fetchAdvancePayments, type AdvancePayment } from '@/api/modules/advance_payment'

const router = useRouter()
const activeNames = ref<string[]>(['ar', 'ap'])

// AR
const loadingAR = ref(false)
const arList = ref<ArInvoice[]>([])
// AP
const loadingAP = ref(false)
const apList = ref<ApInvoice[]>([])
// Advance Receipt
const loadingAdvRec = ref(false)
const advRecList = ref<AdvanceReceipt[]>([])
// Advance Payment
const loadingAdvPay = ref(false)
const advPayList = ref<AdvancePayment[]>([])

const globals = computed(() => {
  const arAmt = arList.value.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  const apAmt = apList.value.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  const advRecAmt = advRecList.value.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  const advPayAmt = advPayList.value.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  return {
    arCount: arList.value.length,
    apCount: apList.value.length,
    advRecCount: advRecList.value.length,
    advPayCount: advPayList.value.length,
    arTotal: arAmt.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    apTotal: apAmt.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    advRecAmt: advRecAmt.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    advPayAmt: advPayAmt.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
  }
})

async function loadAll() {
  loadingAR.value = true
  loadingAP.value = true
  loadingAdvRec.value = true
  loadingAdvPay.value = true
  try {
    const [arRes, apRes, advRecRes, advPayRes] = await Promise.all([
      fetchArInvoices(),
      fetchApInvoices(),
      fetchAdvanceReceipts(),
      fetchAdvancePayments(),
    ])
    const extract = (res: any) => {
      const d = res?.data || res
      return d?.list || []
    }
    arList.value = extract(arRes)
    apList.value = extract(apRes)
    advRecList.value = extract(advRecRes)
    advPayList.value = extract(advPayRes)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loadingAR.value = false
    loadingAP.value = false
    loadingAdvRec.value = false
    loadingAdvPay.value = false
  }
}

function fmt(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function outstandingCls(val: any): string {
  return (Number(val) || 0) > 0 ? 'amt-outstanding' : 'amt-cleared'
}

// AR status helpers
function arStatusLabel(s: string): string {
  const m: Record<string, string> = { draft: '草稿', confirmed: '已确认', partially_paid: '部分核销', paid: '已核销', reversed: '已冲销' }
  return m[s] || s
}
function arStatusTag(s: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  const m: Record<string, any> = { draft: 'warning', confirmed: 'success', partially_paid: 'warning', paid: 'success', reversed: 'info' }
  return m[s] || 'info'
}

// AP status helpers (same statuses)
const apStatusLabel = arStatusLabel
const apStatusTag = arStatusTag

// Advance status helpers
function advStatusLabel(s: string): string {
  const m: Record<string, string> = { draft: '草稿', confirmed: '已确认', partially_allocated: '部分已分配', fully_allocated: '已全部分配', reversed: '已冲销' }
  return m[s] || s
}
function advStatusTag(s: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  const m: Record<string, any> = { draft: 'warning', confirmed: 'primary', partially_allocated: 'warning', fully_allocated: 'success', reversed: 'info' }
  return m[s] || 'info'
}

function goToArDetail(row: ArInvoice) {
  router.push({ path: '/ar-invoices', query: { highlight: row.id } })
}
function goToApDetail(row: ApInvoice) {
  router.push({ path: '/ap-invoices', query: { highlight: row.id } })
}
function goToAdvRecDetail(row: AdvanceReceipt) {
  router.push({ path: '/advance-receipts', query: { highlight: row.id } })
}
function goToAdvPayDetail(row: AdvancePayment) {
  router.push({ path: '/advance-payments', query: { highlight: row.id } })
}

onMounted(loadAll)
</script>

<style scoped lang="scss">
.unified-page {
  .page-header {
    margin-bottom: 12px;
    h3 { font-size: 18px; margin: 0 0 4px; }
    .page-hint { font-size: 12px; color: #999; margin: 0; }
  }
  .global-stat-row { margin-bottom: 12px; }
  .stat-card {
    text-align: center;
    .stat-num { font-size: 20px; font-weight: 700; margin-bottom: 2px; color: #333; }
    .stat-label { font-size: 12px; color: #999; }
    &.ar .stat-num { color: #d4380d; }
    &.ap .stat-num { color: #096dd9; }
    &.adv-rec .stat-num { color: #d48806; }
    &.adv-pay .stat-num { color: #389e0d; }
  }
  .collapse-header {
    font-weight: 600;
    &.ar-header { color: #d4380d; }
    &.ap-header { color: #096dd9; }
    &.adv-rec-header { color: #d48806; }
    &.adv-pay-header { color: #389e0d; }
  }
  :deep(.el-collapse-item__header) { font-weight: 600; }
  .amt-amount { color: #d4380d; font-weight: 600; }
  .amt-paid { color: #389e0d; }
  .amt-outstanding { color: #d4380d; font-weight: 600; }
  .amt-cleared { color: #52c41a; }
}
</style>
