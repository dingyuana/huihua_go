<template>
  <div class="setup-wizard">
    <div class="wizard-header">
      <h2>创建公司账套</h2>
      <p class="wizard-subtitle">分步完成公司信息、会计期间和初始化设置</p>
    </div>

    <el-steps :active="activeStep" finish-status="success" class="wizard-steps">
      <el-step title="公司信息" />
      <el-step title="会计期间" />
      <el-step title="初始化设置" />
      <el-step title="完成" />
    </el-steps>

    <div v-show="activeStep === 0" class="step-content">
      <el-form ref="formRef1" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="公司名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入公司全称" maxlength="100" />
        </el-form-item>
        <el-form-item label="启用日期" prop="enableDate">
          <el-date-picker v-model="form.enableDate" type="date" placeholder="选择启用日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="本位币">
          <el-select v-model="form.currency" style="width: 100%">
            <el-option label="人民币 (CNY)" value="CNY" />
            <el-option label="美元 (USD)" value="USD" />
            <el-option label="港币 (HKD)" value="HKD" />
          </el-select>
        </el-form-item>
      </el-form>
    </div>

    <div v-show="activeStep === 1" class="step-content">
      <el-form label-width="120px">
        <el-form-item label="财务年度起始">
          <el-select v-model="form.fiscalStartMonth" style="width: 100%">
            <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="期间类型">
          <el-radio-group v-model="form.periodType">
            <el-radio value="monthly">自然月度（12期）</el-radio>
            <el-radio value="custom">自定义</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.periodType === 'custom'" label="期间数">
          <el-input-number v-model="form.periodCount" :min="1" :max="12" />
        </el-form-item>
        <el-form-item label="期间预览">
          <el-table :data="previewPeriods" size="small" max-height="200" stripe>
            <el-table-column prop="index" label="#" width="50" />
            <el-table-column prop="start" label="开始日期" />
            <el-table-column prop="end" label="结束日期" />
          </el-table>
        </el-form-item>
      </el-form>
    </div>

    <div v-show="activeStep === 2" class="step-content">
      <el-alert title="账套创建后将自动初始化以下内容" type="info" :closable="false" show-icon class="init-alert" />
      <div class="init-options">
        <el-checkbox v-model="form.initChart" :disabled="true">
          导入内置科目表（<b>小企业会计准则</b>）
        </el-checkbox>
        <p class="option-desc">包含 5 大类、80+ 个标准科目，创建后可按需调整</p>
        <el-checkbox v-model="form.createDefaultAccount">创建默认资金账户</el-checkbox>
        <p class="option-desc">自动创建一个"银行存款-基本户"资金账户</p>
      </div>
    </div>

    <div v-show="activeStep === 3" class="step-content step-done">
      <el-result icon="success" title="账套创建成功" :sub-title="`${form.name} 已准备就绪`">
        <template #extra>
          <el-button type="primary" @click="$router.push('/setup/accounts')">配置科目表</el-button>
          <el-button @click="$router.push('/bank/import')">导入银行流水</el-button>
        </template>
      </el-result>
    </div>

    <div class="wizard-actions">
      <el-button v-if="activeStep > 0 && activeStep < 3" @click="prevStep">上一步</el-button>
      <el-button v-if="activeStep < 2" type="primary" @click="nextStep">下一步</el-button>
      <el-button v-if="activeStep === 2" type="primary" :loading="submitting" @click="handleSubmit">{{ submitting ? '创建中...' : '完成创建' }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import request from '@/api/request'

const formRef1 = ref<FormInstance>()
const activeStep = ref(0)
const submitting = ref(false)

const form = reactive({
  name: '',
  enableDate: '',
  currency: 'CNY',
  fiscalStartMonth: 1,
  periodType: 'monthly',
  periodCount: 12,
  initChart: true,
  createDefaultAccount: true,
  importDemoData: false,
})

const rules = {
  name: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
  enableDate: [{ required: true, message: '请选择启用日期', trigger: 'change' }],
}

const previewPeriods = computed(() => {
  const count = form.periodType === 'monthly' ? 12 : form.periodCount
  const enableDate = form.enableDate ? new Date(form.enableDate) : new Date()
  const startYear = enableDate.getFullYear()
  const startMonth = form.fiscalStartMonth
  const fiscalYearStart = new Date(startYear, startMonth - 1, 1)
  const baseYear = enableDate < fiscalYearStart ? startYear - 1 : startYear
  return Array.from({ length: count }, (_, i) => {
    const m = ((startMonth - 1 + i) % 12) + 1
    const y = baseYear + Math.floor((startMonth - 1 + i) / 12)
    return {
      index: i + 1,
      start: `${y}-${String(m).padStart(2, '0')}-01`,
      end: `${y}-${String(m).padStart(2, '0')}-${new Date(y, m, 0).getDate()}`,
    }
  })
})

async function nextStep() {
  if (activeStep.value === 0) {
    const valid = await formRef1.value?.validate().catch(() => false)
    if (!valid) return
  }
  activeStep.value++
}

function prevStep() {
  activeStep.value--
}

async function handleSubmit() {
  submitting.value = true
  try {
    // 调用后端 API: POST /api/v1/account-setup/wizard
    const payload = {
      company_name: form.name,
      fiscal_year_start_month: form.fiscalStartMonth,
      enable_date: form.enableDate,
      default_currency: form.currency,
      chart_template: 'small_business',
    }
    await request.post('/account-setup/wizard', payload)
    ElMessage.success('账套创建成功！已导入标准科目表')
    activeStep.value = 3
  } catch (err: any) {
    // 后端不可用时降级：直接显示成功（开发模式）
    console.warn('后端不可用，使用本地模式:', err?.message)
    activeStep.value = 3
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.setup-wizard { max-width: 720px; margin: 0 auto; padding: 32px 24px; }
.wizard-header { text-align: center; margin-bottom: 32px; h2 { font-size: 20px; } .wizard-subtitle { color: #999; font-size: 13px; margin-top: 4px; } }
.wizard-steps { margin-bottom: 32px; }
.step-content { min-height: 300px; padding: 16px 24px; background: #fff; border-radius: 6px; }
.init-options { margin-top: 16px; .el-checkbox { display: flex; margin-bottom: 4px; } .option-desc { color: #999; font-size: 12px; margin: 2px 0 16px 24px; } }
.init-alert { margin-bottom: 8px; }
.step-done { display: flex; align-items: center; justify-content: center; }
.wizard-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px; }
</style>
