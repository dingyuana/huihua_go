<template>
  <div class="payment-form-page">
    <div class="page-header">
      <h3>{{ isEdit ? '编辑收付款单' : '新增收付款单' }}</h3>
    </div>

    <el-card>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
        size="default"
        @submit.prevent="handleSubmit"
      >
        <el-form-item label="类型" prop="payment_type">
          <el-radio-group v-model="form.payment_type">
            <el-radio value="receive">收款</el-radio>
            <el-radio value="pay">付款</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="对方单位名称" prop="counterparty_name">
          <el-input v-model="form.counterparty_name" placeholder="请输入对方单位名称" maxlength="100" style="width: 360px" />
        </el-form-item>

        <el-form-item label="金额" prop="paid_amount">
          <el-input
            v-model="form.paid_amount"
            placeholder="请输入金额"
            style="width: 240px"
            type="text"
          >
            <template #prefix>¥</template>
          </el-input>
        </el-form-item>

        <el-form-item label="日期" prop="posting_date">
          <el-date-picker
            v-model="form.posting_date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>

        <el-form-item label="付款方式" prop="payment_method">
          <el-select v-model="form.payment_method" placeholder="请选择付款方式" style="width: 240px">
            <el-option label="银行转账" value="bank" />
            <el-option label="现金" value="cash" />
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>

        <el-form-item label="参考号" prop="reference_no">
          <el-input v-model="form.reference_no" placeholder="请输入参考号（可选）" maxlength="64" style="width: 360px" />
        </el-form-item>

        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注（可选）"
            maxlength="500"
            show-word-limit
            style="width: 480px"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="submitting">保存</el-button>
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createPayment, updatePayment, fetchPaymentDetail } from '@/api/modules/payment'
import type { PaymentEntry } from '@/types/models/payment'

const route = useRoute()
const router = useRouter()

const isEdit = !!route.params.id
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  payment_type: 'receive',
  counterparty_name: '',
  paid_amount: '',
  posting_date: '',
  payment_method: 'bank',
  reference_no: '',
  remark: '',
})

const rules: FormRules = {
  payment_type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  counterparty_name: [{ required: true, message: '请输入对方单位名称', trigger: 'blur' }],
  paid_amount: [
    { required: true, message: '请输入金额', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: Function) => {
        if (value && isNaN(Number(value))) {
          callback(new Error('金额必须为数字'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  posting_date: [{ required: true, message: '请选择日期', trigger: 'change' }],
  payment_method: [{ required: true, message: '请选择付款方式', trigger: 'change' }],
}

onMounted(async () => {
  // 支持路由参数 ?type=receive|pay 默认选中
  const typeParam = route.query.type as string
  if (typeParam === 'receive' || typeParam === 'pay') {
    form.payment_type = typeParam
  }

  // 编辑模式：加载已有数据
  if (isEdit) {
    try {
      const res: any = await fetchPaymentDetail(route.params.id as string)
      const data: PaymentEntry = res?.data || res
      form.payment_type = data.payment_type
      form.counterparty_name = data.counterparty_name || ''
      form.paid_amount = data.paid_amount || ''
      form.posting_date = data.posting_date || ''
      form.payment_method = data.payment_method || 'bank'
      form.reference_no = data.reference_no || ''
      form.remark = (data as any).remark || ''
    } catch (e: any) {
      ElMessage.error('加载数据失败')
    }
  }
})

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = {
      payment_type: form.payment_type,
      counterparty_name: form.counterparty_name,
      paid_amount: form.paid_amount,
      posting_date: form.posting_date,
      payment_method: form.payment_method,
      reference_no: form.reference_no || undefined,
      remark: form.remark || undefined,
    }

    if (isEdit) {
      await updatePayment(route.params.id as string, payload as Partial<PaymentEntry>)
      ElMessage.success('更新成功')
    } else {
      await createPayment({
        bank_transaction_id: '',
        payment_type: form.payment_type,
        posting_date: form.posting_date,
        remark: form.remark,
      })
      ElMessage.success('创建成功')
    }
    router.push('/payments')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || (isEdit ? '更新失败' : '创建失败'))
  } finally {
    submitting.value = false
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/payments')
  }
}
</script>

<style scoped lang="scss">
.payment-form-page {
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    h3 { font-size: 18px; }
  }
}
</style>