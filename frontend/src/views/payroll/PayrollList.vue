<template>
  <div class="payroll-page">
    <div class="page-header">
      <h3>工资单</h3>
      <div class="header-actions">
        <el-button type="primary" @click="goCreate">新建工资单</el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="工资期间">
          <el-input v-model="filter.periodNo" placeholder="如 202506" clearable style="width: 120px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 130px" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已提交" value="submitted" />
            <el-option label="已审核" value="approved" />
            <el-option label="已过账" value="posted" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="员工姓名/部门" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="payrolls" border stripe size="small" v-loading="loading">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="period_no" label="工资期间" width="100">
          <template #default="{ row }">{{ String(row.period_no).slice(0, 4) }}-{{ String(row.period_no).slice(4, 6) }}</template>
        </el-table-column>
        <el-table-column prop="employee_name" label="员工姓名" min-width="100" />
        <el-table-column prop="department_name" label="部门" min-width="100" />
        <el-table-column prop="gross_salary" label="应发工资" width="120" align="right">
          <template #default="{ row }">{{ formatAmount(row.gross_salary) }}</template>
        </el-table-column>
        <el-table-column label="代扣合计" width="120" align="right">
          <template #default="{ row }">{{ formatAmount(calcDeductions(row)) }}</template>
        </el-table-column>
        <el-table-column prop="net_salary" label="实发工资" width="120" align="right">
          <template #default="{ row }">
            <b>{{ formatAmount(row.net_salary) }}</b>
          </template>
        </el-table-column>
        <el-table-column prop="payment_date" label="发放日期" width="100" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.doc_status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
            <el-button v-if="row.doc_status === 0" link type="warning" size="small" @click="goEdit(row)">编辑</el-button>
            <el-button v-if="row.doc_status === 0" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { fetchPayrollList, deletePayroll } from '@/api/modules/payroll'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { Payroll } from '@/types/models/payroll'

const router = useRouter()
const loading = ref(false)
const payrolls = ref<Payroll[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive({
  periodNo: '',
  status: '',
  keyword: '',
})

function calcDeductions(row: Payroll): string {
  return (parseFloat(row.individual_tax || '0') +
    parseFloat(row.social_security || '0') +
    parseFloat(row.housing_fund || '0') +
    parseFloat(row.other_deductions || '0')).toFixed(2)
}

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchPayrollList({
      page: page.value,
      pageSize: pageSize.value,
      period_no: filter.periodNo || undefined,
      status: filter.status || undefined,
      keyword: filter.keyword || undefined,
    })
    payrolls.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total || payrolls.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    payrolls.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.periodNo = ''
  filter.status = ''
  filter.keyword = ''
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}

function goCreate() {
  router.push('/payroll/new')
}

function goDetail(row: Payroll) {
  router.push(`/payroll/${row.id}`)
}

function goEdit(row: Payroll) {
  router.push(`/payroll/${row.id}/edit`)
}

async function handleDelete(row: Payroll) {
  try {
    await ElMessageBox.confirm('确认删除该工资单？', '删除确认', { type: 'warning' })
    await deletePayroll(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.response?.data?.error || '删除失败')
    }
  }
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
</style>