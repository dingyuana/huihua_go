<template>
  <div class="reimbursement-form-page">
    <div class="page-header">
      <h3>{{ isEdit ? '编辑报销单' : '新建报销单' }}</h3>
    </div>

    <el-card>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" size="small">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="申请人" prop="employee_name">
              <el-input v-model="form.employee_name" placeholder="请输入申请人姓名" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="部门" prop="department">
              <el-input v-model="form.department" placeholder="请输入部门" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="报销金额" prop="amount">
              <el-input-number
                v-model="form.amount"
                :min="0"
                :precision="2"
                :controls="false"
                style="width: 100%"
                placeholder="0.00"
                :disabled="readonly"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="说明" prop="description">
              <el-input v-model="form.description" placeholder="请输入报销说明" :disabled="readonly" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注" prop="description">
              <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选" :disabled="readonly" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="form-actions">
        <el-button @click="goBack">返回</el-button>
        <template v-if="!readonly">
          <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
        </template>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchReimbursementDetail, createReimbursement, updateReimbursement } from '@/api/modules/reimbursement'
import type { Reimbursement } from '@/api/modules/reimbursement'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id && route.path.endsWith('/edit'))
const readonly = computed(() => {
  const p = route.params.id
  return !!p && route.path !== `/expense/reimbursement/${p}/edit`
})

const formRef = ref()
const saving = ref(false)

const form = reactive({
  employee_name: '',
  department: '',
  amount: 0,
  description: '',
  remark: '',
})

const rules = {
  employee_name: [{ required: true, message: '请输入申请人姓名', trigger: 'blur' }],
  department: [{ required: true, message: '请输入部门', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入报销金额', trigger: 'blur' }],
  description: [{ required: true, message: '请输入报销说明', trigger: 'blur' }],
}

async function handleSave() {
  try {
    await formRef.value.validate()
    saving.value = true
    const payload: Partial<Reimbursement> = {
      ...form,
      amount: String(form.amount),
    }
    if (isEdit.value) {
      await updateReimbursement(route.params.id as string, payload)
      ElMessage.success('更新成功')
    } else {
      await createReimbursement(payload)
      ElMessage.success('创建成功')
    }
    goBack()
  } catch (e: any) {
    if (e !== false) {
      ElMessage.error(e?.response?.data?.error || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(async () => {
  if (route.params.id) {
    try {
      const res: any = await fetchReimbursementDetail(route.params.id as string)
      const data: Reimbursement = res?.data || res
      Object.assign(form, {
        employee_name: data.employee_name,
        department: data.department,
        amount: parseFloat(data.amount) || 0,
        description: data.description,
        remark: data.remark || '',
      })
    } catch {
      ElMessage.error('加载数据失败')
    }
  }
})
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>