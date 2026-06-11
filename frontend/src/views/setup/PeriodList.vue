<template>
  <div class="period-list">
    <PageLayout title="会计期间" icon="📅" subtitle="管理会计期间" iconBg="linear-gradient(135deg, #667eea, #764ba2)">
      <template #actions>
        <el-date-picker
          v-model="filterYear"
          type="year"
          placeholder="选择年份"
          value-format="YYYY"
          style="width: 140px"
          @change="loadPeriods"
        />
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
      </template>

      <el-table
        :data="periods"
        border
        stripe
        size="small"
        v-loading="loading"
        row-key="id"
      >
        <el-table-column prop="period_no" label="期间编号" width="120" />
        <el-table-column prop="period_name" label="期间名称" min-width="180" />
        <el-table-column label="开始日期" width="130">
          <template #default="{ row }">
            {{ formatDate(row.start_date) }}
          </template>
        </el-table-column>
        <el-table-column label="结束日期" width="130">
          <template #default="{ row }">
            {{ formatDate(row.end_date) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'open' ? 'success' : 'info'" size="small">
              {{ row.status === 'open' ? '启用' : '已关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="editPeriod(row)">编辑</el-button>
            <el-switch
              v-if="row.status === 'open' || row.status === 'closed'"
              :model-value="row.status === 'open'"
              size="small"
              :active-text="row.status === 'open' ? '启用' : '已关闭'"
              style="margin: 0 8px"
              @change="() => toggleStatus(row)"
            />
            <el-button link type="danger" size="small" @click="deletePeriod(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </PageLayout>

    <el-dialog v-model="showDialog" :title="editingPeriod ? '编辑期间' : '新增期间'" width="520px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px">
        <el-form-item label="期间编号" prop="period_no">
          <el-input v-model="form.period_no" placeholder="如：2024-01" />
        </el-form-item>
        <el-form-item label="期间名称" prop="period_name">
          <el-input v-model="form.period_name" placeholder="如：2024年1月" />
        </el-form-item>
        <el-form-item label="开始日期" prop="start_date">
          <el-date-picker
            v-model="form.start_date"
            type="date"
            placeholder="选择开始日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束日期" prop="end_date">
          <el-date-picker
            v-model="form.end_date"
            type="date"
            placeholder="选择结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="启用状态" prop="status">
          <el-switch
            v-model="form.status"
            :active-value="'open'"
            :inactive-value="'closed'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="savePeriod">保存</el-button>
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

interface PeriodItem {
  id: string
  period_no: string
  period_name: string
  start_date: string
  end_date: string
  status: 'open' | 'closed'
  created_at?: string
}

const periods = ref<PeriodItem[]>([])
const loading = ref(false)
const saving = ref(false)
const filterYear = ref<string>('')

const showDialog = ref(false)
const editingPeriod = ref<PeriodItem | null>(null)
const formRef = ref<FormInstance>()

const defaultForm = {
  period_no: '',
  period_name: '',
  start_date: '',
  end_date: '',
  status: 'open' as 'open' | 'closed'
}

const form = reactive({ ...defaultForm })

const formRules = {
  period_no: [{ required: true, message: '请输入期间编号', trigger: 'blur' }],
  period_name: [{ required: true, message: '请输入期间名称', trigger: 'blur' }],
  start_date: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  end_date: [{ required: true, message: '请选择结束日期', trigger: 'change' }]
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return dateStr
}

async function loadPeriods() {
  loading.value = true
  try {
    const params: Record<string, any> = {}
    if (filterYear.value) {
      params.year = filterYear.value
    }
    const res: any = await request.get('/periods', { params })
    const list = Array.isArray(res?.periods) ? res.periods : (Array.isArray(res?.data) ? res.data : [])
    periods.value = list
  } catch (e) {
    console.warn('加载会计期间失败', e)
    ElMessage.error('加载会计期间失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingPeriod.value = null
  Object.assign(form, defaultForm)
  showDialog.value = true
}

function editPeriod(period: PeriodItem) {
  editingPeriod.value = period
  Object.assign(form, {
    period_no: period.period_no,
    period_name: period.period_name,
    start_date: period.start_date,
    end_date: period.end_date,
    status: period.status
  })
  showDialog.value = true
}

async function savePeriod() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (editingPeriod.value && editingPeriod.value.id) {
      await request.put(`/periods/${editingPeriod.value.id}`, {
        period_no: form.period_no,
        period_name: form.period_name,
        start_date: form.start_date,
        end_date: form.end_date,
        status: form.status
      })
      ElMessage.success('期间已更新')
    } else {
      await request.post('/periods', {
        period_no: form.period_no,
        period_name: form.period_name,
        start_date: form.start_date,
        end_date: form.end_date,
        status: form.status
      })
      ElMessage.success('期间已创建')
    }
    showDialog.value = false
    await loadPeriods()
  } catch (e: any) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function toggleStatus(period: PeriodItem) {
  if (!period.id) return
  const nextStatus = period.status === 'open' ? 'closed' : 'open'
  try {
    await request.put(`/periods/${period.id}/toggle-status`)
    period.status = nextStatus
    ElMessage.success(nextStatus === 'open' ? '期间已启用' : '期间已关闭')
  } catch (e) {
    console.error(e)
  }
}

async function deletePeriod(period: PeriodItem) {
  if (!period.id) return
  try {
    await ElMessageBox.confirm('确定要删除该会计期间吗？', '提示', { type: 'warning' })
    await request.delete(`/periods/${period.id}`)
    ElMessage.success('期间已删除')
    await loadPeriods()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(loadPeriods)
</script>

<style scoped lang="scss">
.period-list {
  padding: 24px;
}
</style>
