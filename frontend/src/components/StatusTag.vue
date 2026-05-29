<template>
  <n-tag :type="tagType" size="small" round>
    {{ label }}
  </n-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NTag } from 'naive-ui'
import { VOUCHER_STATUS_MAP } from '@/types'

const props = defineProps<{
  status: number
  type?: 'voucher' | 'invoice'
}>()

const statusMap = computed(() => {
  const type = props.type || 'voucher'
  return type === 'voucher' ? VOUCHER_STATUS_MAP : {
    0: { label: '草稿', color: 'default' },
    1: { label: '已审核', color: 'success' },
    2: { label: '已作废', color: 'warning' }
  }
})

const label = computed(() => statusMap.value[props.status]?.label || '未知')

const tagType = computed(() => {
  const color = statusMap.value[props.status]?.color || 'default'
  switch (color) {
    case 'success': return 'success'
    case 'warning': return 'warning'
    case 'error': return 'error'
    case 'info': return 'info'
    default: return 'default'
  }
})
</script>