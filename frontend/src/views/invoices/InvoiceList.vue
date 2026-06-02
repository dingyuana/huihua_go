<template>
  <div class="invoice-page">
    <div class="page-header">
      <h3>发票管理</h3>
      <el-button
        type="primary"
        @click="showUpload = true"
      >
        + 上传发票
      </el-button>
    </div>

    <!-- 进项税务池 Tabs -->
    <el-card class="tax-tabs-card">
      <el-tabs v-model="invoiceView">
        <el-tab-pane
          label="发票列表"
          name="list"
        />
        <el-tab-pane
          label="进项税务池"
          name="tax-pool"
        >
          <el-table
            :data="taxPool"
            border
            stripe
            size="small"
          >
            <el-table-column
              prop="invoice_no"
              label="发票号"
              width="140"
            />
            <el-table-column
              prop="supplier"
              label="供应商"
              min-width="140"
            />
            <el-table-column
              prop="amount"
              label="金额"
              width="120"
              align="right"
            />
            <el-table-column
              label="认证状态"
              width="130"
            >
              <template #default="{ row }">
                <el-tag
                  :type="taxStatusTag(row.taxStatus)"
                  size="small"
                >
                  {{ row.taxStatusLabel }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="due_date"
              label="到期日"
              width="100"
            >
              <template #default="{ row }">
                <span :class="{ expiring: row.daysLeft > 0 && row.daysLeft < 30 }">{{ row.due_date }}</span>
                <el-tag
                  v-if="row.daysLeft > 0 && row.daysLeft < 30"
                  type="danger"
                  size="small"
                  style="margin-left:4px"
                >
                  {{ row.daysLeft }}天
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              label="操作"
              width="80"
            >
              <template #default="{ row }">
                <el-button
                  v-if="row.taxStatus === 'unverified'"
                  size="small"
                  type="primary"
                  link
                  @click="markVerified(row)"
                >
                  认证
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 筛选 -->
    <el-card
      v-if="invoiceView === 'list'"
      class="filter-card"
    >
      <el-form
        :inline="true"
        size="small"
      >
        <el-form-item label="状态">
          <el-select
            v-model="filter.status"
            placeholder="全部"
            style="width: 130px"
            clearable
          >
            <el-option
              label="待核销"
              value="unpaid"
            />
            <el-option
              label="部分核销"
              value="partially_paid"
            />
            <el-option
              label="已核销"
              value="paid"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select
            v-model="filter.type"
            placeholder="全部"
            style="width: 120px"
            clearable
          >
            <el-option
              label="销项"
              value="sale"
            />
            <el-option
              label="进项"
              value="purchase"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker
            v-model="filter.dateRange"
            type="daterange"
            range-separator="~"
            start-placeholder="开始"
            end-placeholder="结束"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="filter.keyword"
            placeholder="发票号/对方名称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            @click="loadData"
          >
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card v-if="invoiceView === 'list'">
      <el-table
        :data="invoices"
        border
        stripe
        size="small"
      >
        <el-table-column
          prop="invoice_no"
          label="发票号"
          width="140"
        />
        <el-table-column
          label="类型"
          width="70"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.type === 'sale' ? 'success' : 'warning'"
              size="small"
            >
              {{ row.type === 'sale' ? '销项' : '进项' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="customer_name"
          label="对方单位"
          min-width="150"
        />
        <el-table-column
          prop="posting_date"
          label="开票日期"
          width="100"
        />
        <el-table-column
          prop="total_amount"
          label="价税合计"
          width="120"
          align="right"
        />
        <el-table-column
          label="未核销"
          width="120"
          align="right"
        >
          <template #default="{ row }">
            <span
              v-if="parseFloat(row.outstanding) > 0"
              class="amount-positive"
            >{{ row.outstanding }}</span>
            <span
              v-else
              class="amount-negative"
            >已核销</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="due_date"
          label="到期日"
          width="100"
        >
          <template #default="{ row }">
            <span :class="{ 'expiring': isExpiring(row.due_date) }">{{ row.due_date }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="状态"
          width="80"
        >
          <template #default="{ row }">
            <el-tag
              :type="statusTag(row.status)"
              size="small"
            >
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="80"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              @click="showDetail = row"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog
      v-model="showUpload"
      title="上传发票"
      width="500px"
    >
      <el-tabs v-model="uploadTab">
        <el-tab-pane
          label="OCR识别"
          name="ocr"
        >
          <el-upload
            drag
            accept=".pdf,.ofd,.jpg,.png"
            :auto-upload="false"
            :on-change="handleUpload"
            class="upload-area"
          >
            <el-icon :size="40">
              <UploadFilled />
            </el-icon>
            <p>拖拽发票文件或点击上传</p>
            <p class="upload-hint">
              支持 PDF / OFD / 图片格式
            </p>
          </el-upload>
          <div
            v-if="ocrResult"
            class="ocr-result"
          >
            <h4>OCR 识别结果</h4>
            <el-descriptions
              :column="2"
              size="small"
              border
            >
              <el-descriptions-item label="发票号">
                {{ ocrResult.invoice_no }}
              </el-descriptions-item>
              <el-descriptions-item label="金额">
                {{ ocrResult.amount }}
              </el-descriptions-item>
              <el-descriptions-item label="开票日期">
                {{ ocrResult.date }}
              </el-descriptions-item>
              <el-descriptions-item label="对方">
                {{ ocrResult.party }}
              </el-descriptions-item>
              <el-descriptions-item label="置信度">
                <el-tag
                  :type="ocrResult.confidence > 85 ? 'success' : 'warning'"
                  size="small"
                >
                  {{ ocrResult.confidence }}%
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
            <el-alert
              v-if="fieldErrors.length"
              :title="`发现 ${fieldErrors.length} 个字段问题`"
              type="warning"
              :closable="false"
              show-icon
              style="margin-top:8px"
            >
              <ul>
                <li
                  v-for="(err, i) in fieldErrors"
                  :key="i"
                >
                  {{ err }}
                </li>
              </ul>
            </el-alert>
          </div>
        </el-tab-pane>
        <el-tab-pane
          label="批量导入"
          name="batch"
        >
          <el-upload
            drag
            accept=".xlsx,.xls,.csv"
            :auto-upload="false"
            :on-change="handleBatchUpload"
            :limit="1"
            :show-file-list="false"
            class="upload-area"
          >
            <el-icon :size="40">
              <UploadFilled />
            </el-icon>
            <p>拖拽 Excel/CSV 文件或点击上传</p>
            <p class="upload-hint">
              支持 .xlsx / .xls / .csv 格式
            </p>
          </el-upload>
          <div
            v-if="importResult"
            class="import-result"
          >
            <el-result
              :icon="importResult.failed === 0 ? 'success' : 'warning'"
              :title="`导入完成：成功 ${importResult.imported} 条`"
              :sub-title="importResult.failed > 0 ? `失败 ${importResult.failed} 条` : ''"
            />
            <div
              v-if="importResult.failed_rows && importResult.failed_rows.length"
              style="margin-top:8px"
            >
              <h4 style="color:#e6a23c;margin-bottom:8px;">
                失败详情
              </h4>
              <el-table
                :data="importResult.failed_rows"
                size="small"
                border
                max-height="200"
              >
                <el-table-column
                  prop="row"
                  label="行号"
                  width="60"
                />
                <el-table-column
                  prop="date"
                  label="发票号"
                  width="140"
                />
                <el-table-column
                  prop="reason"
                  label="原因"
                  min-width="160"
                />
              </el-table>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="closeUpload">
          取消
        </el-button>
        <el-button
          v-if="uploadTab === 'ocr'"
          type="primary"
          :disabled="!ocrResult"
          @click="confirmUpload"
        >
          确认保存
        </el-button>
        <el-button
          v-else
          type="primary"
          :loading="batchLoading"
          :disabled="batchLoading"
          @click="closeUpload"
        >
          关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 详情抽屉 -->
    <el-drawer
      v-model="showDetail"
      :title="`发票 ${showDetail?.invoice_no}`"
      size="400px"
    >
      <template v-if="showDetail">
        <el-descriptions
          :column="1"
          border
          size="small"
        >
          <el-descriptions-item label="发票号">
            {{ showDetail.invoice_no }}
          </el-descriptions-item>
          <el-descriptions-item label="类型">
            {{ showDetail.type === 'sale' ? '销项' : '进项' }}
          </el-descriptions-item>
          <el-descriptions-item label="对方单位">
            {{ showDetail.customer_name }}
          </el-descriptions-item>
          <el-descriptions-item label="开票日期">
            {{ showDetail.posting_date }}
          </el-descriptions-item>
          <el-descriptions-item label="到期日">
            {{ showDetail.due_date }}
          </el-descriptions-item>
          <el-descriptions-item label="价税合计">
            {{ showDetail.total_amount }}
          </el-descriptions-item>
          <el-descriptions-item label="税额">
            {{ showDetail.tax_amount }}
          </el-descriptions-item>
          <el-descriptions-item label="不含税金额">
            {{ showDetail.net_amount }}
          </el-descriptions-item>
          <el-descriptions-item label="未核销金额">
            {{ showDetail.outstanding }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag
              :type="statusTag(showDetail.status)"
              size="small"
            >
              {{ statusLabel(showDetail.status) }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <div style="margin-top:16px;text-align:center">
          <el-button
            type="primary"
            :loading="genLoading"
            @click="handleGenerateVoucher(showDetail)"
          >
            生成凭证
          </el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import { generateVoucherFromInvoice, uploadInvoice, importInvoicesFile } from '@/api/modules/invoice'

const invoiceView = ref('list')

interface TaxPoolItem {
  invoice_no: string
  supplier: string
  amount: string
  taxStatus: string
  taxStatusLabel: string
  due_date: string
  daysLeft: number
}

const taxPool = ref<TaxPoolItem[]>([])

function taxStatusTag(s: string) {
  return { unverified: 'warning', verified: 'success', expired: 'danger', deducting: 'primary' }[s] || 'info'
}

function markVerified(item: TaxPoolItem) {
  item.taxStatus = 'verified'
  item.taxStatusLabel = '已认证待抵扣'
  ElMessage.success(`发票 ${item.invoice_no} 已标记为已认证待抵扣`)
}

interface InvoiceItem {
  invoice_no: string
  type: string
  customer_name: string
  posting_date: string
  due_date: string
  total_amount: string
  tax_amount: string
  net_amount: string
  outstanding: string
  status: string
}

const filter = reactive({
  status: '',
  type: '',
  dateRange: null as [string, string] | null,
  keyword: '',
})

const invoices = ref<InvoiceItem[]>([])

onMounted(async () => {
  try {
    const res: any = await request.get('/invoices')
    const list = res?.data?.list || res?.data
    if (Array.isArray(list)) { invoices.value = list }
  } catch { /* no data */ }
})

const showUpload = ref(false)
const uploadTab = ref('ocr')
const showDetail = ref<InvoiceItem | null>(null)
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
  } catch {
    ElMessage.error('凭证生成失败')
  }
  genLoading.value = false
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

function loadData() { /* computed filters */ }

function isExpiring(date: string) {
  if (!date) return false
  const days = (new Date(date).getTime() - Date.now()) / (1000 * 86400)
  return days > 0 && days < 30
}

function statusTag(s: string) {
  const map: Record<string, string> = { unpaid: 'danger', partially_paid: 'warning', paid: 'success' }
  return map[s] || 'info'
}

function statusLabel(s: string) {
  const map: Record<string, string> = { unpaid: '待核销', partially_paid: '部分核销', paid: '已核销' }
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

async function handleBatchUpload(file: any) {
  if (!file.raw) return
  batchLoading.value = true
  importResult.value = null
  try {
    const res = await importInvoicesFile(file.raw)
    importResult.value = res.data
    ElMessage.success(`成功导入 ${res.data.imported} 条发票`)
    if (res.data.imported > 0) loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '批量导入失败')
  }
  batchLoading.value = false
}

function closeUpload() {
  showUpload.value = false
  uploadTab.value = 'ocr'
  ocrResult.value = null
  importResult.value = null
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
</style>
