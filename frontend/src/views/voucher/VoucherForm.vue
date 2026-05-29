<template>
  <Layout>
    <div class="voucher-form">
      <n-card :title="isEdit ? '编辑凭证' : '新增凭证'">
        <n-form ref="formRef" :model="formData" label-placement="left" label-width="100">
          <n-form-item label="凭证日期">
            <n-date-picker v-model:value="formData.dateTimestamp" type="date" style="width: 200px;" />
          </n-form-item>

          <n-form-item label="凭证分录">
            <n-data-table
              :columns="lineColumns"
              :data="formData.lines"
              :bordered="true"
              size="small"
              :max-height="400"
            />
          </n-form-item>

          <n-form-item label="">
            <n-space>
              <n-button @click="addLine">添加分录</n-button>
              <n-button type="primary" @click="handleSave" :loading="saving">保存</n-button>
              <n-button @click="$router.back()">取消</n-button>
            </n-space>
          </n-form-item>
        </n-form>

        <!-- 科目选择弹窗 -->
        <n-modal v-model:show="showAccountModal" preset="card" title="选择科目" style="width: 600px;">
          <n-tree
            :data="accountTree"
            :render-suffix="renderSuffix"
            block-line
            @update:selected-keys="handleAccountSelect"
          />
        </n-modal>
      </n-card>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NCard, NForm, NDatePicker, NDataTable, NSpace, NButton, NModal, NTree, NIcon, useMessage, type DataTableColumns } from 'naive-ui'
import { AddOutline, RemoveOutline } from '@vicons/ionicons5'
import Layout from '@/components/Layout.vue'
import { voucherApi, accountApi } from '@/api/adapter/client'
import type { VoucherLine, AccountNode } from '@/types'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const isEdit = computed(() => !!route.params.id)
const formRef = ref()
const saving = ref(false)

const formData = ref({
  dateTimestamp: Date.now(),
  date: computed(() => new Date(formData.value.dateTimestamp).toISOString().split('T')[0]),
  lines: [] as (VoucherLine & { key: number })[]
})

const showAccountModal = ref(false)
const currentLineIndex = ref(-1)
const accountTree = ref<any[]>([])

const lineColumns: DataTableColumns<any> = [
  { title: '科目代码', key: 'account_code', width: 120 },
  { title: '科目名称', key: 'account_name', width: 180 },
  { title: '摘要', key: 'summary', width: 200 },
  { title: '借方', key: 'debit', width: 120, render: (row) => h(NInput, { value: row.debit, onUpdateValue: (v) => updateLine(row.key, 'debit', v) }) },
  { title: '贷方', key: 'credit', width: 120, render: (row) => h(NInput, { value: row.credit, onUpdateValue: (v) => updateLine(row.key, 'credit', v) }) },
  { title: '操作', key: 'action', width: 80, render: (row) => h(NButton, { text: true, onClick: () => removeLine(row.key) }, { default: () => '删除' }) }
]

const updateLine = (key: number, field: string, value: any) => {
  const line = formData.value.lines.find(l => l.key === key)
  if (line) (line as any)[field] = value
}

const addLine = () => {
  formData.value.lines.push({
    key: Date.now(),
    account_id: '',
    account_code: '',
    account_name: '',
    summary: '',
    debit: 0,
    credit: 0
  })
}

const removeLine = (key: number) => {
  formData.value.lines = formData.value.lines.filter(l => l.key !== key)
}

const handleSave = async () => {
  // 验证
  if (!formData.value.lines.length) {
    return message.warning('请添加凭证分录')
  }
  
  const debitTotal = formData.value.lines.reduce((s, l) => s + (l.debit || 0), 0)
  const creditTotal = formData.value.lines.reduce((s, l) => s + (l.credit || 0), 0)
  if (Math.abs(debitTotal - creditTotal) > 0.01) {
    return message.warning('借贷合计必须相等')
  }

  saving.value = true
  try {
    const data = {
      date: formData.value.date,
      lines: formData.value.lines.map(l => ({
        account_id: l.account_id,
        debit: l.debit || 0,
        credit: l.credit || 0,
        summary: l.summary || ''
      }))
    }

    if (isEdit.value) {
      await voucherApi.update(route.params.id as string, data)
    } else {
      await voucherApi.create(data)
    }
    
    message.success('保存成功')
    router.push('/vouchers')
  } catch (e: any) {
    message.error(e?.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleAccountSelect = (keys: string[]) => {
  if (keys.length) {
    const node = findAccountNode(keys[0], accountTree.value)
    if (node && currentLineIndex.value >= 0) {
      const line = formData.value.lines[currentLineIndex.value]
      line.account_id = node.id
      line.account_code = node.code
      line.account_name = node.name
    }
    showAccountModal.value = false
  }
}

const findAccountNode = (id: string, nodes: any[]): AccountNode | null => {
  for (const n of nodes) {
    if (n.key === id) return n
    if (n.children?.length) {
      const found = findAccountNode(id, n.children)
      if (found) return found
    }
  }
  return null
}

const renderSuffix = () => h(NIcon, null, { default: () => h(AddOutline) })

onMounted(async () => {
  try {
    const tree = await accountApi.tree()
    accountTree.value = buildTree(tree)
  } catch (e) {
    console.error('Failed to load account tree:', e)
  }

  if (isEdit.value) {
    try {
      const voucher = await voucherApi.get(route.params.id as string)
      formData.value.dateTimestamp = new Date(voucher.date).getTime()
      formData.value.lines = voucher.lines.map((l, i) => ({
        ...l,
        key: i
      }))
    } catch (e) {
      message.error('加载凭证失败')
    }
  }
})

const buildTree = (accounts: AccountNode[]): any[] => {
  return accounts.map(acc => ({
    key: acc.id,
    label: `${acc.code} - ${acc.name}`,
    code: acc.code,
    name: acc.name,
    children: acc.children?.length ? buildTree(acc.children) : undefined
  }))
}
</script>

<script lang="ts">
import { h } from 'vue'
export default { name: 'VoucherForm' }
</script>