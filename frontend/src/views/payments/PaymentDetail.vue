<template>
  <div class="payment-detail-page">
    <div class="page-header">
      <h3>收付款单详情</h3>
      <DocStatusTag
        v-if="payment"
        :docstatus="payment.docstatus"
        size="default"
      />
    </div>

    <el-card v-loading="loading" v-if="payment">
      <!-- 单据信息 -->
      <div class="section-title">单据信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="单据号">
          <b>{{ payment.payment_no }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="payment.payment_type === 'receive' ? 'success' : 'danger'" size="small">
            {{ payment.payment_type === 'receive' ? '收款' : '付款' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="对方单位">
          {{ payment.counterparty_name || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="付款方式">
          <el-tag size="small" v-if="payment.payment_method">
            {{ paymentMethodLabel(payment.payment_method) }}
          </el-tag>
          <span v-else>—</span>
        </el-descriptions-item>
        <el-descriptions-item label="金额">
          <b :class="payment.payment_type === 'receive' ? 'amount-income' : 'amount-expense'">
            {{ payment.payment_type === 'receive' ? '+' : '-' }}{{ formatAmount(payment.paid_amount) }}
          </b>
        </el-descriptions-item>
        <el-descriptions-item label="参考号">
          {{ payment.reference_no || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="日期">
          {{ payment.posting_date }}
        </el-descriptions-item>
        <el-descriptions-item label="单据状态">
          <DocStatusTag :docstatus="payment.docstatus" />
        </el-descriptions-item>
      </el-descriptions>

      <!-- 凭证信息 -->
      <div class="section-title">凭证信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="关联凭证">
          <template v-if="payment.voucher_id">
            <el-link type="primary" :underline="false" @click="goVoucher(payment.voucher_id)">
              查看凭证
            </el-link>
          </template>
          <span v-else>—</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ payment.created_at || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 发票核销列表 -->
      <div class="section-title">发票核销</div>
      <el-table
        v-if="allocations.length > 0"
        :data="allocations"
        border
        size="small"
        max-height="300"
      >
        <el-table-column prop="invoice_no" label="发票号" min-width="160" />
        <el-table-column label="发票类型" width="100">
          <template #default="{ row }">
            {{ row.invoice_type === 'sale' ? '销项' : row.invoice_type === 'purchase' ? '进项' : row.invoice_type || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="核销金额" width="140" align="right">
          <template #default="{ row }">
            ¥{{ formatAmount(row.allocated_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="核销时间" width="160">
          <template #default="{ row }">
            {{ row.created_at || '—' }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else-if="!loadingAllocations" description="暂无核销记录" />

      <!-- 操作按钮 -->
      <div class="detail-actions">
        <el-button @click="goBack">返回列表</el-button>

        <el-button-group>
          <el-button
            v-if="payment.docstatus === 0"
            type="warning"
            :loading="actionLoading"
            @click="goEdit"
          >
            编辑
          </el-button>

          <el-popconfirm
            v-if="payment.docstatus === 0"
            title="确定提交审核？提交后不可修改。"
            confirm-button-text="确定"
            cancel-button-text="取消"
            @confirm="handleSubmit"
          >
            <template #reference>
              <el-button type="primary" :loading="actionLoading">提交审核</el-button>
            </template>
          </el-popconfirm>

          <el-popconfirm
            v-if="payment.docstatus === 1"
            title="确定审核通过？"
            confirm-button-text="确定"
            cancel-button-text="取消"
            @confirm="handleApprove"
          >
            <template #reference>
              <el-button type="success" :loading="actionLoading">审核通过</el-button>
            </template>
          </el-popconfirm>

          <el-button
            v-if="payment.docstatus === 0 || payment.docstatus === 1"
            type="primary"
            :loading="voucherLoading"
            @click="handleGenerateVoucher"
          >
            生成凭证
          </el-button>

          <el-popconfirm
            v-if="payment.docstatus === 0"
            title="确定删除该单据？删除后不可恢复。"
            confirm-button-text="确定"
            cancel-button-text="取消"
            @confirm="handleDelete"
          >
            <template #reference>
              <el-button type="danger" :loading="actionLoading">删除</el-button>
            </template>
          </el-popconfirm>
        </el-button-group>
      </div>
    </el-card>

    <el-empty v-else-if="!loading" description="未找到收付款单数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  fetchPaymentDetail,
  deletePayment,
  generateVoucherFromPayment,
  updatePayment,
} from '@/api/modules/payment'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { PaymentEntry } from '@/types/models/payment'

const route = useRoute()
const router = useRouter()

const payment = ref<PaymentEntry | null>(null)
const loading = ref(false)
const actionLoading = ref(false)
const voucherLoading = ref(false)

const allocations = ref<any[]>([])
const loadingAllocations = ref(false)

const PAYMENT_METHOD_LABEL: Record<string, string> = {
  bank: '银行转账',
  cash: '现金',
  wechat: '微信',
  alipay: '支付宝',
  other: '其他',
}

function paymentMethodLabel(method?: string): string {
  if (!method) return '—'
  return PAYMENT_METHOD_LABEL[method] || method
}

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchPaymentDetail(route.params.id as string)
    payment.value = res?.data || res
    await loadAllocations()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadAllocations() {
  if (!payment.value?.id) return
  loadingAllocations.value = true
  try {
    // 从 payment detail 或 allocations 端点获取已核销发票列表
    // 若后端有独立端点可替换为 fetcAllocationsByPayment
    const res: any = await fetchPaymentDetail(payment.value.id)
    const data = res?.data || res
    if (data?.allocations) {
      allocations.value = Array.isArray(data.allocations) ? data.allocations : []
    } else {
      allocations.value = []
    }
  } catch {
    allocations.value = []
  } finally {
    loadingAllocations.value = false
  }
}

async function handleSubmit() {
  if (!payment.value) return
  actionLoading.value = true
  try {
    await updatePayment(payment.value.id, { docstatus: 1 } as any)
    ElMessage.success('提交审核成功')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '提交审核失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleApprove() {
  if (!payment.value) return
  actionLoading.value = true
  try {
    await updatePayment(payment.value.id, { docstatus: 2 } as any)
    ElMessage.success('审核通过')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '审核失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleGenerateVoucher() {
  if (!payment.value) return
  voucherLoading.value = true
  try {
    const res: any = await generateVoucherFromPayment(payment.value.id)
    const voucherNo = res?.data?.voucher_no || ''
    ElMessage.success(`凭证已生成: ${voucherNo}`)
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '凭证生成失败')
  } finally {
    voucherLoading.value = false
  }
}

async function handleDelete() {
  if (!payment.value) return
  actionLoading.value = true
  try {
    await deletePayment(payment.value.id)
    ElMessage.success('删除成功')
    goBack()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  } finally {
    actionLoading.value = false
  }
}

function goEdit() {
  if (!payment.value) return
  router.push(`/payments/${payment.value.id}/edit`)
}

function goVoucher(voucherId: string) {
  router.push(`/vouchers/${voucherId}`)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/payments')
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.payment-detail-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
    h3 { font-size: 18px; }
  }

  .section-title {
    font-size: 14px;
    font-weight: 600;
    color: #303133;
    margin: 20px 0 10px;
    padding-left: 8px;
    border-left: 3px solid #409eff;
    line-height: 1.2;

    &:first-child {
      margin-top: 0;
    }
  }

  .detail-actions {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 24px;
  }

  .amount-income { color: #389e0d; font-weight: 600; }
  .amount-expense { color: #cf1322; font-weight: 600; }
}
</style>