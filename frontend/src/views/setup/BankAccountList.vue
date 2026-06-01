<template>
  <div class="bank-account-page">
    <PageLayout title="资金账户管理" icon="💰" subtitle="管理银行存款和库存现金账户，实时监控余额变动">
      <template #actions>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新增账户
        </el-button>
      </template>

      <el-row :gutter="16" class="balance-cards">
        <el-col v-for="acct in accounts" :key="acct.id" :span="6">
          <el-card shadow="hover" :class="['balance-card', acct.bank_account_type]">
            <div class="card-header">
              <span class="bank-name">{{ acct.bank_name }}</span>
              <el-tag v-if="!acct.is_active" size="small" type="danger">已停用</el-tag>
            </div>
            <p class="account-number">{{ maskAccount(acct.account_number) }}</p>
            <p class="balance">{{ acct.balance || '0.00' }}</p>
            <p class="balance-label">{{ acct.currency }} 余额</p>
          </el-card>
        </el-col>
      </el-row>

      <el-table :data="accounts" border stripe size="small">
        <el-table-column prop="bank_name" label="账户名称" width="140" />
        <el-table-column prop="account_number" label="账号" width="180">
          <template #default="{ row }">{{ maskAccount(row.account_number) }}</template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="row.bank_account_type === 'bank' ? 'primary' : 'success'" size="small">
              {{ row.bank_account_type === 'bank' ? '银行存款' : '库存现金' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关联科目" width="160">
          <template #default="{ row }">{{ row.clearing_account_code || '-' }}</template>
        </el-table-column>
        <el-table-column v-if="hasCashAccount" label="保管人/位置" width="200">
          <template #default="{ row }">
            <template v-if="row.bank_account_type === 'cash'">
              <div style="line-height: 1.4">
                <div>👤 {{ row.custodian || '未指定' }}</div>
                <div style="font-size: 12px; color: #999">📍 {{ row.location || '未指定' }}</div>
              </div>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="currency" label="币种" width="60" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">{{ row.is_active ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editAccount(row)">编辑</el-button>
            <el-button link :type="row.is_active ? 'warning' : 'success'" size="small" @click="toggleActive(row)">
              {{ row.is_active ? '停用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </PageLayout>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑账户' : '新增资金账户'" width="560px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="账户名称" prop="bankName">
          <el-input v-model="form.bankName" placeholder="如：工商银行-基本户" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="账号" prop="accountNumber">
              <el-input v-model="form.accountNumber" placeholder="银行账号" maxlength="30" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="账户类型">
              <el-radio-group v-model="form.accountType">
                <el-radio value="bank">银行存款</el-radio>
                <el-radio value="cash">库存现金</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16" v-if="form.accountType === 'bank'">
          <el-col :span="12">
            <el-form-item label="IBAN">
              <el-input v-model="form.iban" placeholder="国际银行账户号码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SWIFT Code">
              <el-input v-model="form.swiftCode" placeholder="如：ICBKCNBJ" maxlength="11" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16" v-if="form.accountType === 'cash'">
          <el-col :span="12">
            <el-form-item label="保管人">
              <el-input v-model="form.custodian" placeholder="如：出纳姓名" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="存放位置">
              <el-input v-model="form.location" placeholder="如：财务部保险柜" maxlength="100" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="关联GL科目">
          <AccountSelector v-model="form.clearingAccount" :ledger-only="true" :disabled="!!editingId" />
          <p class="form-hint">必须选择资产类 Ledger 科目{{ editingId ? '（创建后不可修改）' : '' }}</p>
        </el-form-item>
        <el-form-item label="币种">
          <el-select v-model="form.currency" style="width: 120px">
            <el-option label="CNY" value="CNY" />
            <el-option label="USD" value="USD" />
            <el-option label="HKD" value="HKD" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveAccount">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import type { Account } from '@/types/models/account'
import request from '@/api/request'
import PageLayout from '@/components/app/PageLayout.vue'

interface BankAccountItem {
  id: string
  bank_name: string
  account_number: string
  bank_account_type: 'bank' | 'cash'
  clearing_account_code?: string
  clearing_account?: Account
  iban?: string
  swift_code?: string
  currency: string
  is_active: boolean
  is_cash?: boolean
  custodian?: string
  location?: string
  balance?: string
}

const accounts = ref<BankAccountItem[]>([])

const hasCashAccount = computed(() => accounts.value.some(a => a.bank_account_type === 'cash'))

async function loadBankAccounts() {
  try {
    const res: any = await request.get('/bank-accounts')
    const list = res?.data?.list !== undefined && res?.data?.list !== null ? res.data.list : (res?.data !== undefined && res?.data !== null ? res.data : res)
    accounts.value = Array.isArray(list) ? list : []
  } catch (e) {
    console.warn('后端资金账户接口不可用', e)
    accounts.value = []
  }
}
onMounted(loadBankAccounts)

const showDialog = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()
const saving = ref(false)

const defaultForm = {
  bankName: '', accountNumber: '', accountType: 'bank',
  iban: '', swiftCode: '', clearingAccount: null as Account | null,
  currency: 'CNY',
  custodian: '', location: '',
}

const form = reactive({ ...defaultForm })

const formRules = {
  bankName: [{ required: true, message: '请输入账户名称', trigger: 'blur' }],
  accountNumber: [{ required: true, message: '请输入账号', trigger: 'blur' }],
}

function maskAccount(num: string) {
  if (num === '-' || num.length < 8) return num
  return num.slice(0, 4) + ' **** **** ' + num.slice(-4)
}

function openCreate() {
  editingId.value = ''
  Object.assign(form, defaultForm)
  showDialog.value = true
}

function editAccount(item: BankAccountItem) {
  editingId.value = item.id
  form.bankName = item.bank_name
  form.accountNumber = item.account_number
  form.accountType = item.bank_account_type
  form.iban = item.iban || ''
  form.swiftCode = item.swift_code || ''
  form.currency = item.currency
  form.clearingAccount = item.clearing_account || null
  form.custodian = item.custodian || ''
  form.location = item.location || ''
  showDialog.value = true
}

async function saveAccount() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!form.clearingAccount && !editingId.value) {
    ElMessage.warning('请选择关联 GL 科目')
    return
  }
  saving.value = true

  try {
    if (editingId.value) {
      await request.put(`/bank-accounts/${editingId.value}`, {
        bank_name: form.bankName,
        account_number: form.accountNumber,
        bank_account_type: form.accountType,
        iban: form.iban || undefined,
        swift_code: form.swiftCode || undefined,
        currency: form.currency,
        clearing_account_id: form.clearingAccount?.id || undefined,
        is_cash: form.accountType === 'cash',
        custodian: form.accountType === 'cash' ? form.custodian : undefined,
        location: form.accountType === 'cash' ? form.location : undefined,
      })
      await loadBankAccounts()
      ElMessage.success('账户已更新')
      showDialog.value = false
    } else {
      const res: any = await request.post('/bank-accounts', {
        bank_name: form.bankName,
        account_number: form.accountNumber,
        bank_account_type: form.accountType,
        iban: form.iban || undefined,
        swift_code: form.swiftCode || undefined,
        currency: form.currency,
        clearing_account_id: form.clearingAccount!.id,
        is_cash: form.accountType === 'cash',
        custodian: form.accountType === 'cash' ? form.custodian : undefined,
        location: form.accountType === 'cash' ? form.location : undefined,
      })
      const created = res?.data
      if (created) {
        accounts.value.push(created)
      } else {
        await loadBankAccounts()
      }
      ElMessage.success('账户已创建')
      showDialog.value = false
    }
  } catch (e: any) {
    const msg = e?.response?.data?.error || '保存失败'
    ElMessage.error(msg)
  } finally {
    saving.value = false
  }
}

async function toggleActive(item: BankAccountItem) {
  const nextActive = !item.is_active
  try {
    await request.put(`/bank-accounts/${item.id}`, {
      is_active: nextActive,
      bank_name: item.bank_name,
      account_number: item.account_number,
      currency: item.currency,
    })
    item.is_active = nextActive
    ElMessage.success(nextActive ? '账户已启用' : '账户已停用')
  } catch (e: any) {
    const msg = e?.response?.data?.error || '操作失败'
    ElMessage.error(msg)
  }
}
</script>

<style scoped lang="scss">
.bank-account-page {
  padding: 24px;

  .balance-cards { margin-bottom: 16px; }
}

.balance-card {
  .card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
  .bank-name { font-weight: 600; font-size: 14px; }
  .account-number { color: #999; font-family: monospace; font-size: 12px; margin-bottom: 8px; }
  .balance { font-size: 22px; font-weight: 600; }
  .balance-label { font-size: 12px; color: #999; margin-top: 4px; }
  &.cash { border-left: 3px solid #52c41a; }
  &.bank { border-left: 3px solid #1890ff; }
}
.form-hint { font-size: 12px; color: #999; margin-top: 4px; }
</style>
