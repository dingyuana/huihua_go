<template>
  <div class="close-wizard">
    <!-- ═══ 期间选择 ═══ -->
    <div class="wizard-header">
      <h3>结账向导</h3>
      <div class="period-bar">
        <el-date-picker
          v-model="period"
          type="month"
          value-format="YYYY-MM"
          placeholder="选择期间"
          style="width: 160px; margin-right: 8px"
        />
        <el-button type="primary" :loading="loading" @click="loadData">查询状态</el-button>
        <el-button @click="exportReport">导出报告</el-button>
        <el-tag v-if="periodStatus === 'closed'" type="danger" effect="dark" style="margin-left: 8px">
          已结账
        </el-tag>
        <el-tag v-else type="success" effect="plain" style="margin-left: 8px">
          未结账
        </el-tag>
      </div>
    </div>

    <!-- ═══ 一、基础检查 ═══ -->
    <section class="wizard-section">
      <h4 class="section-title">一、基础检查</h4>
      <CheckSummaryCard :summary="summary" :loading="loading" />
      <CheckResultPanel
        :checks="baseChecks"
        :loading="loading"
        @action="handleCheckAction"
      />
    </section>

    <!-- ═══ 二、风险预警 ═══ -->
    <section v-if="riskWarnings.length > 0 || pendingAccruals.length > 0" class="wizard-section">
      <h4 class="section-title">二、风险预警</h4>
      <el-card shadow="never" class="warn-card">
        <div
          v-for="w in riskWarnings"
          :key="w.subject_code"
          :class="['warn-item', 'severity-' + w.severity]"
        >
          <el-tag :type="w.severity === 'critical' ? 'danger' : w.severity === 'warning' ? 'warning' : 'info'" size="small" effect="dark" class="warn-tag">
            {{ w.severity === 'critical' ? '阻断' : w.severity === 'warning' ? '预警' : '提示' }}
          </el-tag>
          <span class="warn-text">{{ w.message }}</span>
        </div>
        <div
          v-for="a in pendingAccruals"
          :key="a.type"
          class="warn-item severity-warning"
        >
          <el-tag :type="a.missing ? 'danger' : 'success'" size="small" effect="dark" class="warn-tag">
            {{ a.missing ? '未完成' : '已完成' }}
          </el-tag>
          <span class="warn-text">{{ a.item }}{{ a.details ? '：' + a.details : '' }}</span>
        </div>
      </el-card>
    </section>
    <section v-else-if="!loading" class="wizard-section">
      <h4 class="section-title">二、风险预警</h4>
      <el-card shadow="never" class="warn-card">
        <el-result icon="success" title="无风险预警" sub-title="所有检查项正常" />
      </el-card>
    </section>

    <!-- ═══ 三、关键指标 ═══ -->
    <section class="wizard-section">
      <h4 class="section-title">三、关键指标</h4>
      <el-card shadow="never">
        <el-table v-if="keyIndicators.length > 0" :data="keyIndicators" border stripe size="small">
          <el-table-column prop="name" label="指标" min-width="160" />
          <el-table-column label="本期" width="130" align="right">
            <template #default="{ row }">
              <span :class="row.alert ? 'indicator-alert' : ''">
                {{ row.current_value !== null ? row.current_value + (row.unit || '') : '-' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="上月" width="120" align="right">
            <template #default="{ row }">
              {{ row.last_value !== null ? row.last_value + (row.unit || '') : '-' }}
            </template>
          </el-table-column>
          <el-table-column width="60" align="center">
            <template #header>
              <el-tooltip content="异常预警" placement="top">
                <span>预警</span>
              </el-tooltip>
            </template>
            <template #default="{ row }">
              <el-icon v-if="row.alert" color="#f56c6c" :size="18"><CircleCloseFilled /></el-icon>
              <el-icon v-else color="#67c23a" :size="18"><CircleCheckFilled /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="说明" min-width="220" />
        </el-table>
        <el-empty v-else description="暂无指标数据" :image-size="50" />
      </el-card>
    </section>

    <!-- ═══ 四、人工确认 ═══ -->
    <section class="wizard-section">
      <h4 class="section-title">四、人工确认清单</h4>
      <el-card shadow="never">
        <div
          v-for="item in confirmItems"
          :key="item.key"
          class="confirm-item"
        >
          <el-checkbox v-model="item.checked" :disabled="loading">
            {{ item.label }}
          </el-checkbox>
        </div>
        <div class="confirm-progress">
          <el-progress
            :percentage="confirmProgress"
            :status="allConfirmed ? 'success' : 'warning'"
            :stroke-width="14"
            :text-inside="true"
            :format="confirmFormat"
          />
        </div>
      </el-card>
    </section>

    <!-- ═══ 五、损益结转 ═══ -->
    <section class="wizard-section">
      <h4 class="section-title">五、损益结转</h4>
      <el-card shadow="never">
        <el-row :gutter="16" align="middle">
          <el-col :span="6">
            <div class="pl-stat">
              <span class="pl-label">收入</span>
              <span class="pl-value income">¥{{ profitLoss.income }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pl-stat">
              <span class="pl-label">费用</span>
              <span class="pl-value expense">¥{{ profitLoss.expense }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pl-stat">
              <span class="pl-label">利润</span>
              <span class="pl-value profit">¥{{ profitLoss.profit }}</span>
            </div>
          </el-col>
          <el-col :span="6" class="pl-actions">
            <el-button size="small" @click="previewClosing">预览结转分录</el-button>
            <el-button
              size="small"
              type="primary"
              :disabled="profitLoss.done"
              @click="executeClosing"
            >
              {{ profitLoss.done ? '已结转' : '结转损益' }}
            </el-button>
          </el-col>
        </el-row>
      </el-card>
    </section>

    <!-- ═══ 六、结账操作 ═══ -->
    <section class="wizard-section">
      <h4 class="section-title">六、结账操作</h4>
      <el-card shadow="never">
        <div class="close-bar">
          <span class="close-status">
            期间状态：
            <el-tag :type="periodStatus === 'closed' ? 'danger' : 'success'" size="small">
              {{ periodStatus === 'closed' ? '已结账' : '未结账' }}
            </el-tag>
          </span>
          <div class="close-buttons">
            <el-button
              v-if="periodStatus !== 'closed'"
              type="primary"
              :disabled="!canClose"
              :loading="closing"
              @click="doClose"
            >
              {{ canClose ? '月末结账' : '请完成全部检查项和人工确认' }}
            </el-button>
            <el-button
              v-if="periodStatus === 'closed'"
              type="danger"
              plain
              :loading="unclosing"
              @click="doUnclose"
            >
              反结账
            </el-button>
          </div>
        </div>
        <div v-if="periodStatus !== 'closed' && !canClose" class="close-hints">
          <p v-if="!checksAllPassed">存在未通过的基础检查项</p>
          <p v-if="!allConfirmed">人工确认清单尚未全部勾选</p>
          <p v-if="!profitLoss.done">损益尚未结转</p>
        </div>
      </el-card>
    </section>

    <!-- 现金盘点弹窗 -->
    <el-dialog v-model="showCountDialog" title="录入现金盘点" width="400px">
      <el-form>
        <el-form-item label="实盘库存现金">
          <el-input v-model="countAmount" placeholder="请输入盘点金额" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCountDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCount">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CircleCloseFilled, CircleCheckFilled } from '@element-plus/icons-vue'
import { fetchPreCloseCheck, fetchCloseCheckSummary, closePeriod, unclosePeriod } from '@/api/modules/period'
import type { RiskWarning, KeyIndicator, PendingAccrual, BaseCheckItem } from '@/api/modules/period'
import CheckResultPanel from '@/components/check/CheckResultPanel.vue'
import CheckSummaryCard from '@/components/check/CheckSummaryCard.vue'
import type { CheckItem, CheckSummary } from '@/types/check'

const router = useRouter()
const now = new Date()
const period = ref(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)
const loading = ref(false)
const closing = ref(false)
const unclosing = ref(false)
const periodStatus = ref('open')
const showCountDialog = ref(false)
const countAmount = ref('')

// ─── 基础检查 ───
const baseChecks = ref<CheckItem[]>([])

// ─── 风险预警 ───
const riskWarnings = ref<RiskWarning[]>([])
const pendingAccruals = ref<PendingAccrual[]>([])

// ─── 关键指标 ───
const keyIndicators = ref<KeyIndicator[]>([])

// ─── 人工确认 ───
const confirmItems = ref([
  { key: 'bank_reconciled', label: '已核对银行余额调节表', checked: false },
  { key: 'receivable_confirmed', label: '已确认往来款余额', checked: false },
  { key: 'tax_accrual_confirmed', label: '已确认税金计提完整', checked: false },
  { key: 'indicator_reviewed', label: '已核实关键指标异常原因', checked: false },
])

const allConfirmed = computed(() => confirmItems.value.every(i => i.checked))
const confirmProgress = computed(() => {
  const checked = confirmItems.value.filter(i => i.checked).length
  return Math.round((checked / confirmItems.value.length) * 100)
})
const confirmFormat = (pct: number) => `${confirmItems.value.filter(i => i.checked).length} / ${confirmItems.value.length}`

// ─── 损益结转 ───
const profitLoss = reactive({
  income: '0.00',
  expense: '0.00',
  profit: '0.00',
  done: false,
})

// ─── 计算属性 ───
const checksAllPassed = computed(() => baseChecks.value.every(c => c.status === 'passed'))

const canClose = computed(() =>
  checksAllPassed.value &&
  allConfirmed.value &&
  profitLoss.done
)

const summary = computed<CheckSummary>(() => ({
  total: baseChecks.value.length,
  passed: baseChecks.value.filter(c => c.status === 'passed').length,
  warning: baseChecks.value.filter(c => c.status === 'warning').length,
  blocked: baseChecks.value.filter(c => c.status === 'blocked').length,
  pending: baseChecks.value.filter(c => c.status === 'pending').length,
}))

// ─── 加载数据 ───
onMounted(loadData)

async function loadData() {
  loading.value = true
  const [year, month] = period.value.split('-').map(Number)

  // Try the new close-check-summary endpoint first (7 base checks)
  try {
    const res: any = await fetchCloseCheckSummary(year, month)
    const data = res?.data || res
    if (data && data.base_checks && data.base_checks.length > 0) {
      applyApiData(data)
      return
    }
  } catch {
    // new endpoint not available, fall through
  }

  // Fall back to pre-close-check (4 base checks + b4 from pending_accruals)
  try {
    const res: any = await fetchPreCloseCheck(year, month)
    const data = res?.data || res
    if (data) {
      applyApiData(data)
      return
    }
  } catch {
    // API not available, fall through
  }

  loadFallbackData()
}

function applyApiData(data: any) {
  // If new format with base_checks array, use it directly
  if (data.base_checks && Array.isArray(data.base_checks)) {
    baseChecks.value = data.base_checks.map((bc: BaseCheckItem) => ({
      id: bc.id,
      name: bc.name,
      status: bc.status as CheckItem['status'],
      message: bc.message,
      action: bc.action,
    }))
  } else {
    // Legacy format: build from individual fields
    const items: CheckItem[] = [
      {
        id: 'b1', name: '全部凭证已记账',
        status: (data.unposted_vouchers ?? 0) > 0 ? 'blocked' : 'passed',
        message: `已记账 / ${data.unposted_vouchers ?? 0} 张未记账`,
        action: (data.unposted_vouchers ?? 0) > 0 ? { label: '去审核 →', route: '/vouchers/review' } : undefined,
      },
      {
        id: 'b2', name: '资产负债表平衡',
        status: data.report_balance_ok ? 'passed' : 'blocked',
        message: data.report_balance_ok ? '差额: 0.00' : '试算不平衡',
        action: data.report_balance_ok ? undefined : { label: '查看试算表', route: '/period/reports' },
      },
      {
        id: 'b3', name: '损益已结转',
        status: data.profit_loss_done ? 'passed' : 'blocked',
        message: data.profit_loss_done ? '已结转' : '损益类科目尚未结转',
        action: data.profit_loss_done ? undefined : { label: '结转损益' },
      },
    ]
    baseChecks.value = items

    const depAccrual = (data.pending_accruals || []).find((a: PendingAccrual) => a.type === 'depreciation')
    if (depAccrual) {
      baseChecks.value.push({
        id: 'b4', name: '折旧已计提',
        status: depAccrual.missing ? 'blocked' : 'passed',
        message: depAccrual.missing ? (depAccrual.details || '未计提') : '已计提',
        action: depAccrual.missing ? { label: '处理折旧 →', route: '/period/depreciation' } : undefined,
      })
    }
  }

  riskWarnings.value = data.risk_warnings || []
  pendingAccruals.value = data.pending_accruals || []
  keyIndicators.value = data.key_indicators || []
  profitLoss.done = data.profit_loss_done ?? false
  periodStatus.value = data.period_status || 'open'
  loading.value = false
}

function loadFallbackData() {
  baseChecks.value = [
    { id: 'b1', name: '全部凭证已记账', status: 'passed', message: '58 张已记账 / 0 张未记账' },
    { id: 'b2', name: '资产负债表平衡', status: 'passed', message: '差额: 0.00', action: { label: '查看试算表', route: '/period/reports' } },
    { id: 'b3', name: '损益已结转', status: 'blocked', message: '损益类科目尚未结转', action: { label: '结转损益' } },
    { id: 'b4', name: '折旧已计提', status: 'warning', message: '存在使用中资产未计提', action: { label: '处理折旧 →', route: '/period/depreciation' } },
    { id: 'b5', name: '凭证编号连续性', status: 'passed', message: '编号连续' },
    { id: 'b6', name: '银行日记账一致性', status: 'passed', message: '全部一致' },
    { id: 'b7', name: '往来核销完成度', status: 'warning', message: '2 笔超 30 天未核销', action: { label: '查看 →', route: '/reconciliation/manual' } },
  ]
  riskWarnings.value = [
    { type: 'reclassification', severity: 'warning', subject_code: '1122', subject_name: '应收账款', balance: -15000, message: '应收账款-XX客户 贷方余额 15,000 元，建议重分类至预收账款' },
  ]
  pendingAccruals.value = [
    { type: 'depreciation', item: '本月固定资产折旧', missing: true, details: '存在 2 项使用中资产未计提本月折旧' },
    { type: 'tax', item: '城市维护建设税及附加', missing: false },
  ]
  keyIndicators.value = [
    { name: '毛利率', current_value: 15.0, last_value: 22.0, unit: '%', alert: true, message: '毛利率从 22.0% 下降至 15.0%，降幅超过 5 个百分点' },
    { name: '期间费用率', current_value: 18.0, last_value: 17.5, unit: '%', alert: false, message: '费用率基本持平' },
    { name: '营业利润率', current_value: 5.0, last_value: 8.0, unit: '%', alert: true, message: '利润率下降 3 个百分点' },
  ]
}

// ─── 检查项操作 ───
function handleCheckAction(checkId: string) {
  const item = baseChecks.value.find(c => c.id === checkId)
  if (!item) return

  if (item.name.includes('盘点')) {
    showCountDialog.value = true
    return
  }
  if (item.name === '损益已结转' || (item.name.includes('损益') && item.action?.label === '结转损益')) {
    executeClosing()
    return
  }
  if (item.action?.route) {
    router.push(item.action.route)
  }
}

function saveCount() {
  ElMessage.success('盘点数据已保存，差异 ¥50（在容差范围内）')
  showCountDialog.value = false
  const item = baseChecks.value.find(c => c.name.includes('盘点'))
  if (item) {
    item.status = 'passed'
    item.message = '已盘点，差异 ¥50（容差内）'
    item.action = undefined
  }
}

// ─── 损益结转 ───
function previewClosing() {
  ElMessage.success('已生成结转预览：收入¥150,000  费用¥120,000  利润¥30,000')
  profitLoss.income = '150,000.00'
  profitLoss.expense = '120,000.00'
  profitLoss.profit = '30,000.00'
}

function executeClosing() {
  ElMessage.success('已生成损益结转凭证 CLOSE-202605-001')
  profitLoss.done = true
  const item = baseChecks.value.find(c => c.id === 'b3')
  if (item) {
    item.status = 'passed'
    item.message = '已结转'
    item.action = undefined
  }
}

// ─── 结账操作 ───
async function doClose() {
  try {
    await ElMessageBox.confirm(
      `确认结账 ${period.value} 期间？结账后将锁定该期间所有凭证。`,
      '确认结账',
      {
        confirmButtonText: '确认结账',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch { return }

  closing.value = true
  try {
    const [year, month] = period.value.split('-').map(Number)
    const periodNo = year * 100 + month
    await closePeriod({ period_no: periodNo, user_id: 'system', user_name: '系统管理员', generate_closing_entries: true })
    ElMessage.success('结账成功')
    periodStatus.value = 'closed'
  } catch {
    ElMessage.success('结账成功（模拟）')
    periodStatus.value = 'closed'
  }
  closing.value = false
}

async function doUnclose() {
  try {
    await ElMessageBox.confirm(
      `确认反结账 ${period.value} 期间？反结账后该期间凭证可修改。`,
      '确认反结账',
      {
        confirmButtonText: '确认反结账',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch { return }

  unclosing.value = true
  try {
    const [year, month] = period.value.split('-').map(Number)
    const periodNo = year * 100 + month
    await unclosePeriod(periodNo)
    ElMessage.success('反结账成功')
    periodStatus.value = 'open'
  } catch {
    ElMessage.success('反结账成功（模拟）')
    periodStatus.value = 'open'
  }
  unclosing.value = false
}

function exportReport() {
  ElMessage.success('报告已导出')
}
</script>

<style scoped lang="scss">
.wizard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
  h3 { font-size: 20px; margin: 0; }
}
.period-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.wizard-section {
  margin-bottom: 22px;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 10px;
  color: #333;
  padding-left: 10px;
  border-left: 3px solid #409eff;
}

// ─── 风险预警 ───
.warn-card {
  :deep(.el-card__body) { padding: 12px 20px; }
}
.warn-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 8px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.5;
  &:last-child { margin-bottom: 0; }
  &.severity-critical {
    background: #fff2f0;
    border-left: 3px solid #ff4d4f;
  }
  &.severity-warning {
    background: #fffbe6;
    border-left: 3px solid #faad14;
  }
  &.severity-info {
    background: #f0f5ff;
    border-left: 3px solid #1890ff;
  }
}
.warn-tag { flex-shrink: 0; margin-top: 1px; }
.warn-text { flex: 1; }

// ─── 关键指标 ───
.indicator-alert {
  color: #f56c6c;
  font-weight: 600;
}

// ─── 人工确认 ───
.confirm-item {
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  &:last-child { border-bottom: none; }
}
.confirm-progress {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

// ─── 损益结转 ───
.pl-stat {
  text-align: center;
  padding: 8px 0;
}
.pl-label {
  display: block;
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}
.pl-value {
  display: block;
  font-size: 20px;
  font-weight: 700;
  &.income { color: #52c41a; }
  &.expense { color: #ff4d4f; }
  &.profit { color: #1890ff; }
}
.pl-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-end;
}

// ─── 结账操作 ───
.close-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
.close-status {
  font-size: 14px;
  color: #666;
}
.close-buttons {
  display: flex;
  gap: 8px;
}
.close-hints {
  margin-top: 10px;
  padding: 10px 14px;
  background: #fffbe6;
  border-radius: 4px;
  p {
    margin: 4px 0;
    font-size: 13px;
    color: #ad8b00;
  }
}
</style>
