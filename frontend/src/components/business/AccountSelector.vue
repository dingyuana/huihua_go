<template>
  <el-cascader
    :model-value="cascaderValue"
    :options="cascaderOptions"
    :props="cascaderProps"
    :placeholder="placeholder"
    :disabled="disabled"
    :clearable="!disabled"
    :show-all-levels="false"
    :filterable="true"
    style="width: 100%"
    @change="handleCascaderChange"
  >
    <template #default="{ data }">
      <span class="cascader-node">
        <span class="node-code">{{ data.code }}</span>
        <span class="node-name">{{ data.name }}</span>
        <el-tag v-if="data.is_group" size="small" type="info" effect="plain">汇总</el-tag>
      </span>
    </template>
  </el-cascader>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface AccountNode {
  id: string
  code: string
  name: string
  is_group: boolean
  parent_id?: string | null
  children?: AccountNode[]
}

const props = withDefaults(defineProps<{
  modelValue?: AccountNode | string | null
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '请选择科目（一级必选，二三级可选）',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: AccountNode | null]
}>()

const treeData = ref<AccountNode[]>([])
const loading = ref(false)

const cascaderProps = {
  value: 'id',
  label: 'name',
  children: 'children',
  checkStrictly: true,
  emitPath: false,
  expandTrigger: 'hover' as const,
}

const cascaderValue = computed(() => {
  if (!props.modelValue) return null
  if (typeof props.modelValue === 'string') return props.modelValue
  return props.modelValue.id ?? null
})

const cascaderOptions = computed<AccountNode[]>(() => {
  const root = treeData.value.find(n => !n.parent_id || n.parent_id === '')
  if (!root) return treeData.value
  return (root.children || []).map(toCascaderOption)
})

function toCascaderOption(node: AccountNode): AccountNode {
  return {
    ...node,
    children: node.children && node.children.length > 0
      ? node.children.map(toCascaderOption)
      : undefined,
  }
}

const displayText = computed(() => {
  if (!props.modelValue) return ''
  if (typeof props.modelValue === 'string') return props.modelValue
  return `${props.modelValue.code} ${props.modelValue.name}`
})

async function loadTree() {
  if (treeData.value.length > 0) return
  loading.value = true
  try {
    const res: any = await request.get('/accounts/tree')
    const list = res?.data
    treeData.value = Array.isArray(list) ? list : []
  } catch (e: any) {
    ElMessage.error(e?.message || '加载科目失败')
    treeData.value = []
  } finally {
    loading.value = false
  }
}

function findNodeById(nodes: AccountNode[], id: string): AccountNode | null {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children) {
      const found = findNodeById(n.children, id)
      if (found) return found
    }
  }
  return null
}

function handleCascaderChange(id: string | null) {
  if (!id) {
    emit('update:modelValue', null)
    return
  }
  const node = findNodeById(treeData.value, id)
  if (!node) {
    emit('update:modelValue', null)
    return
  }
  if (node.is_group) {
    ElMessage.warning('一级汇总科目不可直接入账，请选择二级或三级明细科目')
    return
  }
  emit('update:modelValue', node)
}

onMounted(() => {
  loadTree()
})
</script>

<style scoped lang="scss">
.cascader-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;

  .node-code {
    font-family: monospace;
    color: #666;
    min-width: 64px;
  }
  .node-name {
    flex: 1;
  }
}
</style>
