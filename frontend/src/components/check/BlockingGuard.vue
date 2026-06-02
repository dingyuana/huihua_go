<template>
  <el-tooltip
    v-if="blocked"
    content="尚有 {{ blockedCount }} 项阻断待处理"
    placement="top"
  >
    <el-button
      type="primary"
      disabled
    >
      请先修复阻断项
    </el-button>
  </el-tooltip>
  <slot v-else />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  blocked: boolean
  blockedCount?: number
}>(), {
  blockedCount: 0,
})
</script>
