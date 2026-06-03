<template>
  <div class="manual-match">
    <div class="page-header"><h3>手工核销</h3></div>
    <el-row :gutter="16">
      <el-col :span="10">
        <el-card><template #header>可选发票</template>
          <el-table :data="availableInvoices" size="small" border @row-click="addAllocation">
            <el-table-column prop="invoice_no" label="发票号" width="120" />
            <el-table-column prop="outstanding" label="未结清" width="100" align="right" />
            <el-table-column label="可选" width="60">
              <template #default="{ row }"><el-checkbox :model-value="isSelected(row)" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card><template #header>已分配核销</template>
          <el-table :data="allocations" size="small" border>
            <el-table-column prop="invoice_no" label="发票号" width="120" />
            <el-table-column prop="outstanding" label="未结清" width="90" align="right" />
            <el-table-column label="本次核销" width="140">
              <template #default="{ row }">
                <AmountInput v-model="row.amount" :max="row.outstanding" />
              </template>
            </el-table-column>
            <el-table-column label="核销后余额" width="100" align="right">
              <template #default="{ row }">¥{{ (parseFloat(row.outstanding.replace(/,/g, '')) - parseFloat(row.amount || '0')).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="60">
              <template #default="{ $index }"><el-button link type="danger" size="small" @click="allocations.splice($index, 1)">移除</el-button></template>
            </el-table-column>
          </el-table>
          <div class="allocation-summary">
            <span>合计核销: <b>¥{{ totalAmount }}</b></span>
            <span class="remaining">剩余可分配: <b>¥{{ remaining }}</b></span>
          </div>
          <el-button type="primary" class="exec-btn" :disabled="allocations.length === 0" @click="execute">执行核销</el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface AvailableInvoice { id: string; invoice_no: string; outstanding: string }
const availableInvoices = ref<AvailableInvoice[]>([])

interface Allocation { invoice_no: string; outstanding: string; amount: string }
const allocations = ref<Allocation[]>([])

const currentPaymentAmount = ref(0)

async function loadInvoices() {
  try {
    const res: any = await request.get('/invoices', { params: { status: 'unpaid' } })
    const list: any[] = res.data?.items ?? res.data ?? []
    availableInvoices.value = list.map(inv => ({
      id: inv.id,
      invoice_no: inv.invoice_number ?? inv.invoice_no ?? '',
      outstanding: (inv.outstanding_amount ?? inv.outstanding ?? 0).toString(),
    }))
  } catch {
    availableInvoices.value = []
  }
}

onMounted(() => {
  loadInvoices()
})

const totalAmount = computed(() => {
  const sum = allocations.value.reduce((a, b) => a + parseFloat(b.amount || '0'), 0)
  return sum.toFixed(2)
})
const remaining = computed(() => (currentPaymentAmount.value - parseFloat(totalAmount.value)).toFixed(2))

function isSelected(row: any) { return allocations.value.some(a => a.invoice_no === row.invoice_no) }

function addAllocation(row: any) {
  if (isSelected(row)) return
  allocations.value.push({ invoice_no: row.invoice_no, outstanding: row.outstanding, amount: '' })
}

async function execute() {
  if (allocations.value.length === 0) return
  try {
    await request.post('/reconciliation/manual', {
      allocations: allocations.value.map(a => ({
        invoice_no: a.invoice_no,
        amount: parseFloat(a.amount || '0'),
      })),
    })
    ElMessage.success('核销执行成功！')
    allocations.value = []
    loadInvoices()
  } catch (e: any) {
    ElMessage.error(e?.message || '核销失败')
  }
}
</script>
<style scoped>
.page-header h3 { font-size: 18px; margin-bottom: 16px; }
.allocation-summary { margin-top: 12px; display: flex; gap: 24px; font-size: 14px; }
.remaining { color: #999; }
.exec-btn { margin-top: 12px; }
</style>
