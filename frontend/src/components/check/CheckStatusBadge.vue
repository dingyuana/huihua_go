<template>
  <el-tag
    :type="tagType"
    :size="size"
    effect="plain"
  >
    {{ label }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CheckStatusLabel, CheckStatusTagType } from '@/types/check'
import type { CheckStatus } from '@/types/check'

const props = withDefaults(defineProps<{
  status: CheckStatus
  size?: 'small' | 'default' | 'large'
}>(), {
  size: 'small',
})

const tagType = computed(() => CheckStatusTagType[props.status])
const label = computed(() => CheckStatusLabel[props.status])
</script>
