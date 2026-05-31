<template>
  <div class="check-result-panel">
    <h4 v-if="title" class="panel-title">{{ title }}</h4>

    <!-- 加载中 -->
    <template v-if="loading">
      <el-skeleton :rows="4" animated />
    </template>

    <!-- 空状态 -->
    <template v-else-if="checks.length === 0">
      <el-empty description="暂无检查数据" :image-size="60" />
    </template>

    <!-- 检查结果列表 -->
    <el-table
      v-else
      :data="checks"
      border
      stripe
      size="small"
      :row-class-name="rowClass"
    >
      <el-table-column label="#" width="50">
        <template #default="{ $index }">{{ $index + 1 }}</template>
      </el-table-column>
      <el-table-column prop="name" label="检查项" min-width="160" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <CheckStatusBadge :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column prop="message" label="详情" min-width="240" />
      <el-table-column label="操作" width="130">
        <template #default="{ row }">
          <el-button
            v-if="row.action"
            link
            type="primary"
            size="small"
            @click="$emit('action', row.id)"
          >
            {{ row.action.label || '立即处理 →' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import CheckStatusBadge from './CheckStatusBadge.vue'
import type { CheckItem } from '@/types/check'

withDefaults(defineProps<{
  checks: CheckItem[]
  loading?: boolean
  title?: string
}>(), {
  loading: false,
  title: '',
})

defineEmits<{
  action: [checkId: string]
}>()

function rowClass({ row }: { row: CheckItem }) {
  if (row.status === 'blocked') return 'row-blocked'
  if (row.status === 'warning') return 'row-warning'
  if (row.status === 'passed') return 'row-passed'
  return ''
}
</script>

<style scoped lang="scss">
.panel-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}
</style>

<style lang="scss">
.el-table .row-blocked {
  background-color: #fff2f0;
}
.el-table .row-warning {
  background-color: #fffbe6;
}
.el-table .row-passed {
  background-color: #f6ffed;
}
</style>
