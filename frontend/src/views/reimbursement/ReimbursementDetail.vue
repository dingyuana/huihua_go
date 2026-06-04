<template>
  <div class="reimbursement-detail-page">
    <div class="page-header">
      <h3>报销单详情</h3>
      <DocStatusTag v-if="reimbursement" :docstatus="reimbursement.doc_status" size="default" />
    </div>

    <el-card v-if="reimbursement">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="申请人">{{ reimbursement.applicant_name }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ reimbursement.department_name }}</el-descriptions-item>
        <el-descriptions-item label="报销金额">
          <b>¥{{ formatAmount(reimbursement.amount) }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="说明">{{ reimbursement.description }}</el-descriptions-item>
        <el-descriptions-item label="单据状态">
          <DocStatusTag :docstatus="reimbursement.doc_status" />
        </el-descriptions-item>
        <el-descriptions-item label="凭证ID">
          <template v-if="reimbursement.voucher_id">
            <el-link type="primary" :underline="false" @click="goVoucher(reimbursement.voucher_id)">
              {{ reimbursement.voucher_id }}
            </el-link>
          </template>
          <span v-else>—</span>
        </el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ reimbursement.remark || '—' }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ reimbursement.created_by }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ reimbursement.created_at }}</el-descriptions-item>
      </el-descriptions>

      <div class="detail-actions">
        <el-button @click="$router.back()">返回</el-button>

        <!-- 草稿: 提交按钮 -->
        <template v-if="reimbursement.doc_status === 0">
          <el-button type="primary" :loading="actionLoading" @click="handleSubmit">提交</el-button>
          <el-button type="warning" :loading="actionLoading" @click="$router.push(`/expense/reimbursement/${reimbursement!.id}/edit`)">编辑</el-button>
        </template>

        <!-- 已提交: 审核按钮 -->
        <template v-if="reimbursement.doc_status === 1">
          <el-button type="success" :loading="actionLoading" @click="handleApprove">审核通过</el-button>
          <el-button type="danger" :loading="actionLoading" @click="handleReject">驳回</el-button>
        </template>

        <!-- 已审核且有凭证ID: 查看凭证 -->
        <template v-if="reimbursement.doc_status >= 2 && reimbursement.voucher_id">
          <el-button type="primary" @click="goVoucher(reimbursement.voucher_id)">查看凭证</el-button>
        </template>
      </div>
    </el-card>

    <el-empty v-else-if="!loading" description="未找到报销单数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchReimbursementDetail, submitReimbursement, approveReimbursement, rejectReimbursement } from '@/api/modules/reimbursement'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { Reimbursement } from '@/api/modules/reimbursement'

const route = useRoute()
const router = useRouter()
const reimbursement = ref<Reimbursement | null>(null)
const loading = ref(false)
const actionLoading = ref(false)

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchReimbursementDetail(route.params.id as string)
    reimbursement.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    await submitReimbursement(reimbursement.value.id)
    ElMessage.success('提交成功')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '提交失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleApprove() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    const res: any = await approveReimbursement(reimbursement.value.id)
    ElMessage.success('审核通过，凭证已生成')
    if (res?.data?.voucher_id) {
      reimbursement.value.voucher_id = res.data.voucher_id
    }
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '审核失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleReject() {
  if (!reimbursement.value) return
  actionLoading.value = true
  try {
    await rejectReimbursement(reimbursement.value.id)
    ElMessage.success('已驳回')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '驳回失败')
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
.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}
</style>