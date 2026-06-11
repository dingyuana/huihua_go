<template>
  <div class="voucher-type-list">
    <PageLayout title="凭证类型" icon="📋" subtitle="管理记账凭证分类" iconBg="linear-gradient(135deg, #43e97b, #38f9d7)">
      <template #actions>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
      </template>

      <el-table
        :data="voucherTypes"
        border
        stripe
        size="small"
        row-key="id"
        v-loading="loading"
      >
        <el-table-column prop="code" label="类型编码" width="120" />
        <el-table-column prop="name" label="类型名称" min-width="140" />
        <el-table-column prop="short_name" label="简称" width="100" />
        <el-table-column prop="voucher_prefix" label="凭证字" width="90">
          <template #default="{ row }">
            <el-tag size="small">{{ row.voucher_prefix }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序号" width="80" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editVoucherType(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteVoucherType(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </PageLayout>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑凭证类型' : '新增凭证类型'" width="520px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="类型编码" prop="code">
          <el-input v-model="form.code" placeholder="如：JE" :disabled="!!editingId" />
        </el-form-item>
        <el-form-item label="类型名称" prop="name">
          <el-input v-model="form.name" placeholder="如：记账凭证" />
        </el-form-item>
        <el-form-item label="简称" prop="short_name">
          <el-input v-model="form.short_name" placeholder="如：记账" />
        </el-form-item>
        <el-form-item label="凭证字" prop="voucher_prefix">
          <el-input v-model="form.voucher_prefix" placeholder="如：记、转、银、现" maxlength="4" />
        </el-form-item>
        <el-form-item label="排序号" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="启用状态" prop="is_active">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveVoucherType">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import request from '@/api/request'
import PageLayout from '@/components/app/PageLayout.vue'

interface VoucherType {
  id: string
  code: string
  name: string
  short_name: string
  voucher_prefix: string
  sort_order: number
  is_active: boolean
}

const voucherTypes = ref<VoucherType[]>([])
const loading = ref(false)
const saving = ref(false)

const showDialog = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()

const defaultForm = {
  code: '',
  name: '',
  short_name: '',
  voucher_prefix: '',
  sort_order: 1,
  is_active: true,
}

const form = reactive({ ...defaultForm })

const formRules = {
  code: [{ required: true, message: '请输入类型编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入类型名称', trigger: 'blur' }],
}

async function loadVoucherTypes() {
  loading.value = true
  try {
    const res: any = await request.get('/voucher-types')
    if (Array.isArray(res?.data)) {
      voucherTypes.value = res.data
    }
  } catch (e) {
    console.warn('加载凭证类型失败', e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  Object.assign(form, defaultForm)
  showDialog.value = true
}

function editVoucherType(item: VoucherType) {
  editingId.value = item.id
  form.code = item.code
  form.name = item.name
  form.short_name = item.short_name
  form.voucher_prefix = item.voucher_prefix
  form.sort_order = item.sort_order
  form.is_active = item.is_active
  showDialog.value = true
}

async function saveVoucherType() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (editingId.value) {
      await request.put(`/voucher-types/${editingId.value}`, { ...form })
      ElMessage.success('凭证类型已更新')
    } else {
      await request.post('/voucher-types', { ...form })
      ElMessage.success('凭证类型已创建')
    }
    showDialog.value = false
    await loadVoucherTypes()
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function deleteVoucherType(item: VoucherType) {
  if (!item.id) return
  try {
    await ElMessageBox.confirm('确定要删除该凭证类型吗？删除后不可恢复。', '提示', { type: 'warning' })
    await request.delete(`/voucher-types/${item.id}`)
    ElMessage.success('凭证类型已删除')
    await loadVoucherTypes()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(loadVoucherTypes)
</script>

<style scoped lang="scss">
.voucher-type-list {
  padding: 24px;
}
</style>
