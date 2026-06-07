<template>
  <div class="voucher-template-page">
    <div class="page-header">
      <h3>凭证模板配置</h3>
      <div>
        <el-button @click="seedDefaults">初始化默认模板</el-button>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新建模板
        </el-button>
      </div>
    </div>

    <el-alert type="info" :closable="false" show-icon style="margin-bottom: 12px">
      <p>凭证模板用于在自动生成凭证时按 <b>分类标识</b> 选择借/贷方科目。一级借/贷方科目在模板行中指定；占位符 <code v-text="'{{amount}}'"></code> 表示自动填入流水金额，<code v-text="'{{party}}'"></code> 表示对方户名。</p>
      <p>每个分类下只能有一个 <b>启用</b> 模板。</p>
    </el-alert>

    <el-table :data="templates" border stripe size="small" v-loading="loading">
      <el-table-column prop="name" label="模板名称" width="200" />
      <el-table-column label="分类" width="160">
        <template #default="{ row }">
          <el-tag v-if="row.classification" size="small" type="primary">{{ row.classification }}</el-tag>
          <span v-else class="text-muted">通用</span>
        </template>
      </el-table-column>
      <el-table-column prop="number_prefix" label="编号前缀" width="100" />
      <el-table-column label="模板行" width="80">
        <template #default="{ row }">
          <el-tag size="small">{{ (row.lines || []).length }} 行</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批流" width="120">
        <template #default="{ row }">
          <span v-if="row.approval_flow_id" class="text-muted">已绑定</span>
          <span v-else class="text-muted">未绑定</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">{{ row.is_active ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="160" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="editTemplate(row)">编辑</el-button>
          <el-button link :type="row.is_active ? 'warning' : 'success'" size="small" @click="toggleActive(row)">
            {{ row.is_active ? '停用' : '启用' }}
          </el-button>
          <el-button link type="danger" size="small" @click="deleteTemplate(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑凭证模板' : '新建凭证模板'" width="720px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="form.name" placeholder="如：银行手续费" maxlength="100" />
        </el-form-item>
        <el-form-item label="分类标识" prop="classification">
          <el-select v-model="form.classification" clearable placeholder="选择分类（必填，启用后该分类下唯一）" style="width: 100%">
            <el-option v-for="opt in classificationOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="编号前缀">
              <el-input v-model="form.number_prefix" placeholder="如：PZ" maxlength="20" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用">
              <el-switch v-model="form.is_active" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="说明">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="200" />
        </el-form-item>
        <el-form-item label="模板行" required>
          <el-table :data="form.lines" border size="small">
            <el-table-column label="行序" width="70">
              <template #default="{ $index }">{{ $index + 1 }}</template>
            </el-table-column>
            <el-table-column label="会计科目" min-width="220">
              <template #default="{ $index }">
                <AccountSelector v-model="form.lines[$index].account" />
              </template>
            </el-table-column>
            <el-table-column label="借方金额" width="130">
              <template #default="{ $index }">
                <el-input v-model="form.lines[$index].dr_amount_template" placeholder="如：{{amount}}" />
              </template>
            </el-table-column>
            <el-table-column label="贷方金额" width="130">
              <template #default="{ $index }">
                <el-input v-model="form.lines[$index].cr_amount_template" placeholder="如：0" />
              </template>
            </el-table-column>
            <el-table-column label="摘要" min-width="140">
              <template #default="{ $index }">
                <el-input v-model="form.lines[$index].summary_template" placeholder="如：手续费" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="60">
              <template #default="{ $index }">
                <el-button link type="danger" size="small" @click="form.lines.splice($index, 1)">×</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top: 8px">
            <el-button text type="primary" @click="addLine">+ 添加模板行</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTemplate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import type { Account } from '@/types/models/account'
import request from '@/api/request'
import AccountSelector from '@/components/business/AccountSelector.vue'

interface TemplateLine {
  id?: string
  account: any
  dr_amount_template: string
  cr_amount_template: string
  summary_template: string
  line_order: number
}

interface Template {
  id: string
  name: string
  description: string
  number_prefix: string
  is_active: boolean
  classification: string | null
  approval_flow_id: string | null
  lines: TemplateLine[]
}

const templates = ref<Template[]>([])
const loading = ref(false)
const saving = ref(false)
const showDialog = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()

const classificationOptions = [
  { value: 'bank_fee',          label: '银行手续费 (bank_fee)' },
  { value: 'interest_income',   label: '利息收入 (interest_income)' },
  { value: 'business_receipt',  label: '业务收款 (business_receipt)' },
  { value: 'business_payment',  label: '业务付款 (business_payment)' },
  { value: 'internal_transfer', label: '内部转账 (internal_transfer)' },
  { value: 'tax_payment',       label: '税务缴费 (tax_payment)' },
  { value: 'social_security',   label: '社保缴费 (social_security)' },
]

const defaultForm = {
  name: '',
  classification: '' as string,
  number_prefix: 'PZ',
  description: '',
  is_active: true,
  lines: [] as TemplateLine[],
}

const form = reactive({ ...defaultForm })

const formRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
}

async function loadTemplates() {
  loading.value = true
  try {
    const res: any = await request.get('/voucher-templates')
    const list = Array.isArray(res?.data) ? res.data : []
    const detailed: Template[] = []
    for (const t of list) {
      try {
        const detail: any = await request.get(`/voucher-templates/${t.id}`)
        detailed.push(detail?.data ?? t)
      } catch {
        detailed.push(t)
      }
    }
    templates.value = detailed
  } catch (e) {
    ElMessage.error('加载凭证模板失败')
    templates.value = []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  Object.assign(form, defaultForm, { lines: [] })
  addLine()
  addLine()
  showDialog.value = true
}

function editTemplate(row: Template) {
  editingId.value = row.id
  form.name = row.name
  form.classification = row.classification || ''
  form.number_prefix = row.number_prefix
  form.description = row.description || ''
  form.is_active = row.is_active
  form.lines = (row.lines || []).map((l: any) => ({
    id: l.id,
    account: l.account_id ? { id: l.account_id, code: l.account_code, name: l.account_name } as Account : null,
    dr_amount_template: l.dr_amount_template || '',
    cr_amount_template: l.cr_amount_template || '',
    summary_template: l.summary_template || '',
    line_order: l.line_order || 0,
  }))
  if (form.lines.length === 0) addLine()
  showDialog.value = true
}

function addLine() {
  form.lines.push({
    account: null,
    dr_amount_template: '',
    cr_amount_template: '',
    summary_template: '',
    line_order: form.lines.length + 1,
  })
}

async function saveTemplate() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (form.lines.length === 0) {
    ElMessage.warning('请至少添加一行模板行')
    return
  }
  if (form.lines.some(l => !l.account)) {
    ElMessage.warning('所有模板行都必须选择会计科目')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      number_prefix: form.number_prefix,
      is_active: form.is_active,
      classification: form.classification || null,
      lines: form.lines.map((l, idx) => ({
        account_id: l.account!.id,
        dr_amount_template: l.dr_amount_template,
        cr_amount_template: l.cr_amount_template,
        summary_template: l.summary_template,
        line_order: idx + 1,
      })),
    }
    if (editingId.value) {
      await request.put(`/voucher-templates/${editingId.value}`, payload)
      ElMessage.success('模板已更新')
    } else {
      await request.post('/voucher-templates', payload)
      ElMessage.success('模板已创建')
    }
    showDialog.value = false
    await loadTemplates()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

async function toggleActive(row: Template) {
  const nextActive = !row.is_active
  try {
    await request.put(`/voucher-templates/${row.id}`, {
      name: row.name,
      description: row.description,
      number_prefix: row.number_prefix,
      is_active: nextActive,
      classification: row.classification,
      lines: (row.lines || []).map((l: any, idx: number) => ({
        account_id: l.account_id,
        dr_amount_template: l.dr_amount_template,
        cr_amount_template: l.cr_amount_template,
        summary_template: l.summary_template,
        line_order: idx + 1,
      })),
    })
    row.is_active = nextActive
    ElMessage.success(nextActive ? '已启用' : '已停用')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '操作失败')
  }
}

async function deleteTemplate(row: Template) {
  try {
    await ElMessageBox.confirm(`确定删除模板「${row.name}」？此操作不可撤销。`, '确认删除', {
      type: 'warning',
    })
  } catch { return }
  try {
    await request.delete(`/voucher-templates/${row.id}`)
    ElMessage.success('已删除')
    await loadTemplates()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '删除失败')
  }
}

async function seedDefaults() {
  try {
    await ElMessageBox.confirm('将为所有未配置模板的分类创建默认模板。继续？', '初始化默认模板', {
      type: 'info',
    })
  } catch { return }
  try {
    const accounts: any = await request.get('/accounts/tree')
    const findByCode = (code: string): string | null => {
      const stack: any[] = Array.isArray(accounts?.data) ? accounts.data : []
      while (stack.length) {
        const n = stack.shift()
        if (n.code === code) return n.id
        if (n.children?.length) stack.push(...n.children)
      }
      return null
    }
    const acct5602 = findByCode('5602')
    const acct1002 = findByCode('1002')
    const acct5601 = findByCode('5601')
    const acct1122 = findByCode('1122')
    const acct1001 = findByCode('1001')
    const acct2221 = findByCode('2221')
    const acct2211 = findByCode('2211')
    if (!acct1002) {
      ElMessage.error('找不到科目 1002 银行存款，请先初始化账套')
      return
    }

    const defaults = [
      { name: '银行手续费模板',     classification: 'bank_fee',          lines: [{ account_id: acct5602, dr: '{{amount}}', cr: '' }, { account_id: acct1002, dr: '', cr: '{{amount}}' }] },
      { name: '利息收入模板',       classification: 'interest_income',   lines: [{ account_id: acct1002, dr: '{{amount}}', cr: '' }, { account_id: acct5601, dr: '', cr: '{{amount}}' }] },
      { name: '业务收款模板',       classification: 'business_receipt',  lines: [{ account_id: acct1002, dr: '{{amount}}', cr: '' }, { account_id: acct1122, dr: '', cr: '{{amount}}' }] },
      { name: '业务付款模板',       classification: 'business_payment',  lines: [{ account_id: acct1122, dr: '{{amount}}', cr: '' }, { account_id: acct1002, dr: '', cr: '{{amount}}' }] },
      { name: '内部转账模板',       classification: 'internal_transfer', lines: [{ account_id: acct1001, dr: '{{amount}}', cr: '' }, { account_id: acct1002, dr: '', cr: '{{amount}}' }] },
    ]
    if (acct2221) defaults.push({ name: '税务缴费模板', classification: 'tax_payment', lines: [{ account_id: acct2221, dr: '{{amount}}', cr: '' }, { account_id: acct1002, dr: '', cr: '{{amount}}' }] })
    if (acct2211) defaults.push({ name: '社保缴费模板', classification: 'social_security', lines: [{ account_id: acct2211, dr: '{{amount}}', cr: '' }, { account_id: acct1002, dr: '', cr: '{{amount}}' }] })

    let ok = 0, fail = 0
    for (const d of defaults) {
      try {
        await request.post('/voucher-templates', {
          name: d.name,
          description: `默认模板 - ${d.classification}`,
          number_prefix: 'PZ',
          is_active: true,
          classification: d.classification,
          lines: d.lines.map((l, idx) => ({
            account_id: l.account_id,
            dr_amount_template: l.dr,
            cr_amount_template: l.cr,
            summary_template: '',
            line_order: idx + 1,
          })),
        })
        ok++
      } catch {
        fail++
      }
    }
    ElMessage.success(`已创建 ${ok} 个模板${fail ? `，${fail} 个失败` : ''}`)
    await loadTemplates()
  } catch (e: any) {
    ElMessage.error(e?.message || '初始化失败')
  }
}

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped lang="scss">
.voucher-template-page { padding: 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; margin: 0; } }
.text-muted { color: #999; }
code { background: #f5f5f5; padding: 2px 6px; border-radius: 3px; font-size: 12px; }
</style>
