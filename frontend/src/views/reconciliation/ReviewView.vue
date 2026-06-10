<template>
  <div class="review-view">
    <div class="page-header"><h3>核销审批</h3></div>

    <!-- 统计 -->
    <el-row :gutter="16" class="mb-16">
      <el-col :span="6">
        <el-statistic title="待审批" :value="pendingCount" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="本月已审批通过" :value="approvedCount" />
      </el-col>
    </el-row>

    <!-- 操作栏 -->
    <el-row :gutter="8" class="toolbar mb-16">
      <el-col :span="4">
        <el-select v-model="filterStatus" placeholder="状态" clearable @change="loadList">
          <el-option label="待审批" value="pending_review" />
          <el-option label="已通过" value="executed" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
      </el-col>
      <el-col :span="6">
        <el-button type="primary" :disabled="!selectedIds.length" @click="batchApprove">
          批量审核通过 ({{ selectedIds.length }})
        </el-button>
        <el-button :disabled="!selectedIds.length" @click="batchReject">
          批量拒绝
        </el-button>
      </el-col>
      <el-col :span="2" :push="12" style="text-align:right">
        <el-button @click="loadList">刷新</el-button>
      </el-col>
    </el-row>

    <!-- 列表 -->
    <el-table :data="list" v-loading="loading" @selection-change="onSelectionChange" border size="small" highlight-current-row>
      <el-table-column type="selection" width="40" />
      <el-table-column prop="source_id" label="收款单" width="200">
        <template #default="{ row }">{{ row.source_name || row.source_id?.slice(0,8) || '-' }}</template>
      </el-table-column>
      <el-table-column prop="target_name" label="发票" width="200">
        <template #default="{ row }">{{ row.target_name || row.target_id?.slice(0,8) || '-' }}</template>
      </el-table-column>
      <el-table-column prop="amount" label="金额" width="120" align="right">
        <template #default="{ row }">¥{{ row.amount }}</template>
      </el-table-column>
      <el-table-column label="匹配等级" width="90">
        <template #default="{ row }">
          <el-tag :type="row.match_level === 'manual' ? 'warning' : 'info'" size="small">
            {{ row.match_level }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag v-if="row.status==='pending_review'" type="warning">待审批</el-tag>
          <el-tag v-else-if="row.status==='executed'" type="success">已通过</el-tag>
          <el-tag v-else-if="row.status==='rejected'" type="danger">已拒绝</el-tag>
          <el-tag v-else>{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="170">
        <template #default="{ row }">{{ row.created_at?.slice(0,16) || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending_review'">
            <el-button type="success" link size="small" @click="approveOne(row.id)">通过</el-button>
            <el-button type="danger" link size="small" @click="rejectOne(row.id)">拒绝</el-button>
          </template>
          <el-tag v-else-if="row.status==='executed'" type="success" size="small">已批准</el-tag>
          <el-tag v-else-if="row.status==='rejected'" type="danger" size="small">已退回</el-tag>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="prev, pager, next, total"
        @current-change="loadList"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const list = ref<any[]>([])
const loading = ref(false)
const selectedIds = ref<string[]>([])
const filterStatus = ref('pending_review')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const pendingCount = ref(0)
const approvedCount = ref(0)

onMounted(() => { loadList() })

async function loadList() {
  loading.value = true
  try {
    const res: any = await request.get('/reconciliation/pairs', {
      params: {
        status: filterStatus.value || undefined,
        page: page.value,
        page_size: pageSize.value,
      },
    })
    const data = res?.data ?? res ?? {}
    list.value = data.list ?? data ?? []
    total.value = data.total ?? list.value.length
    pendingCount.value = data.pending_count ?? 0
    approvedCount.value = data.approved_count ?? 0
  } catch {
    list.value = []
  }
  finally { loading.value = false }
}

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map(r => r.id)
}

async function approveOne(id: string) {
  await doApprove([id])
}
async function batchApprove() {
  if (!selectedIds.value.length) return
  await doApprove(selectedIds.value)
}
async function doApprove(ids: string[]) {
  try {
    await ElMessageBox.confirm(`确认审核通过 ${ids.length} 条核销记录？通过后将写入实际数据且不可撤回。`, '确认审批', {
      confirmButtonText: '通过',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch { return }

  try {
    const res: any = await request.post('/reconciliation/approve', { pair_ids: ids })
    if (res?.data?.executed_count > 0) {
      ElMessage.success(`已通过 ${res.data.executed_count} 条核销记录`)
    } else {
      ElMessage.warning('审批失败，请重试')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '审批失败')
  }
  loadList()
}

async function rejectOne(id: string) {
  await doReject([id])
}
async function batchReject() {
  if (!selectedIds.value.length) return
  await doReject(selectedIds.value)
}
async function doReject(ids: string[]) {
  try {
    await ElMessageBox.confirm(`确认拒绝 ${ids.length} 条核销记录？被拒绝的记录将移至人工处理。`, '确认拒绝', {
      confirmButtonText: '拒绝',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch { return }

  try {
    const res: any = await request.post('/reconciliation/reject', { pair_ids: ids })
    if (res?.data?.executed_count > 0) {
      ElMessage.success(`已拒绝 ${res.data.executed_count} 条`)
    } else {
      ElMessage.warning('操作失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '操作失败')
  }
  loadList()
}
</script>

<style scoped>
.mb-16 { margin-bottom: 16px; }
.toolbar { margin-bottom: 16px; }
.pagination-wrap { margin-top: 16px; text-align: right; }
</style>