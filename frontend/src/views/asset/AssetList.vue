<template>
  <div class="asset-list-page">
    <div class="page-header">
      <h3>固定资产</h3>
      <div class="header-actions">
        <el-button type="primary" @click="goDetail(null)">新增资产</el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="类别">
          <el-input v-model="filter.category" placeholder="资产类别" clearable style="width: 140px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filter.depreciation_status" placeholder="全部" style="width: 130px" clearable>
            <el-option label="正常" value="normal" />
            <el-option label="已提足" value="fully_depreciated" />
            <el-option label="已处置" value="disposed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="filter.keyword" placeholder="资产名称/编号" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="assets" border stripe size="small" v-loading="loading">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="asset_code" label="资产编号" width="120" />
        <el-table-column prop="asset_name" label="资产名称" min-width="140" />
        <el-table-column prop="category" label="资产类别" width="120" />
        <el-table-column prop="purchase_date" label="购置日期" width="100" />
        <el-table-column prop="original_value" label="原值" width="120" align="right">
          <template #default="{ row }">{{ formatAmount(row.original_value) }}</template>
        </el-table-column>
        <el-table-column prop="accumulated_depreciation" label="累计折旧" width="120" align="right">
          <template #default="{ row }">{{ formatAmount(row.accumulated_depreciation) }}</template>
        </el-table-column>
        <el-table-column label="净值" width="120" align="right">
          <template #default="{ row }">
            <b>{{ formatAmount(calcNetValue(row)) }}</b>
          </template>
        </el-table-column>
        <el-table-column label="折旧状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.depreciation_status === 'normal'" type="success" size="small">正常</el-tag>
            <el-tag v-else-if="row.depreciation_status === 'fully_depreciated'" type="info" size="small">已提足</el-tag>
            <el-tag v-else type="danger" size="small">已处置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="goDetail(row.id)">详情</el-button>
            <el-button link type="warning" size="small" @click="goDepreciationRun(row)">执行折旧</el-button>
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
import { fetchAssetSummary } from '@/api/modules/asset'
import type { AssetSummary } from '@/api/modules/asset'

const router = useRouter()
const loading = ref(false)
const assets = ref<AssetSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filter = reactive({
  category: '',
  depreciation_status: '',
  keyword: '',
})

function calcNetValue(row: AssetSummary): string {
  const orig = parseFloat(row.original_value) || 0
  const accum = parseFloat(row.accumulated_depreciation) || 0
  return (orig - accum).toFixed(2)
}

function formatAmount(val: any): string {
  const n = parseFloat(val) || 0
  return '¥' + n.toLocaleString('en', { minimumFractionDigits: 2 })
}

async function loadData() {
  loading.value = true
  try {
    const res: any = await fetchAssetSummary({
      page: page.value,
      pageSize: pageSize.value,
      category: filter.category || undefined,
      depreciation_status: filter.depreciation_status || undefined,
      keyword: filter.keyword || undefined,
    })
    assets.value = res?.data?.list || res?.data || []
    total.value = res?.data?.total || assets.value.length
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '加载失败')
    assets.value = []
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.category = ''
  filter.depreciation_status = ''
  filter.keyword = ''
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}

function goDetail(id: string | null) {
  if (id) {
    router.push(`/asset/${id}`)
  } else {
    router.push('/asset/new')
  }
}

function goDepreciationRun(row: AssetSummary) {
  router.push({ path: '/depreciation/run', query: { asset_id: row.id } })
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