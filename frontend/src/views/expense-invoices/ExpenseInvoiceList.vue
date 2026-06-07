<template>
  <div class="expense-invoice-page">
    <div class="page-header">
      <h3>进项发票</h3>
      <div class="header-actions">
        <el-button type="primary" @click="goCreate">新增发票</el-button>
        <el-button @click="goImport">导入向导</el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="单据状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 140px" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已确认" value="confirmed" />
          </el-select>
        </el-form-item>
        <el-form-item label="开票日期">
          <el-date-picker
            v-model="filter.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input
            v-model="filter.keyword"
            placeholder="发票号/供应商/备注"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="invoices" border stripe size="small" v-loading="loading">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="invoice_no" label="发票号码" min-width="160" show-overflow-tooltip />
        <el-table-column prop="invoice_code" label="发票代码" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.invoice_code || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="invoice_date" label="开票日期" width="120" />
        <el-table-column label="发票类型" width="120">
          <template #default="{ row }">
            {{ invoiceKindLabel(row.invoice_kind) }}
          </template>
        </el-table-column>
        <el-table-column prop="vendor_name" label="供应商" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.vendor_name || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="价税合计" width="140" align="right">
          <template #default="{ row }">
            <b>¥{{ formatAmount(row.total_amount) }}</b>
          </template>
        </el-table-column>
        <el-table-column label="验真状态" width="100">
          <template #default="{ row }">
            <el-tag :type="verifyStatusTag(row.verify_status)" size="small">
              {{ verifyStatusLabel(row.verify_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="单据状态" width="100">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.docstatus" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
            <el-button
              v-if="row.docstatus === 0"
              link
              type="warning"
              size="small"
              @click="goEdit(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              v-if="row.docstatus === 0"
              title="确认删除该进项发票？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button link type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
            <el-button
              v-if="row.docstatus === 0"
              link
              type="success"
              size="small"
              @click="handleConfirm(row)"
            >
              确认
            </el-button>
            <el-button link type="primary" size="small" @click="handleVerify(row)">验真</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import {
  fetchExpenseInvoiceList,
  deleteExpenseInvoice,
  confirmExpenseInvoice,
  verifyExpenseInvoice,
  type ExpenseInvoice,
} from '@/api/modules/expense-invoice'

const router = useRouter()
const loading = ref(false)
const invoices = ref<ExpenseInvoice[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive<{
  status: string
  dateRange: [string, string] | null
  start_date: string
  end_date: string
  keyword: string
}>({
  status: '',
  dateRange: null,
  start_date: '',
  end_date: '',
  keyword: '',
})

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function invoiceKindLabel(k: any): string {
  switch (k) {
    case 'paper_normal':
      return '纸质普票'
    case 'paper_special':
      return '纸质专票'
    case 'electronic_normal':
      return '电子普票'
    case 'electronic_special':
      return '电子专票'
    default:
      return k || '—'
  }
}

function verifyStatusLabel(s: any): string {
  switch (s) {
    case 'verified':
      return '已验真'
    case 'invalid':
      return '验真失败'
    case 'unverified':
      return '未验真'
    default:
      return s || '未验真'
  }
}

function verifyStatusTag(s: any): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  switch (s) {
    case 'verified':
      return 'success'
    case 'invalid':
      return 'danger'
    case 'unverified':
      return 'info'
    default:
      return 'info'
  }
}

async function loadData() {
  loading.value = true
  // 同步日期范围到 start_date/end_date
  if (filter.dateRange && filter.dateRange.length === 2) {
    filter.start_date = filter.dateRange[0]
    filter.end_date = filter.dateRange[1]
  } else {
    filter.start_date = ''
    filter.end_date = ''
  }
  try {
    const res: any = await fetchExpenseInvoiceList({
      page: page.value,
      pageSize: pageSize.value,
      status: filter.status || undefined,
      start_date: filter.start_date || undefined,
      end_date: filter.end_date || undefined,
      keyword: filter.keyword || undefined,
    })
    invoices.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total ?? invoices.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    invoices.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.status = ''
  filter.dateRange = null
  filter.start_date = ''
  filter.end_date = ''
  filter.keyword = ''
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}

function goCreate() {
  router.push('/expense-invoices/create')
}

function goImport() {
  // 导入向导页面尚未实现，先占位跳转
  ElMessage.info('导入向导页面待开发')
  // router.push('/expense-invoices/import')
}

function goDetail(row: ExpenseInvoice) {
  router.push(`/expense-invoices/detail/${row.id}`)
}

function goEdit(row: ExpenseInvoice) {
  router.push(`/expense-invoices/edit/${row.id}`)
}

async function handleDelete(row: ExpenseInvoice) {
  try {
    await deleteExpenseInvoice(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  }
}

async function handleConfirm(row: ExpenseInvoice) {
  try {
    await confirmExpenseInvoice(row.id)
    ElMessage.success('确认成功')
    loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '确认失败')
  }
}

async function handleVerify(row: ExpenseInvoice) {
  try {
    const res: any = await verifyExpenseInvoice(row.id)
    const data = res?.data || res
    ElMessage.success(`验真完成：${data?.verify_status || '已发起'}`)
    loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '验真失败')
  }
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.expense-invoice-page {
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    h3 { font-size: 18px; margin: 0; }
    .header-actions { display: flex; gap: 8px; }
  }
  .filter-card { margin-bottom: 12px; }
}
</style>
