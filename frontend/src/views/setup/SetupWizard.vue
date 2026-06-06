<template>
  <div class="setup-wizard">
    <div class="wizard-header">
      <div class="header-icon">
        <span class="icon-bg">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M3 21h18M5 21V7l7-4 7 4v14M9 21v-6h6v6M9 9h1M14 9h1M9 13h1M14 13h1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </span>
      </div>
      <h2>创建公司账套</h2>
      <p class="wizard-subtitle">分步完成公司信息、会计期间和初始化设置</p>
    </div>

    <!-- 已初始化提示 -->
    <div v-if="alreadyInitialized" class="initialized-card">
      <div class="card-header">
        <div class="status-badge">
          <svg viewBox="0 0 24 24" fill="none" width="18" height="18">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span>账套已创建</span>
        </div>
        <span class="badge">已完成</span>
      </div>
      <div class="card-body">
        <div class="info-grid">
          <div class="info-item">
            <span class="label">公司名称</span>
            <span class="value">{{ savedCompany.company_name }}</span>
          </div>
          <div class="info-item">
            <span class="label">启用日期</span>
            <span class="value">{{ savedCompany.enable_date }}</span>
          </div>
          <div class="info-item">
            <span class="label">本位币</span>
            <span class="value currency-badge">{{ savedCompany.default_currency }}</span>
          </div>
          <div class="info-item">
            <span class="label">会计期间</span>
            <span class="value">{{ savedCompany.periods_count }} 个期间</span>
          </div>
        </div>
        <div class="card-actions">
          <el-button type="primary" @click="$router.push('/setup/accounts')">
            <el-icon><Setting /></el-icon>
            配置科目表
          </el-button>
          <el-button @click="$router.push('/bank/import')">
            <el-icon><Upload /></el-icon>
            导入银行流水
          </el-button>
          <el-button :loading="seeding" @click="handleSeedAccounts">
            <el-icon><Refresh /></el-icon>
            {{ seeding ? '初始化中...' : '初始化科目表' }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- Danger zone: 清空数据 (dev/test only — backend gates via cfg.App.Mode) -->
    <el-card class="danger-zone-card" shadow="never">
      <div class="danger-collapse-bar" @click="dangerExpanded = !dangerExpanded">
        <div class="danger-collapse-left">
          <el-icon class="danger-icon"><Warning /></el-icon>
          <span class="danger-title">数据管理 — 危险操作</span>
          <el-tag size="small" type="info">仅开发/测试环境</el-tag>
          <span class="danger-hint-short">点击展开销毁性操作</span>
        </div>
        <el-button text size="small">
          <el-icon>
            <component :is="dangerExpanded ? 'ArrowDown' : 'ArrowRight'" />
          </el-icon>
          {{ dangerExpanded ? '收起' : '展开' }}
        </el-button>
      </div>

      <div v-show="dangerExpanded" class="danger-expanded-body">
        <p class="danger-desc">
          以下操作不可恢复，会永久删除当前账套的数据。请确认后再执行。
        </p>
        <div class="danger-actions">
          <el-button
            type="warning"
            :loading="clearingBusiness"
            @click="handleClearBusiness"
          >
            <el-icon><Delete /></el-icon>
            清空业务数据
          </el-button>
          <el-button
            type="danger"
            :loading="clearingBasic"
            @click="handleClearBasic"
          >
            <el-icon><Delete /></el-icon>
            清空基本信息
          </el-button>
        </div>
        <p class="danger-hint">
          <strong>清空业务数据</strong>：删除发票、银行流水、收付款单、凭证、银行余额等业务流水（保留科目/客商/模板等基础设置）。<br>
          <strong>清空基本信息</strong>：删除公司信息、银行账户、科目表、客商档案、期间、模板、规则等（请先清空业务数据）。
        </p>
      </div>
    </el-card>

    <!-- 向导步骤 -->
    <el-card v-if="!alreadyInitialized" class="wizard-card" shadow="never">
      <el-steps :active="activeStep" finish-status="success" align-center class="wizard-steps">
        <el-step title="公司信息" />
        <el-step title="会计期间" />
        <el-step title="初始化设置" />
        <el-step title="完成" />
      </el-steps>

      <div v-show="activeStep === 0" class="step-content">
        <el-form ref="formRef1" :model="form" :rules="rules" label-width="100px">
          <el-form-item label="公司名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入公司全称" maxlength="100" size="large" />
          </el-form-item>
          <el-form-item label="启用日期" prop="enableDate">
            <el-date-picker v-model="form.enableDate" type="date" placeholder="选择启用日期" size="large" style="width: 100%" />
          </el-form-item>
          <el-form-item label="本位币">
            <el-select v-model="form.currency" size="large" style="width: 100%">
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
            <el-select v-model="form.fiscalStartMonth" size="large" style="width: 100%">
              <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
            </el-select>
          </el-form-item>
          <el-form-item label="期间类型">
            <el-radio-group v-model="form.periodType" size="large">
              <el-radio-button value="monthly">自然月度（12期）</el-radio-button>
              <el-radio-button value="custom">自定义</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="form.periodType === 'custom'" label="期间数">
            <el-input-number v-model="form.periodCount" :min="1" :max="12" size="large" />
          </el-form-item>
          <el-form-item label="期间预览">
            <el-table :data="previewPeriods" size="small" max-height="240" stripe border>
              <el-table-column prop="index" label="#" width="60" align="center" />
              <el-table-column prop="start" label="开始日期" align="center" />
              <el-table-column prop="end" label="结束日期" align="center" />
            </el-table>
          </el-form-item>
        </el-form>
      </div>

      <div v-show="activeStep === 2" class="step-content">
        <div class="init-features">
          <div class="feature-item">
            <div class="feature-icon green">
              <svg viewBox="0 0 24 24" fill="none" width="22" height="22">
                <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <div class="feature-text">
              <strong>导入内置科目表</strong>
              <span>小企业会计准则 · 5 大类 · 80+ 标准科目</span>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-icon blue">
              <svg viewBox="0 0 24 24" fill="none" width="22" height="22">
                <path d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <div class="feature-text">
              <strong>创建默认资金账户</strong>
              <span>自动创建"银行存款-基本户"资金账户</span>
            </div>
          </div>
        </div>
      </div>

      <div v-show="activeStep === 3" class="step-content step-done">
        <div class="done-animation">
          <div class="success-circle">
            <svg viewBox="0 0 64 64" width="64" height="64">
              <circle cx="32" cy="32" r="28" fill="#e8f5e9" stroke="#4caf50" stroke-width="2"/>
              <path d="M20 32l8 8 16-16" stroke="#4caf50" stroke-width="3" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
        </div>
        <h3 class="done-title">账套创建成功</h3>
        <p class="done-desc">{{ form.name }} 已准备就绪</p>
        <div class="done-actions">
          <el-button type="primary" size="large" @click="$router.push('/setup/accounts')">
            <el-icon><Setting /></el-icon>
            配置科目表
          </el-button>
          <el-button size="large" @click="$router.push('/bank/import')">
            <el-icon><Upload /></el-icon>
            导入银行流水
          </el-button>
        </div>
      </div>

      <div class="wizard-actions">
        <el-button v-if="activeStep > 0 && activeStep < 3" @click="prevStep">上一步</el-button>
        <div class="spacer" />
        <el-button v-if="activeStep < 2" type="primary" @click="nextStep">
          下一步
          <el-icon><ArrowRight /></el-icon>
        </el-button>
        <el-button v-if="activeStep === 2" type="primary" size="large" :loading="submitting" @click="handleSubmit">
          {{ submitting ? '创建中...' : '完成创建' }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting, Upload, ArrowRight, ArrowDown, Refresh, Delete, Warning } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import request from '@/api/request'
import { clearBusinessData, clearBasicInfo } from '@/api/modules/clearData'

const formRef1 = ref<FormInstance>()
const activeStep = ref(0)
const submitting = ref(false)
const seeding = ref(false)
const clearingBusiness = ref(false)
const clearingBasic = ref(false)
const dangerExpanded = ref(false)
const alreadyInitialized = ref(false)
const savedCompany = reactive({
  id: '',
  company_name: '',
  enable_date: '',
  default_currency: '',
  periods_count: 0,
})

const form = reactive({
  name: '',
  enableDate: null as Date | null,
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

async function loadStatus() {
  try {
    const res = await request.get('/account-setup/status')
    const data = res.data || res
    if (data.initialized && data.company) {
      alreadyInitialized.value = true
      savedCompany.id = data.company.id || ''
      savedCompany.company_name = data.company.company_name
      savedCompany.enable_date = data.company.enable_date
      savedCompany.default_currency = data.company.default_currency
      savedCompany.periods_count = data.periods_count || 0
    }
  } catch {
    // Not initialized or error, continue with empty form
  }
}

async function handleSeedAccounts() {
  if (!savedCompany.id) {
    ElMessage.warning('未找到公司信息，请刷新后重试')
    return
  }
  seeding.value = true
  try {
    await request.post('/accounts/init-seed', { company_id: savedCompany.id })
    ElMessage.success('科目表初始化完成！')
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '初始化失败'
    ElMessage.error(msg)
  } finally {
    seeding.value = false
  }
}

function summarizeResult(r: Record<string, number>): string {
  const top = Object.entries(r)
    .filter(([, n]) => n > 0)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([t, n]) => `${t}=${n}`)
    .join(', ')
  const total = Object.values(r).reduce((s, n) => s + n, 0)
  return `共 ${total} 行${top ? ` (主要: ${top})` : ''}`
}

async function handleClearBusiness() {
  try {
    await ElMessageBox.confirm(
      '此操作将清空当前账套的所有业务数据（发票/银行流水/收付款单/凭证/银行余额等），且不可恢复。确认执行？',
      '清空业务数据',
      { type: 'warning', confirmButtonText: '确认清空', cancelButtonText: '取消', confirmButtonClass: 'el-button--warning' }
    )
  } catch { return }
  clearingBusiness.value = true
  try {
    const res = await clearBusinessData()
    const data = (res as any).data || res
    ElMessage.success(`业务数据已清空 — ${summarizeResult(data.result)}`)
    await loadStatus()
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '清空失败'
    ElMessage.error(msg)
  } finally {
    clearingBusiness.value = false
  }
}

async function handleClearBasic() {
  try {
    await ElMessageBox.confirm(
      '此操作将清空当前账套的所有基本信息（公司信息/银行账户/科目表/客商档案/期间/模板/规则等），且不可恢复。\n\n请确认已先清空业务数据（否则会因外键约束失败）。\n\n确认执行？',
      '清空基本信息',
      { type: 'error', confirmButtonText: '我已确认，清空', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' }
    )
  } catch { return }
  clearingBasic.value = true
  try {
    const res = await clearBasicInfo()
    const data = (res as any).data || res
    ElMessage.success(`基本信息已清空 — ${summarizeResult(data.result)}`)
    alreadyInitialized.value = false
    await loadStatus()
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '清空失败'
    ElMessage.error(msg)
  } finally {
    clearingBasic.value = false
  }
}

onMounted(loadStatus)

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
  const valid = await formRef1.value?.validate().catch(() => false)
  if (!valid) {
    ElMessage.warning('请完善公司信息')
    activeStep.value = 0
    return
  }

  submitting.value = true
  try {
    const enableDate = form.enableDate
      ? form.enableDate.toISOString().split('T')[0]
      : ''

    const payload = {
      company_name: form.name,
      fiscal_year_start_month: form.fiscalStartMonth,
      enable_date: enableDate,
      default_currency: form.currency,
      chart_template: 'small_business',
    }
    await request.post('/account-setup/wizard', payload)
    ElMessage.success('账套创建成功！已导入标准科目表')
    activeStep.value = 3
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '创建失败，请稍后重试'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.setup-wizard {
  max-width: 760px;
  margin: 0 auto;
  padding: 40px 24px;
}

.wizard-header {
  text-align: center;
  margin-bottom: 36px;

  .header-icon {
    display: flex;
    justify-content: center;
    margin-bottom: 16px;

    .icon-bg {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 56px;
      height: 56px;
      border-radius: 16px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;

      svg {
        width: 28px;
        height: 28px;
      }
    }
  }

  h2 {
    font-size: 22px;
    font-weight: 600;
    margin: 0;
    color: #1d2129;
  }

  .wizard-subtitle {
    color: #86909c;
    font-size: 14px;
    margin-top: 8px;
    margin-bottom: 0;
  }
}

/* 已初始化卡片 */
.initialized-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  overflow: hidden;

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%);
    border-bottom: 1px solid #e8f5e9;

    .status-badge {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #2e7d32;
      font-weight: 600;
      font-size: 15px;

      svg {
        color: #4caf50;
      }
    }

    .badge {
      padding: 4px 12px;
      border-radius: 20px;
      background: #4caf50;
      color: #fff;
      font-size: 12px;
      font-weight: 500;
    }
  }

  .card-body {
    padding: 24px;

    .info-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;
      margin-bottom: 24px;

      .info-item {
        display: flex;
        flex-direction: column;
        gap: 4px;

        .label {
          font-size: 12px;
          color: #86909c;
        }

        .value {
          font-size: 14px;
          color: #1d2129;
          font-weight: 500;

          &.currency-badge {
            display: inline-flex;
            align-items: center;
            padding: 2px 10px;
            border-radius: 4px;
            background: #f0f5ff;
            color: #2f54eb;
            font-weight: 600;
            width: fit-content;
          }
        }
      }
    }

    .card-actions {
      display: flex;
      gap: 12px;
    }
  }
}

/* 向导卡片 */
.wizard-card {
  border-radius: 12px;
  border: none;

  :deep(.el-card__body) {
    padding: 28px 24px;
  }
}

.wizard-steps {
  margin-bottom: 32px;

  :deep(.el-step__title) {
    font-size: 13px;
  }
}

.step-content {
  min-height: 300px;
  padding: 8px 0;

  :deep(.el-form-item) {
    margin-bottom: 22px;
  }

  :deep(.el-input__wrapper),
  :deep(.el-select .el-input__wrapper) {
    box-shadow: 0 0 0 1px #e5e6eb inset;
    transition: all 0.2s;

    &:hover {
      box-shadow: 0 0 0 1px #c9cdd4 inset;
    }
  }
}

/* 初始化功能列表 */
.init-features {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .feature-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 18px 20px;
    border-radius: 10px;
    background: #f7f8fa;
    border: 1px solid #e5e6eb;
    transition: all 0.2s;

    &:hover {
      border-color: #c9cdd4;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    }

    .feature-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 44px;
      height: 44px;
      border-radius: 10px;
      flex-shrink: 0;

      &.green {
        background: #e8f5e9;
        color: #4caf50;
      }

      &.blue {
        background: #e8f0fe;
        color: #2f54eb;
      }
    }

    .feature-text {
      display: flex;
      flex-direction: column;
      gap: 2px;

      strong {
        font-size: 14px;
        color: #1d2129;
      }

      span {
        font-size: 12px;
        color: #86909c;
      }
    }
  }
}

/* 完成页面 */
.step-done {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 340px;
  text-align: center;

  .done-animation {
    margin-bottom: 20px;

    .success-circle {
      animation: scaleIn 0.4s ease-out;
    }
  }

  .done-title {
    font-size: 20px;
    font-weight: 600;
    color: #1d2129;
    margin: 0 0 8px;
  }

  .done-desc {
    font-size: 14px;
    color: #86909c;
    margin: 0 0 28px;
  }

  .done-actions {
    display: flex;
    gap: 16px;
  }
}

@keyframes scaleIn {
  from {
    transform: scale(0.5);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

/* 底部操作 */
.wizard-actions {
  display: flex;
  align-items: center;
  padding-top: 24px;
  margin-top: 8px;
  border-top: 1px solid #f0f0f0;

  .spacer {
    flex: 1;
  }
}
</style>
