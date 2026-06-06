<template>
  <div class="account-selector">
    <el-select
      v-model="level1Id"
      :placeholder="level1Placeholder"
      :disabled="disabled"
      filterable
      style="width: 33%"
      @change="handleLevel1Change"
    >
      <el-option
        v-for="node in level1Options"
        :key="node.id"
        :label="`${node.code} ${node.name}`"
        :value="node.id"
      >
        <span class="node-code">{{ node.code }}</span>
        <span class="node-name">{{ node.name }}</span>
        <el-tag v-if="node.is_group" size="small" type="info" effect="plain" style="margin-left: 8px">汇总</el-tag>
      </el-option>
    </el-select>

    <el-select
      v-model="level2Id"
      :placeholder="level2Placeholder"
      :disabled="disabled || !level1Id"
      filterable
      style="width: 33%"
      @change="handleLevel2Change"
    >
      <el-option
        v-for="node in level2Options"
        :key="node.id"
        :label="`${node.code} ${node.name}`"
        :value="node.id"
      >
        <span class="node-code">{{ node.code }}</span>
        <span class="node-name">{{ node.name }}</span>
        <el-tag v-if="node.is_group" size="small" type="info" effect="plain" style="margin-left: 8px">汇总</el-tag>
      </el-option>
    </el-select>

    <el-select
      v-model="level3Id"
      :placeholder="level3Placeholder"
      :disabled="disabled || !level2Id"
      filterable
      style="width: 33%"
      @change="handleLevel3Change"
    >
      <el-option
        v-for="node in level3Options"
        :key="node.id"
        :label="`${node.code} ${node.name}`"
        :value="node.id"
      >
        <span class="node-code">{{ node.code }}</span>
        <span class="node-name">{{ node.name }}</span>
        <el-tag v-if="node.is_group" size="small" type="info" effect="plain" style="margin-left: 8px">汇总</el-tag>
      </el-option>
    </el-select>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface AccountNode {
  id: string
  code: string
  name: string
  is_group: boolean
  parent_id?: string | null
  children?: AccountNode[]
  level: number
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

const level1Id = ref<string | null>(null)
const level2Id = ref<string | null>(null)
const level3Id = ref<string | null>(null)

const level1Placeholder = computed(() => props.placeholder || '请选择一级科目')
const level2Placeholder = computed(() => !level1Id.value ? '请先选择一级科目' : '请选择二级科目')
const level3Placeholder = computed(() => !level2Id.value ? '请先选择二级科目' : '请选择三级科目')

const level1Options = computed<AccountNode[]>(() => {
  const root = treeData.value.find(n => !n.parent_id || n.parent_id === '')
  if (!root) return []

  // Skip the "title wrapper" group (e.g., "0000 会计科目总表") which only
  // contains the real level-1 accounts. We want the actual level-1 accounts
  // (流动资产/流动负债/...) in the dropdown, not the title.
  const titleWrapper = (root.children || []).find(c =>
    c.is_group && (c.code === '0000' || /会计科目总表|全部科目/.test(c.name))
  )
  if (titleWrapper && titleWrapper.children?.length) {
    return titleWrapper.children.map(node => ({ ...node, level: 1 }))
  }

  // Fallback: original flatten logic if no title wrapper is found
  const children = root.children || []
  if (children.length > 0 && children.every(n => n.is_group)) {
    const result: AccountNode[] = []
    for (const groupNode of children) {
      if (groupNode.children) {
        for (const child of groupNode.children) {
          const extended = { ...child, level: 1, _originalGroupId: groupNode.id }
          result.push(extended)
        }
      }
    }
    return result
  }

  return children.map(node => ({ ...node, level: 1 }))
})

const level2Options = computed<AccountNode[]>(() => {
  if (!level1Id.value) return []
  
  const level1Node = findNodeById(treeData.value, level1Id.value)
  if (!level1Node) return []
  
  return (level1Node.children || []).map(node => ({ ...node, level: 2 }))
})

const level3Options = computed<AccountNode[]>(() => {
  if (!level2Id.value) return []
  const level2Node = findNodeById(treeData.value, level2Id.value)
  if (!level2Node) return []
  return (level2Node.children || []).map(node => ({ ...node, level: 3 }))
})

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

function getNodePath(node: AccountNode): AccountNode[] {
  const path: AccountNode[] = []
  let current: AccountNode | null = node
  while (current) {
    path.unshift(current)
    if (current.parent_id) {
      current = findNodeById(treeData.value, current.parent_id!)
    } else {
      current = null
    }
  }
  return path
}

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

function handleLevel1Change() {
  level2Id.value = null
  level3Id.value = null
  checkSelection()
}

function handleLevel2Change() {
  level3Id.value = null
  checkSelection()
}

function handleLevel3Change() {
  checkSelection()
}

function checkSelection() {
  // Determine the deepest selected level
  const deepestId = level3Id.value || level2Id.value || level1Id.value

  if (!deepestId) {
    emit('update:modelValue', null)
    return
  }

  const deepestNode = findNodeById(treeData.value, deepestId)
  if (!deepestNode) {
    emit('update:modelValue', null)
    return
  }

  // If the deepest selected node is a group, user hasn't finished picking yet
  // → don't emit, let them continue drilling down
  if (deepestNode.is_group) {
    return
  }

  // Leaf account selected → emit it
  emit('update:modelValue', deepestNode)
}

function applyModelValue(newVal: AccountNode | string | null) {
  if (!newVal) {
    level1Id.value = null
    level2Id.value = null
    level3Id.value = null
    return
  }

  let id: string
  if (typeof newVal === 'string') {
    id = newVal
  } else {
    id = newVal.id
  }

  const node = findNodeById(treeData.value, id)
  if (node) {
    const path = getNodePath(node)
    // level1Options is flattened to root's grandchildren when all root.children are groups
    // (see level1Options computed). So we need to find which path index corresponds
    // to a level1 option, instead of blindly using path[1].
    const l1Ids = new Set(level1Options.value.map(n => n.id))
    let l1Idx = path.findIndex(n => l1Ids.has(n.id))
    if (l1Idx === -1) l1Idx = 1 // fallback: assume original 3-level structure
    level1Id.value = path[l1Idx]?.id || null
    level2Id.value = path[l1Idx + 1]?.id || null
    level3Id.value = path[l1Idx + 2]?.id || null
  }
}

watch(() => props.modelValue, (newVal) => {
  applyModelValue(newVal ?? null)
}, { immediate: true })

// When tree data loads asynchronously after modelValue is already set,
// re-apply the modelValue so the selects can resolve the account id.
watch(treeData, () => {
  if (props.modelValue != null) {
    applyModelValue(props.modelValue ?? null)
  }
})

onMounted(() => {
  loadTree()
})
</script>

<style scoped lang="scss">
.account-selector {
  display: flex;
  gap: 8px;
  width: 100%;
}

.node-code {
  font-family: monospace;
  color: #666;
  min-width: 64px;
  margin-right: 8px;
}

.node-name {
  flex: 1;
}
</style>
