<template>
  <div class="invoice-page">
    <div class="page-header">
      <h3>发票工作台</h3>
      <div class="header-actions">
        <template v-if="activeTab === 'sales'">
          <el-button type="primary" @click="showUpload = true">导入销售发票</el-button>
          <el-button type="default" @click="showManualCreate = true">手工录入</el-button>
        </template>
        <template v-else>
          <el-button type="primary" @click="showManualCreate = true">手工录入</el-button>
          <el-button type="default" @click="showUpload = true">OCR识别录入</el-button>
        </template>
      </div>
    </div>

    <!-- 发票类型Tabs -->
    <el-card class="tax-tabs-card">
      <el-tabs v-model="activeTab" @change="handleTabChange">
        <el-tab-pane :label="`销售发票 ${salesBadge > 0 ? '(' + salesBadge + ')' : ''}`" name="sales" />
        <el-tab-pane :label="`采购发票 ${purchaseBadge > 0 ? '(' + purchaseBadge + ')' : ''}`" name="purchase" />
        <el-tab-pane :label="`费用发票 ${expenseBadge > 0 ? '(' + expenseBadge + ')' : ''}`" name="expense" />
      </el-tabs>
    </el-card>

    <!-- 状态统计 -->
    <el-card class="stats-card">
      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-label">草稿</span>
          <span class="stat-value draft">{{ draftCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">待确认</span>
          <span class="stat-value pending">{{ pendingCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">已确认</span>
          <span class="stat-value confirmed">{{ confirmedCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">部分核销</span>
          <span class="stat-value partial">{{ partialCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">已核销</span>
          <span class="stat-value paid">{{ paidCount }}</span>
        </div>
      </div>
    </el-card>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="状态">
          <el-select v-model="filter.status" placeholder="全部" style="width: 130px" clearable>
            <el-option label="待核销" value="unpaid" />
            <el-option label="部分核销" value="partially_paid" />
            <el-option label="已核销" value="paid" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filter.type" placeholder="全部" style="width: 120px" clearable>
            <el-option label="销项" value="sale" />
            <el-option label="进项" value="purchase" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.isReturn" placeholder="全部" style="width: 120px" clearable>
            <el-option label="正常" :value="false" />
            <el-option label="红字" :value="true" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="filter.dateRange" type="daterange" range-separator="~" start-placeholder="开始" end-placeholder="结束" style="width: 240px" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="发票号/对方名称" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="invoices" border stripe size="small" :expand-row-keys="expandedRows" @expand-change="handleExpandChange" row-key="id">
        <el-table-column type="expand">
          <template #expand-icon="{ expanded, row }">
            <el-icon v-if="row.line_items && row.line_items.length > 1" :size="14">
              <component :is="expanded ? 'ArrowDown' : 'ArrowRight'" />
            </el-icon>
            <span v-else>&nbsp;&nbsp;&nbsp;</span>
          </template>
          <template #default="{ row }">
            <div v-if="row.line_items && row.line_items.length > 0">
              <el-table :data="row.line_items" border size="mini" style="width: 100%">
                <el-table-column prop="item_code" label="商品编码" width="120" />
                <el-table-column prop="description" label="商品名称" min-width="200" />
                <el-table-column prop="quantity" label="数量" width="80" align="right" />
                <el-table-column prop="unit" label="单位" width="60" />
                <el-table-column prop="unit_price" label="单价" width="100" align="right" />
                <el-table-column prop="net_amount" label="不含税金额" width="120" align="right" />
                <el-table-column prop="tax_rate" label="税率" width="80" align="right" />
                <el-table-column prop="tax_amount" label="税额" width="100" align="right" />
                <el-table-column prop="total_amount" label="价税合计" width="120" align="right" />
              </el-table>
            </div>
            <div v-else style="padding: 16px; text-align: center; color: #999;">
              无明细行项
            </div>
          </template>
        </el-table-column>
        <el-table-column type="selection" width="55" />
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="invoice_no" label="发票号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="invoice_category" label="票种" min-width="120" show-overflow-tooltip />
        <el-table-column label="类型" width="70">
          <template #default="{ row }">
            <el-tag :type="row.type === 'sale' ? 'success' : row.type === 'purchase' ? 'warning' : 'info'" size="small">
              {{ row.type === 'sale' ? '销项' : row.type === 'purchase' ? '进项' : '费用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_return" type="danger" size="small">红字</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="对应蓝字发票" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <template v-if="row.is_return">
              <el-link v-if="row.source_red_invoice_no" type="primary" :underline="false" size="small">
                {{ row.source_red_invoice_no }}
              </el-link>
              <span v-else>—</span>
            </template>
            <span v-else class="no-col">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="customer_name" label="对方单位" min-width="150" />
        <el-table-column prop="tax_id" label="税号" width="140" />
        <el-table-column prop="posting_date" label="开票日期" width="100" />
        <el-table-column prop="total_amount" label="价税合计" width="120" align="right" />
        <el-table-column prop="tax_amount" label="税额" width="100" align="right" />
        <el-table-column prop="net_amount" label="不含税金额" width="120" align="right" />
        <el-table-column label="未核销" width="120" align="right">
          <template #default="{ row }">
            <span v-if="parseFloat(row.outstanding) > 0" class="amount-positive">{{ row.outstanding }}</span>
            <span v-else class="amount-negative">已核销</span>
          </template>
        </el-table-column>
        <el-table-column prop="due_date" label="到期日" width="100">
          <template #default="{ row }">
            <span :class="{ 'expiring': isExpiring(row.due_date) }">{{ row.due_date }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail = row">详情</el-button>
            <el-button v-if="row.status === 'draft'" link type="warning" size="small" @click="handleConfirmInvoice(row)">确认</el-button>
            <el-button v-if="row.status === 'verified'" link type="success" size="small" @click="handleGenerateVoucher(row)">生成凭证</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top:12px; display:flex; justify-content:flex-between; align-items:center;">
        <span>已选 {{ selectedInvoices.length }} 条</span>
        <div>
          <el-button v-if="selectedInvoices.length > 0" type="primary" size="small" @click="batchConfirm">批量确认</el-button>
          <el-button v-if="selectedInvoices.length > 0" type="danger" size="small" @click="batchDelete">批量删除</el-button>
        </div>
      </div>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog v-model="showUpload" title="上传发票" width="700px">
      <el-tabs v-model="uploadTab">
        <el-tab-pane label="OCR识别" name="ocr">
          <el-upload drag accept=".pdf,.ofd,.jpg,.png" :auto-upload="false" :on-change="handleUpload" class="upload-area">
            <el-icon :size="40"><UploadFilled /></el-icon>
            <p>拖拽发票文件或点击上传</p>
            <p class="upload-hint">支持 PDF / OFD / 图片格式</p>
          </el-upload>
          <div v-if="ocrResult" class="ocr-result">
            <h4>OCR 识别结果</h4>
            <el-descriptions :column="2" size="small" border>
              <el-descriptions-item label="发票号">{{ ocrResult.invoice_no }}</el-descriptions-item>
              <el-descriptions-item label="金额">{{ ocrResult.amount }}</el-descriptions-item>
              <el-descriptions-item label="开票日期">{{ ocrResult.date }}</el-descriptions-item>
              <el-descriptions-item label="对方">{{ ocrResult.party }}</el-descriptions-item>
              <el-descriptions-item label="置信度">
                <el-tag :type="ocrResult.confidence > 85 ? 'success' : 'warning'" size="small">
                  {{ ocrResult.confidence }}%
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
            <el-alert v-if="fieldErrors.length" :title="`发现 ${fieldErrors.length} 个字段问题`" type="warning" :closable="false" show-icon style="margin-top:8px">
              <ul>
                <li v-for="(err, i) in fieldErrors" :key="i">{{ err }}</li>
              </ul>
            </el-alert>
          </div>
        </el-tab-pane>
        <el-tab-pane label="批量导入" name="batch">
          <!-- 上传文件 -->
          <div v-if="!showMapping && !showPreview" class="upload-step">
            <el-upload drag accept=".xlsx,.xls,.csv" :auto-upload="false" :on-change="handleBatchFileChange" :limit="1" :show-file-list="false" class="upload-area">
              <el-icon :size="40"><UploadFilled /></el-icon>
              <p>拖拽 Excel/CSV 文件或点击上传</p>
              <p class="upload-hint">支持 .xlsx / .xls / .csv 格式</p>
            </el-upload>
            <el-alert type="info" :closable="false" show-icon style="margin-top:12px">
              <p><b>新增列（可选）：</b>状态、备注、对应蓝字发票号、是否红字</p>
              <p>红字发票：是否红字填"是"，对应蓝字发票号填原蓝字发票号，红冲时自动建立 link 关系</p>
            </el-alert>
            <div v-if="uploadedFile" class="file-info">
              <el-tag type="success" size="small">已选择</el-tag>
              <span class="file-name">{{ uploadedFile.name }}</span>
              <span class="file-size">({{ (uploadedFile.size / 1024).toFixed(1) }} KB)</span>
              <el-tag :type="formatTagType" size="small" style="margin:0 8px">{{ detectedFormat }}</el-tag>
              <el-button text type="primary" size="small" @click="handleBatchPreview">预览并解析文件</el-button>
            </div>
          </div>

          <!-- 字段映射 -->
          <div v-if="showMapping" class="mapping-step">
            <div class="step-title">字段映射确认</div>
            <p class="step-hint">系统已从文件中读取列名，请确认映射关系（自动匹配相似列名）</p>
            <el-table :data="fieldMappings" size="small" border>
              <el-table-column prop="field" label="系统字段" width="120" />
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
            <div style="margin-top:12px; display:flex; gap:8px;">
              <el-button type="primary" size="small" @click="showPreview = true; showMapping = false">确认映射，预览数据</el-button>
              <el-button size="small" @click="resetBatchImport">重新上传</el-button>
            </div>
          </div>

          <!-- 数据预览 -->
          <div v-if="showPreview" class="preview-step">
            <div class="step-title">数据预览</div>
            <div class="preview-stats">
              <span>总识别: <b>{{ totalRows }}</b> 条</span>
              <span>正常: <b class="success">{{ normalCount }}</b> 条</span>
              <span>异常: <b class="danger">{{ parseErrors.length }}</b> 条</span>
            </div>
            <el-tabs v-model="previewTab">
              <el-tab-pane label="全部数据" name="all">
                <el-table :data="previewData" size="small" border stripe max-height="300">
                  <el-table-column prop="invoice_no" label="发票号" width="140" />
                  <el-table-column prop="invoice_type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag :type="row.invoice_type === 'sale' ? 'success' : 'warning'" size="small">
                        {{ row.invoice_type === 'sale' ? '销项' : '进项' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="customer_name" label="对方单位" min-width="150" />
                  <el-table-column prop="posting_date" label="开票日期" width="100" />
                  <el-table-column prop="total_amount" label="价税合计" width="120" align="right" />
                  <el-table-column prop="tax_amount" label="税额" width="100" align="right" />
                  <el-table-column prop="net_amount" label="不含税金额" width="120" align="right" />
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
          </div>

          <!-- 导入结果 -->
          <div v-if="importResult" class="import-result">
            <el-result :icon="importResult.failed === 0 ? 'success' : 'warning'"
              :title="`导入完成：成功 ${importResult.imported} 条`"
              :sub-title="importResult.failed > 0 ? `失败 ${importResult.failed} 条` : ''">
            </el-result>
            <div v-if="importResult.failed_rows && importResult.failed_rows.length" style="margin-top:8px">
              <h4 style="color:#e6a23c;margin-bottom:8px;">失败详情</h4>
              <el-table :data="importResult.failed_rows" size="small" border max-height="200">
                <el-table-column prop="row" label="行号" width="60" />
                <el-table-column prop="date" label="发票号" width="140" />
                <el-table-column prop="reason" label="原因" min-width="160" />
              </el-table>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="closeUpload">取消</el-button>
        <el-button v-if="uploadTab === 'ocr'" type="primary" :disabled="!ocrResult" @click="confirmUpload">确认保存</el-button>
        <el-button v-else-if="showPreview && !importResult" type="primary" :loading="batchLoading" @click="confirmBeforeImport">确认导入 ({{ normalCount }} 条)</el-button>
        <el-button v-else-if="importResult" type="primary" @click="closeUpload">关闭</el-button>
        <el-button v-else type="primary" :disabled="!uploadedFile" @click="handleBatchPreview">预览文件</el-button>
      </template>
    </el-dialog>

    <!-- 导入确认弹窗 -->
    <el-dialog v-model="showConfirmDialog" title="确认导入" width="420px">
      <div style="padding:8px 0">
        <el-descriptions :column="1" size="small" border>
          <el-descriptions-item label="导入文件">{{ uploadedFile?.name }}</el-descriptions-item>
          <el-descriptions-item label="文件格式">{{ detectedFormat }}</el-descriptions-item>
          <el-descriptions-item label="导入条数">{{ normalCount }} 条</el-descriptions-item>
        </el-descriptions>
        <el-alert v-if="parseErrors.length" title="存在异常记录将被跳过" type="warning" :description="`${parseErrors.length} 条记录无法解析`" show-icon style="margin-top:12px" :closable="false" />
      </div>
      <template #footer>
        <el-button @click="showConfirmDialog = false">取消</el-button>
        <el-button type="primary" :loading="batchLoading" @click="handleBatchImport">确认导入</el-button>
      </template>
    </el-dialog>

    <!-- 详情抽屉 -->
    <el-drawer v-model="showDetail" :title="`发票 ${showDetail?.invoice_no}`" size="400px">
      <template v-if="showDetail">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="发票号">{{ showDetail.invoice_no }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ showDetail.type === 'sale' ? '销项' : '进项' }}</el-descriptions-item>
          <el-descriptions-item label="对方单位">{{ showDetail.customer_name }}</el-descriptions-item>
          <el-descriptions-item label="开票日期">{{ showDetail.posting_date }}</el-descriptions-item>
          <el-descriptions-item label="到期日">{{ showDetail.due_date }}</el-descriptions-item>
          <el-descriptions-item label="价税合计">{{ showDetail.total_amount }}</el-descriptions-item>
          <el-descriptions-item label="税额">{{ showDetail.tax_amount }}</el-descriptions-item>
          <el-descriptions-item label="不含税金额">{{ showDetail.net_amount }}</el-descriptions-item>
          <el-descriptions-item label="未核销金额">{{ showDetail.outstanding }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(showDetail.status)" size="small">{{ statusLabel(showDetail.status) }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <div style="margin-top:16px;text-align:center">
          <el-button type="primary" :loading="genLoading" @click="handleGenerateVoucher(showDetail)">
            生成凭证
          </el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import request from '@/api/request'
import { generateVoucherFromInvoice, uploadInvoice, importInvoicesFile, previewInvoiceExcel, confirmSalesInvoice } from '@/api/modules/invoice'

const activeTab = ref('sales')

interface InvoiceLineItem {
  id: string
  item_code: string
  description: string
  quantity: string
  unit: string
  unit_price: string
  net_amount: string
  tax_rate: string
  tax_amount: string
  total_amount: string
}

interface InvoiceItem {
  id: string
  invoice_no: string
  invoice_code?: string
  invoice_category?: string
  type: string
  customer_name: string
  tax_id: string
  posting_date: string
  due_date: string
  total_amount: string
  tax_amount: string
  net_amount: string
  outstanding: string
  status: string
  is_return?: boolean
  source_red_invoice_no?: string
  remark?: string
  line_items?: InvoiceLineItem[]
}

const filter = reactive({
  status: '',
  type: '',
  isReturn: null as boolean | null,
  dateRange: null as [string, string] | null,
  keyword: '',
})

const invoices = ref<InvoiceItem[]>([])
const selectedInvoices = ref<InvoiceItem[]>([])
const confirmLoading = ref(false)
const expandedRows = ref<string[]>([])

function handleExpandChange(row: InvoiceItem, expanded: boolean) {
  const index = expandedRows.value.indexOf(row.id)
  if (expanded && index === -1) {
    expandedRows.value.push(row.id)
  } else if (!expanded && index > -1) {
    expandedRows.value.splice(index, 1)
  }
}

const salesBadge = computed(() => invoices.value.filter(i => i.type === 'sale' && i.status === 'draft').length)
const purchaseBadge = computed(() => invoices.value.filter(i => i.type === 'purchase' && i.status === 'draft').length)
const expenseBadge = computed(() => invoices.value.filter(i => i.type === 'expense' && i.status === 'draft').length)

const draftCount = computed(() => invoices.value.filter(i => i.status === 'draft').length)
const pendingCount = computed(() => invoices.value.filter(i => i.status === 'submitted').length)
const confirmedCount = computed(() => invoices.value.filter(i => i.status === 'verified').length)
const partialCount = computed(() => invoices.value.filter(i => i.status === 'partially_paid').length)
const paidCount = computed(() => invoices.value.filter(i => i.status === 'paid').length)

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    const params: any = {}
    if (filter.status) params.status = filter.status
    if (filter.type) params.type = filter.type
    if (filter.isReturn !== null) params.is_return = filter.isReturn
    if (filter.keyword) params.keyword = filter.keyword
    if (filter.dateRange && filter.dateRange.length === 2) {
      params.from_date = filter.dateRange[0]
      params.to_date = filter.dateRange[1]
    }
    if (activeTab.value === 'sales') params.type = 'sale'
    else if (activeTab.value === 'purchase') params.type = 'purchase'
    else if (activeTab.value === 'expense') params.type = 'expense'

    const res: any = await request.get('/invoices', { params })
    const list = res?.list
    if (Array.isArray(list)) { invoices.value = list }
  } catch { /* no data */ }
}

function handleTabChange() {
  loadData()
}

const showUpload = ref(false)
const uploadTab = ref('batch')
const showDetail = ref<InvoiceItem | null>(null)
const showManualCreate = ref(false)
const ocrResult = ref<any>(null)
const fieldErrors = ref<string[]>([])
const genLoading = ref(false)
const batchLoading = ref(false)
const importResult = ref<{
  total_rows: number
  imported: number
  failed: number
  failed_rows?: Array<{ row: number; date?: string; reason: string }>
} | null>(null)

async function handleGenerateVoucher(invoice: InvoiceItem | null) {
  if (!invoice) return
  genLoading.value = true
  try {
    const res = await generateVoucherFromInvoice(invoice.invoice_no)
    ElMessage.success(`凭证已生成: ${res?.data?.voucher_no || ''}`)
    await loadData()
  } catch {
    ElMessage.error('凭证生成失败')
  }
  genLoading.value = false
}

async function handleConfirmInvoice(invoice: InvoiceItem) {
  confirmLoading.value = true
  try {
    await confirmSalesInvoice(invoice.id)
    ElMessage.success(`发票 ${invoice.invoice_no} 已确认，将生成应收账款记录`)
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '确认失败')
  }
  confirmLoading.value = false
}

async function batchConfirm() {
  if (selectedInvoices.value.length === 0) {
    ElMessage.warning('请先选择要确认的发票')
    return
  }
  confirmLoading.value = true
  let successCount = 0
  for (const invoice of selectedInvoices.value) {
    try {
      await confirmSalesInvoice(invoice.id)
      successCount++
    } catch { /* continue */ }
  }
  ElMessage.success(`成功确认 ${successCount} 条发票`)
  selectedInvoices.value = []
  await loadData()
  confirmLoading.value = false
}

async function batchDelete() {
  if (selectedInvoices.value.length === 0) {
    ElMessage.warning('请先选择要删除的发票')
    return
  }
  confirmLoading.value = true
  let successCount = 0
  for (const invoice of selectedInvoices.value) {
    try {
      await request.delete(`/invoices/${invoice.id}`)
      successCount++
    } catch { /* continue */ }
  }
  ElMessage.success(`成功删除 ${successCount} 条发票`)
  selectedInvoices.value = []
  await loadData()
  confirmLoading.value = false
}

function handleSelectionChange(val: InvoiceItem[]) {
  selectedInvoices.value = val
}

/** 字段逻辑校验：金额+税额=价税合计、发票号格式、日期合理性 */
function validateInvoiceFields(data: any): string[] {
  const errors: string[] = []
  // 发票号格式（增值税发票20位）
  if (!/^\d{20}$/.test(data.invoice_no)) {
    errors.push('发票号格式不正确（应为20位数字）')
  }
  // 金额+税额=价税合计
  const net = parseFloat(data.net_amount?.replace(/,/g, '')) || 0
  const tax = parseFloat(data.tax_amount?.replace(/,/g, '')) || 0
  const total = parseFloat(data.total_amount?.replace(/,/g, '')) || 0
  if (Math.abs(net + tax - total) > 0.01) {
    errors.push(`金额校验失败：${net} + ${tax} ≠ ${total}`)
  }
  // 日期合理性
  if (data.date && new Date(data.date) > new Date()) {
    errors.push('开票日期不能晚于当前日期')
  }
  return errors
}

function isExpiring(date: string) {
  if (!date) return false
  const days = (new Date(date).getTime() - Date.now()) / (1000 * 86400)
  return days > 0 && days < 30
}

function statusTag(s: string) {
  const map: Record<string, string> = { 
    draft: 'info', 
    submitted: 'warning', 
    verified: 'success',
    unpaid: 'danger', 
    partially_paid: 'warning', 
    paid: 'success' 
  }
  return map[s] || 'info'
}

function statusLabel(s: string) {
  const map: Record<string, string> = { 
    draft: '草稿',
    submitted: '待确认',
    verified: '已确认',
    unpaid: '待核销', 
    partially_paid: '部分核销', 
    paid: '已核销' 
  }
  return map[s] || s
}

async function handleUpload(file: any) {
  try {
    const res = await uploadInvoice(file.raw || file)
    ocrResult.value = res?.data
    fieldErrors.value = ocrResult.value ? validateInvoiceFields(ocrResult.value) : []
  } catch {
    ElMessage.warning('OCR 识别暂不可用，请手动输入发票信息')
  }
}

function confirmUpload() {
  if (!ocrResult.value) {
    ElMessage.warning('请先上传发票文件')
    return
  }
  ElMessage.success('发票上传成功')
  showUpload.value = false
  ocrResult.value = null
}

// Excel预览相关状态
const previewLoading = ref(false)
const uploadedFile = ref<File | null>(null)
const detectedFormat = ref('')
const showMapping = ref(false)
const showPreview = ref(false)
const showConfirmDialog = ref(false)
const fileColumns = ref<string[]>([])
const sampleData = ref<string[][]>([])
const totalRows = ref(0)
const previewTab = ref('all')

interface FieldMapping {
  field: string
  fieldKey: string
  required: boolean
  matched: boolean
  mappedColumn: string
  sampleValue: string
}

const defaultMappings: FieldMapping[] = [
  { field: '发票号', fieldKey: 'invoice_no', required: true, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '发票类型', fieldKey: 'invoice_type', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '开票日期', fieldKey: 'posting_date', required: true, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '对方单位', fieldKey: 'customer_name', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '价税合计', fieldKey: 'total_amount', required: true, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '税额', fieldKey: 'tax_amount', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '不含税金额', fieldKey: 'net_amount', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '税率', fieldKey: 'tax_rate', required: false, matched: false, mappedColumn: '', sampleValue: '' },
  { field: '备注', fieldKey: 'remark', required: false, matched: false, mappedColumn: '', sampleValue: '' },
]

const fieldMappings = ref<FieldMapping[]>([...defaultMappings])
const parseErrors = ref<{ row: number; field: string; value: string; issue: string }[]>([])

const formatTagType = computed(() => {
  const map: Record<string, string> = { 'CSV': '', 'Excel': 'success' }
  return map[detectedFormat.value] || ''
})

const normalCount = computed(() => totalRows.value - parseErrors.value.length)

function handleBatchFileChange(file: any) {
  uploadedFile.value = file.raw
  const name = file.name.toLowerCase()
  if (name.endsWith('.csv')) detectedFormat.value = 'CSV'
  else if (name.endsWith('.xlsx') || name.endsWith('.xls')) detectedFormat.value = 'Excel'
}

async function handleBatchPreview() {
  if (!uploadedFile.value) {
    ElMessage.warning('请先上传文件')
    return
  }

  previewLoading.value = true
  try {
    const res = await previewInvoiceExcel(uploadedFile.value)
    let actualData = res as any
    if (actualData?.data?.columns) {
      actualData = actualData.data
    } else if (actualData?.columns) {
      actualData = res
    }

    if (actualData?.columns && Array.isArray(actualData.columns)) {
      fileColumns.value = actualData.columns
      sampleData.value = actualData.sample || []
      totalRows.value = actualData.total_rows || 0
      autoMatchColumns()
      showMapping.value = true
      showPreview.value = false
      ElMessage.success(`文件解析成功，共识别 ${totalRows.value} 条记录`)
    } else {
      ElMessage.error('文件解析失败：列名数据格式不正确')
    }
  } catch (e) {
    console.error('Preview failed:', e)
    ElMessage.error('预览文件失败，请检查文件格式')
  }
  previewLoading.value = false
}

function autoMatchColumns() {
  const columns = fileColumns.value.map(c => c.toLowerCase().trim())

  fieldMappings.value.forEach(mapping => {
    const fieldLower = mapping.field.toLowerCase()
    let matchIdx = columns.findIndex(col => col === fieldLower)

    if (matchIdx === -1) {
      matchIdx = columns.findIndex(col =>
        col.includes(fieldLower) || fieldLower.includes(col)
      )
    }

    if (matchIdx === -1) {
      const synonyms: Record<string, string[]> = {
        invoice_no: ['发票号', '发票号码', '数电发票号码', 'invoice no', 'invoice_number'],
        invoice_type: ['发票类型', '类型', 'type', '发票票种', '票种'],
        posting_date: ['开票日期', '日期', '开票日', 'date'],
        customer_name: ['对方单位', '购买方名称', '客户名称', '购方名称', '供应商', '客户', 'customer'],
        total_amount: ['价税合计', '金额', '合计', 'total', '含税金额'],
        tax_amount: ['税额', '税金', '税', 'tax'],
        net_amount: ['不含税金额', '金额', '净额', 'net'],
        tax_rate: ['税率', 'tax rate', 'rate'],
        remark: ['备注', '说明', 'remark', 'remarks'],
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

const previewData = computed(() => {
  if (sampleData.value.length === 0) return []

  const idxMap: Record<string, number> = {}
  fieldMappings.value.forEach(mapping => {
    if (mapping.mappedColumn) {
      idxMap[mapping.fieldKey] = fileColumns.value.indexOf(mapping.mappedColumn)
    }
  })

  return sampleData.value.map((row, idx) => ({
    invoice_no: idxMap['invoice_no'] !== undefined && row[idxMap['invoice_no']] ? row[idxMap['invoice_no']] : '-',
    invoice_type: idxMap['invoice_type'] !== undefined && row[idxMap['invoice_type']] ? (row[idxMap['invoice_type']].includes('销') ? 'sale' : 'purchase') : 'purchase',
    customer_name: idxMap['customer_name'] !== undefined && row[idxMap['customer_name']] ? row[idxMap['customer_name']] : '-',
    posting_date: idxMap['posting_date'] !== undefined && row[idxMap['posting_date']] ? row[idxMap['posting_date']] : '-',
    total_amount: idxMap['total_amount'] !== undefined && row[idxMap['total_amount']] ? row[idxMap['total_amount']] : '0',
    tax_amount: idxMap['tax_amount'] !== undefined && row[idxMap['tax_amount']] ? row[idxMap['tax_amount']] : '0',
    net_amount: idxMap['net_amount'] !== undefined && row[idxMap['net_amount']] ? row[idxMap['net_amount']] : '0',
  }))
})

function skipRow(idx: number) {
  parseErrors.value.splice(idx, 1)
  totalRows.value--
  ElMessage.info('已跳过该异常记录')
}

function editRow(idx: number) {
  ElMessage.success('请在发票管理页面补录该记录')
}

function resetBatchImport() {
  uploadedFile.value = null
  showMapping.value = false
  showPreview.value = false
  showConfirmDialog.value = false
  importResult.value = null
  parseErrors.value = []
  detectedFormat.value = ''
  fileColumns.value = []
  sampleData.value = []
  totalRows.value = 0
  fieldMappings.value = [...defaultMappings]
}

function confirmBeforeImport() {
  if (!uploadedFile.value) {
    ElMessage.warning('请先上传文件')
    return
  }
  showConfirmDialog.value = true
}

async function handleBatchImport() {
  if (!uploadedFile.value) {
    ElMessage.warning('请先上传文件')
    return
  }
  showConfirmDialog.value = false
  batchLoading.value = true
  try {
    const res = await importInvoicesFile(uploadedFile.value)
    const result = (res as any).data || res
    importResult.value = result
    showPreview.value = false
    ElMessage.success(`成功导入 ${result.imported} 条发票`)
    if (result.imported > 0) loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '批量导入失败')
  }
  batchLoading.value = false
}

function closeUpload() {
  showUpload.value = false
  uploadTab.value = 'batch'
  ocrResult.value = null
  importResult.value = null
  resetBatchImport()
}
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.stats-card { margin-bottom: 16px; }
.stats-row {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  align-items: center;
}
.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
  .stat-label { color: #666; font-size: 14px; }
  .stat-value {
    font-size: 18px;
    font-weight: 600;
    min-width: 24px;
    text-align: center;
    &.draft { color: #909399; }
    &.pending { color: #e6a23c; }
    &.confirmed { color: #409eff; }
    &.partial { color: #faad14; }
    &.paid { color: #67c23a; }
  }
}
.filter-card {
  margin-bottom: 16px;
}
.upload-area { text-align: center; p { margin: 8px 0; } .upload-hint { color: #999; font-size: 12px; } }
.ocr-result {
  margin-top: 16px;
  h4 { margin-bottom: 12px; }
}
.import-result { margin-top: 16px; }
.expiring { color: #faad14; font-weight: 600; }

.file-info { margin-top: 12px; display: flex; align-items: center; flex-wrap: wrap; gap: 4px; .file-name { font-weight: 500; } .file-size { color: #999; font-size: 12px; } }
.step-title { font-weight: 600; margin-bottom: 12px; font-size: 14px; color: #333; }
.step-hint { color: #999; font-size: 12px; margin-bottom: 12px; }
.preview-stats { display: flex; gap: 20px; margin-bottom: 12px; font-size: 13px; .success { color: #52c41a; } .danger { color: #ff4d4f; } }
.sample-value { font-size: 12px; color: #666; }
.no-sample { color: #999; }
</style>
