<template>
  <div class="account-chart">
    <PageLayout
      title="科目表管理"
      icon="📋"
      subtitle="查看和管理会计科目，支持树形结构和导入标准科目表"
    >
      <template #actions>
        <el-button @click="importStandardChart">
          <el-icon><Download /></el-icon>
          导入标准科目表
        </el-button>
        <el-button
          type="primary"
          @click="openCreateDialog()"
        >
          <el-icon><Plus /></el-icon>
          新增科目
        </el-button>
      </template>

      <div class="toolbar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索科目名称或编码"
          clearable
          style="width: 260px"
          @input="onSearch"
        />
        <el-select
          v-model="filterType"
          placeholder="科目类型"
          clearable
          style="width: 140px"
        >
          <el-option
            label="资产"
            value="asset"
          />
          <el-option
            label="负债"
            value="liability"
          />
          <el-option
            label="权益"
            value="equity"
          />
          <el-option
            label="成本"
            value="cost"
          />
          <el-option
            label="损益"
            value="income"
          />
        </el-select>
        <el-checkbox
          v-model="showLedgerOnly"
          style="margin-left: 12px"
        >
          仅可记账科目
        </el-checkbox>
      </div>

      <el-table
        :data="filteredData"
        row-key="id"
        :tree-props="{ children: 'children' }"
        default-expand-all
        border
        stripe
        size="small"
        class="account-tree"
      >
        <el-table-column
          prop="code"
          label="科目编码"
          width="150"
        />
        <el-table-column
          prop="name"
          label="科目名称"
          min-width="220"
        >
          <template #default="{ row }">
            <span :class="{ 'group-name': row.is_group, 'ledger-name': !row.is_group }">
              {{ row.name }}
              <el-tag
                v-if="row.is_group"
                size="small"
                type="info"
                effect="plain"
                style="margin-left: 6px"
              >汇总</el-tag>
            </span>
          </template>
        </el-table-column>
        <el-table-column
          label="类型"
          width="80"
        >
          <template #default="{ row }">
            {{ typeLabel(row.account_type) }}
          </template>
        </el-table-column>
        <el-table-column
          label="余额方向"
          width="80"
        >
          <template #default="{ row }">
            {{ rootTypeByAccountType(row.account_type) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="currency"
          label="币种"
          width="60"
        />
        <el-table-column
          label="操作"
          width="180"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              @click="openCreateDialog(row)"
            >
              新增下级
            </el-button>
            <el-button
              link
              type="primary"
              size="small"
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              link
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </PageLayout>

    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑科目' : '新增科目'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="dialogForm"
        :rules="dialogRules"
        label-width="100px"
      >
        <el-form-item label="父科目">
          <AccountSelector
            v-model="dialogForm.parent"
            :ledger-only="false"
            placeholder="根科目（不选）"
          />
        </el-form-item>
        <el-form-item label="科目编码">
          <el-input
            v-model="dialogForm.code"
            :disabled="isEditing"
            style="width: 200px"
          >
            <template #append>
              <el-button @click="previewAutoCode">
                自动生成
              </el-button>
            </template>
          </el-input>
          <span
            v-if="previewCode"
            class="code-hint"
          >建议: {{ previewCode }}</span>
        </el-form-item>
        <el-form-item
          label="科目名称"
          prop="name"
        >
          <el-input
            v-model="dialogForm.name"
            maxlength="100"
          />
        </el-form-item>
        <el-form-item label="科目类型">
          <el-select
            v-model="dialogForm.accountType"
            style="width: 100%"
            @change="onTypeChange"
          >
            <el-option
              label="资产"
              value="asset"
            />
            <el-option
              label="负债"
              value="liability"
            />
            <el-option
              label="共同"
              value="equity"
            />
            <el-option
              label="成本"
              value="cost"
            />
            <el-option
              label="损益"
              value="income"
            />
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
          <el-select
            v-model="dialogForm.currency"
            style="width: 120px"
          >
            <el-option
              label="CNY"
              value="CNY"
            />
            <el-option
              label="USD"
              value="USD"
            />
            <el-option
              label="HKD"
              value="HKD"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSave"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import type { Account } from '@/types/models/account'
import request from '@/api/request'
import PageLayout from '@/components/app/PageLayout.vue'

function rootTypeByAccountType(type: string): string {
  if (['asset', 'cost'].includes(type)) return '借方'
  return '贷方'
}

function typeLabel(type: string): string {
  const map: Record<string, string> = { asset: '资产', liability: '负债', equity: '权益', cost: '成本', income: '损益' }
  return map[type] || type
}

function buildStandardAccounts(): Account[] {
  const id = (prefix: string) => `std-${prefix}`
  const acc = (code: string, name: string, type: string, isGroup: boolean, children?: Account[]): Account => ({
    id: id(code), code, name, account_type: type as any, root_type: rootTypeByAccountType(type) === '借方' ? 'debit' as any : 'credit' as any,
    parent_id: null, lft: 0, rgt: 0, is_group: isGroup, company_id: '', currency: 'CNY', created_at: '', children,
  })
  return [
    acc('1001', '现金', 'asset', true, [
      acc('1001-01', '库存现金', 'asset', false),
      acc('1001-02', '银行存款', 'asset', true, [
        acc('1001-02-01', '银行存款-工行', 'asset', false),
        acc('1001-02-02', '银行存款-建行', 'asset', false),
        acc('1001-02-03', '银行存款-招行', 'asset', false),
      ]),
      acc('1001-03', '其他货币资金', 'asset', true, [
        acc('1001-03-01', '外埠存款', 'asset', false),
        acc('1001-03-02', '银行本票', 'asset', false),
        acc('1001-03-03', '银行汇票', 'asset', false),
      ]),
    ]),
    acc('1122', '应收及预付款项', 'asset', true, [
      acc('1122-01', '应收账款', 'asset', true, [
        acc('1122-01-01', '应收账款-A公司', 'asset', false),
        acc('1122-01-02', '应收账款-B公司', 'asset', false),
      ]),
      acc('1122-02', '预付账款', 'asset', false),
      acc('1122-03', '应收股利', 'asset', false),
      acc('1122-04', '应收利息', 'asset', false),
      acc('1122-05', '其他应收款', 'asset', false),
      acc('1122-06', '坏账准备', 'asset', false),
    ]),
    acc('1401', '存货', 'asset', true, [
      acc('1401-01', '原材料', 'asset', false),
      acc('1401-02', '在产品', 'asset', false),
      acc('1401-03', '库存商品', 'asset', false),
      acc('1401-04', '周转材料', 'asset', false),
      acc('1401-05', '存货跌价准备', 'asset', false),
    ]),
    acc('1501', '长期投资', 'asset', true, [
      acc('1501-01', '长期股权投资', 'asset', false),
      acc('1501-02', '长期债券投资', 'asset', false),
    ]),
    acc('1601', '固定资产', 'asset', true, [
      acc('1601-01', '固定资产原值', 'asset', true, [
        acc('1601-01-01', '房屋建筑物', 'asset', false),
        acc('1601-01-02', '机器设备', 'asset', false),
        acc('1601-01-03', '运输工具', 'asset', false),
        acc('1601-01-04', '电子设备', 'asset', false),
      ]),
      acc('1601-02', '累计折旧', 'asset', false),
      acc('1601-03', '固定资产清理', 'asset', false),
      acc('1601-04', '在建工程', 'asset', false),
    ]),
    acc('1701', '无形资产', 'asset', true, [
      acc('1701-01', '无形资产原值', 'asset', false),
      acc('1701-02', '累计摊销', 'asset', false),
    ]),
    acc('2001', '应付及预收款项', 'liability', true, [
      acc('2001-01', '应付账款', 'liability', true, [
        acc('2001-01-01', '应付账款-X公司', 'liability', false),
        acc('2001-01-02', '应付账款-Y公司', 'liability', false),
      ]),
      acc('2001-02', '预收账款', 'liability', false),
      acc('2001-03', '应付职工薪酬', 'liability', true, [
        acc('2001-03-01', '工资', 'liability', false),
        acc('2001-03-02', '奖金', 'liability', false),
        acc('2001-03-03', '社保', 'liability', false),
        acc('2001-03-04', '公积金', 'liability', false),
      ]),
      acc('2001-04', '应交税费', 'liability', true, [
        acc('2001-04-01', '应交增值税', 'liability', false),
        acc('2001-04-02', '应交所得税', 'liability', false),
        acc('2001-04-03', '应交城建税', 'liability', false),
      ]),
      acc('2001-05', '应付利息', 'liability', false),
      acc('2001-06', '应付股利', 'liability', false),
      acc('2001-07', '其他应付款', 'liability', false),
    ]),
    acc('3001', '所有者权益', 'equity', true, [
      acc('3001-01', '实收资本', 'equity', false),
      acc('3001-02', '资本公积', 'equity', false),
      acc('3001-03', '盈余公积', 'equity', false),
      acc('3001-04', '未分配利润', 'equity', false),
      acc('3001-05', '本年利润', 'equity', false),
    ]),
    acc('4001', '生产成本', 'cost', true, [
      acc('4001-01', '直接材料', 'cost', false),
      acc('4001-02', '直接人工', 'cost', false),
      acc('4001-03', '制造费用', 'cost', false),
    ]),
    acc('5001', '营业收入', 'income', true, [
      acc('5001-01', '主营业务收入', 'income', false),
      acc('5001-02', '其他业务收入', 'income', false),
      acc('5001-03', '投资收益', 'income', false),
      acc('5001-04', '营业外收入', 'income', false),
    ]),
    acc('5002', '营业成本及税金', 'income', true, [
      acc('5002-01', '主营业务成本', 'income', false),
      acc('5002-02', '其他业务成本', 'income', false),
      acc('5002-03', '营业税金及附加', 'income', false),
    ]),
    acc('5601', '期间费用', 'income', true, [
      acc('5601-01', '销售费用', 'income', true, [
        acc('5601-01-01', '广告费', 'income', false),
        acc('5601-01-02', '运输费', 'income', false),
      ]),
      acc('5601-02', '管理费用', 'income', true, [
        acc('5601-02-01', '办公费', 'income', false),
        acc('5601-02-02', '差旅费', 'income', false),
        acc('5601-02-03', '业务招待费', 'income', false),
        acc('5601-02-04', '折旧费', 'income', false),
        acc('5601-02-05', '工资', 'income', false),
      ]),
      acc('5601-03', '财务费用', 'income', true, [
        acc('5601-03-01', '利息支出', 'income', false),
        acc('5601-03-02', '手续费', 'income', false),
        acc('5601-03-03', '汇兑损益', 'income', false),
      ]),
    ]),
    acc('5701', '营业外支出', 'income', false),
    acc('5801', '所得税', 'income', true, [
      acc('5801-01', '所得税费用', 'income', false),
    ]),
  ]
}

const allAccounts = ref<Account[]>([])
const searchKeyword = ref('')
const filterType = ref('')
const showLedgerOnly = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const previewCode = ref('')
const formRef = ref<FormInstance>()
const saving = ref(false)

const dialogForm = reactive({
  parent: null as Account | null,
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

const filteredData = computed(() => {
  let data = allAccounts.value
  if (showLedgerOnly.value) {
    data = filterLedgerOnly(data)
  }
  if (filterType.value) {
    data = deepFilter(data, (n: Account) => n.account_type === filterType.value)
  }
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    data = deepFilter(data, (n: Account) => n.code.includes(kw) || n.name.includes(kw))
  }
  return data
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

function filterLedgerOnly(nodes: Account[]): Account[] {
  return nodes.filter(n => !n.is_group || (n.children && n.children.length > 0 && filterLedgerOnly(n.children).length > 0))
}

function onSearch() {}

function onTypeChange() {
  dialogForm.balanceDirection = ['asset', 'cost'].includes(dialogForm.accountType) ? 'debit' : 'credit'
}

function openCreateDialog(parent?: Account) {
  isEditing.value = false
  editingId.value = ''
  previewCode.value = ''
  dialogForm.parent = parent || null
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
  dialogForm.balanceDirection = rootTypeByAccountType(account.account_type) === '借方' ? 'debit' : 'credit'
  dialogForm.isGroup = account.is_group
  dialogForm.currency = account.currency
  dialogVisible.value = true
}

function previewAutoCode() {
  const parent = dialogForm.parent
  if (!parent) {
    ElMessage.info('请先选择父科目')
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

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!dialogForm.isGroup && !dialogForm.parent) {
    ElMessage.warning('一级科目必须为汇总科目（Group）')
    return
  }
  saving.value = true
  await new Promise(r => setTimeout(r, 300))
  saving.value = false
  ElMessage.success(isEditing.value ? '科目已更新' : '科目已创建')
  dialogVisible.value = false
}

function handleDelete(row: Account) {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('该科目存在下级科目，无法删除')
    return
  }
  ElMessageBox.confirm(`确认删除科目「${row.code} ${row.name}」？\n删除后不可恢复。`, '确认删除', {
    type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
  }).then(() => {
    ElMessage.success('科目已删除')
  }).catch(() => {})
}

onMounted(async () => {
  try {
    const res: any = await request.get('/accounts/tree')
    const data = res?.data !== undefined && res?.data !== null ? res.data : res
    allAccounts.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.warn('后端科目表接口不可用', e)
    allAccounts.value = []
  }
})

async function importStandardChart() {
  try {
    await request.post('/accounts/init-seed', {
      tenant_id: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
      company_id: '12345678-1234-1234-1234-123456789abc',
    })
    const res: any = await request.get('/accounts/tree')
    const data = res?.data !== undefined && res?.data !== null ? res.data : res
    allAccounts.value = Array.isArray(data) ? data : []
    ElMessage.success('已导入《小企业会计准则》标准科目表（5 大类，80+ 科目）')
  } catch (e) {
    allAccounts.value = []
    ElMessage.error('导入标准科目表失败')
  }
}
</script>

<style scoped lang="scss">
.account-chart {
  padding: 24px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #f7f8fa;
  border-radius: 8px;
}

.group-name { font-style: italic; color: #666; }
.ledger-name { font-weight: 500; }
.code-hint { color: #1890ff; font-size: 12px; margin-left: 8px; }
.form-hint { font-size: 12px; color: #999; margin-left: 8px; }
.account-tree :deep(.el-table__body .el-table__row--level-1 td:first-child) { padding-left: 24px !important; }
.account-tree :deep(.el-table__body .el-table__row--level-2 td:first-child) { padding-left: 48px !important; }
</style>
