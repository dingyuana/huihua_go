<template>
  <div class="rule-library">
    <PageLayout title="智能分类规则库" icon="⚙️" subtitle="配置银行流水自动分类规则，支持正则和关键词匹配">
      <template #actions>
        <el-button @click="seedRules">
          <el-icon><Download /></el-icon>
          初始化默认规则
        </el-button>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新建规则
        </el-button>
      </template>

      <el-table
        :data="rules"
        border
        stripe
        size="small"
        row-key="id"
        v-loading="loading"
      >
        <el-table-column label="优先级" width="80">
          <template #default="{ $index }">
            <el-tag size="small" type="info">{{ $index + 1 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column label="匹配模式" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getRuleTypeTag(row.rule_type)">
              {{ getRuleTypeLabel(row.rule_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配字段" width="100">
          <template #default="{ row }">
            {{ row.match_field === 'counterparty' ? '对方户名' : '摘要' }}
          </template>
        </el-table-column>
        <el-table-column prop="pattern" label="匹配内容" min-width="200">
          <template #default="{ row }">
            <code class="pattern-code">{{ row.pattern }}</code>
          </template>
        </el-table-column>
        <el-table-column label="金额方向" width="90">
          <template #default="{ row }">
            {{ row.direction === 'in' ? '收款' : row.direction === 'out' ? '付款' : '不限' }}
          </template>
        </el-table-column>
        <el-table-column label="目标分类" width="130">
          <template #default="{ row }">
            <el-tag :type="getClassificationTagType(row.classification)" size="small">
              {{ getClassificationLabel(row.classification) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.is_active"
              size="small"
              @change="toggleRuleStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="testPattern(row)">测试</el-button>
            <el-button link type="primary" size="small" @click="editRule(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </PageLayout>

    <el-dialog v-model="showTestDialog" :title="`测试规则: ${testRule?.name}`" width="520px">
      <p class="test-label">输入测试文本：</p>
      <el-input
        v-model="testInput"
        type="textarea"
        :rows="3"
        placeholder="输入摘要或对方户名进行匹配测试"
      />
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

    <el-dialog v-model="showDialog" :title="editingRule ? '编辑规则' : '新建规则'" width="560px">
      <el-form :model="form" label-width="110px" :rules="formRules" ref="formRef">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="如：银行手续费匹配" />
        </el-form-item>
        <el-form-item label="规则类型" prop="rule_type">
          <el-radio-group v-model="form.rule_type">
            <el-radio value="keyword">关键词</el-radio>
            <el-radio value="keyword_regex">正则表达式</el-radio>
            <el-radio value="counterparty_match">客户匹配</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="匹配字段" prop="match_field">
          <el-select v-model="form.match_field" style="width: 100%">
            <el-option label="摘要" value="description" />
            <el-option label="对方户名" value="counterparty" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配内容" prop="pattern">
          <el-input
            v-model="form.pattern"
            :placeholder="form.rule_type === 'keyword_regex' ? '如：手续费|管理费|年费' : '如：银行手续费'"
          />
          <p v-if="form.rule_type === 'keyword_regex'" class="form-hint">
            使用 | 分隔多个模式；支持正则表达式语法
          </p>
        </el-form-item>
        <el-form-item label="金额方向" prop="direction">
          <el-select v-model="form.direction" style="width: 100%">
            <el-option label="不限" value="" />
            <el-option label="收款" value="in" />
            <el-option label="付款" value="out" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标分类" prop="classification">
          <el-select v-model="form.classification" style="width: 100%">
            <el-option label="业务收款" value="business_receipt" />
            <el-option label="业务付款" value="business_payment" />
            <el-option label="银行手续费" value="bank_fee" />
            <el-option label="利息收入" value="interest_income" />
            <el-option label="内部转账" value="internal_transfer" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-input-number v-model="form.priority" :min="1" :max="999" style="width: 100%" />
          <p class="form-hint">数字越小优先级越高；匹配成功后不再检测后续规则</p>
        </el-form-item>
        <el-form-item label="启用状态" prop="is_active">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Download } from '@element-plus/icons-vue'
import PageLayout from '@/components/app/PageLayout.vue'
import {
  fetchClassificationRules,
  createClassificationRule,
  updateClassificationRule,
  deleteClassificationRule,
  reorderClassificationRules,
  seedClassificationRules,
  type ClassificationRule
} from '@/api/modules/classification-rule'

const rules = ref<ClassificationRule[]>([])
const loading = ref(false)
const saving = ref(false)

const showDialog = ref(false)
const editingRule = ref<ClassificationRule | null>(null)
const formRef = ref()
const defaultForm: Partial<ClassificationRule> = {
  name: '',
  rule_type: 'keyword',
  match_field: 'description',
  pattern: '',
  direction: '',
  classification: 'business_receipt',
  priority: 1,
  is_active: true
}
const form = reactive({ ...defaultForm })
const formRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  pattern: [{ required: true, message: '请输入匹配内容', trigger: 'blur' }],
  classification: [{ required: true, message: '请选择目标分类', trigger: 'change' }]
}

const showTestDialog = ref(false)
const testRule = ref<ClassificationRule | null>(null)
const testInput = ref('')
const testResult = ref<boolean | null>(null)

function getRuleTypeLabel(type: string) {
  const map: Record<string, string> = {
    'keyword': '关键词',
    'keyword_regex': '正则表达式',
    'counterparty_match': '客户匹配'
  }
  return map[type] || type
}

function getRuleTypeTag(type: string) {
  const map: Record<string, string> = {
    'keyword': '',
    'keyword_regex': 'primary',
    'counterparty_match': 'success'
  }
  return map[type] || ''
}

function getClassificationLabel(classification: string) {
  const map: Record<string, string> = {
    'business_receipt': '业务收款',
    'business_payment': '业务付款',
    'bank_fee': '银行手续费',
    'interest_income': '利息收入',
    'internal_transfer': '内部转账'
  }
  return map[classification] || classification
}

function getClassificationTagType(classification: string) {
  const map: Record<string, string> = {
    'business_receipt': 'success',
    'business_payment': 'danger',
    'bank_fee': 'warning',
    'interest_income': 'primary',
    'internal_transfer': 'info'
  }
  return map[classification] || ''
}

async function loadRules() {
  loading.value = true
  try {
    const res = await fetchClassificationRules()
    if (Array.isArray(res?.data)) {
      rules.value = res.data
    }
  } catch (e) {
    console.warn('加载规则失败', e)
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingRule.value = null
  Object.assign(form, defaultForm)
  showDialog.value = true
}

function editRule(rule: ClassificationRule) {
  editingRule.value = rule
  Object.assign(form, {
    name: rule.name,
    rule_type: rule.rule_type,
    match_field: rule.match_field,
    pattern: rule.pattern,
    direction: rule.direction,
    classification: rule.classification,
    priority: rule.priority,
    is_active: rule.is_active
  })
  showDialog.value = true
}

async function saveRule() {
  await formRef.value?.validate()
  saving.value = true
  try {
    if (editingRule.value && editingRule.value.id) {
      await updateClassificationRule(editingRule.value.id, form)
      ElMessage.success('规则已更新')
    } else {
      await createClassificationRule(form as ClassificationRule)
      ElMessage.success('规则已创建')
    }
    showDialog.value = false
    await loadRules()
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function toggleRuleStatus(rule: ClassificationRule) {
  if (!rule.id) return
  try {
    await updateClassificationRule(rule.id, { is_active: rule.is_active })
    ElMessage.success(rule.is_active ? '规则已启用' : '规则已禁用')
  } catch (e) {
    rule.is_active = !rule.is_active
  }
}

async function deleteRule(rule: ClassificationRule) {
  await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
    type: 'warning'
  })
  if (!rule.id) return
  try {
    await deleteClassificationRule(rule.id)
    ElMessage.success('规则已删除')
    await loadRules()
  } catch (e) {
    console.error(e)
  }
}

async function seedRules() {
  try {
    await ElMessageBox.confirm('将初始化默认分类规则（仅首次调用有效），确认？', '确认初始化')
    await seedClassificationRules()
    ElMessage.success('默认规则已初始化')
    await loadRules()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('初始化失败')
    }
  }
}

function testPattern(rule: ClassificationRule) {
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
    const ruleType = testRule.value.rule_type

    if (ruleType === 'keyword_regex') {
      const patterns = pattern.split('|')
      testResult.value = patterns.some(p => text.toLowerCase().includes(p.trim().toLowerCase()))
    } else {
      testResult.value = text.toLowerCase().includes(pattern.toLowerCase())
    }
  } catch {
    testResult.value = false
    ElMessage.warning('匹配测试失败')
  }
}

onMounted(loadRules)
</script>

<style scoped lang="scss">
.rule-library {
  padding: 24px;
}

.pattern-code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  color: #cf1322;
  font-family: monospace;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.test-label {
  font-size: 13px;
  color: #666;
  margin-bottom: 8px;
}

.test-result {
  margin-top: 12px;
  text-align: center;
}
</style>
