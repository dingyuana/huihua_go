<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    filterable
    remote
    :remote-method="handleSearch"
    :loading="loading"
    clearable
    @change="handleChange"
    @clear="$emit('update:modelValue', null)"
  >
    <el-option
      v-for="item in options"
      :key="item.id"
      :label="`${item.name} (${item.tax_id})`"
      :value="item"
    >
      <div class="party-option">
        <span class="party-name">{{ item.name }}</span>
        <span class="party-tax">{{ item.tax_id }}</span>
        <span v-if="item.bank_name" class="party-bank">{{ item.bank_name }}</span>
      </div>
    </el-option>
  </el-select>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Party } from '@/types/models/party'
import { PartyType } from '@/types/enums'

withDefaults(defineProps<{
  modelValue?: Party | string | null
  placeholder?: string
  disabled?: boolean
  partyType?: string
}>(), {
  placeholder: '搜索客商（输入2字符以上）',
  partyType: 'both',
})

const emit = defineEmits<{
  'update:modelValue': [value: Party | null]
}>()

const loading = ref(false)
const options = ref<Party[]>([])

function handleSearch(query: string) {
  if (query.length < 2) return
  loading.value = true
  // 实际从 API 搜索
  setTimeout(() => {
    options.value = []
    loading.value = false
  }, 300)
}

function handleChange(val: Party | string) {
  emit('update:modelValue', typeof val === 'string' ? null : val)
}
</script>

<style scoped lang="scss">
.party-option {
  display: flex;
  gap: 12px;
  font-size: 13px;
  .party-name { flex: 1; }
  .party-tax { color: #999; font-family: monospace; }
  .party-bank { color: #999; }
}
</style>
