<template>
  <div class="match-view">
    <div class="page-header"><h3>智能匹配推荐</h3></div>

    <!-- 选择期间 + 运行 -->
    <el-row :gutter="16" class="toolbar">
      <el-col :span="4">
        <el-input-number v-model="periodYear" :min="2020" :max="2030" controls-position="right" style="width:100%" />
      </el-col>
      <el-col :span="3">
        <el-select v-model="periodMonth" style="width:100%">
          <el-option v-for="m in 12" :key="m" :label="m + '月'" :value="m" />
        </el-select>
      </el-col>
      <el-col :span="10">
        <el-button type="primary" :loading="running" @click="runMatch">运行自动匹配</el-button>
        <el-button :disabled="matchedPairs.length === 0" @click="confirmAll">确认全部</el-button>
        <el-button type="success" :disabled="confirmedPairs.length === 0" :loading="executing" @click="executeAll">执行核销</el-button>
      </el-col>
    </el-row>

    <!-- 匹配结果概览 -->
    <el-row v-if="result" :gutter="16" class="summary-row">
      <el-col :span="6"><el-statistic title="扫描总数" :value="result.total_scanned" /></el-col>
      <el-col :span="6"><el-statistic title="已匹配" :value="result.matched" /></el-col>
      <el-col :span="6"><el-statistic title="未匹配" :value="result.unmatched" /></el-col>
      <el-col :span="6"><el-statistic title="已确认" :value="confirmedPairs.length" /></el-col>
    </el-row>

    <!-- 匹配结果按等级展示 -->
    <div v-for="grp in matchGroups" :key="grp.level" class="match-level" :class="'level-' + grp.levelClass">
      <h4>{{ grp.title }}</h4>
      <el-table :data="grp.items" size="small" border style="margin-top:8px">
        <el-table-column prop="match_level" label="等级" width="60" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'confirmed'" size="small" type="success">已确认</el-tag>
            <el-tag v-else-if="row.status === 'matched'" size="small" type="warning">待确认</el-tag>
            <el-tag v-else size="small" type="info">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button v-if="row.status === 'matched'" size="small" type="primary" @click="confirmPair(row)">确认</el-button>
            <el-tag v-else-if="row.status === 'confirmed'" size="small" type="success">✅</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 未匹配项 -->
    <el-card v-if="result && result.unmatched_items && result.unmatched_items.length" shadow="never" class="unmatched-card">
      <template #header>未匹配项</template>
      <el-table :data="result.unmatched_items" size="small" border>
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="party_name" label="对方" width="150" />
        <el-table-column prop="summary" label="摘要" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const periodYear = ref(new Date().getFullYear())
const periodMonth = ref(new Date().getMonth() + 1)

const running = ref(false)
const executing = ref(false)
const result = ref<any>(null)
const allPairs = ref<any[]>([])

interface MatchItem {
  pairId: string
  match_level: string
  amount: string
  status: string
}

const levelMap: Record<string, string> = {
  L1: 'L1 精确匹配',
  L2: 'L2 发票号匹配',
  L3: 'L3 客商+金额+日期',
  L4: 'L4 金额匹配',
  L5: 'L5 容差匹配',
}
const levelClassMap: Record<string, string> = {
  L1: '1', L2: '2', L3: '3', L4: '4', L5: '5',
}

const matchGroups = computed(() => {
  const groups: { level: string; levelClass: string; title: string; items: MatchItem[] }[] = []
  for (const lvl of ['L1', 'L2', 'L3', 'L4', 'L5']) {
    const items = allPairs.value.filter((p: any) => (p.match_level || '').toUpperCase() === lvl)
    if (items.length) {
      groups.push({
        level: lvl,
        levelClass: levelClassMap[lvl] || lvl,
        title: levelMap[lvl] || lvl,
        items,
      })
    }
  }
  return groups
})

const matchedPairs = computed(() => allPairs.value.filter((p: any) => p.status === 'matched'))
const confirmedPairs = computed(() => allPairs.value.filter((p: any) => p.status === 'confirmed'))

async function runMatch() {
  running.value = true
  result.value = null
  allPairs.value = []
  try {
    const res: any = await request.post('/reconciliation/run', null, {
      params: { period_no: periodYear.value * 100 + periodMonth.value },
    })
    const data = res?.data || res
    result.value = data
    allPairs.value = (data?.pairs || []).map((p: any) => ({
      pairId: p.id,
      match_level: p.match_level || 'L3',
      amount: p.amount || '0.00',
      status: p.status || 'matched',
    }))
  } catch (e: any) {
    ElMessage.error(e?.message || '自动匹配失败')
  } finally {
    running.value = false
  }
}

async function confirmPair(row: MatchItem) {
  try {
    await request.post(`/reconciliation/pairs/${row.pairId}/confirm`)
    row.status = 'confirmed'
    ElMessage.success('已确认')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || e?.message || '确认失败')
  }
}

async function confirmAll() {
  const all = matchedPairs.value
  for (const row of all) {
    try {
      await request.post(`/reconciliation/pairs/${row.pairId}/confirm`)
      row.status = 'confirmed'
    } catch { /* skip */ }
  }
  ElMessage.success(`已确认 ${all.length} 条`)
}

async function executeAll() {
  const all = confirmedPairs.value
  if (!all.length) return
  executing.value = true
  try {
    const pairIds = all.map(p => p.pairId)
    await request.post('/reconciliation/execute', { pair_ids: pairIds })
    ElMessage.success(`已提交审核，共 ${pairIds.length} 条，等待审批`)
    allPairs.value = []
    result.value = null
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || e?.message || '核销执行失败')
  } finally {
    executing.value = false
  }
}
</script>

<style scoped>
.page-header h3 { font-size: 18px; margin-bottom: 16px; }
.toolbar { margin-bottom: 16px; }
.summary-row { margin-bottom: 16px; }
.match-level { margin-bottom: 20px; padding: 16px; border: 1px solid #e8e8e8; border-radius: 6px; }
.match-level h4 { margin-bottom: 8px; font-size: 14px; }
.level-1 { border-left: 3px solid #52c41a; }
.level-1 h4 { color: #52c41a; }
.level-2 { border-left: 3px solid #1890ff; }
.level-2 h4 { color: #1890ff; }
.level-3 { border-left: 3px solid #faad14; }
.level-3 h4 { color: #faad14; }
.level-4 { border-left: 3px solid #909399; }
.level-4 h4 { color: #909399; }
.unmatched-card { margin-top: 8px; }
</style>
