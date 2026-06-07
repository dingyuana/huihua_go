<template>
  <div class="expense-invoice-detail-page">
    <div class="page-header">
      <h3>进项发票详情</h3>
      <DocStatusTag
        v-if="invoice"
        :docstatus="invoice.docstatus"
        size="default"
      />
    </div>

    <el-card v-loading="loading" v-if="invoice">
      <!-- 基本信息 -->
      <div class="section-title">基本信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="发票号码">
          <b>{{ invoice.invoice_no }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="发票代码">
          {{ invoice.invoice_code || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="开票日期">
          {{ invoice.invoice_date }}
        </el-descriptions-item>
        <el-descriptions-item label="发票类型">
          {{ invoiceKindLabel(invoice.invoice_kind) }}
        </el-descriptions-item>
        <el-descriptions-item label="费用类别">
          {{ categoryLabel(invoice.category) }}
        </el-descriptions-item>
        <el-descriptions-item label="来源">
          <el-tag size="small" :type="sourceTagType(invoice.source)" effect="plain">
            {{ sourceLabel(invoice.source) }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 金额信息 -->
      <div class="section-title">金额信息</div>
      <el-descriptions :column="3" border size="small">
        <el-descriptions-item label="不含税金额">
          ¥{{ formatAmount(invoice.amount) }}
        </el-descriptions-item>
        <el-descriptions-item label="税额">
          ¥{{ formatAmount(invoice.tax_amount) }}
        </el-descriptions-item>
        <el-descriptions-item label="价税合计">
          <b style="color: #f56c6c">¥{{ formatAmount(invoice.total_amount) }}</b>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 供应商信息 -->
      <div class="section-title">供应商信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="供应商名称">
          {{ invoice.vendor_name || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="纳税人识别号">
          {{ invoice.tax_id || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">
          {{ invoice.description || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 验真信息 -->
      <div class="section-title">验真信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="验真状态">
          <el-tag size="small" :type="verifyTagType(invoice.verify_status)" effect="plain">
            {{ verifyLabel(invoice.verify_status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="验真时间">
          {{ invoice.verified_at || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="验真结果" :span="2">
          {{ invoice.verify_result || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 凭证信息 -->
      <div class="section-title">凭证信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="凭证ID" :span="2">
          <template v-if="invoice.voucher_id">
            <el-link
              type="primary"
              :underline="false"
              @click="goVoucher(invoice.voucher_id)"
            >
              {{ invoice.voucher_id }}
            </el-link>
          </template>
          <span v-else>—</span>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 元信息 -->
      <div class="section-title">元信息</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="单据状态">
          <DocStatusTag :docstatus="invoice.docstatus" />
        </el-descriptions-item>
        <el-descriptions-item label="业务状态">
          {{ invoice.status === 'confirmed' ? '已确认' : '草稿' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建人">
          {{ invoice.created_by || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ invoice.created_at }}
        </el-descriptions-item>
        <el-descriptions-item label="更新人">
          {{ invoice.updated_by || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ invoice.updated_at || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="detail-actions">
        <el-button @click="goBack">返回</el-button>

        <el-button-group>
          <!-- 编辑：草稿时显示 -->
          <el-button
            v-if="invoice.docstatus === 0"
            type="warning"
            :loading="actionLoading"
            @click="goEdit"
          >
            编辑
          </el-button>

          <!-- 确认入库：草稿时显示 -->
          <el-button
            v-if="invoice.docstatus === 0"
            type="primary"
            :loading="actionLoading"
            @click="handleConfirm"
          >
            确认入库
          </el-button>

          <!-- 验真：未验真时显示 -->
          <el-button
            v-if="invoice.verify_status === 'unverified' || !invoice.verify_status"
            type="success"
            :loading="actionLoading"
            @click="handleVerify"
          >
            验真
          </el-button>

          <!-- 删除：草稿时显示，弹 popconfirm -->
          <el-popconfirm
            v-if="invoice.docstatus === 0"
            title="确定删除该进项发票？删除后不可恢复。"
            confirm-button-text="确定"
            cancel-button-text="取消"
            @confirm="handleDelete"
          >
            <template #reference>
              <el-button type="danger" :loading="actionLoading">删除</el-button>
            </template>
          </el-popconfirm>
        </el-button-group>

        <!-- 有凭证ID 时直接显示查看凭证按钮 -->
        <el-button
          v-if="invoice.voucher_id"
          type="primary"
          @click="goVoucher(invoice.voucher_id)"
        >
          查看凭证
        </el-button>
      </div>
    </el-card>

    <el-empty v-else-if="!loading" description="未找到进项发票数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  fetchExpenseInvoiceDetail,
  confirmExpenseInvoice,
  verifyExpenseInvoice,
  deleteExpenseInvoice,
} from '@/api/modules/expense-invoice'
import type { ExpenseInvoice } from '@/api/modules/expense-invoice'
import DocStatusTag from '@/components/business/DocStatusTag.vue'

const route = useRoute()
const router = useRouter()
const invoice = ref<ExpenseInvoice | null>(null)
const loading = ref(false)
const actionLoading = ref(false)

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

/* ********** 枚举显示 ********** */
const INVOICE_KIND_MAP: Record<string, string> = {
  paper_normal: '纸质普票',
  paper_special: '纸质专票',
  electronic_normal: '电子普票',
  electronic_special: '电子专票',
}
function invoiceKindLabel(v?: string): string {
  if (!v) return '—'
  return INVOICE_KIND_MAP[v] || v
}

const CATEGORY_MAP: Record<string, string> = {
  transport: '交通',
  office: '办公',
  travel: '差旅',
  entertain: '招待',
  communication: '通讯',
  training: '培训',
  welfare: '福利',
  other: '其他',
}
function categoryLabel(v?: string): string {
  if (!v) return '—'
  return CATEGORY_MAP[v] || v
}

const SOURCE_MAP: Record<string, string> = {
  manual: '手工录入',
  import: '批量导入',
  ocr: 'OCR 识别',
}
function sourceLabel(v?: string): string {
  if (!v) return '—'
  return SOURCE_MAP[v] || v
}
function sourceTagType(v?: string): 'primary' | 'success' | 'warning' | 'info' {
  switch (v) {
    case 'manual': return 'primary'
    case 'import': return 'success'
    case 'ocr': return 'warning'
    default: return 'info'
  }
}

const VERIFY_MAP: Record<string, string> = {
  unverified: '未验真',
  verified: '已验真',
  invalid: '验真失败',
}
function verifyLabel(v?: string): string {
  if (!v) return '未验真'
  return VERIFY_MAP[v] || v
}
function verifyTagType(v?: string): 'primary' | 'success' | 'danger' | 'info' {
  switch (v) {
    case 'verified': return 'success'
    case 'invalid': return 'danger'
    case 'unverified':
    default: return 'info'
  }
}

/* ********** 加载与操作 ********** */
async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchExpenseInvoiceDetail(route.params.id as string)
    invoice.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleConfirm() {
  if (!invoice.value) return
  actionLoading.value = true
  try {
    await confirmExpenseInvoice(invoice.value.id)
    ElMessage.success('确认入库成功')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '确认入库失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleVerify() {
  if (!invoice.value) return
  actionLoading.value = true
  try {
    // 当前为 Mock 实现：后端固定返回 verified
    const res: any = await verifyExpenseInvoice(invoice.value.id)
    const data = res?.data || res
    ElMessage.success(
      data?.verify_status === 'invalid' ? '验真失败' : '验真成功',
    )
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '验真失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleDelete() {
  if (!invoice.value) return
  actionLoading.value = true
  try {
    await deleteExpenseInvoice(invoice.value.id)
    ElMessage.success('删除成功')
    goBack()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  } finally {
    actionLoading.value = false
  }
}

function goEdit() {
  if (!invoice.value) return
  router.push(`/expense-invoices/edit/${invoice.value.id}`)
}

function goVoucher(voucherId: string) {
  router.push(`/vouchers/${voucherId}`)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/expense-invoices/list')
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.expense-invoice-detail-page {
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
}
</style>
