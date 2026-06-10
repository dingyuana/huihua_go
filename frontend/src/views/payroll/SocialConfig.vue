<template>
  <div class="social-config-page">
    <div class="page-header">
      <h3>社保配置</h3>
    </div>

    <el-card>
      <el-table :data="configs" border stripe size="small" v-loading="loading">
        <el-table-column prop="insurance_name" label="险种名称" min-width="140" />
        <el-table-column label="单位比例 (%)" width="140" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row._company_rate"
              :min="0"
              :max="100"
              :precision="4"
              :step="0.0001"
              size="small"
              style="width: 120px"
              controls-position="right"
            />
          </template>
        </el-table-column>
        <el-table-column label="个人比例 (%)" width="140" align="center">
          <template #default="{ row }">
            <el-input-number
              v-model="row._personal_rate"
              :min="0"
              :max="100"
              :precision="4"
              :step="0.0001"
              size="small"
              style="width: 120px"
              controls-position="right"
            />
          </template>
        </el-table-column>
        <el-table-column label="启用状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch v-model="row._is_active" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              :loading="row._saving"
              :disabled="!row._dirty"
              @click="handleSave(row)"
            >
              保存
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchSocialConfig, updateSocialConfig } from '@/api/modules/payroll'
import type { SocialConfig } from '@/types/models/social-config'

const loading = ref(false)
const configs = ref<(SocialConfig & {
  _company_rate: number
  _personal_rate: number
  _is_active: boolean
  _dirty: boolean
  _saving: boolean
})[]>([])

async function loadConfigs() {
  loading.value = true
  try {
    const res: any = await fetchSocialConfig()
    const list: SocialConfig[] = res?.data?.configs || res?.configs || []
    configs.value = list.map(item => {
      const compRate = parseFloat(item.company_rate || '0') * 100
      const persRate = parseFloat(item.personal_rate || '0') * 100
      return {
        ...item,
        _company_rate: compRate,
        _personal_rate: persRate,
        _is_active: item.is_active,
        _dirty: false,
        _saving: false,
      }
    })
  } catch (e: any) {
    ElMessage.error('加载社保配置失败：' + (e?.response?.data?.error || e?.message || e))
  } finally {
    loading.value = false
  }
}

// 监听编辑变更，标记脏状态
watch(
  configs,
  (list) => {
    list.forEach(item => {
      const origComp = parseFloat(item.company_rate || '0') * 100
      const origPers = parseFloat(item.personal_rate || '0') * 100
      const changed =
        item._company_rate !== origComp ||
        item._personal_rate !== origPers ||
        item._is_active !== item.is_active
      if (changed !== item._dirty) {
        item._dirty = changed
      }
    })
  },
  { deep: true },
)

async function handleSave(row: any) {
  row._saving = true
  try {
    await updateSocialConfig(row.id, {
      company_rate: (row._company_rate / 100).toFixed(6),
      personal_rate: (row._personal_rate / 100).toFixed(6),
      is_active: row._is_active,
    })
    // 同步原始值
    row.company_rate = (row._company_rate / 100).toFixed(6)
    row.personal_rate = (row._personal_rate / 100).toFixed(6)
    row.is_active = row._is_active
    row._dirty = false
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error('保存失败：' + (e?.response?.data?.error || e?.message || e))
  } finally {
    row._saving = false
  }
}

onMounted(loadConfigs)
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  h3 { font-size: 18px; }
}
</style>
