<template>
  <el-tag :type="tagType" :size="size" effect="plain">
    {{ label }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { DocStatus, DocStatusLabel } from '@/types/enums'

const props = withDefaults(defineProps<{
  docstatus?: number
  size?: 'small' | 'default' | 'large'
}>(), {
  size: 'small',
})

const tagType = computed(() => {
  switch (props.docstatus) {
    case DocStatus.Draft: return 'info'
    case DocStatus.Submitted: return 'primary'
    case DocStatus.Cancelled: return 'danger'
    default: return 'warning'
  }
})

const label = computed(() => {
  if (props.docstatus === undefined) return '待处理'
  return DocStatusLabel[props.docstatus as DocStatus] || '未知'
})
</script>
