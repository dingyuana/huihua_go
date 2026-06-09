<template>
  <div class="payroll-page">
    <div class="page-header">
      <h3>工资单</h3>
      <div class="header-actions">
        <el-button @click="showGenerateDialog = true">生成薪资凭证</el-button>
        <el-button @click="showSocialDialog = true">社保计提</el-button>
        <el-button @click="showTaxDialog = true">个税计提</el-button>
        <el-button @click="handleExportSalary">导出工资单</el-button>
        <el-button @click="handleExportTax">导出一个税明细</el-button>
        <el-button type="primary" @click="goCreate">新建工资单</el-button>
      </div>
    </div>

    <!-- 生成薪资凭证对话框 -->
    <el-dialog v-model="showGenerateDialog" title="生成薪资凭证" width="420px">
      <el-form :model="generateForm" label-width="100px">
        <el-form-item label="会计年份" required>
          <el-select v-model="generateForm.year" placeholder="选择年份" style="width: 100%">
            <el-option v-for="y in availableYears" :key="y" :label="`${y}年`" :value="y" />
          </el-select>
        </el-form-item>
        <el-form-item label="会计月份" required>
          <el-select v-model="generateForm.month" placeholder="选择月份" style="width: 100%">
            <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" :loading="generating" @click="handleGeneratePeriodVouchers">确认生成</el-button>
      </template>
    </el-dialog>

    <!-- 社保计提对话框 -->
    <el-dialog v-model="showSocialDialog" title="社保计提" width="420px">
      <el-form :model="socialForm" label-width="100px">
        <el-form-item label="会计年份" required>
          <el-select v-model="socialForm.year" placeholder="选择年份" style="width: 100%">
            <el-option v-for="y in availableYears" :key="y" :label="`${y}年`" :value="y" />
          </el-select>
        </el-form-item>
        <el-form-item label="会计月份" required>
          <el-select v-model="socialForm.month" placeholder="选择月份" style="width: 100%">
            <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
          </el-select>
        </el-form-item>
      </el-form>
      <div v-if="socialResult" style="margin: 0 0 16px 20px; padding: 12px; background: #f5f7fa; border-radius: 6px; line-height: 2;">
        <div>养老保险（单位）：<b>¥{{ socialResult.total_social }}</b></div>
        <div>住房公积金（单位）：<b>¥{{ socialResult.total_housing }}</b></div>
        <div>单位合计：<b>¥{{ socialResult.total_employer }}</b></div>
        <div v-if="socialResult.voucher">凭证号：<b>{{ socialResult.voucher.voucher_no }}</b></div>
      </div>
      <template #footer>
        <el-button @click="showSocialDialog = false">取消</el-button>
        <el-button type="primary" :loading="socialCalculating" @click="handleCalculateSocial">确认计提</el-button>
      </template>
    </el-dialog>

    <!-- 个税计提对话框 -->
    <el-dialog v-model="showTaxDialog" title="个税计提" width="520px">
      <el-form :model="taxForm" label-width="100px">
        <el-form-item label="会计年份" required>
          <el-select v-model="taxForm.year" placeholder="选择年份" style="width: 100%">
            <el-option v-for="y in availableYears" :key="y" :label="`${y}年`" :value="y" />
          </el-select>
        </el-form-item>
        <el-form-item label="会计月份" required>
          <el-select v-model="taxForm.month" placeholder="选择月份" style="width: 100%">
            <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
          </el-select>
        </el-form-item>
      </el-form>
      <div v-if="taxResult" style="margin: 0 0 16px 20px; padding: 12px; background: #f5f7fa; border-radius: 6px; line-height: 2;">
        <div>计提总人数：<b>{{ taxResult.total_employees }}</b></div>
        <div>个税总额：<b>¥{{ taxResult.total_tax }}</b></div>
        <div v-if="taxResult.voucher">凭证号：<b>{{ taxResult.voucher.voucher_no }}</b></div>
        <el-collapse style="margin-top: 8px">
          <el-collapse-item title="查看详情" v-if="taxResult.details && taxResult.details.length">
            <el-table :data="taxResult.details" size="small" border stripe max-height="300">
              <el-table-column prop="employee_name" label="员工姓名" width="100" />
              <el-table-column prop="taxable_income" label="应纳税所得额" width="130" align="right">
                <template #default="{ row }">{{ formatAmount(row.taxable_income) }}</template>
              </el-table-column>
              <el-table-column prop="tax_rate" label="税率" width="70" align="right">
                <template #default="{ row }">{{ row.tax_rate }}%</template>
              </el-table-column>
              <el-table-column prop="tax_amount" label="税额" width="110" align="right">
                <template #default="{ row }">{{ formatAmount(row.tax_amount) }}</template>
              </el-table-column>
            </el-table>
          </el-collapse-item>
        </el-collapse>
      </div>
      <template #footer>
        <el-button @click="showTaxDialog = false">取消</el-button>
        <el-button type="primary" :loading="taxCalculating" @click="handleCalculateTax">确认计提</el-button>
      </template>
    </el-dialog>

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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { fetchPayrollList, deletePayroll, generatePeriodVouchers, calculatePeriodSocial, calculatePeriodTax, exportSalaryExcel, exportTaxExcel } from '@/api/modules/payroll'
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

// 生成薪资凭证对话框
const showGenerateDialog = ref(false)
const generating = ref(false)
const generateForm = reactive({
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
})
const availableYears = computed(() => {
  const y = new Date().getFullYear()
  return [y - 2, y - 1, y, y + 1]
})

async function handleGeneratePeriodVouchers() {
  if (!generateForm.year || !generateForm.month) {
    ElMessage.warning('请选择会计年份和月份')
    return
  }
  const periodNo = generateForm.year * 100 + generateForm.month
  generating.value = true
  try {
    const res: any = await generatePeriodVouchers(periodNo)
    const vouchers = res?.vouchers || []
    ElMessage.success(`已生成 ${vouchers.length} 张凭证：${vouchers.map((v: any) => v.voucher_no).join('、')}`)
    showGenerateDialog.value = false
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '生成失败')
  } finally {
    generating.value = false
  }
}

// 社保计提对话框
const showSocialDialog = ref(false)
const socialCalculating = ref(false)
const socialResult = ref<any>(null)
const socialForm = reactive({
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
})

async function handleCalculateSocial() {
  if (!socialForm.year || !socialForm.month) {
    ElMessage.warning('请选择会计年份和月份')
    return
  }
  const periodNo = socialForm.year * 100 + socialForm.month
  socialCalculating.value = true
  socialResult.value = null
  try {
    const res: any = await calculatePeriodSocial(periodNo)
    socialResult.value = res?.result || res
    ElMessage.success('社保计提完成')
  } catch (e: any) {
    ElMessage.error('社保计提失败：' + (e?.response?.data?.error || e?.message || e))
  } finally {
    socialCalculating.value = false
  }
}

// 个税计提对话框
const showTaxDialog = ref(false)
const taxCalculating = ref(false)
const taxResult = ref<any>(null)
const taxForm = reactive({
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
})

async function handleCalculateTax() {
  if (!taxForm.year || !taxForm.month) {
    ElMessage.warning('请选择会计年份和月份')
    return
  }
  const periodNo = taxForm.year * 100 + taxForm.month
  taxCalculating.value = true
  taxResult.value = null
  try {
    const res: any = await calculatePeriodTax(periodNo)
    taxResult.value = res?.result || res
    ElMessage.success('个税计提完成')
  } catch (e: any) {
    ElMessage.error('个税计提失败：' + (e?.response?.data?.error || e?.message || e))
  } finally {
    taxCalculating.value = false
  }
}

function getPeriodNo(): number {
  if (filter.periodNo) {
    return parseInt(filter.periodNo, 10)
  }
  const now = new Date()
  return now.getFullYear() * 100 + (now.getMonth() + 1)
}

function handleExportSalary() {
  exportSalaryExcel(getPeriodNo())
}

function handleExportTax() {
  exportTaxExcel(getPeriodNo())
}

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
    payrolls.value = res?.payrolls || []
    total.value = res?.total || payrolls.value.length
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