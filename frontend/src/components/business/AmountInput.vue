<template>
  <el-input
    :model-value="formattedValue"
    :placeholder="placeholder"
    :disabled="disabled"
    @input="handleInput"
    @blur="handleBlur"
  >
    <template #prefix>
      <el-tag size="small" type="info" effect="plain">{{ currencySymbol }}</el-tag>
    </template>
    <template #suffix>
      <span v-if="showBalance && balanceText" class="balance-text">
        余额: {{ balanceText }}
      </span>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  disabled?: boolean
  currency?: string
  max?: string
  showBalance?: boolean
  balanceText?: string
}>(), {
  placeholder: '请输入金额',
  currency: 'CNY',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const rawValue = ref(props.modelValue || '')

const currencySymbol = computed(() => {
  const map: Record<string, string> = { CNY: '¥', USD: '$', EUR: '€', HKD: 'HK$' }
  return map[props.currency] || props.currency
})

const formattedValue = computed(() => {
  if (!rawValue.value) return ''
  const num = rawValue.value.replace(/,/g, '')
  if (!/^\d+(\.\d{0,2})?$/.test(num)) return rawValue.value
  const parts = num.split('.')
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return parts.join('.')
})

function handleInput(val: string) {
  // 只允许数字、小数点、负号
  const cleaned = val.replace(/[^\d.-]/g, '')
  rawValue.value = cleaned
  emit('update:modelValue', cleaned)
}

function handleBlur() {
  // 超过 max 时提示
  if (props.max && rawValue.value) {
    const num = parseFloat(rawValue.value)
    const maxNum = parseFloat(props.max)
    if (num > maxNum) {
      // Element Plus 会自动显示验证样式
    }
  }
}
</script>

<style scoped lang="scss">
.balance-text {
  font-size: 12px;
  color: #999;
}
</style>
