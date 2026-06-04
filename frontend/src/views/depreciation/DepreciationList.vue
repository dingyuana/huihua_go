<template>
  <div class="depreciation-list-page">
    <div class="page-header">
      <h3>折旧/摊销执行记录</h3>
      <div class="header-actions">
        <el-button type="primary" @click="goRun">执行折旧</el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="会计期间">
          <el-input v-model="filter.period" placeholder="如 202506" clearable style="width: 120px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.doc_status" placeholder="全部" style="width: 130px" clearable>
            <el-option label="草稿" :value="0" />
            <el-option label="已提交" :value="1" />
            <el-option label="已审核" :value="2" />
            <el-option label="已过账" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="runs" border stripe size="small" v-loading="loading">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="period" label="会计期间" width="100" />
        <el-table-column prop="asset_name" label="资产名称" min-width="140" />
        <el-table-column prop="depreciation_amount" label="折旧金额" width="120" align="right">
          <template #default="{ row }">{{ formatAmount(row.depreciation_amount) }}</template>
        </el-table-column>
        <el-table-column prop="run_date" label="执行日期" width="100" />
        <el-table-column label="凭证号" width="120">
          <template #default="{ row }">
            <span v-if="row.voucher_no">{{ row.voucher_no }}</span>
            <span v-else style="color:#999">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <DocStatusTag :docstatus="row.doc_status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleGenerateVoucher(row)" :disabled="row.doc_status !== 0">
              生成凭证
            </el-button>
            <el-button v-if="row.voucher_id" link type="primary" size="small" @click="goVoucher(row.voucher_id)">
              查看凭证
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > 0"
        style="margin-top: 12px; justify-content: flex-end"
        background
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        layout="total, prev, pager, next"
        @current-change="onPageChange"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchDepreciationRuns, generateDepreciationVoucher } from '@/api/modules/asset'
import DocStatusTag from '@/components/business/DocStatusTag.vue'
import type { DepreciationRun } from '@/api/modules/asset'

const router = useRouter()
const loading = ref(false)
const runs = ref<DepreciationRun[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive({
  period: '',
  doc_status: undefined as number | undefined,
})

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchDepreciationRuns({
      page: page.value,
      pageSize: pageSize.value,
      period: filter.period || undefined,
      doc_status: filter.doc_status,
    })
    runs.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total || runs.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    runs.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.period = ''
  filter.doc_status = undefined
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}

function goRun() {
  router.push('/depreciation/run')
}

async function handleGenerateVoucher(row: DepreciationRun) {
  try {
    const res: any = await generateDepreciationVoucher({ run_ids: [row.id] })
    ElMessage.success('凭证已生成')
    await loadData()
    if (res?.data?.voucher_id) {
      router.push(`/vouchers/${res.data.voucher_id}`)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '生成失败')
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
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.filter-card { margin-bottom: 12px; }
</style>