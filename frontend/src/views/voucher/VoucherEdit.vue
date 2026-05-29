<template>
  <div class="voucher-edit">
    <div class="page-header">
      <h3>{{ isNew ? '新增凭证' : (readonly ? '查看凭证' : '编辑凭证') }}</h3>
      <DocStatusTag v-if="!isNew" :docstatus="docstatus" size="default" />
    </div>

    <el-card>
      <!-- 头部信息 -->
      <el-form :inline="true" size="small">
        <el-form-item label="日期">
          <el-date-picker v-model="form.date" type="date" style="width: 140px" :disabled="readonly" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 80px" :disabled="readonly">
            <el-option v-for="t in ['记','银','现','转']" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="编号">
          <el-input :model-value="voucherNo" disabled style="width: 160px" />
        </el-form-item>
        <el-form-item v-if="!isNew && reversalId" label="关联冲销">
          <el-tag type="warning" size="small">冲销凭证: {{ reversalId }}</el-tag>
        </el-form-item>
        <el-form-item label="摘要" style="flex:1">
          <el-input v-model="form.remark" :disabled="readonly" placeholder="请输入摘要" />
        </el-form-item>
      </el-form>

      <!-- 分录行表格 -->
      <el-table :data="lines" size="small" border>
        <el-table-column label="科目" min-width="240">
          <template #default="{ $index }">
            <AccountSelector v-model="lines[$index].account" :disabled="readonly" />
          </template>
        </el-table-column>
        <el-table-column label="借方金额" width="180">
          <template #default="{ $index }">
            <AmountInput v-model="lines[$index].debit" :disabled="readonly" :max="''" @update:model-value="onLineChange($index, 'debit')" />
          </template>
        </el-table-column>
        <el-table-column label="贷方金额" width="180">
          <template #default="{ $index }">
            <AmountInput v-model="lines[$index].credit" :disabled="readonly" :max="''" @update:model-value="onLineChange($index, 'credit')" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60">
          <template #default="{ $index }">
            <el-button v-if="!readonly" link type="danger" size="small" @click="lines.splice($index, 1)">×</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!readonly" class="add-line">
        <el-button text type="primary" @click="addLine">+ 添加分录行</el-button>
      </div>

      <!-- 合计 -->
      <div class="totals">
        <span>借方合计: <b class="amount-positive">¥{{ totalDebit }}</b></span>
        <span>贷方合计: <b class="amount-negative">¥{{ totalCredit }}</b></span>
        <el-tag v-if="totalDebit === '0.00' && totalCredit === '0.00'" type="info" size="large">请添加分录</el-tag>
        <el-tag v-else-if="isBalanced" type="success" size="large">✅ 借贷平衡</el-tag>
        <el-tag v-else type="danger" size="large">❌ 借贷不平衡 (差额 ¥{{ diff }})</el-tag>
      </div>

      <!-- 操作按钮 -->
      <div class="edit-actions">
        <el-button @click="$router.back()">返回</el-button>

        <!-- 草稿状态: 可编辑 -->
        <template v-if="!readonly">
          <el-button @click="saveDraft">保存草稿</el-button>
          <el-button type="primary" :disabled="!canSubmit" @click="submit">提交审核</el-button>
        </template>

        <!-- 已审核状态: 可红字冲销 -->
        <template v-if="docstatus === 1">
          <el-button type="warning" @click="showReverseDialog = true">🔴 红字冲销</el-button>
        </template>
      </div>
    </el-card>

    <!-- 红字冲销弹窗 -->
    <el-dialog v-model="showReverseDialog" title="红字冲销" width="480px">
      <el-alert type="warning" :closable="false" show-icon>
        <p>冲销后将生成一张与原凭证借贷方向完全相反的新凭证。</p>
        <p>原凭证标记为「已作废」，两张凭证均保留可查。</p>
      </el-alert>
      <el-form class="reverse-form">
        <el-form-item label="冲销日期" required>
          <el-date-picker v-model="reverseDate" type="date" style="width: 100%" />
        </el-form-item>
        <el-form-item label="冲销原因" required>
          <el-input v-model="reverseReason" type="textarea" :rows="3" placeholder="请详细说明冲销原因" />
        </el-form-item>
      </el-form>

      <!-- 冲销分录预览 -->
      <el-divider>冲销分录预览</el-divider>
      <el-table :data="reverseLines" size="small" border>
        <el-table-column prop="account_name" label="科目" min-width="160" />
        <el-table-column label="原借方" width="100" align="right">
          <template #default="{ row }">{{ row.original_debit }}</template>
        </el-table-column>
        <el-table-column label="原贷方" width="100" align="right">
          <template #default="{ row }">{{ row.original_credit }}</template>
        </el-table-column>
        <el-table-column label="冲销借方" width="100" align="right">
          <template #default="{ row }">{{ row.reverse_debit }}</template>
        </el-table-column>
        <el-table-column label="冲销贷方" width="100" align="right">
          <template #default="{ row }">{{ row.reverse_credit }}</template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="showReverseDialog = false">取消</el-button>
        <el-button type="warning" :disabled="!reverseReason || !reverseDate" @click="executeReverse">确认冲销</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const route = useRoute()
const router = useRouter()
const isNew = ref(!route.params.id)

interface LineItem {
  account: any
  debit: string
  credit: string
  account_name?: string
  original_debit?: string
  original_credit?: string
  reverse_debit?: string
  reverse_credit?: string
}

const docstatus = ref(0)
const readonly = computed(() => docstatus.value === 1)
const reversalId = ref('')

const form = reactive({ date: '', type: '记', remark: '' })
const lines = ref<LineItem[]>([
  { account: null, debit: '', credit: '' },
  { account: null, debit: '', credit: '' },
])

const voucherNo = computed(() => {
  if (isNew.value) return '(自动生成)'
  return route.params.id as string || '记-2026-05-XXXX'
})

/** 编辑模式：从后端加载凭证数据 */
onMounted(async () => {
  if (isNew.value) {
    form.date = new Date().toISOString().slice(0, 10)
    return
  }
  try {
    const res: any = await request.get(`/vouchers/${route.params.id}`)
    const data = res?.data || res
    if (data) {
      form.date = data.posting_date || ''
      form.type = data.voucher_type || '记'
      form.remark = data.remark || ''
      docstatus.value = data.docstatus ?? 0
      reversalId.value = data.reversal_id || ''
      if (data.lines?.length) {
        lines.value = data.lines.map((l: any) => ({
          account: { id: l.account_id, code: l.account_code, name: l.account_name },
          debit: l.debit || '',
          credit: l.credit || '',
        }))
      }
    }
  } catch {
    form.date = new Date().toISOString().slice(0, 10)
  }
})

const totalDebit = computed(() =>
  lines.value.reduce((s, l) => s + (parseFloat(l.debit) || 0), 0).toFixed(2)
)
const totalCredit = computed(() =>
  lines.value.reduce((s, l) => s + (parseFloat(l.credit) || 0), 0).toFixed(2)
)
const isBalanced = computed(() =>
  totalDebit.value === totalCredit.value && totalDebit.value !== '0.00'
)
const diff = computed(() =>
  (parseFloat(totalDebit.value) - parseFloat(totalCredit.value)).toFixed(2)
)
const canSubmit = computed(() => isBalanced.value)

// 红字冲销
const showReverseDialog = ref(false)
const reverseDate = ref('')
const reverseReason = ref('')

const reverseLines = computed<LineItem[]>(() =>
  lines.value
    .filter(l => l.account)
    .map(l => ({
      account: l.account,
      account_name: l.account?.name || '',
      debit: '',
      credit: '',
      original_debit: l.debit || '0.00',
      original_credit: l.credit || '0.00',
      reverse_debit: l.credit || '0.00',
      reverse_credit: l.debit || '0.00',
    }))
)

function addLine() {
  lines.value.push({ account: null, debit: '', credit: '' })
}

function onLineChange(index: number, field: string) {
  const line = lines.value[index]
  if (field === 'debit' && line.debit) line.credit = ''
  if (field === 'credit' && line.credit) line.debit = ''
}

function saveDraft() {
  if (!validateLines()) return
  ElMessage.success('草稿已保存')
}

function submit() {
  // 强制借贷平衡校验
  if (!isBalanced.value) {
    ElMessage.error(`借贷不平衡，无法提交审核（借方 ${totalDebit.value} ≠ 贷方 ${totalCredit.value}）`)
    return
  }
  if (lines.value.filter(l => l.account).length < 2) {
    ElMessage.warning('至少需要两条分录行')
    return
  }
  ElMessageBox.confirm('确认提交审核？提交后不可直接修改。', '确认', {
    confirmButtonText: '提交', cancelButtonText: '取消', type: 'info',
  }).then(() => {
    ElMessage.success('凭证已提交审核')
    docstatus.value = 1
  }).catch(() => {})
}

function executeReverse() {
  ElMessageBox.confirm(
    `确认对凭证 ${voucherNo.value} 执行红字冲销？\n冲销原因：${reverseReason.value}`,
    '确认冲销', { confirmButtonText: '确认冲销', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    ElMessage.success(`冲销成功！已生成冲销凭证，原凭证已作废`)
    showReverseDialog.value = false
    docstatus.value = 2
    router.push('/vouchers')
  }).catch(() => {})
}

function validateLines(): boolean {
  if (lines.value.filter(l => l.account).length === 0) {
    ElMessage.warning('请至少添加一条分录')
    return false
  }
  return true
}
</script>

<style scoped>
.page-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; h3 { font-size: 18px; } }
.add-line { margin: 12px 0; }
.totals { display: flex; align-items: center; gap: 24px; padding: 12px 0; border-top: 1px solid #e8e8e8; font-size: 15px; }
.edit-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
.reverse-form { margin-top: 16px; }
</style>
