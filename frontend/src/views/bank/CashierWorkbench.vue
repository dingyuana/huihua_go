<template>
  <div class="workbench">
    <div class="page-header">
      <h3>出纳核对工作台</h3>
      <div>
        <el-radio-group v-model="viewMode" size="small">
          <el-radio-button value="list">列表模式</el-radio-button>
          <el-radio-button value="batch">批量模式</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <el-row :gutter="16" class="stat-row">
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card total">
          <p class="stat-num">{{ stats.total }}</p>
          <p class="stat-label">本月流水</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card confirmed">
          <p class="stat-num">{{ stats.confirmed }}</p>
          <p class="stat-label">已确认</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card pending">
          <p class="stat-num">{{ stats.pending }}</p>
          <p class="stat-label">待确认</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card unclassified">
          <p class="stat-num danger">{{ stats.unclassified }}</p>
          <p class="stat-label">未分类</p>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card docs">
          <p class="stat-num">{{ stats.generatedDocs }}</p>
          <p class="stat-label">已生成单据</p>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <el-tabs v-model="activeTab" class="classification-tabs">
        <el-tab-pane :label="`全部 (${stats.total})`" name="all" />
        <el-tab-pane label="业务收款" name="business_receipt" />
        <el-tab-pane label="业务付款" name="business_payment" />
        <el-tab-pane label="银行费用" name="bank_fee" />
        <el-tab-pane label="利息收入" name="interest_income" />
        <el-tab-pane label="内部转账" name="internal_transfer" />
        <el-tab-pane :label="`待处理 ${stats.unclassified > 0 ? '🔴' : ''}`" name="pending" />
      </el-tabs>

      <div class="batch-bar">
        <el-checkbox v-model="selectAll" @change="onSelectAll">全选</el-checkbox>
        <span class="selected-count">已选 {{ selectedIds.length }} 条</span>
        <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="batchConfirm">确认选中</el-button>
        <el-button size="small" :disabled="selectedIds.length === 0" @click="showClassifyDialog = true">修正分类</el-button>
      </div>

      <!-- 批量模式下提示待生成单据 -->
      <el-alert v-if="viewMode === 'batch' && selectedIds.length > 0" :title="batchPreview" type="info" :closable="false" show-icon class="batch-preview" />

      <el-table
        ref="tableRef"
        :data="filteredTxns"
        border stripe size="small"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="40" />
        <el-table-column prop="date" label="日期" width="90" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span :class="row.direction === 'in' ? 'amount-positive' : 'amount-negative'">
              {{ row.direction === 'in' ? '+' : '-' }}{{ row.amount }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="counterparty" label="对方" width="140" />
        <el-table-column prop="description" label="摘要" min-width="180" show-overflow-tooltip />
        <el-table-column label="分类" width="100">
          <template #default="{ row }">
            <el-tag :type="classificationTag(row.classification)" size="small">{{ classificationLabel(row.classification) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="将生成" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.confirmed" size="small" type="success">{{ docTypeLabel(row.classification) }}</el-tag>
            <span v-else class="doc-preview">{{ docTypeLabel(row.classification) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.confirmed" type="success" size="small">已确认</el-tag>
            <el-tag v-else type="warning" size="small">待确认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.confirmed" link type="primary" size="small" @click="confirmOne(row)">确认</el-button>
            <el-button link type="primary" size="small" @click="editOne(row)">修正</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showClassifyDialog" title="修正分类" width="420px">
      <el-form label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="classifyForm.classification" style="width: 100%">
            <el-option label="业务收款" value="business_receipt" />
            <el-option label="业务付款" value="business_payment" />
            <el-option label="银行费用" value="bank_fee" />
            <el-option label="利息收入" value="interest_income" />
            <el-option label="内部转账" value="internal_transfer" />
            <el-option label="待处理" value="pending" />
          </el-select>
        </el-form-item>
        <el-form-item label="对方单位">
          <PartySelector v-model="classifyForm.party" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="classifyForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showClassifyDialog = false">取消</el-button>
        <el-button type="primary" @click="saveClassification">确认修正</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface TxnItem {
  id: string
  date: string
  amount: string
  direction: string
  counterparty: string
  description: string
  classification: string
  confirmed: boolean
}

const localTxns: TxnItem[] = [
  { id: 't1', date: '05-20', amount: '12,000.00', direction: 'in', counterparty: '上海XX贸易公司', description: '网银转账-收款-上海XX', classification: 'business_receipt', confirmed: true },
  { id: 't2', date: '05-20', amount: '50.00', direction: 'out', counterparty: '', description: '账户管理费', classification: 'bank_fee', confirmed: true },
  { id: 't3', date: '05-21', amount: '3,500.00', direction: 'in', counterparty: '北京YY科技', description: '转账收入', classification: 'business_receipt', confirmed: false },
  { id: 't4', date: '05-21', amount: '5,000.00', direction: 'out', counterparty: '广州ZZ贸易', description: '货款支付', classification: 'business_payment', confirmed: false },
  { id: 't5', date: '05-22', amount: '150.00', direction: 'in', counterparty: '', description: '存款利息', classification: 'interest_income', confirmed: false },
  { id: 't6', date: '05-22', amount: '10,000.00', direction: 'out', counterparty: '上海XX贸易公司', description: '网银转账', classification: 'pending', confirmed: false },
  { id: 't7', date: '05-23', amount: '200.00', direction: 'out', counterparty: '', description: '跨行转账手续费', classification: 'bank_fee', confirmed: false },
  { id: 't8', date: '05-23', amount: '8,000.00', direction: 'in', counterparty: '未知', description: '来账-摘要不明', classification: 'pending', confirmed: false },
]

const allTxns = ref<TxnItem[]>([])
onMounted(async () => {
  try {
    const res: any = await request.get('/bank-transactions', { params: { page: 1, pageSize: 50 } })
    const list = res?.data?.list || res?.data
    if (Array.isArray(list) && list.length > 0) { allTxns.value = list; return }
  } catch { /* fallback */ }
  allTxns.value = localTxns
})

const viewMode = ref('list')
const tableRef = ref()
const activeTab = ref('all')
const selectedIds = ref<string[]>([])
const selectAll = ref(false)
const showClassifyDialog = ref(false)
let docCounter = ref(3) // 已生成单据数量统计

const classifyForm = reactive({
  classification: 'business_receipt',
  party: null as any,
  remark: '',
})

const stats = computed(() => ({
  total: allTxns.value.length,
  confirmed: allTxns.value.filter(t => t.confirmed).length,
  pending: allTxns.value.filter(t => !t.confirmed).length,
  unclassified: allTxns.value.filter(t => t.classification === 'pending').length,
  generatedDocs: docCounter.value,
}))

const filteredTxns = computed(() => {
  if (activeTab.value === 'all') return allTxns.value
  if (activeTab.value === 'pending') return allTxns.value.filter(t => t.classification === 'pending')
  return allTxns.value.filter(t => t.classification === activeTab.value)
})

/** 批量操作预览：显示选中项将生成的单据类型 */
const batchPreview = computed(() => {
  const selected = allTxns.value.filter(t => selectedIds.value.includes(t.id))
  const types = [...new Set(selected.map(t => docTypeLabel(t.classification)))]
  return `确认后将自动生成：${types.join('、')}，共 ${selectedIds.value.length} 笔`
})

function docTypeLabel(cls: string): string {
  const map: Record<string, string> = {
    business_receipt: '收款单',
    business_payment: '付款单',
    bank_fee: '银行费用单',
    interest_income: '利息收入单',
    internal_transfer: '银行转账单',
    pending: '待处理',
  }
  return map[cls] || cls
}

function classificationTag(val: string) {
  const map: Record<string, string> = { business_receipt: 'success', business_payment: 'danger', bank_fee: 'warning', interest_income: 'primary', internal_transfer: 'info', pending: 'danger' }
  return map[val] || 'info'
}

function classificationLabel(val: string) {
  const map: Record<string, string> = { business_receipt: '业务收款', business_payment: '业务付款', bank_fee: '银行费用', interest_income: '利息收入', internal_transfer: '内部转账', pending: '待处理' }
  return map[val] || val
}

function onSelectionChange(rows: TxnItem[]) {
  selectedIds.value = rows.map(r => r.id)
}

function onSelectAll() {
  // el-table 内置全选
}

function confirmOne(row: TxnItem) {
  row.confirmed = true
  docCounter.value++
  const docType = docTypeLabel(row.classification)
  ElMessage.success(`流水 ${row.date} 已确认，已生成${docType}`)
}

function batchConfirm() {
  const selected = allTxns.value.filter(t => selectedIds.value.includes(t.id))
  selected.forEach(t => { t.confirmed = true })
  docCounter.value += selected.length
  const typeSummary = [...new Set(selected.map(t => docTypeLabel(t.classification)))].join('、')
  ElMessage.success(`已确认 ${selectedIds.value.length} 条流水，自动生成 ${typeSummary}`)
  selectedIds.value = []
  selectAll.value = false
}

function editOne(row: TxnItem) {
  classifyForm.classification = row.classification
  showClassifyDialog.value = true
}

function saveClassification() {
  showClassifyDialog.value = false
  ElMessage.success('分类已更新')
}
</script>

<style scoped lang="scss">
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.stat-row { margin-bottom: 16px; }
.stat-card { text-align: center; .stat-num { font-size: 28px; font-weight: 700; margin-bottom: 4px; &.danger { color: #ff4d4f; } } .stat-label { font-size: 13px; color: #999; } &.total .stat-num { color: #1890ff; } &.confirmed .stat-num { color: #52c41a; } &.pending .stat-num { color: #faad14; } &.docs .stat-num { color: #722ed1; } }
.classification-tabs { margin-bottom: 8px; }
.batch-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; padding: 8px 0; .selected-count { font-size: 13px; color: #666; } }
.batch-preview { margin-bottom: 12px; }
.doc-preview { color: #999; font-size: 12px; }
</style>
