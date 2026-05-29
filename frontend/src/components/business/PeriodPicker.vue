<template>
  <el-date-picker
    :model-value="modelValue"
    type="month"
    :placeholder="placeholder"
    :disabled="disabled"
    :disabled-date="disabledDate"
    value-format="YYYY-MM"
    @change="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '选择会计期间',
})

defineEmits<{
  'update:modelValue': [value: string]
}>()

function disabledDate(time: Date) {
  return time.getTime() > Date.now()
}
</script>
