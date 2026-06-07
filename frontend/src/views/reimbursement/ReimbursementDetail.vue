<template>
  <div class="reimbursement-detail-page">
    <div class="page-header">
      <h3>报销单详情</h3>
      <DocStatusTag v-if="reimbursement" :docstatus="reimbursement.doc_status" size="default" />
    </div>

    <el-tabs v-if="reimbursement" v-model="activeTab" class="detail-tabs">
      <!-- ============== Tab 1: 基本信息 ============== -->
      <el-tab-pane label="基本信息" name="basic">
        <el-card>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="申请人">{{ reimbursement.applicant_name }}</el-descriptions-item>
            <el-descriptions-item label="部门">{{ reimbursement.department_name }}</el-descriptions-item>
            <el-descriptions-item label="报销金额">
              <b>¥{{ formatAmount(reimbursement.amount) }}</b>
            </el-descriptions-item>
            <el-descriptions-item label="说明">{{ reimbursement.description }}</el-descriptions-item>
            <el-descriptions-item label="单据状态">
              <DocStatusTag :docstatus="reimbursement.doc_status" />
            </el-descriptions-item>
            <el-descriptions-item label="凭证ID">
              <template v-if="reimbursement.voucher_id">
                <el-link type="primary" :underline="false" @click="goVoucher(reimbursement.voucher_id)">
                  {{ reimbursement.voucher_id }}
                </el-link>
              </template>
              <span v-else>—</span>
            </el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ reimbursement.remark || '—' }}</el-descriptions-item>
            <el-descriptions-item label="创建人">{{ reimbursement.created_by }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ reimbursement.created_at }}</el-descriptions-item>
          </el-descriptions>

          <div class="detail-actions">
            <el-button @click="$router.back()">返回</el-button>

            <!-- 草稿: 提交按钮 -->
            <template v-if="reimbursement.doc_status === 0">
              <el-button type="primary" :loading="actionLoading" @click="handleSubmit">提交</el-button>
              <el-button type="warning" :loading="actionLoading" @click="$router.push(`/expense/reimbursement/${reimbursement!.id}/edit`)">编辑</el-button>
            </template>

            <!-- 已提交: 审核按钮 -->
            <template v-if="reimbursement.doc_status === 1">
              <el-button type="success" :loading="actionLoading" @click="handleApprove">审核通过</el-button>
              <el-button type="danger" :loading="actionLoading" @click="handleReject">驳回</el-button>
            </template>

            <!-- 已审核且有凭证ID: 查看凭证 -->
            <template v-if="reimbursement.doc_status >= 2 && reimbursement.voucher_id">
              <el-button type="primary" @click="goVoucher(reimbursement.voucher_id)">查看凭证</el-button>
            </template>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- ============== Tab 2: 附件管理 ============== -->
      <el-tab-pane label="附件管理" name="attachments">
        <el-card>
          <el-upload
            class="upload-dragger"
            drag
            multiple
            :show-file-list="false"
            :http-request="handleUploadAttachment"
            accept=".jpg,.png,.pdf"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持 jpg / png / pdf 格式
              </div>
            </template>
          </el-upload>

          <el-table
            v-loading="attachmentLoading"
            :data="attachments"
            border
            stripe
            style="margin-top: 16px;"
            empty-text="暂无附件"
          >
            <el-table-column prop="file_name" label="文件名" min-width="200" show-overflow-tooltip />
            <el-table-column label="大小" width="120" align="right">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column prop="mime_type" label="类型" width="140" show-overflow-tooltip />
            <el-table-column prop="uploaded_by" label="上传人" width="120" />
            <el-table-column prop="uploaded_at" label="上传时间" width="180" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleDownloadAttachment(row)">下载</el-button>
                <el-popconfirm
                  title="确定删除该附件？"
                  confirm-button-text="确定"
                  cancel-button-text="取消"
                  @confirm="handleDeleteAttachment(row)"
                >
                  <template #reference>
                    <el-button link type="danger">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- ============== Tab 3: 发票关联 ============== -->
      <el-tab-pane label="发票关联" name="invoices">
        <el-card>
          <div class="tab-toolbar">
            <el-input
              v-model="linkedKeyword"
              placeholder="搜索已关联发票号/供应商"
              clearable
              style="width: 280px;"
              @keyup.enter="loadLinkedInvoices"
              @clear="loadLinkedInvoices"
            >
              <template #prefix>
                <el-icon><search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" :icon="Plus" @click="openLinkDialog">添加关联</el-button>
            <el-button :icon="Refresh" @click="loadLinkedInvoices">刷新</el-button>
          </div>

          <el-table
            v-loading="linkedLoading"
            :data="filteredLinkedInvoices"
            border
            stripe
            empty-text="暂未关联进项发票"
          >
            <el-table-column prop="invoice_no" label="发票号" min-width="180" show-overflow-tooltip />
            <el-table-column prop="invoice_date" label="开票日期" width="120" />
            <el-table-column prop="vendor_name" label="供应商" min-width="180" show-overflow-tooltip />
            <el-table-column label="价税合计" width="140" align="right">
              <template #default="{ row }">¥{{ formatAmount(row.total_amount) }}</template>
            </el-table-column>
            <el-table-column label="验真状态" width="120" align="center">
              <template #default="{ row }">
                <el-tag :type="verifyStatusTag(row.verify_status)">{{ verifyStatusLabel(row.verify_status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-popconfirm
                  title="确定取消该发票关联？"
                  confirm-button-text="确定"
                  cancel-button-text="取消"
                  @confirm="handleUnlinkInvoice(row)"
                >
                  <template #reference>
                    <el-button link type="danger">取消关联</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-empty v-else-if="!loading" description="未找到报销单数据" />

    <!-- 关联进项发票弹窗 -->
    <el-dialog
      v-model="linkDialogVisible"
      title="添加关联进项发票"
      width="900px"
      destroy-on-close
      @closed="resetLinkDialog"
    >
      <div class="link-toolbar">
        <el-input
          v-model="availableKeyword"
          placeholder="搜索可关联发票号/供应商"
          clearable
          style="width: 280px;"
          @keyup.enter="loadAvailableInvoices"
          @clear="loadAvailableInvoices"
        >
          <template #prefix>
            <el-icon><search /></el-icon>
          </template>
        </el-input>
        <el-button :icon="Refresh" @click="loadAvailableInvoices">刷新</el-button>
        <span class="link-tip">已选择 {{ selectedInvoiceIds.length }} 条</span>
      </div>

      <el-table
        v-loading="availableLoading"
        :data="filteredAvailableInvoices"
        border
        stripe
        height="420"
        empty-text="暂无可关联的进项发票"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="invoice_no" label="发票号" min-width="180" show-overflow-tooltip />
        <el-table-column prop="invoice_date" label="开票日期" width="120" />
        <el-table-column prop="vendor_name" label="供应商" min-width="200" show-overflow-tooltip />
        <el-table-column label="价税合计" width="140" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.total_amount) }}</template>
        </el-table-column>
        <el-table-column label="验真状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="verifyStatusTag(row.verify_status)">{{ verifyStatusLabel(row.verify_status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="linkDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="linkSubmitting"
          :disabled="selectedInvoiceIds.length === 0"
          @click="handleConfirmLink"
        >
          确定关联 ({{ selectedInvoiceIds.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UploadFilled, Search, Plus, Refresh } from '@element-plus/icons-vue'
import {
  fetchReimbursementDetail,
  submitReimbursement,
  approveReimbursement,
  rejectReimbursement,
  uploadReimbursementAttachment,
  listReimbursementAttachments,
  deleteReimbursementAttachment,
  downloadReimbursementAttachment,
  listAvailableInvoices,
  linkInvoice,
  unlinkInvoice,
  listLinkedInvoices,
  type ReimbursementAttachment,
  type LinkedInvoice,
} from '@/api/modules/reimbursement'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { Reimbursement } from '@/api/modules/reimbursement'

const route = useRoute()
const router = useRouter()
const reimbursement = ref<Reimbursement | null>(null)
const loading = ref(false)
const actionLoading = ref(false)
const activeTab = ref('basic')

/* ============== 附件 ============== */
const attachments = ref<ReimbursementAttachment[]>([])
const attachmentLoading = ref(false)

async function loadAttachments() {
  if (!reimbursement.value) return
  attachmentLoading.value = true
  try {
    const res: any = await listReimbursementAttachments(reimbursement.value.id)
    attachments.value = (res?.data || res || []) as ReimbursementAttachment[]
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载附件失败')
  } finally {
    attachmentLoading.value = false
  }
}

async function handleUploadAttachment(option: any) {
  if (!reimbursement.value) return
  const file: File | undefined = option?.file
  if (!file) return
  try {
    await uploadReimbursementAttachment(reimbursement.value.id, file)
    ElMessage.success(`上传成功: ${file.name}`)
    await loadAttachments()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || `上传失败: ${file.name}`)
  }
}

async function handleDownloadAttachment(row: ReimbursementAttachment) {
  if (!reimbursement.value) return
  try {
    const blob: Blob = await downloadReimbursementAttachment(reimbursement.value.id, row.id)
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = row.file_name || '附件'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '下载失败')
  }
}

async function handleDeleteAttachment(row: ReimbursementAttachment) {
  if (!reimbursement.value) return
  try {
    await deleteReimbursementAttachment(reimbursement.value.id, row.id)
    ElMessage.success('删除成功')
    await loadAttachments()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  }
}

function formatFileSize(bytes: any): string {
  const n = Number(bytes) || 0
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(2)} KB`
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(2)} MB`
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`
}

/* ============== 发票关联 ============== */
const linkedInvoices = ref<LinkedInvoice[]>([])
const linkedLoading = ref(false)
const linkedKeyword = ref('')

const filteredLinkedInvoices = computed(() => {
  const kw = linkedKeyword.value.trim().toLowerCase()
  if (!kw) return linkedInvoices.value
  return linkedInvoices.value.filter(
    (it) =>
      (it.invoice_no || '').toLowerCase().includes(kw) ||
      (it.vendor_name || '').toLowerCase().includes(kw),
  )
})

async function loadLinkedInvoices() {
  if (!reimbursement.value) return
  linkedLoading.value = true
  try {
    const res: any = await listLinkedInvoices(reimbursement.value.id)
    linkedInvoices.value = (res?.data || res || []) as LinkedInvoice[]
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载已关联发票失败')
  } finally {
    linkedLoading.value = false
  }
}

async function handleUnlinkInvoice(row: LinkedInvoice) {
  if (!reimbursement.value) return
  try {
    await unlinkInvoice(reimbursement.value.id, row.id)
    ElMessage.success('已取消关联')
    await loadLinkedInvoices()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '取消关联失败')
  }
}

/* ----- 关联弹窗 ----- */
const linkDialogVisible = ref(false)
const availableInvoices = ref<LinkedInvoice[]>([])
const availableLoading = ref(false)
const availableKeyword = ref('')
const selectedInvoiceIds = ref<string[]>([])
const linkSubmitting = ref(false)

const filteredAvailableInvoices = computed(() => {
  const kw = availableKeyword.value.trim().toLowerCase()
  if (!kw) return availableInvoices.value
  return availableInvoices.value.filter(
    (it) =>
      (it.invoice_no || '').toLowerCase().includes(kw) ||
      (it.vendor_name || '').toLowerCase().includes(kw),
  )
})

async function openLinkDialog() {
  linkDialogVisible.value = true
  await loadAvailableInvoices()
}

async function loadAvailableInvoices() {
  if (!reimbursement.value) return
  availableLoading.value = true
  try {
    const res: any = await listAvailableInvoices(reimbursement.value.id, { pageSize: 200 })
    const data = res?.data
    if (Array.isArray(data)) {
      availableInvoices.value = data as LinkedInvoice[]
    } else if (data && Array.isArray(data.list)) {
      availableInvoices.value = data.list as LinkedInvoice[]
    } else {
      availableInvoices.value = []
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载可关联发票失败')
  } finally {
    availableLoading.value = false
  }
}

function onSelectionChange(rows: LinkedInvoice[]) {
  selectedInvoiceIds.value = rows.map((r) => r.id)
}

function resetLinkDialog() {
  availableInvoices.value = []
  availableKeyword.value = ''
  selectedInvoiceIds.value = []
}

async function handleConfirmLink() {
  if (!reimbursement.value || selectedInvoiceIds.value.length === 0) return
  linkSubmitting.value = true
  try {
    let ok = 0
    for (const invId of selectedInvoiceIds.value) {
      try {
        await linkInvoice(reimbursement.value.id, invId)
        ok++
      } catch (e: any) {
        ElMessage.warning(e?.response?.data?.error || `关联 ${invId} 失败`)
      }
    }
    ElMessage.success(`成功关联 ${ok} 张发票`)
    linkDialogVisible.value = false
    await loadLinkedInvoices()
  } finally {
    linkSubmitting.value = false
  }
}

/* ----- 验真状态展示 ----- */
function verifyStatusLabel(s: any): string {
  const v = String(s || '').toLowerCase()
  if (v === 'verified' || v === 'success' || v === 'passed') return '已验真'
  if (v === 'failed' || v === 'fail') return '验真失败'
  if (v === 'pending' || v === 'processing') return '验真中'
  return s || '未验真'
}
function verifyStatusTag(s: any): 'success' | 'danger' | 'warning' | 'info' {
  const v = String(s || '').toLowerCase()
  if (v === 'verified' || v === 'success' || v === 'passed') return 'success'
  if (v === 'failed' || v === 'fail') return 'danger'
  if (v === 'pending' || v === 'processing') return 'warning'
  return 'info'
}

/* ============== 基本信息 ============== */
function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchReimbursementDetail(route.params.id as string)
    reimbursement.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    await submitReimbursement(reimbursement.value.id)
    ElMessage.success('提交成功')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '提交失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleApprove() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    const res: any = await approveReimbursement(reimbursement.value.id)
    ElMessage.success('审核通过，凭证已生成')
    if (res?.data?.voucher_id) {
      reimbursement.value.voucher_id = res.data.voucher_id
    }
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '审核失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleReject() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    await rejectReimbursement(reimbursement.value.id)
    ElMessage.success('已驳回')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '驳回失败')
  } finally {
    actionLoading.value = false
  }
}

function goVoucher(voucherId: string) {
  router.push(`/vouchers/${voucherId}`)
}

onMounted(async () => {
  await loadData()
  if (reimbursement.value) {
    await Promise.all([loadAttachments(), loadLinkedInvoices()])
  }
})
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.detail-tabs {
  :deep(.el-tabs__content) {
    overflow: visible;
  }
}
.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}
.upload-dragger {
  :deep(.el-upload) {
    width: 100%;
  }
  :deep(.el-upload-dragger) {
    width: 100%;
  }
}
.tab-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.link-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  .link-tip {
    margin-left: auto;
    color: #909399;
    font-size: 13px;
  }
}
</style>
