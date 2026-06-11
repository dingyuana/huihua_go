<template>
  <div class="subject-tree">
    <PageLayout title="科目管理">
      <template #actions>
        <el-button type="primary" @click="openCreateDialog()">
          <el-icon><Plus /></el-icon>
          新增根科目
        </el-button>
      </template>

      <div class="subject-layout">
        <!-- Left Panel -->
        <div class="left-panel">
          <div class="left-toolbar">
            <el-input v-model="searchKeyword" placeholder="搜索科目" clearable size="small" @input="onSearch" />
          </div>
          <el-tree
            ref="treeRef"
            :data="filteredTreeData"
            node-key="id"
            highlight-current
            default-expand-all
            :props="{ children: 'children', label: 'name' }"
            @node-click="onNodeClick"
            @node-contextmenu="onContextMenu"
          >
            <template #default="{ data }">
              <span class="tree-node-label">
                <span>{{ data.code }} - {{ data.name }}</span>
                <el-tag v-if="data.is_group" size="small" type="info" effect="plain" style="margin-left: 6px">汇总</el-tag>
              </span>
            </template>
          </el-tree>
        </div>

        <!-- Right Panel -->
        <div class="right-panel">
          <div v-if="!selectedAccount" class="detail-placeholder">
            <el-empty description="请从左侧选择一个科目" />
          </div>
          <div v-else class="detail-form">
            <div class="detail-header">
              <span class="detail-title">科目详情</span>
              <el-button size="small" @click="openEditDialog(selectedAccount)">编辑</el-button>
            </div>
            <el-form label-width="100px" size="small">
              <el-form-item label="科目编码">
                <el-input v-model="selectedAccount.code" readonly />
              </el-form-item>
              <el-form-item label="科目名称">
                <el-input v-model="selectedAccount.name" />
              </el-form-item>
              <el-form-item label="科目类型">
                <el-select v-model="selectedAccount.account_type" @change="onDetailTypeChange">
                  <el-option label="资产" value="asset" />
                  <el-option label="负债" value="liability" />
                  <el-option label="共同" value="equity" />
                  <el-option label="成本" value="cost" />
                  <el-option label="损益" value="income" />
                </el-select>
              </el-form-item>
              <el-form-item label="余额方向">
                <el-tag :type="selectedAccount.root_type === 'debit' ? 'danger' : 'success'">
                  {{ selectedAccount.root_type === 'debit' ? '借方' : '贷方' }}
                </el-tag>
              </el-form-item>
              <el-form-item label="汇总科目">
                <el-switch v-model="selectedAccount.is_group" />
              </el-form-item>
              <el-form-item label="币种">
                <el-select v-model="selectedAccount.currency">
                  <el-option label="CNY" value="CNY" />
                  <el-option label="USD" value="USD" />
                  <el-option label="HKD" value="HKD" />
                </el-select>
              </el-form-item>
            </el-form>
            <div class="detail-actions">
              <el-button type="primary" @click="handleSaveDetail">保存修改</el-button>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>

    <!-- Context Menu -->
    <div
      v-if="contextMenuVisible"
      class="context-menu"
      :style="{ top: contextMenuY + 'px', left: contextMenuX + 'px' }"
      @click.stop
    >
      <div class="context-menu-item" @click="onContextMenuAction('create')">新增下级</div>
      <div class="context-menu-item" @click="onContextMenuAction('edit')">编辑</div>
      <div class="context-menu-item danger" @click="onContextMenuAction('delete')">删除</div>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEditing ? '编辑科目' : '新增科目'" width="520px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="dialogForm" :rules="dialogRules" label-width="100px">
        <el-form-item label="父科目">
          <el-tree-select
            v-model="dialogForm.parentId"
            :data="allAccountsTree"
            :props="{ children: 'children', label: 'name', value: 'id' }"
            check-strictly
            placeholder="根科目（不选）"
            clearable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="科目编码">
          <el-input v-model="dialogForm.code" :disabled="isEditing" style="width: 200px">
            <template #append>
              <el-button @click="previewAutoCode">自动生成</el-button>
            </template>
          </el-input>
          <span v-if="previewCode" class="code-hint">建议: {{ previewCode }}</span>
        </el-form-item>
        <el-form-item label="科目名称" prop="name">
          <el-input v-model="dialogForm.name" maxlength="100" />
        </el-form-item>
        <el-form-item label="科目类型">
          <el-select v-model="dialogForm.accountType" style="width: 100%" @change="onTypeChange">
            <el-option label="资产" value="asset" />
            <el-option label="负债" value="liability" />
            <el-option label="共同" value="equity" />
            <el-option label="成本" value="cost" />
            <el-option label="损益" value="income" />
          </el-select>
        </el-form-item>
        <el-form-item label="余额方向">
          <el-tag :type="dialogForm.balanceDirection === 'debit' ? 'danger' : 'success'">
            {{ dialogForm.balanceDirection === 'debit' ? '借方' : '贷方' }}
          </el-tag>
          <span class="form-hint">根据科目类型自动确定: 资产/成本类→借方, 负债/权益/损益类→贷方</span>
        </el-form-item>
        <el-form-item label="汇总科目">
          <el-switch v-model="dialogForm.isGroup" />
          <span class="form-hint">汇总科目不可用于记账，仅用于组织科目层级</span>
        </el-form-item>
        <el-form-item label="币种">
          <el-select v-model="dialogForm.currency" style="width: 120px">
            <el-option label="CNY" value="CNY" />
            <el-option label="USD" value="USD" />
            <el-option label="HKD" value="HKD" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import type { Account } from '@/types/models/account'
import { fetchAccountTree, createAccount, updateAccount, deleteAccount } from '@/api/modules/account'
import PageLayout from '@/components/app/PageLayout.vue'

const treeRef = ref()
const allAccounts = ref<Account[]>([])
const searchKeyword = ref('')
const selectedAccount = ref<Account | null>(null)

// Context menu state
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuAccount = ref<Account | null>(null)

// Dialog state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const previewCode = ref('')
const formRef = ref<FormInstance>()
const saving = ref(false)

const dialogForm = reactive({
  parentId: null as string | null,
  code: '',
  name: '',
  accountType: 'asset',
  balanceDirection: 'debit' as string,
  isGroup: false,
  currency: 'CNY',
})

const dialogRules = {
  name: [{ required: true, message: '请输入科目名称', trigger: 'blur' }],
}

const allAccountsTree = computed(() => allAccounts.value)

const filteredTreeData = computed(() => {
  if (!searchKeyword.value) {
    return allAccounts.value
  }
  return deepFilter(allAccounts.value, (n: Account) =>
    n.code.includes(searchKeyword.value) || n.name.includes(searchKeyword.value)
  )
})

function deepFilter(nodes: Account[], predicate: (n: Account) => boolean): Account[] {
  return nodes.reduce<Account[]>((acc, node) => {
    const children = node.children ? deepFilter(node.children, predicate) : undefined
    if (predicate(node) || (children && children.length > 0)) {
      acc.push({ ...node, children })
    }
    return acc
  }, [])
}

function rootTypeByAccountType(type: string): string {
  if (['asset', 'cost'].includes(type)) return '借方'
  return '贷方'
}

function onSearch() {
  // Trigger filteredTreeData recompute
}

function onNodeClick(data: Account) {
  selectedAccount.value = { ...data }
}

function onContextMenu(event: MouseEvent, data: Account) {
  event.preventDefault()
  contextMenuAccount.value = data
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

function onContextMenuAction(action: 'create' | 'edit' | 'delete') {
  if (!contextMenuAccount.value) return
  if (action === 'create') {
    openCreateDialog(contextMenuAccount.value)
  } else if (action === 'edit') {
    openEditDialog(contextMenuAccount.value)
  } else if (action === 'delete') {
    handleDelete(contextMenuAccount.value)
  }
  contextMenuVisible.value = false
}

function hideContextMenu() {
  contextMenuVisible.value = false
}

function onTypeChange() {
  dialogForm.balanceDirection = ['asset', 'cost'].includes(dialogForm.accountType) ? 'debit' : 'credit'
}

function openCreateDialog(parent?: Account) {
  isEditing.value = false
  editingId.value = ''
  previewCode.value = ''
  dialogForm.parentId = parent?.id || null
  dialogForm.code = ''
  dialogForm.name = ''
  dialogForm.accountType = parent?.account_type || 'asset'
  dialogForm.balanceDirection = rootTypeByAccountType(dialogForm.accountType) === '借方' ? 'debit' : 'credit'
  dialogForm.isGroup = false
  dialogForm.currency = 'CNY'
  dialogVisible.value = true
}

function openEditDialog(account: Account) {
  isEditing.value = true
  editingId.value = account.id
  dialogForm.code = account.code
  dialogForm.name = account.name
  dialogForm.accountType = account.account_type
  dialogForm.balanceDirection = account.root_type === 'debit' ? 'debit' : 'credit'
  dialogForm.isGroup = account.is_group
  dialogForm.currency = account.currency
  dialogForm.parentId = account.parent_id
  dialogVisible.value = true
}

function previewAutoCode() {
  if (!dialogForm.parentId) {
    ElMessage.info('请先选择父科目')
    return
  }
  const parent = findNodeById(allAccounts.value, dialogForm.parentId)
  if (!parent) {
    ElMessage.info('未找到父科目')
    return
  }
  const parts = parent.code.split('-')
  if (parts.length >= 4) {
    ElMessage.warning('已达最大层级（4 层）')
    return
  }
  const siblings = allAccounts.value
    .flatMap(n => n.children || [])
    .filter(c => c.parent_id === parent.id || (c.code.startsWith(parent.code) && c.code !== parent.code))
  const maxSeq = siblings.reduce((max, a) => {
    const p = a.code.split('-')
    return Math.max(max, parseInt(p[p.length - 1]) || 0)
  }, 0)
  const nextSeq = String(maxSeq + 1).padStart(2, '0')
  previewCode.value = `${parent.code}-${nextSeq}`
  dialogForm.code = previewCode.value
}

function findNodeById(nodes: Account[], id: string): Account | null {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children) {
      const found = findNodeById(node.children, id)
      if (found) return found
    }
  }
  return null
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!dialogForm.isGroup && !dialogForm.parentId) {
    ElMessage.warning('一级科目必须为汇总科目（Group）')
    return
  }
  saving.value = true
  try {
    if (isEditing.value) {
      await updateAccount(editingId.value, {
        name: dialogForm.name,
        account_type: dialogForm.accountType as any,
        root_type: dialogForm.balanceDirection === 'debit' ? 'debit' as any : 'credit' as any,
        is_group: dialogForm.isGroup,
        currency: dialogForm.currency,
      })
      ElMessage.success('科目已更新')
    } else {
      await createAccount({
        parent_id: dialogForm.parentId || '',
        code: dialogForm.code,
        name: dialogForm.name,
        account_type: dialogForm.accountType as any,
        root_type: dialogForm.balanceDirection === 'debit' ? 'debit' as any : 'credit' as any,
        is_group: dialogForm.isGroup,
        currency: dialogForm.currency,
        company_id: '',
      } as any)
      ElMessage.success('科目已创建')
    }
    dialogVisible.value = false
    await loadTree()
  } catch (e) {
    ElMessage.error(isEditing.value ? '更新失败' : '创建失败')
  } finally {
    saving.value = false
  }
}

function onDetailTypeChange() {
  if (selectedAccount.value) {
    selectedAccount.value.root_type = (['asset', 'cost'].includes(selectedAccount.value.account_type) ? 'debit' : 'credit') as any
  }
}

async function handleSaveDetail() {
  if (!selectedAccount.value) return
  try {
    await updateAccount(selectedAccount.value.id, {
      name: selectedAccount.value.name,
      account_type: selectedAccount.value.account_type,
      root_type: selectedAccount.value.root_type === 'debit' ? 'debit' as any : 'credit' as any,
      is_group: selectedAccount.value.is_group,
      currency: selectedAccount.value.currency,
    })
    ElMessage.success('保存成功')
    await loadTree()
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

function handleDelete(account: Account) {
  if (account.children && account.children.length > 0) {
    ElMessage.warning('该科目存在下级科目，无法删除')
    return
  }
  ElMessageBox.confirm(`确认删除科目「${account.code} ${account.name}」？\n删除后不可恢复。`, '确认删除', {
    type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
  }).then(async () => {
    try {
      await deleteAccount(account.id)
      ElMessage.success('科目已删除')
      if (selectedAccount.value?.id === account.id) {
        selectedAccount.value = null
      }
      await loadTree()
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

async function loadTree() {
  try {
    const res = await fetchAccountTree()
    const data = (res as any)?.data !== undefined && (res as any)?.data !== null ? (res as any).data : res
    allAccounts.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.warn('加载科目失败', e)
    allAccounts.value = []
  }
}

onMounted(async () => {
  await loadTree()
  document.addEventListener('click', hideContextMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', hideContextMenu)
})
</script>

<style scoped lang="scss">
.subject-tree {
  padding: 24px;
}

.subject-layout {
  display: flex;
  flex-direction: row;
  gap: 16px;
  min-height: 500px;
}

.left-panel {
  width: 280px;
  flex-shrink: 0;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  background: #fafafa;

  .left-toolbar {
    margin-bottom: 12px;
  }
}

.right-panel {
  flex: 1;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 20px;
  background: #fff;
}

.detail-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.detail-form {
  .detail-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;

    .detail-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .detail-actions {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid #ebeef5;
  }
}

.tree-node-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.context-menu {
  position: absolute;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 4px 0;
  z-index: 1000;

  .context-menu-item {
    padding: 8px 16px;
    font-size: 14px;
    color: #606266;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    &.danger {
      color: #f56c6c;
    }
  }
}

.code-hint {
  color: #1890ff;
  font-size: 12px;
  margin-left: 8px;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-left: 8px;
}

.group-name {
  font-style: italic;
  color: #666;
}
</style>