<template>
  <div class="reimbursement-page">
    <div class="page-header">
      <h3>报销单</h3>
      <div class="header-actions">
        <el-button type="primary" @click="goCreate">新建报销单</el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="申请人">
          <el-input v-model="filter.employee_name" placeholder="申请人姓名" clearable style="width: 130px" />
        </el-form-item>
        <el-form-item label="部门">
          <el-input v-model="filter.department" placeholder="部门" clearable style="width: 130px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 130px" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已提交" value="submitted" />
            <el-option label="已审核" value="approved" />
            <el-option label="已驳回" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="备注/说明" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="reimbursements" border stripe size="small" v-loading="loading">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="employee_name" label="申请人" min-width="100" />
        <el-table-column prop="department" label="部门" min-width="120" />
        <el-table-column label="报销金额" width="130" align="right">
          <template #default="{ row }">
            <b>{{ formatAmount(row.amount) }}</b>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.docstatus" />
          </template>
        </el-table-column>
        <el-table-column prop="voucher_id" label="凭证" width="100">
          <template #default="{ row }">
            <template v-if="row.voucher_id">
              <el-link type="primary" :underline="false" size="small" @click="goVoucher(row.voucher_id)">
                {{ row.voucher_id }}
              </el-link>
            </template>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
            <el-button v-if="row.docstatus === 0" link type="warning" size="small" @click="goEdit(row)">编辑</el-button>
            <el-button v-if="row.docstatus === 0" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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
import { fetchReimbursementList, deleteReimbursement } from '@/api/modules/reimbursement'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { Reimbursement } from '@/api/modules/reimbursement'

const router = useRouter()
const loading = ref(false)
const reimbursements = ref<Reimbursement[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive({
  employee_name: '',
  department: '',
  status: '',
  keyword: '',
})

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchReimbursementList({
      page: page.value,
      pageSize: pageSize.value,
      employee_name: filter.employee_name || undefined,
      department: filter.department || undefined,
      status: filter.status || undefined,
      keyword: filter.keyword || undefined,
    })
    reimbursements.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total || reimbursements.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    reimbursements.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.employee_name = ''
  filter.department = ''
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
  router.push('/expense/reimbursement/new')
}

function goDetail(row: Reimbursement) {
  router.push(`/expense/reimbursement/${row.id}`)
}

function goEdit(row: Reimbursement) {
  router.push(`/expense/reimbursement/${row.id}/edit`)
}

function goVoucher(voucherId: string) {
  router.push(`/vouchers/${voucherId}`)
}

async function handleDelete(row: Reimbursement) {
  try {
    await ElMessageBox.confirm('确认删除该报销单？', '删除确认', { type: 'warning' })
    await deleteReimbursement(row.id)
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