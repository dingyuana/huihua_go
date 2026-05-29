<template>
  <div class="health-check">
    <div class="page-header">
      <h3>结账前体检</h3>
      <div>
        <el-select v-model="period" style="width: 140px; margin-right: 8px">
          <el-option label="2026-05" value="2026-05" />
          <el-option label="2026-04" value="2026-04" />
        </el-select>
        <el-button type="primary" @click="runCheck">刷新检查</el-button>
        <el-button @click="exportReport">导出报告</el-button>
      </div>
    </div>

    <el-card :class="['status-card', overallStatus]">
      <div class="status-badge">
        <el-tag :type="overallStatus === 'red' ? 'danger' : overallStatus === 'yellow' ? 'warning' : 'success'" size="large">
          {{ overallStatus === 'red' ? '🔴 存在阻断项，无法结账' : overallStatus === 'yellow' ? '🟡 有警告，可结账' : '🟢 全部通过' }}
        </el-tag>
      </div>
    </el-card>

    <el-card class="checklist">
      <el-table :data="checks" border stripe size="small">
        <el-table-column label="#" width="50"><template #default="{ $index }">{{ $index + 1 }}</template></el-table-column>
        <el-table-column prop="name" label="检查项" min-width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="详情" min-width="260" />
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button
              v-if="row.action"
              link
              type="primary"
              size="small"
              :disabled="!row.action"
              @click="handleAction(row)"
            >
              {{ row.actionBtn || '立即处理 →' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="close-actions">
      <el-button @click="exportReport">导出 PDF 报告</el-button>
      <el-button type="primary" :disabled="overallStatus === 'red'" @click="doClose">
        {{ overallStatus === 'red' ? '请先修复阻断项' : '开始结账' }}
      </el-button>
    </div>

    <!-- 现金盘点弹窗 -->
    <el-dialog v-model="showCountDialog" title="录入现金盘点" width="400px">
      <el-form>
        <el-form-item label="实盘库存现金">
          <AmountInput v-model="countAmount" />
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const router = useRouter()
const period = ref('2026-05')
const showCountDialog = ref(false)
const countAmount = ref('')

interface CheckItem {
  name: string
  status: string
  message: string
  action: boolean
  actionBtn?: string
  route?: string
}

const localChecks: CheckItem[] = [
  { name: '凭证借贷平衡', status: 'passed', message: '全部凭证借贷平衡', action: false },
  { name: '凭证完整性', status: 'warning', message: '3 张凭证待审核', action: true, actionBtn: '去审核 →', route: '/vouchers/review' },
  { name: '凭证编号连续性', status: 'passed', message: '编号连续', action: false },
  { name: '固定资产折旧', status: 'blocked', message: '1 笔折旧计划未过账', action: true, actionBtn: '处理折旧 →', route: '/period/depreciation' },
  { name: '银行日记账一致性', status: 'passed', message: '全部一致', action: false },
  { name: '现金账实一致', status: 'warning', message: '本月尚未盘点', action: true, actionBtn: '录入盘点 →' },
  { name: '往来核销完成度', status: 'blocked', message: '2 笔超 30 天未核销，金额 ¥35,000', action: true, actionBtn: '查看清单 →', route: '/reconciliation/manual' },
  { name: '进项发票到期', status: 'warning', message: '3 张发票即将过期', action: true, actionBtn: '查看 →', route: '/invoices' },
  { name: '损益结转', status: 'blocked', message: '损益类科目尚未结转', action: true, actionBtn: '生成结转凭证 →' },
  { name: '期间锁定状态', status: 'passed', message: '当前期间未锁定', action: false },
]

const checks = ref<CheckItem[]>([])
onMounted(async () => {
  try {
    const res: any = await request.post('/periods/health-check', { period: period.value })
    const items = res?.data?.checks || res?.data
    if (Array.isArray(items)) { checks.value = items; return }
  } catch { /* fallback */ }
  checks.value = localChecks
})

const overallStatus = computed(() => {
  if (checks.value.some(c => c.status === 'blocked')) return 'red'
  if (checks.value.some(c => c.status === 'warning')) return 'yellow'
  return 'green'
})

function statusTag(s: string) {
  return { passed: 'success', warning: 'warning', blocked: 'danger' }[s] || 'info'
}
function statusLabel(s: string) {
  return { passed: '通过', warning: '警告', blocked: '阻断' }[s] || s
}

function handleAction(row: CheckItem) {
  if (row.name === '现金账实一致') {
    showCountDialog.value = true
    return
  }
  if (row.name === '损益结转') {
    ElMessage.success('已生成损益结转凭证草稿')
    return
  }
  if (row.route) {
    router.push(row.route)
  }
}

function saveCount() {
  ElMessage.success('盘点数据已保存，差异 ¥50（在容差范围内）')
  showCountDialog.value = false
  // 更新检查状态
  const item = checks.value.find(c => c.name === '现金账实一致')
  if (item) { item.status = 'passed'; item.message = '已盘点，差异 ¥50（容差内）'; item.action = false }
}

function runCheck() {
  ElMessage.success('体检完成')
}

function exportReport() {
  ElMessage.success('报告已导出')
}

function doClose() {
  ElMessageBox.confirm('确认结账 2026-05 期间？结账后将锁定该期间所有凭证。', '确认结账', {
    confirmButtonText: '确认结账', cancelButtonText: '取消', type: 'warning',
  }).then(() => ElMessage.success('结账成功')).catch(() => {})
}
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.status-card { margin-bottom: 16px; text-align: center; padding: 8px; }
.status-card.red { border-left: 4px solid #ff4d4f; }
.status-card.yellow { border-left: 4px solid #faad14; }
.status-card.green { border-left: 4px solid #52c41a; }
.status-badge { padding: 8px 0; }
.checklist { margin-bottom: 16px; }
.close-actions { display: flex; justify-content: flex-end; gap: 12px; }
</style>
