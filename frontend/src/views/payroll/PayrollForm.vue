<template>
  <div class="payroll-form-page">
    <div class="page-header">
      <h3>{{ isEdit ? '编辑工资单' : '新建工资单' }}</h3>
    </div>

    <el-card>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" size="small">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="员工姓名" prop="employee_name">
              <el-input v-model="form.employee_name" placeholder="请输入员工姓名" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="部门" prop="department_name">
              <el-input v-model="form.department_name" placeholder="请输入部门" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工资期间" prop="period_no">
              <el-input v-model="form.period_no" placeholder="如 202506" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发放日期" prop="payment_date">
              <el-date-picker v-model="form.payment_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="应发工资" prop="gross_salary">
              <el-input-number v-model="form.gross_salary" :min="0" :precision="2" :controls="false" style="width: 100%" placeholder="0.00" :disabled="readonly" @change="calcNetSalary" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代扣个税" prop="individual_tax">
              <el-input-number v-model="form.individual_tax" :min="0" :precision="2" :controls="false" style="width: 100%" placeholder="0.00" :disabled="readonly" @change="calcNetSalary" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代扣社保" prop="social_security">
              <el-input-number v-model="form.social_security" :min="0" :precision="2" :controls="false" style="width: 100%" placeholder="0.00" :disabled="readonly" @change="calcNetSalary" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代扣公积金" prop="housing_fund">
              <el-input-number v-model="form.housing_fund" :min="0" :precision="2" :controls="false" style="width: 100%" placeholder="0.00" :disabled="readonly" @change="calcNetSalary" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="其他扣款" prop="other_deductions">
              <el-input-number v-model="form.other_deductions" :min="0" :precision="2" :controls="false" style="width: 100%" placeholder="0.00" :disabled="readonly" @change="calcNetSalary" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="银行卡号">
              <el-input v-model="form.bank_account_no" placeholder="请输入银行卡号" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="可选" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="实发工资">
              <b class="net-salary-display">¥{{ computedNetSalary }}</b>
              <span class="net-hint">（自动计算：应发 - 个税 - 社保 - 公积金 - 其他）</span>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="form-actions">
        <el-button @click="goBack">返回</el-button>
        <template v-if="!readonly">
          <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
        </template>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchPayrollDetail, createPayroll, updatePayroll } from '@/api/modules/payroll'
import type { Payroll } from '@/types/models/payroll'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id && route.path.endsWith('/edit'))
const readonly = computed(() => {
  const p = route.params.id
  return !!p && route.path !== `/payroll/${p}/edit`
})

const formRef = ref()
const saving = ref(false)

const form = reactive({
  employee_name: '',
  department_name: '',
  period_no: '',
  gross_salary: 0,
  individual_tax: 0,
  social_security: 0,
  housing_fund: 0,
  other_deductions: 0,
  payment_date: '',
  bank_account_no: '',
  remark: '',
  net_salary: 0,
})

const rules = {
  employee_name: [{ required: true, message: '请输入员工姓名', trigger: 'blur' }],
  department_name: [{ required: true, message: '请输入部门', trigger: 'blur' }],
  period_no: [{ required: true, message: '请输入工资期间', trigger: 'blur' }],
  gross_salary: [{ required: true, message: '请输入应发工资', trigger: 'blur' }],
  payment_date: [{ required: true, message: '请选择发放日期', trigger: 'change' }],
}

const computedNetSalary = computed(() => {
  const gross = form.gross_salary || 0
  const tax = form.individual_tax || 0
  const ss = form.social_security || 0
  const hf = form.housing_fund || 0
  const other = form.other_deductions || 0
  return (gross - tax - ss - hf - other).toLocaleString('en', { minimumFractionDigits: 2 })
})

function calcNetSalary() {
  form.net_salary = parseFloat(computedNetSalary.value.replace(/,/g, '')) || 0
}

async function handleSave() {
  try {
    await formRef.value.validate()
    saving.value = true
    const payload: Partial<Payroll> = {
      ...form,
      period_no: Number(form.period_no),
      gross_salary: String(form.gross_salary),
      individual_tax: String(form.individual_tax),
      social_security: String(form.social_security),
      housing_fund: String(form.housing_fund),
      other_deductions: String(form.other_deductions),
      net_salary: String(form.net_salary || parseFloat(computedNetSalary.value.replace(/,/g, ''))),
    }
    if (isEdit.value) {
      await updatePayroll(route.params.id as string, payload)
      ElMessage.success('更新成功')
    } else {
      await createPayroll(payload)
      ElMessage.success('创建成功')
    }
    goBack()
  } catch (e: any) {
    if (e !== false) {
      ElMessage.error(e?.response?.data?.error || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(async () => {
  if (route.params.id) {
    try {
      const res: any = await fetchPayrollDetail(route.params.id as string)
      const data: Payroll = res?.data || res
      Object.assign(form, {
        employee_name: data.employee_name,
        department_name: data.department_name,
        period_no: String(data.period_no),
        gross_salary: parseFloat(data.gross_salary) || 0,
        individual_tax: parseFloat(data.individual_tax) || 0,
        social_security: parseFloat(data.social_security) || 0,
        housing_fund: parseFloat(data.housing_fund) || 0,
        other_deductions: parseFloat(data.other_deductions) || 0,
        payment_date: data.payment_date,
        bank_account_no: data.bank_account_no || '',
        remark: data.remark || '',
        net_salary: parseFloat(data.net_salary) || 0,
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
.net-salary-display {
  font-size: 20px;
  color: #389e0d;
}
.net-hint {
  margin-left: 12px;
  font-size: 12px;
  color: #999;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>