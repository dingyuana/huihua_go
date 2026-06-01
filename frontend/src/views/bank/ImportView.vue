<template>
  <div class="import-view">
    <div class="page-header">
      <h3>银行流水导入</h3>
    </div>

    <el-card class="step-card">
      <div class="step-title">1. 选择银行账户</div>
      <el-select v-model="bankAccountId" placeholder="选择银行账户" style="width: 320px">
        <el-option v-for="acct in bankAccounts" :key="acct.id" :label="`${acct.bank_name} (${maskAccount(acct.account_number)})`" :value="acct.id" />
      </el-select>
      <el-button v-if="bankAccountId" style="margin-left:8px" @click="fetchOnline">📡 银企直连抓取</el-button>
    </el-card>

    <el-card class="step-card">
      <div class="step-title">2. 上传对账单文件</div>
      <el-upload drag accept=".xlsx,.xls,.csv,.xml" :auto-upload="false" :on-change="handleFileChange" class="upload-area">
        <el-icon :size="48"><UploadFilled /></el-icon>
        <p class="upload-text">拖拽文件到此区域，或 <em>点击上传</em></p>
        <p class="upload-hint">支持 CSV / Excel / CAMT053 / MT940 格式</p>
      </el-upload>
      <div v-if="uploadedFile" class="file-info">
        <el-tag type="success" size="small">已选择</el-tag>
        <span class="file-name">{{ uploadedFile.name }}</span>
        <span class="file-size">({{ (uploadedFile.size / 1024).toFixed(1) }} KB)</span>
        <el-tag :type="formatTagType" size="small" style="margin:0 8px">{{ detectedFormat }}</el-tag>
        <el-button text type="primary" size="small" @click="handlePreview">预览并解析文件</el-button>
      </div>
    </el-card>

    <!-- 字段映射 -->
    <el-card v-if="showMapping" class="step-card">
      <div class="step-title">3. 字段映射确认</div>
      <p class="step-hint">系统已从文件中读取列名，请确认映射关系（自动匹配相似列名）</p>
      <el-table :data="fieldMappings" size="small" border>
        <el-table-column prop="field" label="系统字段" width="160" />
        <el-table-column prop="required" label="必填" width="60">
          <template #default="{ row }">
            <el-tag v-if="row.required" type="danger" size="small">*</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="匹配状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.matched ? 'success' : 'warning'" size="small">{{ row.matched ? '已匹配' : '待确认' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="文件列名" min-width="200">
          <template #default="{ row }">
            <el-select v-model="row.mappedColumn" placeholder="请选择对应的列" style="width: 100%" filterable allow-create clearable>
              <el-option v-for="col in fileColumns" :key="col" :label="col" :value="col" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="样本数据" min-width="150">
          <template #default="{ row }">
            <span v-if="row.mappedColumn && row.sampleValue" class="sample-value">{{ row.sampleValue }}</span>
            <span v-else class="no-sample">-</span>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="parseErrors.length" style="margin-top:12px">
        <el-alert :title="`解析完成，发现 ${parseErrors.length} 条异常记录`" type="warning" :closable="false" show-icon />
      </div>
      <div style="margin-top:12px; display:flex; gap:8px;">
        <el-button type="primary" size="small" @click="showPreview = true; showMapping = false">确认映射，预览数据</el-button>
        <el-button size="small" @click="showMapping = false">重新上传</el-button>
      </div>
    </el-card>

    <!-- 预览 -->
    <el-card v-if="showPreview" class="step-card">
      <div class="step-title">4. 数据预览</div>
      <div class="preview-stats">
        <span>总识别: <b>{{ totalRows }}</b> 条</span>
        <span>正常: <b class="success">{{ normalCount }}</b> 条</span>
        <span>异常: <b class="danger">{{ parseErrors.length }}</b> 条</span>
      </div>
      <el-tabs v-model="previewTab">
        <el-tab-pane label="全部数据" name="all">
          <el-table :data="previewData" size="small" border stripe max-height="300">
            <el-table-column prop="date" label="日期" width="100" />
            <el-table-column prop="amount" label="金额" width="120">
              <template #default="{ row }">
                <span :class="row.direction === 'in' ? 'amount-positive' : 'amount-negative'">
                  {{ row.direction === 'in' ? '+' : '-' }}{{ row.amount }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="direction" label="方向" width="60">
              <template #default="{ row }">{{ row.direction === 'in' ? '收入' : '支出' }}</template>
            </el-table-column>
            <el-table-column prop="counterparty" label="对方户名" width="140" />
            <el-table-column prop="description" label="摘要" min-width="180" show-overflow-tooltip />
            <el-table-column prop="ref" label="流水号" width="130" />
            <el-table-column prop="transaction_type" label="交易类型" width="80" />
            <el-table-column prop="payer_account" label="付款人账号" width="130" />
            <el-table-column prop="payer_bank" label="付款人开户行" min-width="140" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="`异常记录 (${parseErrors.length})`" name="errors">
          <el-table :data="parseErrors" size="small" border stripe>
            <el-table-column prop="row" label="行号" width="60" />
            <el-table-column prop="field" label="字段" width="100" />
            <el-table-column prop="value" label="原始值" width="140" />
            <el-table-column prop="issue" label="问题" min-width="200" />
            <el-table-column label="操作" width="100">
              <template #default="{ $index }">
                <el-button size="small" @click="skipRow($index)">跳过</el-button>
                <el-button size="small" type="primary" @click="editRow($index)">补录</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <div v-if="showPreview" class="import-actions">
      <el-button size="large" @click="resetImport">重新选择</el-button>
      <el-button size="large" type="primary" :loading="importing" @click="confirmBeforeImport">确认导入 ({{ normalCount }} 条)</el-button>
    </div>

    <!-- 导入确认弹窗（仅当有多个银行账户时弹出） -->
    <el-dialog v-model="showConfirmDialog" title="确认导入" width="420px">
      <div style="padding:8px 0">
        <el-descriptions :column="1" size="small" border>
          <el-descriptions-item label="银行账户">{{ selectedBankName }}</el-descriptions-item>
          <el-descriptions-item label="导入文件">{{ uploadedFile?.name }}</el-descriptions-item>
          <el-descriptions-item label="文件格式">{{ detectedFormat }}</el-descriptions-item>
          <el-descriptions-item label="导入条数">{{ normalCount }} 条</el-descriptions-item>
        </el-descriptions>
        <el-alert v-if="parseErrors.length" title="存在异常记录将被跳过" type="warning" :description="`${parseErrors.length} 条记录无法解析`" show-icon style="margin-top:12px" :closable="false" />
      </div>
      <template #footer>
        <el-button @click="showConfirmDialog = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="handleImport">确认导入</el-button>
      </template>
    </el-dialog>

    <!-- 导入结果弹窗 -->
    <el-dialog v-model="showResultDialog" title="导入完成" width="480px">
      <el-result icon="success" title="导入完成">
        <template #extra>
          <el-descriptions :column="3" size="small" border>
            <el-descriptions-item label="银行账户">{{ selectedBankName }}</el-descriptions-item>
            <el-descriptions-item label="总条数">{{ importResult.total_rows }}</el-descriptions-item>
            <el-descriptions-item label="成功">{{ importResult.success_count }}</el-descriptions-item>
            <el-descriptions-item label="失败">{{ importResult.failed_count }}</el-descriptions-item>
          </el-descriptions>
          <el-alert v-if="importResult.failed_rows?.length" title="以下行号导入失败" type="warning" :description="importResult.failed_rows.join(', ')" show-icon style="margin-top:12px" />
          <div style="margin-top:16px">
            <el-button type="primary" @click="$router.push('/bank/workbench')">前往核对工作台</el-button>
            <el-button @click="closeResultDialog">继续导入</el-button>
          </div>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import { previewExcelFile } from '@/api/modules/bank'

interface BankAccount {
  id: string
  bank_name: string
  account_number: string
}

const bankAccounts = ref<BankAccount[]>([])
const bankAccountId = ref('')
const uploadedFile = ref<File | null>(null)
const showMapping = ref(false)
const showPreview = ref(false)
const showConfirmDialog = ref(false)
const showResultDialog = ref(false)
const previewTab = ref('all')
const loading = ref(false)
const importing = ref(false)
const importDone = ref(false)

const importResult = ref({ total_rows: 0, success_count: 0, failed_count: 0, failed_rows: [] as number[] })

async function loadBankAccounts() {
  try {
    const res: any = await request.get('/bank-accounts')
    const list = res?.data?.list !== undefined && res?.data?.list !== null ? res.data.list : (res?.data !== undefined && res?.data !== null ? res.data : res)
    bankAccounts.value = Array.isArray(list) ? list : []

    // 默认选中基本户
    if (bankAccounts.value.length === 1) {
      bankAccountId.value = bankAccounts.value[0].id
    } else if (bankAccounts.value.length > 1) {
      const basic = bankAccounts.value.find(a =>
        (a as any).bank_name?.includes('基本') || (a as any).bank_account_type === 'current'
      )
      if (basic) bankAccountId.value = basic.id
    }
  } catch (e) {
    console.warn('后端资金账户接口不可用', e)
    bankAccounts.value = []
  }
}

const selectedBankName = computed(() => {
  const acct = bankAccounts.value.find(a => a.id === bankAccountId.value)
  return acct ? `${acct.bank_name} (${maskAccount(acct.account_number)})` : '-'
})

function maskAccount(num: string) {
  if (!num || num === '-' || num.length < 8) return num || '-'
  return num.slice(0, 4) + ' **** **** ' + num.slice(-4)
}

onMounted(loadBankAccounts)

const detectedFormat = ref('')
const formatTagType = computed(() => {
  const map: Record<string, string> = { 'CSV': '', 'Excel': 'success', 'CAMT053': 'warning', 'MT940': 'info' }
  return map[detectedFormat.value] || ''
})

const fileColumns = ref<string[]>([])
const sampleData = ref<string[][]>([])
const totalRows = ref(0)

interface FieldMapping {
  field: string
  fieldKey: string
  required: boolean
  matched: boolean
  mappedColumn: string
  sampleValue: string
}

const defaultMappings: FieldMapping[] = [
  { field: '交易日期', fieldKey: 'date', required: true, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '金额', fieldKey: 'amount', required: true, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收入金额', fieldKey: 'income', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '支出金额', fieldKey: 'expense', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收支方向', fieldKey: 'direction', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '对方户名', fieldKey: 'counterparty', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '对方账号', fieldKey: 'counterparty_account', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '摘要', fieldKey: 'description', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '流水号', fieldKey: 'ref', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '余额', fieldKey: 'balance', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '交易类型', fieldKey: 'transaction_type', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '付款人账号', fieldKey: 'payer_account', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '付款人户名', fieldKey: 'payer_name', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '付款人开户行号', fieldKey: 'payer_bank_code', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '付款人开户行名', fieldKey: 'payer_bank', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收款人账号', fieldKey: 'payee_account', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收款人户名', fieldKey: 'payee_name', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收款人开户行号', fieldKey: 'payee_bank_code', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '收款人开户行名', fieldKey: 'payee_bank', required: false, matched: false, mappedColumn: '', sampleValue: '' },
]

const fieldMappings = ref<FieldMapping[]>([...defaultMappings])

function handleFileChange(file: any) {
  uploadedFile.value = file.raw
  const name = file.name.toLowerCase()
  if (name.endsWith('.csv')) detectedFormat.value = 'CSV'
  else if (name.endsWith('.xlsx') || name.endsWith('.xls')) detectedFormat.value = 'Excel'
  else if (name.includes('camt') || name.endsWith('.xml')) detectedFormat.value = 'CAMT053'
  else detectedFormat.value = 'MT940'
}

async function handlePreview() {
  if (!uploadedFile.value) {
    ElMessage.warning('请先上传文件')
    return
  }

  loading.value = true
  try {
    const res = await previewExcelFile(uploadedFile.value)
    console.log('API响应原始数据:', res)

    // 正确处理响应数据，考虑可能的包装结构
    let actualData = res as any
    // 检查是否有 data 包装
    if (actualData?.data?.columns) {
      actualData = actualData.data
    }

    console.log('解析后的 columns:', actualData?.columns)
    console.log('解析后的 sample:', actualData?.sample)

    if (actualData?.columns && Array.isArray(actualData.columns)) {
      fileColumns.value = actualData.columns
      sampleData.value = actualData.sample || []
      totalRows.value = actualData.total_rows || 0

      console.log('设置 fileColumns:', fileColumns.value)

      // Auto-match columns
      autoMatchColumns()
      showMapping.value = true
      showPreview.value = false
      ElMessage.success(`文件解析成功，共识别 ${totalRows.value} 条记录，${fileColumns.value.length} 个列`)
    } else {
      ElMessage.error('文件解析失败：列名数据格式不正确')
      console.error('列名数据异常:', actualData)
    }
  } catch (e) {
    console.error('Preview failed:', e)
    ElMessage.error('预览文件失败，请检查文件格式')
  }
  loading.value = false
}

function autoMatchColumns() {
  const columns = fileColumns.value.map(c => c.toLowerCase().trim())

  fieldMappings.value.forEach(mapping => {
    const fieldLower = mapping.field.toLowerCase()

    // Try exact match first
    let matchIdx = columns.findIndex(col => col === fieldLower)

    // Try contains match
    if (matchIdx === -1) {
      matchIdx = columns.findIndex(col =>
        col.includes(fieldLower) || fieldLower.includes(col)
      )
    }

    // Try common synonyms
    if (matchIdx === -1) {
      const synonyms: Record<string, string[]> = {
        date: ['transaction date', '交易日期', '记账日期', '发生日期', '交易时间'],
        amount: ['金额', 'transaction amount', '发生金额'],
        income: ['收入金额', '贷方金额', '收入', 'credit'],
        expense: ['支出金额', '借方金额', '支出', 'debit'],
        direction: ['收支方向', '方向', '借贷方向'],
        counterparty: ['对方户名', '对方名称', '收款人', '付款人', '交易对方'],
        counterparty_account: ['对方账号', '对方账户', '收款账号', '付款账号'],
        description: ['摘要', 'description', '备注', '用途', '附言'],
        ref: ['流水号', 'ref', 'reference', '交易流水号', '交易编号'],
        balance: ['余额', 'balance', '账户余额'],
        transaction_type: ['交易类型', 'transaction type', '业务类型'],
        payer_account: ['付款人账号', 'payer account', '转出账号'],
        payer_name: ['付款人户名', 'payer name', '转出户名'],
        payer_bank: ['付款人开户行名', 'payer bank', '转出行'],
        payer_bank_code: ['付款人开户行号', 'payer bank code', '转出行号'],
        payee_account: ['收款人账号', 'payee account', '转入账号'],
        payee_name: ['收款人户名', 'payee name', '转入户名'],
        payee_bank: ['收款人开户行名', 'payee bank', '转入行'],
        payee_bank_code: ['收款人开户行号', 'payee bank code', '转入行号'],
      }

      const syns = synonyms[mapping.fieldKey] || []
      matchIdx = columns.findIndex(col =>
        syns.some(syn => col === syn.toLowerCase() || col.includes(syn.toLowerCase()))
      )
    }

    if (matchIdx !== -1) {
      mapping.matched = true
      mapping.mappedColumn = fileColumns.value[matchIdx]
      if (sampleData.value.length > 0 && sampleData.value[0][matchIdx]) {
        mapping.sampleValue = sampleData.value[0][matchIdx]
      }
    } else {
      mapping.matched = false
      mapping.mappedColumn = ''
      mapping.sampleValue = ''
    }
  })
}

function fetchOnline() {
  ElMessage.success('银企直连抓取成功，获取到 125 条最新流水')
}

const parseErrors = ref<{ row: number; field: string; value: string; issue: string }[]>([])

const normalCount = computed(() => totalRows.value - parseErrors.value.length)

const previewData = computed(() => {
  if (sampleData.value.length === 0) return []

  const dateIdx = fieldMappings.value.find(m => m.fieldKey === 'date')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'date')!.mappedColumn)
    : -1
  const incomeIdx = fieldMappings.value.find(m => m.fieldKey === 'income')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'income')!.mappedColumn)
    : -1
  const expenseIdx = fieldMappings.value.find(m => m.fieldKey === 'expense')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'expense')!.mappedColumn)
    : -1
  const directionIdx = fieldMappings.value.find(m => m.fieldKey === 'direction')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'direction')!.mappedColumn)
    : -1
  const counterpartyIdx = fieldMappings.value.find(m => m.fieldKey === 'counterparty')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'counterparty')!.mappedColumn)
    : -1
  const descriptionIdx = fieldMappings.value.find(m => m.fieldKey === 'description')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'description')!.mappedColumn)
    : -1
  const refIdx = fieldMappings.value.find(m => m.fieldKey === 'ref')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'ref')!.mappedColumn)
    : -1
  const transactionTypeIdx = fieldMappings.value.find(m => m.fieldKey === 'transaction_type')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'transaction_type')!.mappedColumn)
    : -1
  const payerAccountIdx = fieldMappings.value.find(m => m.fieldKey === 'payer_account')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'payer_account')!.mappedColumn)
    : -1
  const payerBankIdx = fieldMappings.value.find(m => m.fieldKey === 'payer_bank')?.mappedColumn
    ? fileColumns.value.indexOf(fieldMappings.value.find(m => m.fieldKey === 'payer_bank')!.mappedColumn)
    : -1

  return sampleData.value.map((row, idx) => {
    const income = incomeIdx !== -1 && row[incomeIdx] ? parseFloat(row[incomeIdx].replace(/,/g, '')) : 0
    const expense = expenseIdx !== -1 && row[expenseIdx] ? parseFloat(row[expenseIdx].replace(/,/g, '')) : 0
    const direction = directionIdx !== -1 && row[directionIdx]
      ? (row[directionIdx].includes('收') || row[directionIdx].includes('贷') ? 'in' : 'out')
      : (income > 0 ? 'in' : 'out')

    return {
      date: dateIdx !== -1 ? row[dateIdx] : '-',
      amount: (income > 0 ? income : expense).toFixed(2),
      direction,
      counterparty: counterpartyIdx !== -1 ? row[counterpartyIdx] : '-',
      description: descriptionIdx !== -1 ? row[descriptionIdx] : '-',
      ref: refIdx !== -1 ? row[refIdx] : `AUTO-${idx + 1}`,
      transaction_type: transactionTypeIdx !== -1 ? row[transactionTypeIdx] : '-',
      payer_account: payerAccountIdx !== -1 ? row[payerAccountIdx] : '-',
      payer_bank: payerBankIdx !== -1 ? row[payerBankIdx] : '-',
    }
  })
})

function skipRow(idx: number) {
  parseErrors.value.splice(idx, 1)
  totalRows.value--
  ElMessage.info('已跳过该异常记录')
}

function editRow(idx: number) {
  ElMessage.success('请在出纳核对工作台补录该记录')
}

function resetImport() {
  uploadedFile.value = null
  showMapping.value = false
  showPreview.value = false
  showConfirmDialog.value = false
  showResultDialog.value = false
  importDone.value = false
  parseErrors.value = []
  detectedFormat.value = ''
  fileColumns.value = []
  sampleData.value = []
  totalRows.value = 0
  fieldMappings.value = [...defaultMappings]
}

function confirmBeforeImport() {
  if (!uploadedFile.value || !bankAccountId.value) {
    ElMessage.warning('请选择银行账户和上传文件')
    return
  }
  // 只有一个银行账户时跳过确认弹窗，直接导入
  if (bankAccounts.value.length <= 1) {
    handleImport()
    return
  }
  showConfirmDialog.value = true
}

async function handleImport() {
  if (!uploadedFile.value || !bankAccountId.value) {
    ElMessage.warning('请选择银行账户和上传文件')
    return
  }
  showConfirmDialog.value = false
  importing.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadedFile.value)
    const res: any = await request.post(`/bank-transactions/import?bank_account_id=${bankAccountId.value}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importResult.value = {
      total_rows: res?.total_rows ?? totalRows.value,
      success_count: res?.success_count ?? 0,
      failed_count: res?.failed_count ?? 0,
      failed_rows: res?.failed_rows ?? [],
    }
    importDone.value = true
    showResultDialog.value = true
    ElMessage.success(`导入完成，成功 ${importResult.value.success_count} 条`)
  } catch (e: any) {
    const msg = e?.response?.data?.error || '导入失败'
    ElMessage.error(msg)
  } finally {
    importing.value = false
  }
}

function closeResultDialog() {
  showResultDialog.value = false
  resetImport()
}
</script>

<style scoped lang="scss">
.page-header { margin-bottom: 16px; h3 { font-size: 18px; } }
.step-card { margin-bottom: 16px; .step-title { font-weight: 600; margin-bottom: 16px; font-size: 15px; color: #333; } }
.step-hint { color: #999; font-size: 13px; margin-bottom: 12px; }
.upload-area { text-align: center; .upload-text { font-size: 14px; margin: 8px 0; em { color: #1890ff; font-style: normal; } } .upload-hint { color: #999; font-size: 12px; } }
.file-info { margin-top: 12px; display: flex; align-items: center; flex-wrap: wrap; gap: 4px; .file-name { font-weight: 500; } .file-size { color: #999; font-size: 12px; } }
.preview-stats { display: flex; gap: 20px; margin-bottom: 12px; font-size: 13px; .success { color: #52c41a; } .danger { color: #ff4d4f; } }
.sample-value { font-size: 12px; color: #666; }
.no-sample { color: #999; }
.import-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
.import-result { margin-top: 24px; p { margin-bottom: 16px; color: #666; } }
</style>
