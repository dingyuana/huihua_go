<template>
  <div class="expense-invoice-import-page">
    <div class="page-header">
      <h3>进项发票导入</h3>
      <div class="header-actions">
        <el-button @click="goList">返回列表</el-button>
      </div>
    </div>

    <el-card>
      <!-- 步骤条 -->
      <el-steps :active="activeStep" finish-status="success" align-center class="import-steps">
        <el-step title="上传文件" description="选择 Excel 文件" />
        <el-step title="预览确认" description="核对数据并选择" />
        <el-step title="完成" description="导入结果" />
      </el-steps>

      <!-- Step 1: 上传 -->
      <div v-show="activeStep === 0" class="step-pane">
        <el-upload
          ref="uploadRef"
          class="upload-dragger"
          drag
          action=""
          :auto-upload="false"
          :show-file-list="false"
          :accept="acceptExt"
          :on-change="onFileChange"
          :before-upload="beforeUpload"
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            将 Excel 文件拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              <p>支持 .xlsx / .xls 格式</p>
              <p>必填列：发票号码 / 开票日期 / 价税合计 / 税额 / 供应商名称</p>
            </div>
          </template>
        </el-upload>

        <div v-if="uploading" class="upload-status">
          <el-icon class="is-loading"><loading /></el-icon>
          <span>正在上传并解析文件…</span>
        </div>

        <div v-if="uploadError" class="upload-error">
          <el-alert :title="uploadError" type="error" show-icon :closable="false" />
        </div>
      </div>

      <!-- Step 2: 预览 -->
      <div v-show="activeStep === 1" class="step-pane">
        <div class="preview-summary">
          <el-tag size="large" type="info">总数：{{ previewData.total_rows ?? previewData.details?.length ?? 0 }}</el-tag>
          <el-tag size="large" type="success">有效行：{{ previewData.valid_rows ?? validCount }}</el-tag>
          <el-tag size="large" type="danger">错误行：{{ previewData.error_rows ?? invalidCount }}</el-tag>
          <el-tag size="large" type="warning">重复行：{{ duplicateCount }}</el-tag>
        </div>

        <el-table
          ref="tableRef"
          :data="previewData.details"
          border
          stripe
          size="small"
          height="520"
          @selection-change="onSelectionChange"
        >
          <el-table-column type="selection" width="48" :selectable="rowSelectable" />
          <el-table-column type="index" label="行号" width="70" />
          <el-table-column prop="invoice_no" label="发票号码" min-width="160" show-overflow-tooltip />
          <el-table-column label="类型" width="110">
            <template #default="{ row }">
              {{ invoiceKindLabel(row.invoice_kind) }}
            </template>
          </el-table-column>
          <el-table-column prop="vendor_name" label="供应商" min-width="180" show-overflow-tooltip />
          <el-table-column prop="invoice_date" label="开票日期" width="120" />
          <el-table-column label="价税合计" width="140" align="right">
            <template #default="{ row }">
              ¥{{ formatAmount(row.total_amount) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="180">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row)" size="small">
                {{ statusLabel(row) }}
              </el-tag>
              <div v-if="row.validation_err" class="row-err">{{ row.validation_err }}</div>
              <div v-else-if="row.is_duplicate && row.duplicate_info" class="row-err">
                重复：{{ row.duplicate_info }}
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="step-actions">
          <el-button @click="reupload" :disabled="confirming">重新上传</el-button>
          <el-button
            type="primary"
            :loading="confirming"
            :disabled="selectedRows.length === 0"
            @click="onConfirm"
          >
            确认导入（{{ selectedRows.length }} 条）
          </el-button>
        </div>
      </div>

      <!-- Step 3: 完成 -->
      <div v-show="activeStep === 2" class="step-pane">
        <el-result
          :icon="resultIcon"
          :title="resultTitle"
          :sub-title="resultSubTitle"
        >
          <template #extra>
            <div class="result-stats">
              <el-statistic title="成功导入" :value="importResult.imported" :value-style="{ color: '#67c23a' }" />
              <el-statistic title="跳过" :value="importResult.skipped" :value-style="{ color: '#909399' }" />
              <el-statistic title="错误" :value="importResult.errors" :value-style="{ color: '#f56c6c' }" />
            </div>

            <div v-if="importResult.failed_rows && importResult.failed_rows.length" class="failed-section">
              <h4>失败行明细</h4>
              <el-table :data="importResult.failed_rows" border size="small" max-height="300">
                <el-table-column prop="row" label="行号" width="100" />
                <el-table-column prop="reason" label="失败原因" min-width="300" show-overflow-tooltip />
              </el-table>
            </div>

            <div class="step-actions">
              <el-button @click="reupload">导入更多</el-button>
              <el-button type="primary" @click="goList">返回列表</el-button>
            </div>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type UploadFile, type UploadRawFile, type UploadInstance, type UploadProps, type TableInstance } from 'element-plus'
import { UploadFilled, Loading } from '@element-plus/icons-vue'
import {
  uploadExpenseInvoiceImport,
  previewExpenseInvoiceImport,
  confirmExpenseInvoiceImport,
} from '@/api/modules/expense-invoice'

const router = useRouter()

// ---------- state ----------
const activeStep = ref(0)
const uploadRef = ref<UploadInstance>()
const tableRef = ref<TableInstance>()
const acceptExt = '.xlsx,.xls'
const uploading = ref(false)
const uploadError = ref('')
const confirming = ref(false)

interface PreviewDetail {
  row_index?: number
  row?: number
  invoice_no?: string
  invoice_kind?: string
  vendor_name?: string
  invoice_date?: string
  amount?: string | number
  tax_amount?: string | number
  total_amount?: string | number
  status?: 'valid' | 'invalid' | 'duplicate' | string
  validation_err?: string
  reason?: string
  is_duplicate?: boolean
  duplicate_info?: string
  valid?: boolean
}

interface PreviewPayload {
  batch_id?: string
  total_rows?: number
  valid_rows?: number
  error_rows?: number
  details: PreviewDetail[]
  // legacy field names
  total?: number
  valid?: number
  invalid?: number
  rows?: PreviewDetail[]
}

const batchId = ref('')
const previewData = ref<PreviewPayload>({ details: [] })
const selectedRows = ref<PreviewDetail[]>([])

interface ImportResultPayload {
  imported: number
  skipped: number
  errors: number
  failed_rows?: Array<{ row: number; reason: string }>
}
const importResult = ref<ImportResultPayload>({ imported: 0, skipped: 0, errors: 0 })

// ---------- computed ----------
const validCount = computed(() =>
  previewData.value.details.filter(d => statusOf(d) === 'valid').length,
)
const invalidCount = computed(() =>
  previewData.value.details.filter(d => statusOf(d) === 'invalid').length,
)
const duplicateCount = computed(() =>
  previewData.value.details.filter(d => statusOf(d) === 'duplicate').length,
)

const resultIcon = computed(() => {
  if (importResult.value.errors === 0) return 'success'
  if (importResult.value.imported === 0) return 'error'
  return 'warning'
})
const resultTitle = computed(() => {
  if (importResult.value.errors === 0) return '导入完成'
  if (importResult.value.imported === 0) return '导入失败'
  return '部分导入成功'
})
const resultSubTitle = computed(() => {
  const { imported, skipped, errors } = importResult.value
  return `成功 ${imported} 条，跳过 ${skipped} 条，失败 ${errors} 条`
})

// ---------- helpers ----------
function statusOf(row: PreviewDetail): 'valid' | 'invalid' | 'duplicate' {
  if (row.status === 'valid' || row.status === 'invalid' || row.status === 'duplicate') {
    return row.status
  }
  if (row.is_duplicate) return 'duplicate'
  if (row.valid === false || row.validation_err || row.reason) return 'invalid'
  return 'valid'
}

function statusTagType(row: PreviewDetail): 'success' | 'danger' | 'warning' {
  const s = statusOf(row)
  if (s === 'valid') return 'success'
  if (s === 'invalid') return 'danger'
  return 'warning'
}

function statusLabel(row: PreviewDetail): string {
  const s = statusOf(row)
  if (s === 'valid') return '有效'
  if (s === 'invalid') return '校验失败'
  return '重复'
}

function rowSelectable(row: PreviewDetail): boolean {
  return statusOf(row) === 'valid'
}

function invoiceKindLabel(kind?: string): string {
  switch (kind) {
    case 'paper_normal': return '纸质普票'
    case 'paper_special': return '纸质专票'
    case 'electronic_normal': return '电子普票'
    case 'electronic_special': return '电子专票'
    default: return kind || '—'
  }
}

function formatAmount(v: string | number | undefined): string {
  if (v === undefined || v === null || v === '') return '0.00'
  const n = Number(v)
  if (Number.isNaN(n)) return String(v)
  return n.toFixed(2)
}

// ---------- upload ----------
const beforeUpload: UploadProps['beforeUpload'] = (file: UploadRawFile) => {
  const name = file.name.toLowerCase()
  if (!name.endsWith('.xlsx') && !name.endsWith('.xls')) {
    ElMessage.error('仅支持 .xlsx / .xls 格式')
    return false
  }
  const maxMB = 10
  if (file.size > maxMB * 1024 * 1024) {
    ElMessage.error(`文件大小不能超过 ${maxMB}MB`)
    return false
  }
  return true
}

function onFileChange(file: UploadFile) {
  if (!file.raw) return
  uploadError.value = ''
  doUpload(file.raw)
}

async function doUpload(file: File) {
  uploading.value = true
  try {
    const res: any = await uploadExpenseInvoiceImport(file)
    const data = res?.data ?? res
    const id = data?.batch_id
    if (!id) {
      uploadError.value = '上传成功但未返回 batch_id'
      return
    }
    batchId.value = String(id)
    await loadPreview()
    activeStep.value = 1
  } catch (e: any) {
    uploadError.value = e?.message || '上传失败，请重试'
    ElMessage.error(uploadError.value)
  } finally {
    uploading.value = false
  }
}

async function loadPreview() {
  if (!batchId.value) return
  const res: any = await previewExpenseInvoiceImport(batchId.value)
  const data = res?.data ?? res
  // Normalize: backend may return details or rows; total_rows or total
  const details: PreviewDetail[] = data?.details || data?.rows || []
  previewData.value = {
    ...data,
    details,
  }
  // default-select all valid rows
  setTimeout(() => {
    const validRows = details.filter(d => statusOf(d) === 'valid')
    if (tableRef.value && validRows.length) {
      tableRef.value.clearSelection()
      validRows.forEach(r => {
        // @ts-ignore - toggleRowSelection is on the table instance
        tableRef.value!.toggleRowSelection(r, true)
      })
    }
  }, 50)
}

function onSelectionChange(rows: PreviewDetail[]) {
  selectedRows.value = rows
}

function rowKey(row: PreviewDetail): string {
  return String(row.row_index ?? row.row ?? row.invoice_no ?? Math.random())
}

// ---------- confirm ----------
async function onConfirm() {
  if (!batchId.value) {
    ElMessage.warning('批次信息丢失，请重新上传')
    activeStep.value = 0
    return
  }
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请至少选择一条有效数据')
    return
  }
  confirming.value = true
  try {
    const selected_ids = selectedRows.value.map(rowKey)
    const res: any = await confirmExpenseInvoiceImport({
      batch_id: batchId.value,
      selected_ids,
    } as any)
    const data = res?.data ?? res
    importResult.value = {
      imported: Number(data?.imported ?? 0),
      skipped: Number(data?.skipped ?? 0),
      errors: Number(data?.errors ?? 0),
      failed_rows: data?.failed_rows || [],
    }
    activeStep.value = 2
  } catch (e: any) {
    ElMessage.error(e?.message || '导入失败')
  } finally {
    confirming.value = false
  }
}

// ---------- navigation ----------
function reupload() {
  activeStep.value = 0
  batchId.value = ''
  previewData.value = { details: [] }
  selectedRows.value = []
  importResult.value = { imported: 0, skipped: 0, errors: 0 }
  uploadError.value = ''
  // clear uploaded file list
  uploadRef.value?.clearFiles()
}

function goList() {
  router.push('/expense-invoices/list')
}

onMounted(() => {
  // no-op
})
</script>

<style scoped>
.expense-invoice-import-page {
  padding: 16px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h3 {
  margin: 0;
}
.import-steps {
  margin-bottom: 24px;
}
.step-pane {
  padding: 8px 0;
}
.upload-dragger :deep(.el-upload-dragger) {
  width: 560px;
  max-width: 100%;
  margin: 24px auto;
}
.el-upload__tip p {
  margin: 2px 0;
  color: #909399;
  font-size: 12px;
}
.upload-status,
.upload-error {
  margin: 12px auto;
  max-width: 560px;
  text-align: center;
}
.upload-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #409eff;
}
.preview-summary {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.row-err {
  margin-top: 4px;
  color: #f56c6c;
  font-size: 12px;
}
.step-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.result-stats {
  display: flex;
  gap: 48px;
  justify-content: center;
  margin: 16px 0 24px;
}
.failed-section {
  margin: 0 auto 16px;
  max-width: 720px;
}
.failed-section h4 {
  text-align: left;
  margin: 8px 0;
}
</style>
