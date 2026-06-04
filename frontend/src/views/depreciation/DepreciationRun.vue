<template>
  <div class="depreciation-run-page">
    <div class="page-header">
      <h3>执行折旧</h3>
      <div class="header-actions">
        <el-button @click="router.back()">返回</el-button>
      </div>
    </div>

    <!-- 执行表单 -->
    <el-card class="run-form-card" style="margin-bottom: 12px">
      <template #header>折旧执行</template>
      <el-form :inline="true" size="small" label-width="100px">
        <el-form-item label="资产">
          <el-select
            v-model="selectedAssetId"
            placeholder="请选择资产（可选）"
            clearable
            filterable
            style="width: 260px"
          >
            <el-option
              v-for="a in assetOptions"
              :key="a.id"
              :label="a.asset_name"
              :value="a.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="会计期间" required>
          <el-input v-model="period" placeholder="如 202506" style="width: 120px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="running" @click="handleRun">执行折旧</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="runResult" type="success" :closable="false" style="margin-top: 8px">
        折旧执行完成，共 {{ runResult.runs?.length || 0 }} 条记录
      </el-alert>
    </el-card>

    <!-- 执行结果 -->
    <el-card v-if="runResult && runResult.runs?.length">
      <template #header>执行结果</template>
      <el-table :data="runResult.runs" border stripe size="small">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="asset_name" label="资产名称" min-width="140" />
        <el-table-column prop="period" label="会计期间" width="100" />
        <el-table-column prop="depreciation_amount" label="折旧金额" width="130" align="right">
          <template #default="{ row }">{{ formatAmount(row.depreciation_amount) }}</template>
        </el-table-column>
        <el-table-column prop="run_date" label="执行日期" width="100" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.doc_status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              :disabled="row.doc_status !== 0"
              @click="handleGenerateVoucher(row)"
            >
              生成凭证
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchAssetSummary, runDepreciation, generateDepreciationVoucher } from '@/api/modules/asset'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { AssetSummary, DepreciationRun } from '@/api/modules/asset'

const router = useRouter()
const route = useRoute()
const running = ref(false)
const selectedAssetId = ref<string | undefined>(route.query.asset_id as string | undefined)
const period = ref('')
const assetOptions = ref<AssetSummary[]>([])
const runResult = ref<{ runs: DepreciationRun[] } | null>(null)

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadAssets() {
  try {
    const res: any = await fetchAssetSummary({ pageSize: 500 })
    assetOptions.value = res?.data?.list || res?.data || []
  } catch {
    // ignore
  }
}

async function handleRun() {
  if (!period.value) {
    ElMessage.warning('请输入会计期间')
    return
  }
  running.value = true
  try {
    const res: any = await runDepreciation({
      asset_id: selectedAssetId.value,
      period: period.value,
    })
    runResult.value = res?.data || res
    ElMessage.success('折旧执行成功')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '执行失败')
  } finally {
    running.value = false
  }
}

async function handleGenerateVoucher(row: DepreciationRun) {
  try {
    const res: any = await generateDepreciationVoucher({ run_ids: [row.id] })
    ElMessage.success('凭证已生成')
    if (res?.data?.voucher_id) {
      router.push(`/vouchers/${res.data.voucher_id}`)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '生成失败')
  }
}

onMounted(loadAssets)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.run-form-card { margin-bottom: 12px; }
</style>