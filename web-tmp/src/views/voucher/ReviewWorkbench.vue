<template>
  <div class="review-page">
    <div class="page-header">
      <h3>凭证审核工作台</h3>
      <el-tag type="warning" size="large">待审核: {{ pendingCount }} 张</el-tag>
    </div>

    <el-card>
      <!-- 批量操作 -->
      <div class="batch-bar">
        <el-checkbox v-model="selectAll" @change="toggleAll">全选</el-checkbox>
        <span class="selected-count">已选 {{ selectedIds.length }} 张</span>
        <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="batchApprove">审核通过</el-button>
        <el-button size="small" :disabled="selectedIds.length === 0" @click="showRejectDialog = true">驳回</el-button>
      </div>

      <el-table :data="pendingVouchers" border stripe size="small" @selection-change="onSelection" @row-click="openDetail">
        <el-table-column type="selection" width="40" @click.stop />
        <el-table-column prop="voucher_no" label="凭证号" width="140" />
        <el-table-column prop="date" label="日期" width="80" />
        <el-table-column prop="remark" label="摘要" min-width="200" show-overflow-tooltip />
        <el-table-column prop="amount" label="金额" width="120" align="right" />
        <el-table-column prop="creator" label="制单人" width="80" />
        <el-table-column label="AI 风控" width="100">
          <template #default="{ row }">
            <el-popover trigger="hover" :width="280">
              <template #reference>
                <el-tag
                  :type="row.risk.level === 'high' ? 'danger' : row.risk.level === 'low' ? 'success' : 'warning'"
                  size="small"
                  style="cursor:pointer"
                >
                  {{ row.risk.level === 'high' ? '高风险' : row.risk.level === 'low' ? '正常' : '关注' }}
                </el-tag>
              </template>
              <div class="risk-card">
                <h4>AI 风控详情</h4>
                <ul>
                  <li v-for="(item, i) in row.risk.items" :key="i" :class="'risk-' + item.severity">
                    <el-tag :type="item.severity === 'error' ? 'danger' : 'warning'" size="small" style="margin-right:4px">
                      {{ item.severity === 'error' ? '风险' : '提示' }}
                    </el-tag>
                    {{ item.message }}
                  </li>
                </ul>
              </div>
            </el-popover>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 详情抽屉：分屏审核 -->
    <el-drawer v-model="showDetail" :title="`审核凭证: ${currentVoucher?.voucher_no}`" size="600px">
      <template v-if="currentVoucher">
        <el-tabs v-model="detailTab">
          <el-tab-pane label="凭证信息" name="voucher">
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="凭证号">{{ currentVoucher.voucher_no }}</el-descriptions-item>
              <el-descriptions-item label="日期">{{ currentVoucher.date }}</el-descriptions-item>
              <el-descriptions-item label="摘要" :span="2">{{ currentVoucher.remark }}</el-descriptions-item>
              <el-descriptions-item label="制单人">{{ currentVoucher.creator }}</el-descriptions-item>
              <el-descriptions-item label="金额">{{ currentVoucher.amount }}</el-descriptions-item>
            </el-descriptions>

            <h4 class="section-title">分录明细</h4>
            <el-table :data="currentVoucher.lines || []" size="small" border>
              <el-table-column prop="account" label="科目" min-width="160" />
              <el-table-column prop="debit" label="借方" width="120" align="right" />
              <el-table-column prop="credit" label="贷方" width="120" align="right" />
            </el-table>

            <!-- AI 风险详情 -->
            <h4 class="section-title">AI 风控分析</h4>
            <el-card v-if="currentVoucher.risk.items.length" shadow="never" class="risk-detail-card">
              <div v-for="(item, i) in currentVoucher.risk.items" :key="i" class="risk-item">
                <el-tag :type="item.severity === 'error' ? 'danger' : 'warning'" size="small">
                  {{ item.severity === 'error' ? '⚠️ 风险' : '💡 提示' }}
                </el-tag>
                <span class="risk-msg">{{ item.message }}</span>
                <p v-if="item.suggestion" class="risk-suggestion">{{ item.suggestion }}</p>
              </div>
            </el-card>
            <el-empty v-else description="AI 风控未发现异常" :image-size="60" />
          </el-tab-pane>

          <el-tab-pane label="原始单据" name="source">
            <el-empty description="关联原始单据（待对接）" :image-size="60" />
            <p class="source-hint">审核时可联查银行流水/发票等原始单据</p>
          </el-tab-pane>
        </el-tabs>
      </template>

      <template #footer>
        <div class="drawer-actions">
          <el-button @click="showDetail = false">关闭</el-button>
          <el-button type="danger" @click="openRejectFromDrawer">驳回</el-button>
          <el-button type="primary" @click="approveCurrent">审核通过</el-button>
        </div>
      </template>
    </el-drawer>

    <!-- 驳回弹窗 -->
    <el-dialog v-model="showRejectDialog" title="驳回凭证" width="420px">
      <el-form>
        <el-form-item label="驳回原因" required>
          <el-input v-model="rejectReason" type="textarea" :rows="3" placeholder="请填写驳回原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectDialog = false">取消</el-button>
        <el-button type="primary" @click="reject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

interface RiskItem {
  severity: 'error' | 'warning'
  message: string
  suggestion?: string
}

interface VoucherItem {
  id: string
  voucher_no: string
  date: string
  remark: string
  amount: string
  creator: string
  risk: { level: string; items: RiskItem[] }
  lines?: { account: string; debit: string; credit: string }[]
}

const pendingVouchers = ref<VoucherItem[]>([
  {
    id: 'v1', voucher_no: '记-2026-05-0010', date: '05-27',
    remark: '收款-上海XX贸易公司', amount: '12,000.00', creator: '李四',
    risk: {
      level: 'low',
      items: [],
    },
    lines: [
      { account: '1001-01 银行存款-工行', debit: '12,000.00', credit: '' },
      { account: '1122  应收账款', debit: '', credit: '12,000.00' },
    ],
  },
  {
    id: 'v2', voucher_no: '记-2026-05-0011', date: '05-27',
    remark: '付款-北京YY科技', amount: '5,000.00', creator: '李四',
    risk: {
      level: 'medium',
      items: [
        { severity: 'warning', message: '摘要含"工资"但借方科目非"应付职工薪酬"', suggestion: '请确认科目选择是否正确' },
      ],
    },
    lines: [
      { account: '6401  主营业务成本', debit: '5,000.00', credit: '' },
      { account: '1001-01 银行存款-工行', debit: '', credit: '5,000.00' },
    ],
  },
  {
    id: 'v3', voucher_no: '记-2026-05-0012', date: '05-26',
    remark: '大额付款-广州ZZ', amount: '200,000.00', creator: '王五',
    risk: {
      level: 'high',
      items: [
        { severity: 'error', message: '大额整数无零头，金额异常（¥200,000.00）', suggestion: '请核实交易真实性，附上合同或审批单' },
        { severity: 'warning', message: '公转私超过限额 ¥50,000', suggestion: '公转私需主管审批' },
      ],
    },
    lines: [
      { account: '2001  应付账款', debit: '200,000.00', credit: '' },
      { account: '1001-01 银行存款-工行', debit: '', credit: '200,000.00' },
    ],
  },
])

const pendingCount = ref(pendingVouchers.value.length)
const selectAll = ref(false)
const selectedIds = ref<string[]>([])
const showRejectDialog = ref(false)
const rejectReason = ref('')

// 详情抽屉
const showDetail = ref(false)
const detailTab = ref('voucher')
const currentVoucher = ref<VoucherItem | null>(null)

function onSelection(rows: any[]) {
  selectedIds.value = rows.map((r: any) => r.id)
}

function toggleAll() { /* handled by el-table */ }

function openDetail(row: VoucherItem) {
  currentVoucher.value = row
  detailTab.value = 'voucher'
  showDetail.value = true
}

function approveCurrent() {
  if (!currentVoucher.value) return
  ElMessage.success(`凭证 ${currentVoucher.value.voucher_no} 已审核通过`)
  pendingVouchers.value = pendingVouchers.value.filter(v => v.id !== currentVoucher.value!.id)
  pendingCount.value = pendingVouchers.value.length
  showDetail.value = false
}

function openRejectFromDrawer() {
  showDetail.value = false
  showRejectDialog.value = true
  selectedIds.value = currentVoucher.value ? [currentVoucher.value.id] : []
}

function batchApprove() {
  ElMessage.success(`已审核通过 ${selectedIds.value.length} 张凭证`)
  pendingVouchers.value = pendingVouchers.value.filter(v => !selectedIds.value.includes(v.id))
  pendingCount.value = pendingVouchers.value.length
  selectedIds.value = []
}

function reject() {
  if (!rejectReason.value) { ElMessage.warning('请填写驳回原因'); return }
  ElMessage.success(`已驳回 ${selectedIds.value.length} 张凭证`)
  showRejectDialog.value = false
  rejectReason.value = ''
}
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.batch-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.selected-count { font-size: 13px; color: #666; }
.section-title { font-size: 14px; font-weight: 600; margin: 16px 0 8px; }
.risk-detail-card { background: #fffbe6; border: 1px solid #ffe58f; }
.risk-item { margin-bottom: 12px; padding: 8px; background: #fff; border-radius: 4px; }
.risk-msg { margin-left: 4px; font-size: 13px; }
.risk-suggestion { color: #999; font-size: 12px; margin-top: 4px; padding-left: 4px; }
.source-hint { color: #999; font-size: 13px; text-align: center; margin-top: 8px; }
.drawer-actions { display: flex; justify-content: flex-end; gap: 12px; }
</style>
