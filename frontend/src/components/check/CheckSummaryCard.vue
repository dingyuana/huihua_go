<template>
  <div class="check-summary">
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card total" v-loading="loading" element-loading-background="rgba(0,0,0,0.02)">
          <p class="stat-val">{{ summary.total }}</p>
          <p class="stat-lbl">总检查项</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card passed" v-loading="loading" element-loading-background="rgba(0,0,0,0.02)">
          <p class="stat-val passed">{{ summary.passed }}</p>
          <p class="stat-lbl">通过</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card warning" v-loading="loading" element-loading-background="rgba(0,0,0,0.02)">
          <p class="stat-val warning">{{ summary.warning }}</p>
          <p class="stat-lbl">警告</p>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card blocked" v-loading="loading" element-loading-background="rgba(0,0,0,0.02)">
          <p class="stat-val blocked">{{ summary.blocked }}</p>
          <p class="stat-lbl">阻断</p>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import type { CheckSummary } from '@/types/check'

withDefaults(defineProps<{
  summary: CheckSummary
  loading?: boolean
}>(), {
  loading: false,
})
</script>

<style scoped lang="scss">
.check-summary {
  margin-bottom: 16px;
}
.stat-card {
  text-align: center;
  padding: 4px 0;
  :deep(.el-card__body) {
    padding: 16px 8px;
  }
}
.stat-val {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 4px;
  line-height: 1.2;
  &.passed { color: #52c41a; }
  &.warning { color: #faad14; }
  &.blocked { color: #ff4d4f; }
}
.stat-lbl {
  font-size: 13px;
  color: #999;
  margin: 0;
}
</style>
