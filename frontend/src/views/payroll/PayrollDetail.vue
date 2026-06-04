<template>
  <div class="payroll-detail-page">
    <div class="page-header">
      <h3>工资单详情</h3>
      <DocStatusTag v-if="payroll" :docstatus="payroll.doc_status" size="default" />
    </div>

    <el-card v-if="payroll">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="员工姓名">{{ payroll.employee_name }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ payroll.department_name }}</el-descriptions-item>
        <el-descriptions-item label="工资期间">
          {{ String(payroll.period_no).slice(0, 4) }}-{{ String(payroll.period_no).slice(4, 6) }}
        </el-descriptions-item>
        <el-descriptions-item label="发放日期">{{ payroll.payment_date }}</el-descriptions-item>
        <el-descriptions-item label="应发工资">
          <b>¥{{ formatAmount(payroll.gross_salary) }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="代扣个税">¥{{ formatAmount(payroll.individual_tax) }}</el-descriptions-item>
        <el-descriptions-item label="代扣社保">¥{{ formatAmount(payroll.social_security) }}</el-descriptions-item>
        <el-descriptions-item label="代扣公积金">¥{{ formatAmount(payroll.housing_fund) }}</el-descriptions-item>
        <el-descriptions-item label="其他扣款">¥{{ formatAmount(payroll.other_deductions) }}</el-descriptions-item>
        <el-descriptions-item label="银行卡号">{{ payroll.bank_account_no || '—' }}</el-descriptions-item>
        <el-descriptions-item label="实发工资">
          <b class="net-salary">¥{{ formatAmount(payroll.net_salary) }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="单据状态">
          <DocStatusTag :docstatus="payroll.doc_status" />
        </el-descriptions-item>
        <el-descriptions-item label="凭证ID" :span="2">
          <template v-if="payroll.voucher_id">
            <el-link type="primary" :underline="false" @click="goVoucher(payroll.voucher_id)">{{ payroll.voucher_id }}</el-link>
          </template>
          <span v-else>—</span>
        </el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ payroll.remark || '—' }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ payroll.created_by }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ payroll.created_at }}</el-descriptions-item>
      </el-descriptions>

      <!-- 工资计算明细 -->
      <el-divider content-position="left">工资计算明细</el-divider>
      <el-table :data="salaryBreakdown" size="small" border>
        <el-table-column prop="item" label="项目" width="200" />
        <el-table-column prop="amount" label="金额（元）" width="160" align="right" />
        <el-table-column prop="remark" label="说明" />
      </el-table>

      <div class="detail-actions">
        <el-button @click="$router.back()">返回</el-button>

        <!-- 草稿: 提交按钮 -->
        <template v-if="payroll.doc_status === 0">
          <el-button type="primary" :loading="actionLoading" @click="handleSubmit">提交</el-button>
          <el-button type="warning" :loading="actionLoading" @click="$router.push(`/payroll/${payroll!.id}/edit`)">编辑</el-button>
        </template>

        <!-- 已提交: 审核按钮 -->
        <template v-if="payroll.doc_status === 1">
          <el-button type="danger" :loading="actionLoading" @click="handleApprove">审核通过</el-button>
        </template>

        <!-- 已审核且有凭证ID: 查看凭证 -->
        <template v-if="payroll.doc_status >= 2 && payroll.voucher_id">
          <el-button type="primary" @click="goVoucher(payroll.voucher_id)">查看凭证</el-button>
        </template>

        <!-- 已审核但无凭证ID: 独立制证 -->
        <template v-if="payroll.doc_status >= 2 && !payroll.voucher_id">
          <el-button type="success" :loading="actionLoading" @click="handleGenerateVoucher">生成凭证</el-button>
        </template>
      </div>
    </el-card>

    <el-empty v-else-if="!loading" description="未找到工资单数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchPayrollDetail, submitPayroll, approvePayroll, generateVoucherFromPayroll } from '@/api/modules/payroll'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { Payroll } from '@/types/models/payroll'

const route = useRoute()
const router = useRouter()
const payroll = ref<Payroll | null>(null)
const loading = ref(false)
const actionLoading = ref(false)

const salaryBreakdown = computed(() => {
  if (!payroll.value) return []
  const p = payroll.value
  return [
    { item: '应发工资', amount: `¥${formatAmount(p.gross_salary)}`, remark: '基本工资 + 绩效 + 奖金等' },
    { item: '代扣个人所得税', amount: `-¥${formatAmount(p.individual_tax)}`, remark: '月度代扣个人所得税' },
    { item: '代扣社保', amount: `-¥${formatAmount(p.social_security)}`, remark: '养老保险、医疗保险等' },
    { item: '代扣公积金', amount: `-¥${formatAmount(p.housing_fund)}`, remark: '住房公积金' },
    { item: '其他扣款', amount: `-¥${formatAmount(p.other_deductions)}`, remark: p.remark || '其他扣除项' },
    { item: '实发工资', amount: `¥${formatAmount(p.net_salary)}`, remark: '应发 - 所有代扣 = 实发' },
  ]
})

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchPayrollDetail(route.params.id as string)
    payroll.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!payroll.value) return
  actionLoading.value = true
  try {
    await submitPayroll(payroll.value.id)
    ElMessage.success('提交成功')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '提交失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleApprove() {
  if (!payroll.value) return
  actionLoading.value = true
  try {
    const res: any = await approvePayroll(payroll.value.id)
    ElMessage.success('审核通过，凭证已生成')
    if (res?.data?.voucher_id) {
      payroll.value.voucher_id = res.data.voucher_id
    }
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '审核失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleGenerateVoucher() {
  if (!payroll.value) return
  actionLoading.value = true
  try {
    const res: any = await generateVoucherFromPayroll(payroll.value.id)
    ElMessage.success('凭证已生成')
    if (res?.data?.voucher_id) {
      payroll.value.voucher_id = res.data.voucher_id
    }
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '制证失败')
  } finally {
    actionLoading.value = false
  }
}

function goVoucher(voucherId: string) {
  router.push(`/vouchers/${voucherId}`)
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.net-salary {
  font-size: 18px;
  color: #389e0d;
}
.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}
</style>