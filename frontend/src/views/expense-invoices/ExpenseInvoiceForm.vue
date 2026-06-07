<template>
  <div class="expense-invoice-form-page">
    <div class="page-header">
      <h3>{{ isEdit ? '编辑进项发票' : '新增进项发票' }}</h3>
    </div>

    <el-card>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" size="small">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="发票号码" prop="invoice_no">
              <el-input v-model="form.invoice_no" placeholder="请输入发票号码" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发票代码" prop="invoice_code">
              <el-input v-model="form.invoice_code" placeholder="选填，电子发票可不填" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="开票日期" prop="invoice_date">
              <el-date-picker
                v-model="form.invoice_date"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="请选择开票日期"
                style="width: 100%"
                :disabled="readonly"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发票类型" prop="invoice_kind">
              <el-select v-model="form.invoice_kind" placeholder="请选择发票类型" style="width: 100%" :disabled="readonly">
                <el-option label="纸质普票" value="paper_normal" />
                <el-option label="纸质专票" value="paper_special" />
                <el-option label="电子普票" value="electronic_normal" />
                <el-option label="电子专票" value="electronic_special" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="费用类别" prop="category">
              <el-select v-model="form.category" placeholder="可选" clearable style="width: 100%" :disabled="readonly">
                <el-option label="交通" value="transport" />
                <el-option label="办公" value="office" />
                <el-option label="差旅" value="travel" />
                <el-option label="招待" value="entertain" />
                <el-option label="通讯" value="communication" />
                <el-option label="培训" value="training" />
                <el-option label="福利" value="welfare" />
                <el-option label="其他" value="other" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="不含税金额" prop="amount">
              <el-input
                v-model="form.amount"
                placeholder="0.00"
                :disabled="readonly"
                @input="onAmountChange"
              >
                <template #append>元</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="税额" prop="tax_amount">
              <el-input
                v-model="form.tax_amount"
                placeholder="0.00"
                :disabled="readonly"
                @input="onTaxAmountChange"
              >
                <template #append>元</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="价税合计" prop="total_amount">
              <el-input
                v-model="form.total_amount"
                placeholder="0.00"
                :disabled="readonly"
              >
                <template #append>元</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="供应商名称" prop="vendor_name">
              <el-input v-model="form.vendor_name" placeholder="选填" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="供应商税号" prop="tax_id">
              <el-input v-model="form.tax_id" placeholder="选填" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注" prop="description">
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="3"
                placeholder="选填"
                :disabled="readonly"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="form-actions">
        <el-button @click="goBack">取消</el-button>
        <template v-if="!readonly">
          <el-button type="primary" :loading="saving" @click="handleSave">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
        </template>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  fetchExpenseInvoiceDetail,
  createExpenseInvoice,
  updateExpenseInvoice,
} from '@/api/modules/expense-invoice'
import type { ExpenseInvoice, ExpenseInvoiceCreateRequest } from '@/api/modules/expense-invoice'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id && route.path.endsWith('/edit'))
const readonly = computed(() => {
  const p = route.params.id
  // 详情页只读，新建/编辑可写
  return !!p && !route.path.endsWith('/edit')
})

const formRef = ref()
const saving = ref(false)

const form = reactive<ExpenseInvoiceCreateRequest>({
  invoice_no: '',
  invoice_code: '',
  invoice_date: '',
  invoice_kind: '',
  category: '',
  amount: '',
  tax_amount: '',
  total_amount: '',
  vendor_name: '',
  tax_id: '',
  description: '',
})

const validateAmountField = (rule: any, value: any, callback: any) => {
  if (value === '' || value === null || value === undefined) {
    callback(new Error('请输入金额'))
    return
  }
  const num = Number(value)
  if (Number.isNaN(num) || num < 0) {
    callback(new Error('金额必须为非负数字'))
    return
  }
  callback()
}

const rules = {
  invoice_no: [{ required: true, message: '请输入发票号码', trigger: 'blur' }],
  invoice_date: [{ required: true, message: '请选择开票日期', trigger: 'change' }],
  invoice_kind: [{ required: true, message: '请选择发票类型', trigger: 'change' }],
  amount: [{ required: true, validator: validateAmountField, trigger: 'blur' }],
  tax_amount: [{ required: true, validator: validateAmountField, trigger: 'blur' }],
  total_amount: [{ required: true, validator: validateAmountField, trigger: 'blur' }],
}

/** 保留 2 位小数 */
function format2(n: number): string {
  if (!Number.isFinite(n)) return '0.00'
  return n.toFixed(2)
}

/** 解析输入为数字（不通过则返回 0） */
function parseAmount(v: string | number): number {
  const num = Number(v)
  return Number.isFinite(num) ? num : 0
}

/** amount 或 tax_amount 变化时自动重算 total_amount */
function recomputeTotal() {
  const sum = parseAmount(form.amount) + parseAmount(form.tax_amount)
  form.total_amount = format2(sum)
}

function onAmountChange() {
  recomputeTotal()
}

function onTaxAmountChange() {
  recomputeTotal()
}

async function handleSave() {
  try {
    await formRef.value.validate()
    saving.value = true
    // 后端要求 string 类型，el-input v-model 已是 string，规范化一下
    const payload: ExpenseInvoiceCreateRequest = {
      ...form,
      amount: format2(parseAmount(form.amount)),
      tax_amount: format2(parseAmount(form.tax_amount)),
      total_amount: format2(parseAmount(form.total_amount)),
    }
    if (isEdit.value) {
      await updateExpenseInvoice(route.params.id as string, payload)
      ElMessage.success('更新成功')
    } else {
      await createExpenseInvoice(payload)
      ElMessage.success('创建成功')
    }
    // 1.5s 后跳到列表页
    setTimeout(() => {
      router.push('/expense-invoices/list')
    }, 1500)
  } catch (e: any) {
    if (e !== false) {
      ElMessage.error(e?.response?.data?.error || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/expense-invoices/list')
  }
}

onMounted(async () => {
  if (route.params.id) {
    try {
      const res: any = await fetchExpenseInvoiceDetail(route.params.id as string)
      const data: ExpenseInvoice = res?.data || res
      Object.assign(form, {
        invoice_no: data.invoice_no || '',
        invoice_code: data.invoice_code || '',
        invoice_date: data.invoice_date || '',
        invoice_kind: data.invoice_kind || '',
        category: data.category || '',
        amount: data.amount || '',
        tax_amount: data.tax_amount || '',
        total_amount: data.total_amount || '',
        vendor_name: data.vendor_name || '',
        tax_id: data.tax_id || '',
        description: data.description || '',
      })
    } catch {
      ElMessage.error('加载数据失败')
    }
  }
})
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>
