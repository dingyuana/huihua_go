<template>
  <div class="advance-receipt-page">
    <div class="page-header">
      <h3>预收款单</h3>
      <p class="page-hint">先款后票流程：客户预付 → 自动生成凭证 → 后续开票时冲抵</p>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 160px" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已确认" value="confirmed" />
            <el-option label="部分已分配" value="partially_allocated" />
            <el-option label="已全部分配" value="fully_allocated" />
            <el-option label="已冲销" value="reversed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
          <el-button type="success" @click="openCreateDialog">新建预收单</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ stats.total }}</p>
          <p class="stat-label">单据总数</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card draft">
          <p class="stat-num">¥{{ stats.totalAmount }}</p>
          <p class="stat-label">预收总额</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card confirmed">
          <p class="stat-num">¥{{ stats.allocatedAmount }}</p>
          <p class="stat-label">已分配</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">¥{{ stats.outstandingAmount }}</p>
          <p class="stat-label">未分配</p>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <el-table :data="filteredList" border stripe size="small" v-loading="loading">
        <el-table-column prop="advance_no" label="预收单号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="customer_id" label="客户ID" width="280" show-overflow-tooltip />
        <el-table-column label="金额" width="140" align="right">
          <template #default="{ row }"><span>¥{{ formatAmount(row.amount) }}</span></template>
        </el-table-column>
        <el-table-column label="已分配" width="120" align="right">
          <template #default="{ row }"><span>¥{{ formatAmount(row.allocated_amount) }}</span></template>
        </el-table-column>
        <el-table-column label="未分配" width="120" align="right">
          <template #default="{ row }"><span class="amount-out">¥{{ formatAmount(row.outstanding_amount) }}</span></template>
        </el-table-column>
        <el-table-column prop="received_date" label="收款日期" width="110" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="voucher_no" label="凭证号" width="120" show-overflow-tooltip />
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'draft'" link type="primary" size="small" @click="handleConfirm(row)">确认</el-button>
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="showDrawer" :title="`预收单 ${currentItem?.advance_no || ''}`" size="500px">
      <template v-if="currentItem">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="预收单号">{{ currentItem.advance_no }}</el-descriptions-item>
          <el-descriptions-item label="客户ID">{{ currentItem.customer_id }}</el-descriptions-item>
          <el-descriptions-item label="金额">¥{{ formatAmount(currentItem.amount) }}</el-descriptions-item>
          <el-descriptions-item label="已分配">¥{{ formatAmount(currentItem.allocated_amount) }}</el-descriptions-item>
          <el-descriptions-item label="未分配">¥{{ formatAmount(currentItem.outstanding_amount) }}</el-descriptions-item>
          <el-descriptions-item label="收款日期">{{ currentItem.received_date }}</el-descriptions-item>
          <el-descriptions-item label="到期日">{{ currentItem.due_date || '—' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="凭证号">{{ currentItem.voucher_no || '—' }}</el-descriptions-item>
          <el-descriptions-item label="备注">{{ currentItem.remark || '—' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentItem.created_at }}</el-descriptions-item>
          <el-descriptions-item label="确认时间">{{ currentItem.confirmed_at || '—' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-drawer>

    <el-dialog v-model="createDialogVisible" title="新建预收单" width="520px">
      <el-form :model="createForm" label-width="100px" size="small">
        <el-form-item label="客户ID" required>
          <el-input v-model="createForm.customer_id" placeholder="客户 UUID" />
        </el-form-item>
        <el-form-item label="公司ID" required>
          <el-input v-model="createForm.company_id" placeholder="公司 UUID" />
        </el-form-item>
        <el-form-item label="金额" required>
          <el-input v-model="createForm.amount" placeholder="例如 1000.00" />
        </el-form-item>
        <el-form-item label="收款日期" required>
          <el-date-picker v-model="createForm.received_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="到期日">
          <el-date-picker v-model="createForm.due_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">保存为草稿</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchAdvanceReceipts, createAdvanceReceipt, confirmAdvanceReceipt,
  type AdvanceReceipt,
} from '@/api/modules/advance_receipt'

const loading = ref(false)
const list = ref<AdvanceReceipt[]>([])
const filter = reactive({ status: '' })

const filteredList = computed(() => {
  if (!filter.status) return list.value
  return list.value.filter(r => r.status === filter.status)
})

const stats = computed(() => {
  const all = filteredList.value
  const totalAmount = all.reduce((s, r) => s + (Number(r.amount) || 0), 0)
  const allocatedAmount = all.reduce((s, r) => s + (Number(r.allocated_amount) || 0), 0)
  const outstandingAmount = all.reduce((s, r) => s + (Number(r.outstanding_amount) || 0), 0)
  return {
    total: all.length,
    totalAmount: totalAmount.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    allocatedAmount: allocatedAmount.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    outstandingAmount: outstandingAmount.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
  }
})

const showDrawer = ref(false)
const currentItem = ref<AdvanceReceipt | null>(null)

const createDialogVisible = ref(false)
const creating = ref(false)
const createForm = reactive({
  customer_id: '', company_id: '', amount: '',
  received_date: '', due_date: '', remark: '',
})

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchAdvanceReceipts()
    const data = res?.data || res
    list.value = data?.list || []
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    list.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() { filter.status = '' }

function showDetail(row: AdvanceReceipt) {
  currentItem.value = row
  showDrawer.value = true
}

function openCreateDialog() {
  Object.assign(createForm, { customer_id: '', company_id: '', amount: '', received_date: '', due_date: '', remark: '' })
  createDialogVisible.value = true
}

async function handleCreate() {
  if (!createForm.customer_id || !createForm.amount || !createForm.received_date) {
    ElMessage.warning('请填写必填项')
    return
  }
  creating.value = true
  try {
    await createAdvanceReceipt({
      customer_id: createForm.customer_id,
      company_id: createForm.company_id,
      amount: createForm.amount,
      received_date: createForm.received_date,
      due_date: createForm.due_date || undefined,
      remark: createForm.remark || undefined,
    })
    ElMessage.success('预收单创建成功')
    createDialogVisible.value = false
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

async function handleConfirm(row: AdvanceReceipt) {
  try {
    await confirmAdvanceReceipt(row.id)
    ElMessage.success('预收单已确认，凭证已生成')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '确认失败')
  }
}

function statusLabel(s: string): string {
  const map: Record<string, string> = {
    draft: '草稿', confirmed: '已确认',
    partially_allocated: '部分已分配', fully_allocated: '已全部分配',
    reversed: '已冲销',
  }
  return map[s] || s
}

function statusTag(s: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' {
  const map: Record<string, 'success' | 'warning' | 'info' | 'primary' | 'danger'> = {
    draft: 'warning', confirmed: 'primary',
    partially_allocated: 'warning', fully_allocated: 'success',
    reversed: 'info',
  }
  return map[s] || 'info'
}

function formatAmount(val: any): string {
  const n = Number(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.advance-receipt-page {
  .page-header { margin-bottom: 16px; h3 { font-size: 18px; margin: 0 0 4px; } .page-hint { font-size: 12px; color: #999; margin: 0; } }
  .filter-card { margin-bottom: 12px; }
  .stat-row { margin-bottom: 12px; }
  .stat-card { text-align: center; .stat-num { font-size: 22px; font-weight: 700; margin-bottom: 4px; color: #333; } .stat-label { font-size: 12px; color: #999; } &.draft .stat-num { color: #d48806; } &.confirmed .stat-num { color: #389e0d; } }
  .amount-out { color: #d4380d; font-weight: 600; }
}
</style>
