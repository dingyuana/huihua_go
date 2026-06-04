<template>
  <div class="opening-balance-page">
    <div class="page-header">
      <h3>期初余额</h3>
      <div class="header-actions">
        <el-button type="primary" @click="handleImport">导入</el-button>
        <el-button @click="loadData">刷新</el-button>
      </div>
    </div>

    <!-- 汇总卡片 -->
    <el-row :gutter="12" class="stat-row">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ summary.total_debit }}</p>
          <p class="stat-label">期初借方合计</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ summary.total_credit }}</p>
          <p class="stat-label">期初贷方合计</p>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <p class="stat-num">{{ summary.account_count }}</p>
          <p class="stat-label">科目数量</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" size="small">
        <el-form-item label="会计期间">
          <el-date-picker
            v-model="period"
            type="month"
            placeholder="选择月份"
            value-format="YYYY-MM"
            style="width: 140px"
          />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filter.keyword" placeholder="科目编码/名称" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表 -->
    <el-card>
      <el-table :data="balances" border stripe size="small" v-loading="loading">
        <el-table-column prop="account_code" label="科目编码" width="120" />
        <el-table-column prop="account_name" label="科目名称" min-width="160" />
        <el-table-column label="期初借方余额" width="160" align="right">
          <template #default="{ row }">
            <span v-if="editingId === row.id" class="edit-cell">
              <el-input-number
                v-model="editForm.opening_debit"
                size="small"
                :precision="2"
                :controls="false"
                style="width: 120px"
                @keyup.enter="handleSave(row)"
              />
            </span>
            <span v-else>{{ formatAmount(row.opening_debit) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="期初贷方余额" width="160" align="right">
          <template #default="{ row }">
            <span v-if="editingId === row.id" class="edit-cell">
              <el-input-number
                v-model="editForm.opening_credit"
                size="small"
                :precision="2"
                :controls="false"
                style="width: 120px"
                @keyup.enter="handleSave(row)"
              />
            </span>
            <span v-else>{{ formatAmount(row.opening_credit) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <span v-if="editingId === row.id">
              <el-button type="primary" size="small" @click="handleSave(row)">保存</el-button>
              <el-button size="small" @click="editingId = ''">取消</el-button>
            </span>
            <el-button v-else type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import {
  fetchOpeningBalances,
  fetchOpeningBalanceSummary,
  updateOpeningBalance,
  type OpeningBalance,
} from '@/api/modules/opening-balance'

const router = useRouter()

const loading = ref(false)
const balances = ref<OpeningBalance[]>([])
const editingId = ref('')
const editForm = reactive({ opening_debit: 0, opening_credit: 0 })

const period = ref(new Date().toISOString().slice(0, 7))
const filter = reactive({ keyword: '' })

const pagination = reactive({
  page: 1,
  pageSize: 50,
  total: 0,
})

const summary = reactive({
  total_debit: '0.00',
  total_credit: '0.00',
  account_count: 0,
})

function formatAmount(val: string | number): string {
  const num = parseFloat(val as string)
  return isNaN(num) ? '0.00' : num.toLocaleString('zh-CN', { minimumFractionDigits: 2 })
}

function parsePeriod(p: string) {
  const [year, month] = p.split('-').map(Number)
  return { year, period_no: month }
}

async function loadData() {
  loading.value = true
  try {
    const { year, period_no } = parsePeriod(period.value)
    const res: any = await fetchOpeningBalances({
      page: pagination.page,
      pageSize: pagination.pageSize,
      year,
      period_no,
      keyword: filter.keyword,
    })
    const data = res?.data || res
    balances.value = data?.list || data || []
    pagination.total = data?.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function loadSummary() {
  try {
    const { year, period_no } = parsePeriod(period.value)
    const res: any = await fetchOpeningBalanceSummary({ year, period_no })
    const data = res?.data || res
    if (data) {
      summary.total_debit = formatAmount(data.total_debit)
      summary.total_credit = formatAmount(data.total_credit)
      summary.account_count = data.account_count || 0
    }
  } catch (e) {
    console.error(e)
  }
}

function resetFilter() {
  filter.keyword = ''
  pagination.page = 1
  loadData()
}

function handleEdit(row: OpeningBalance) {
  editingId.value = row.id
  editForm.opening_debit = parseFloat(row.opening_debit) || 0
  editForm.opening_credit = parseFloat(row.opening_credit) || 0
}

async function handleSave(row: OpeningBalance) {
  try {
    await updateOpeningBalance(row.id, {
      opening_debit: String(editForm.opening_debit),
      opening_credit: String(editForm.opening_credit),
    })
    ElMessage.success('保存成功')
    editingId.value = ''
    loadData()
    loadSummary()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '保存失败')
  }
}

function handleImport() {
  router.push('/opening-balance/import')
}

onMounted(() => {
  loadData()
  loadSummary()
})
</script>

<style scoped lang="scss">
.opening-balance-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    h3 { margin: 0; font-size: 18px; }
  }
  .header-actions { display: flex; gap: 8px; }
}

.stat-row {
  margin-bottom: 12px;
  .stat-card { text-align: center; }
  .stat-num { font-size: 20px; font-weight: bold; margin-bottom: 4px; }
  .stat-label { font-size: 12px; color: #999; margin: 0; }
}

.filter-card { margin-bottom: 12px; }

.pagination-wrap {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.edit-cell {
  display: flex;
  align-items: center;
}
</style>