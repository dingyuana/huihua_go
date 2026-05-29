<template>
  <el-popover
    trigger="click"
    :width="320"
    :visible="visible"
    @show="handleShow"
  >
    <template #reference>
      <el-input
        :model-value="displayText"
        :placeholder="placeholder"
        :disabled="disabled"
        readonly
        clearable
        @clear="handleClear"
        @click="visible = true"
      >
        <template #prefix>
          <el-icon><Collection /></el-icon>
        </template>
      </el-input>
    </template>
    <div class="account-selector">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索科目名称或编码"
        size="small"
        clearable
        class="search-input"
        @input="handleSearch"
      />
      <el-tree
        ref="treeRef"
        :data="filteredTree"
        :props="{ label: 'name', children: 'children' }"
        :filter-node-method="filterNode"
        node-key="id"
        highlight-current
        default-expand-all
        @node-click="handleNodeClick"
      >
        <template #default="{ data }">
          <span :class="['node-label', { 'is-group': data.is_group, 'is-ledger': !data.is_group }]">
            <span class="node-code">{{ data.code }}</span>
            <span class="node-name">{{ data.name }}</span>
            <el-tag v-if="data.is_group" size="small" type="info" effect="plain">汇总</el-tag>
          </span>
        </template>
      </el-tree>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Account } from '@/types/models/account'
import { AccountType, RootType } from '@/types/enums'

const props = withDefaults(defineProps<{
  modelValue?: Account | string | null
  placeholder?: string
  disabled?: boolean
  ledgerOnly?: boolean
}>(), {
  placeholder: '请选择科目',
  ledgerOnly: true,
})

const emit = defineEmits<{
  'update:modelValue': [value: Account | null]
}>()

const visible = ref(false)
const treeRef = ref()
const searchKeyword = ref('')
const treeData = ref<Account[]>([])

// 过滤：Group or Ledger
const filteredTree = computed(() => {
  if (!props.ledgerOnly) return treeData.value
  return filterLedgerOnly(treeData.value)
})

function filterLedgerOnly(nodes: Account[]): Account[] {
  return nodes.map(node => ({
    ...node,
    children: node.children ? filterLedgerOnly(node.children) : undefined,
  })).filter(n => !n.is_group || (n.children && n.children.length > 0))
}

const displayText = computed(() => {
  if (!props.modelValue) return ''
  if (typeof props.modelValue === 'string') return props.modelValue
  return `${props.modelValue.code} ${props.modelValue.name}`
})

function handleShow() {
  // 加载科目树（实际项目从 API 获取）
  if (treeData.value.length === 0) {
    // 临时 Mock
    treeData.value = [
      {
        id: '1', code: '1001', name: '银行存款', account_type: AccountType.Asset, root_type: RootType.Debit,
        parent_id: null, lft: 1, rgt: 10, is_group: true, company_id: '', currency: 'CNY', created_at: '',
        children: [
          { id: '1-1', code: '1001-01', name: '银行存款-工行', account_type: AccountType.Asset, root_type: RootType.Debit, parent_id: '1', lft: 2, rgt: 3, is_group: false, company_id: '', currency: 'CNY', created_at: '' },
          { id: '1-2', code: '1001-02', name: '银行存款-建行', account_type: AccountType.Asset, root_type: RootType.Debit, parent_id: '1', lft: 4, rgt: 5, is_group: false, company_id: '', currency: 'CNY', created_at: '' },
        ],
      },
      {
        id: '2', code: '1122', name: '应收账款', account_type: AccountType.Asset, root_type: RootType.Debit,
        parent_id: null, lft: 11, rgt: 14, is_group: true, company_id: '', currency: 'CNY', created_at: '',
        children: [
          { id: '2-1', code: '1122-01', name: '应收账款-A公司', account_type: AccountType.Asset, root_type: RootType.Debit, parent_id: '2', lft: 12, rgt: 13, is_group: false, company_id: '', currency: 'CNY', created_at: '' },
        ],
      },
    ]
  }
}

function handleSearch() {
  treeRef.value?.filter(searchKeyword.value)
}

function filterNode(value: string, data: Account) {
  if (!value) return true
  return data.code.includes(value) || data.name.includes(value)
}

function handleNodeClick(data: Account) {
  if (data.is_group) {
    // Group 科目不可选
    return
  }
  emit('update:modelValue', data)
  visible.value = false
  searchKeyword.value = ''
}

function handleClear() {
  emit('update:modelValue', null)
}
</script>

<style scoped lang="scss">
.account-selector {
  .search-input {
    margin-bottom: 8px;
  }
}
.node-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;

  &.is-group {
    color: #999;
    font-style: italic;
  }
  &.is-ledger {
    color: #333;
  }

  .node-code {
    font-family: monospace;
    color: #666;
    min-width: 80px;
  }
  .node-name {
    flex: 1;
  }
}
</style>
