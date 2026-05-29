<template>
  <div class="rule-library">
    <div class="page-header">
      <h3>智能分类规则库</h3>
      <el-button type="primary" @click="openCreate">+ 新建规则</el-button>
    </div>

    <el-card>
      <el-table :data="rules" border stripe size="small" row-key="name" @row-drop="handleDrop">
        <el-table-column label="优先级" width="70">
          <template #default="{ $index }">
            <el-tag size="small">{{ $index + 1 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="规则名称" min-width="140" />
        <el-table-column label="匹配字段" width="90">
          <template #default="{ row }">{{ row.matchField === 'description' ? '摘要' : '对方户名' }}</template>
        </el-table-column>
        <el-table-column prop="pattern" label="匹配模式" min-width="200">
          <template #default="{ row }">
            <code class="pattern-code">{{ row.pattern }}</code>
          </template>
        </el-table-column>
        <el-table-column label="测试" width="80">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="testPattern(row)">测试</el-button>
          </template>
        </el-table-column>
        <el-table-column label="目标分类" width="120">
          <template #default="{ row }">
            <el-tag :type="classificationTagType(row.classification)" size="small">{{ classificationLabel(row.classification) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="direction" label="方向" width="60">
          <template #default="{ row }">{{ row.direction === 'in' ? '收款' : row.direction === 'out' ? '付款' : '不限' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }"><el-switch v-model="row.is_active" size="small" /></template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row, $index }">
            <el-button link type="primary" size="small" @click="editRule(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteRule($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 正则测试弹窗 -->
    <el-dialog v-model="showTestDialog" :title="`测试规则: ${testRule?.name}`" width="480px">
      <p class="test-label">输入测试文本：</p>
      <el-input v-model="testInput" type="textarea" :rows="3" placeholder="输入摘要或对方户名进行匹配测试" />
      <div v-if="testResult !== null" class="test-result">
        <el-tag :type="testResult ? 'success' : 'danger'" size="large">
          {{ testResult ? '✅ 匹配成功' : '❌ 未匹配' }}
        </el-tag>
      </div>
      <template #footer>
        <el-button @click="showTestDialog = false">关闭</el-button>
        <el-button type="primary" @click="runTest">执行测试</el-button>
      </template>
    </el-dialog>

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="showDialog" :title="editingIndex >= 0 ? '编辑规则' : '新建规则'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="规则名称" required>
          <el-input v-model="form.name" placeholder="如：手续费匹配" />
        </el-form-item>
        <el-form-item label="匹配字段">
          <el-select v-model="form.matchField" style="width: 100%">
            <el-option label="摘要" value="description" />
            <el-option label="对方户名" value="counterparty" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配模式" required>
          <el-input v-model="form.pattern" placeholder="如：手续费|管理费|年费">
            <template #append>
              <el-select v-model="form.matchType" style="width: 100px">
                <el-option label="正则" value="regex" />
                <el-option label="关键词" value="keyword" />
              </el-select>
            </template>
          </el-input>
          <p class="form-hint">正则：使用 <code>|</code> 分隔多个模式；关键词：输入完整关键词</p>
        </el-form-item>
        <el-form-item label="金额方向">
          <el-select v-model="form.direction" style="width: 100%">
            <el-option label="不限" value="" />
            <el-option label="收款(in)" value="in" />
            <el-option label="付款(out)" value="out" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标分类" required>
          <el-select v-model="form.classification" style="width: 100%">
            <el-option label="业务收款" value="business_receipt" />
            <el-option label="业务付款" value="business_payment" />
            <el-option label="银行费用" value="bank_fee" />
            <el-option label="利息收入" value="interest_income" />
            <el-option label="内部转账" value="internal_transfer" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="1" :max="100" />
          <p class="form-hint">数字越小越优先匹配；匹配成功后不再检测后续规则</p>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

interface Rule {
  name: string
  pattern: string
  matchType: string
  matchField: string
  direction: string
  classification: string
  priority: number
  is_active: boolean
}

const defaultForm = {
  name: '', pattern: '', matchType: 'regex', matchField: 'description',
  direction: '', classification: 'business_receipt', priority: 10,
}

const rules = ref<Rule[]>([
  { name: '手续费匹配', pattern: '手续费|管理费|年费|账户管理', matchType: 'regex', matchField: 'description', direction: 'out', classification: 'bank_fee', priority: 1, is_active: true },
  { name: '利息收入', pattern: '利息|结息', matchType: 'keyword', matchField: 'description', direction: 'in', classification: 'interest_income', priority: 2, is_active: true },
  { name: '内部转账', pattern: '同行.*划转|内部.*调拨', matchType: 'regex', matchField: 'description', direction: '', classification: 'internal_transfer', priority: 3, is_active: true },
  { name: '上海XX回款', pattern: '上海XX贸易公司', matchType: 'keyword', matchField: 'counterparty', direction: 'in', classification: 'business_receipt', priority: 4, is_active: true },
])

const showDialog = ref(false)
const editingIndex = ref(-1)
const form = reactive({ ...defaultForm })

// 正则测试
const showTestDialog = ref(false)
const testRule = ref<Rule | null>(null)
const testInput = ref('')
const testResult = ref<boolean | null>(null)

function classificationTagType(val: string) {
  const map: Record<string, string> = { business_receipt: 'success', business_payment: 'danger', bank_fee: 'warning', interest_income: 'primary', internal_transfer: 'info' }
  return map[val] || ''
}
function classificationLabel(val: string) {
  const map: Record<string, string> = { business_receipt: '业务收款', business_payment: '业务付款', bank_fee: '银行费用', interest_income: '利息收入', internal_transfer: '内部转账' }
  return map[val] || val
}

function openCreate() {
  editingIndex.value = -1
  Object.assign(form, defaultForm)
  showDialog.value = true
}

function editRule(rule: Rule) {
  editingIndex.value = rules.value.indexOf(rule)
  form.name = rule.name
  form.pattern = rule.pattern
  form.matchType = rule.matchType
  form.matchField = rule.matchField
  form.direction = rule.direction
  form.classification = rule.classification
  form.priority = rule.priority
  showDialog.value = true
}

function saveRule() {
  if (editingIndex.value >= 0) {
    const idx = editingIndex.value
    rules.value[idx] = { ...form, is_active: rules.value[idx].is_active }
    ElMessage.success('规则已更新')
  } else {
    rules.value.push({ ...form, is_active: true } as Rule)
    ElMessage.success('规则已创建')
  }
  showDialog.value = false
}

function deleteRule(index: number) {
  rules.value.splice(index, 1)
  ElMessage.success('规则已删除')
}

function testPattern(rule: Rule) {
  testRule.value = rule
  testInput.value = ''
  testResult.value = null
  showTestDialog.value = true
}

function runTest() {
  if (!testRule.value || !testInput.value) {
    testResult.value = false
    return
  }
  try {
    const text = testInput.value
    const pattern = testRule.value.pattern
    const isRegex = testRule.value.matchType === 'regex'
    if (isRegex) {
      testResult.value = new RegExp(pattern, 'i').test(text)
    } else {
      testResult.value = text.toLowerCase().includes(pattern.toLowerCase())
    }
  } catch {
    testResult.value = false
    ElMessage.warning('正则表达式语法错误')
  }
}

function handleDrop() {
  ElMessage.success('优先级已更新')
}
</script>

<style scoped lang="scss">
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.pattern-code { background: #f5f5f5; padding: 2px 6px; border-radius: 3px; font-size: 12px; color: #cf1322; }
.form-hint { font-size: 12px; color: #999; margin-top: 4px; code { background: #f5f5f5; padding: 1px 4px; border-radius: 2px; } }
.test-label { font-size: 13px; color: #666; margin-bottom: 8px; }
.test-result { margin-top: 12px; text-align: center; }
</style>
