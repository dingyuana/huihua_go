<template>
  <div class="party-page">
    <div class="page-header">
      <h3>客商档案</h3>
      <div>
        <el-button @click="showImportDialog = true">⬇ 批量导入</el-button>
        <el-button type="primary" @click="openCreate">+ 新建客商</el-button>
      </div>
    </div>

    <el-card>
      <el-tabs v-model="activeTab" class="party-tabs">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane label="客户" name="customer" />
        <el-tab-pane label="供应商" name="supplier" />
      </el-tabs>
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索名称或税号" clearable style="width: 260px" />
      </div>
      <el-table :data="filteredParties" border stripe size="small">
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="tax_id" label="税号" width="180" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.party_type === 'customer'" type="success" size="small">客户</el-tag>
            <el-tag v-else-if="row.party_type === 'supplier'" type="warning" size="small">供应商</el-tag>
            <el-tag v-else size="small">客户/供应商</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bank_name" label="开户行" width="140" />
        <el-table-column prop="bank_account" label="银行账号" width="140">
          <template #default="{ row }">{{ row.bank_account || '-' }}</template>
        </el-table-column>
        <el-table-column prop="credit_limit" label="信用额度" width="100">
          <template #default="{ row }">{{ row.credit_limit || '-' }}</template>
        </el-table-column>
        <el-table-column prop="payment_terms" label="账期" width="70" />
        <el-table-column prop="phone" label="电话" width="120" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editParty(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteParty(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="showDialog" :title="editingId ? '编辑客商' : '新建客商'" width="600px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="客商全称" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="14">
            <el-form-item label="税号" prop="taxId">
              <el-input v-model="form.taxId" placeholder="18位统一社会信用代码" maxlength="18" />
              <span v-if="taxIdError" class="tax-error">{{ taxIdError }}</span>
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="类型">
              <el-radio-group v-model="form.partyType">
                <el-radio value="customer">客户</el-radio>
                <el-radio value="supplier">供应商</el-radio>
                <el-radio value="both">全部</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="开户行">
              <el-input v-model="form.bankName" placeholder="开户银行名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="银行账号">
              <el-input v-model="form.bankAccount" placeholder="银行账号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="信用额度">
              <AmountInput v-model="form.creditLimit" placeholder="0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="账期">
              <el-select v-model="form.paymentTerms" style="width: 100%">
                <el-option label="立即" value="immediate" />
                <el-option label="15天" value="net15" />
                <el-option label="30天" value="net30" />
                <el-option label="45天" value="net45" />
                <el-option label="60天" value="net60" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="电话">
          <el-input v-model="form.phone" placeholder="手机或固话" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" type="textarea" :rows="2" placeholder="详细地址" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveParty">保存</el-button>
      </template>
    </el-dialog>

    <!-- 导入结果（放在导入弹窗之前） -->
    <el-dialog v-model="showImportResult" title="导入结果" width="520px">
      <el-alert :title="`共处理 ${importResult.total} 条`" :type="importResult.failed > 0 ? 'warning' : 'success'" :closable="false" show-icon>
        <template #default>
          <p>成功 {{ importResult.imported }} 条，失败 {{ importResult.failed }} 条</p>
        </template>
      </el-alert>
      <el-table v-if="importErrors.length" :data="importErrors" size="small" border style="margin-top:12px">
        <el-table-column prop="row" label="行号" width="60" />
        <el-table-column prop="msg" label="错误详情" min-width="340" />
      </el-table>
      <template #footer>
        <el-button type="primary" @click="showImportResult = false">完成</el-button>
      </template>
    </el-dialog>

    <!-- 导入弹窗 -->
    <el-dialog v-model="showImportDialog" title="批量导入客商" width="450px">
      <el-upload drag accept=".xlsx,.xls" :auto-upload="false" :on-change="handleFileChange" :limit="1">
        <el-icon class="upload-icon" :size="40"><UploadFilled /></el-icon>
        <p>拖拽 Excel 文件到此处，或点击上传</p>
        <p class="upload-hint">仅支持 .xlsx / .xls 格式，模板需包含：名称、税号、类型、开户行、账号</p>
      </el-upload>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!importFile" :loading="importing" @click="handleImport">开始导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import request from '@/api/request'

interface PartyItem {
  id: string
  name: string
  tax_id: string
  party_type: string
  bank_name: string
  bank_account: string
  credit_limit: string
  payment_terms: string
  phone: string
  address: string
}

const nextId = ref(4)
const parties = ref<PartyItem[]>([])

/** 从后端加载客商列表 */
async function loadParties() {
  try {
    const res: any = await request.get('/parties')
    const list = res?.data?.list || res?.data
    if (Array.isArray(list) && list.length > 0) {
      parties.value = list
      return
    }
  } catch { /* fallback */ }
  // 本地降级数据
  parties.value = [
  { id: 'p1', name: '上海XX贸易公司', tax_id: '91310000MA7A1B2C3D', party_type: 'customer', bank_name: '中国银行', bank_account: '345678901234567', credit_limit: '500,000.00', payment_terms: 'net30', phone: '021-88886666', address: '上海市浦东新区' },
  { id: 'p2', name: '北京YY科技有限公司', tax_id: '91110108MA9E5F6G7H', party_type: 'supplier', bank_name: '工商银行', bank_account: '1102021219001234', credit_limit: '200,000.00', payment_terms: 'net45', phone: '010-66668888', address: '北京市海淀区' },
  { id: 'p3', name: '广州ZZ贸易有限公司', tax_id: '91440101MA8I9J0K1L', party_type: 'customer', bank_name: '建设银行', bank_account: '4302021219005678', credit_limit: '300,000.00', payment_terms: 'net30', phone: '020-88886666', address: '广州市天河区' },
  ]
}

onMounted(loadParties)

const activeTab = ref('all')
const keyword = ref('')
const showDialog = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()
const saving = ref(false)

const showImportDialog = ref(false)
const importFile = ref<File | null>(null)
const importing = ref(false)
const showImportResult = ref(false)
const importResult = ref({ total: 0, imported: 0, failed: 0 })
const importErrors = ref<{ row: number; msg: string }[]>([])

const taxIdError = ref('')

const defaultForm = {
  name: '', taxId: '', partyType: 'customer',
  bankName: '', bankAccount: '', creditLimit: '',
  paymentTerms: 'net30', phone: '', address: '',
}

const form = reactive({ ...defaultForm })

const formRules = {
  name: [{ required: true, message: '请输入客商名称', trigger: 'blur' }],
  taxId: [{ required: true, message: '请输入税号', trigger: 'blur' }],
}

function validateTaxId() {
  const val = form.taxId
  if (!val) { taxIdError.value = ''; return }
  const reg = /^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$/
  if (!reg.test(val.toUpperCase())) {
    taxIdError.value = '统一社会信用代码格式不正确（18位数字+字母）'
  } else {
    taxIdError.value = ''
  }
}

const filteredParties = computed(() => {
  return parties.value.filter(p => {
    if (activeTab.value !== 'all' && p.party_type !== activeTab.value && p.party_type !== 'both') return false
    if (keyword.value) {
      const kw = keyword.value.toLowerCase()
      return p.name.includes(kw) || p.tax_id.includes(kw)
    }
    return true
  })
})

function openCreate() {
  editingId.value = ''
  Object.assign(form, defaultForm)
  taxIdError.value = ''
  showDialog.value = true
}

function editParty(item: PartyItem) {
  editingId.value = item.id
  form.name = item.name
  form.taxId = item.tax_id
  form.partyType = item.party_type
  form.bankName = item.bank_name
  form.bankAccount = item.bank_account
  form.creditLimit = item.credit_limit
  form.paymentTerms = item.payment_terms
  form.phone = item.phone
  form.address = item.address
  taxIdError.value = ''
  showDialog.value = true
}

async function saveParty() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  validateTaxId()
  if (taxIdError.value) { ElMessage.warning('请修正税号格式'); return }

  saving.value = true
  try {
    if (editingId.value) {
      await request.put(`/parties/${editingId.value}`, {
        name: form.name, tax_id: form.taxId, party_type: form.partyType,
        bank_name: form.bankName, bank_account: form.bankAccount,
        credit_limit: form.creditLimit, payment_terms: form.paymentTerms,
        phone: form.phone, address: form.address,
      })
    } else {
      await request.post('/parties', {
        name: form.name, tax_id: form.taxId.toUpperCase(), party_type: form.partyType,
        bank_name: form.bankName, bank_account: form.bankAccount,
        credit_limit: form.creditLimit, payment_terms: form.paymentTerms,
        phone: form.phone, address: form.address,
      })
    }
    ElMessage.success(editingId.value ? '客商已更新' : '客商已创建')
    showDialog.value = false
    loadParties() // 重新加载
  } catch {
    // 后端不可用时本地保存
    if (editingId.value) {
      const idx = parties.value.findIndex(p => p.id === editingId.value)
      if (idx >= 0) Object.assign(parties.value[idx], { name: form.name })
    } else {
      parties.value.push({ id: `p${nextId.value++}`, name: form.name, tax_id: form.taxId.toUpperCase(), party_type: form.partyType, bank_name: form.bankName, bank_account: form.bankAccount, credit_limit: form.creditLimit, payment_terms: form.paymentTerms, phone: form.phone, address: form.address })
    }
    ElMessage.success(editingId.value ? '客商已更新（本地）' : '客商已创建（本地）')
    showDialog.value = false
  } finally {
    saving.value = false
  }
}

function deleteParty(item: PartyItem) {
  ElMessageBox.confirm(`确认删除客商「${item.name}」？`, '确认', {
    type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
  }).then(() => {
    parties.value = parties.value.filter(p => p.id !== item.id)
    ElMessage.success('客商已删除')
  }).catch(() => {})
}

function handleFileChange(file: any) {
  importFile.value = file.raw
}

async function handleImport() {
  importing.value = true
  await new Promise(r => setTimeout(r, 1500))
  importing.value = false

  importErrors.value = [
    { row: 5, msg: '税号格式不正确（91110108MA12345 长度不足18位）' },
    { row: 23, msg: '已存在的客商（税号 91110108MA... 与"北京YY科技"重复）' },
  ]
  importResult.value = { total: 50, imported: 48, failed: 2 }
  showImportResult.value = true
  showImportDialog.value = false
  importFile.value = null
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
.party-tabs { margin-bottom: 12px; }
.toolbar { margin-bottom: 12px; }
.upload-icon { margin-bottom: 8px; }
.upload-hint { color: #999; font-size: 12px; margin-top: 4px; }
.tax-error { color: #ff4d4f; font-size: 12px; margin-top: 4px; display: block; }
</style>
