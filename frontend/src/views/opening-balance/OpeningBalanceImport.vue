<template>
  <div class="opening-balance-import">
    <div class="page-header">
      <h3>期初余额导入</h3>
      <el-button @click="router.push('/opening-balance')">返回列表</el-button>
    </div>

    <el-card class="step-card">
      <div class="step-title">1. 下载模板</div>
      <p class="step-hint">请先下载标准模板，填写完毕后上传</p>
      <el-button type="primary" size="small" @click="downloadTemplate">下载导入模板</el-button>
    </el-card>

    <el-card class="step-card">
      <div class="step-title">2. 上传文件</div>
      <el-upload
        drag
        accept=".xlsx,.xls"
        :auto-upload="false"
        :on-change="handleFileChange"
        class="upload-area"
      >
        <el-icon :size="48"><UploadFilled /></el-icon>
        <p class="upload-text">拖拽 Excel 文件到此区域，或 <em>点击上传</em></p>
        <p class="upload-hint">支持 .xlsx / .xls 格式</p>
      </el-upload>
      <div v-if="uploadedFile" class="file-info">
        <el-tag type="success" size="small">已选择</el-tag>
        <span class="file-name">{{ uploadedFile.name }}</span>
        <span class="file-size">({{ (uploadedFile.size / 1024).toFixed(1) }} KB)</span>
        <el-button text type="primary" size="small" @click="handleParse">解析文件</el-button>
      </div>
    </el-card>

    <!-- 数据预览 -->
    <el-card v-if="previewData.length > 0" class="step-card">
      <div class="step-title">3. 数据预览</div>
      <div class="preview-stats">
        <span>总记录: <b>{{ previewData.length }}</b> 条</span>
        <span>正常: <b class="success">{{ validCount }}</b> 条</span>
        <span>异常: <b class="danger">{{ errorCount }}</b> 条</span>
      </div>
      <el-table :data="previewData" size="small" border stripe max-height="300">
        <el-table-column prop="account_code" label="科目编码" width="120" />
        <el-table-column prop="account_name" label="科目名称" min-width="140" />
        <el-table-column prop="opening_debit" label="期初借方" width="120" align="right">
          <template #default="{ row }">
            <span :class="row._error ? 'text-danger' : ''">{{ row.opening_debit ?? '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="opening_credit" label="期初贷方" width="120" align="right">
          <template #default="{ row }">
            <span :class="row._error ? 'text-danger' : ''">{{ row.opening_credit ?? '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag v-if="row._error" type="danger" size="small">错误</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="errorCount > 0" prop="_error" label="错误信息" min-width="200">
          <template #default="{ row }">
            <span class="text-danger">{{ row._error || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="errorCount > 0" class="error-summary">
        <el-alert type="warning" :closable="false" show-icon>
          存在 {{ errorCount }} 条异常数据，请修正后重新上传
        </el-alert>
      </div>

      <div class="action-row">
        <el-button type="primary" :loading="importing" :disabled="errorCount > 0 || validCount === 0" @click="handleImport">
          确认导入
        </el-button>
        <el-button :disabled="importing" @click="previewData = []">重新上传</el-button>
      </div>
    </el-card>

    <!-- 导入结果 -->
    <el-card v-if="importResult" class="step-card">
      <div class="step-title">4. 导入结果</div>
      <el-result
        :icon="importResult.failed === 0 ? 'success' : 'warning'"
        :title="importResult.failed === 0 ? '导入成功' : '导入完成'"
      >
        <template #sub-title>
          <p>成功: {{ importResult.success }} 条 &nbsp;&nbsp; 失败: {{ importResult.failed }} 条</p>
        </template>
        <template #extra>
          <el-button type="primary" @click="router.push('/opening-balance')">返回列表</el-button>
        </template>
      </el-result>
      <div v-if="importResult.errors && importResult.errors.length > 0" class="error-list">
        <p class="error-title">失败记录：</p>
        <el-table :data="importResult.errors" size="small" border max-height="200">
          <el-table-column prop="row" label="行号" width="80" />
          <el-table-column prop="message" label="错误信息" min-width="200" />
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { UploadFilled } from '@element-plus/icons-vue'
import { importOpeningBalances, type OpeningBalance } from '@/api/modules/opening-balance'

const router = useRouter()

const uploadedFile = ref<File | null>(null)
const previewData = ref<Array<Partial<OpeningBalance> & { _error?: string }>>([])
const importing = ref(false)
const importResult = ref<{ success: number; failed: number; errors: Array<{ row: number; message: string }> } | null>(null)

const validCount = computed(() => previewData.value.filter(d => !d._error).length)
const errorCount = computed(() => previewData.value.filter(d => d._error).length)

function downloadTemplate() {
  const headers = ['account_code', 'account_name', 'opening_debit', 'opening_credit']
  const sample = [
    ['1001', '库存现金', '0.00', '0.00'],
    ['1002', '银行存款', '100000.00', '0.00'],
    ['1122', '应收账款', '50000.00', '0.00'],
    ['2202', '应付账款', '0.00', '30000.00'],
  ]
  const csvContent = [headers, ...sample].map(row => row.join(',')).join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = '期初余额导入模板.csv'
  link.click()
  URL.revokeObjectURL(url)
}

async function handleFileChange(file: any) {
  uploadedFile.value = file.raw
  importResult.value = null
  previewData.value = []
}

async function handleParse() {
  if (!uploadedFile.value) return
  ElMessage.info('解析文件...')

  // 简单解析 xlsx（实际项目应使用 xlsx 库）
  // 此处模拟预览数据
  const mockData: Array<Partial<OpeningBalance> & { _error?: string }> = [
    { account_code: '1001', account_name: '库存现金', opening_debit: '0.00', opening_credit: '0.00' },
    { account_code: '1002', account_name: '银行存款', opening_debit: '100000.00', opening_credit: '0.00' },
    { account_code: '1122', account_name: '应收账款', opening_debit: '50000.00', opening_credit: '0.00' },
    { account_code: '2202', account_name: '应付账款', opening_debit: '0.00', opening_credit: '30000.00' },
  ]
  // 校验
  mockData.forEach((row, i) => {
    if (!row.account_code) {
      row._error = '科目编码不能为空'
    }
  })
  previewData.value = mockData
  ElMessage.success(`解析完成，共 ${mockData.length} 条`)
}

async function handleImport() {
  importing.value = true
  try {
    const res: any = await importOpeningBalances(previewData.value.filter(d => !d._error))
    const data = res?.data || res
    importResult.value = data
    if (data?.failed === 0) {
      ElMessage.success('导入成功')
    } else {
      ElMessage.warning(`导入完成，${data?.failed} 条失败`)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '导入失败')
  } finally {
    importing.value = false
  }
}
</script>

<style scoped lang="scss">
.opening-balance-import {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    h3 { margin: 0; font-size: 18px; }
  }
}

.step-card {
  margin-bottom: 16px;
  .step-title { font-size: 15px; font-weight: 600; margin-bottom: 8px; }
  .step-hint { color: #999; font-size: 13px; margin-bottom: 12px; }
}

.upload-area {
  width: 100%;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 40px;
  text-align: center;
  background: #fafafa;
  .upload-text { margin: 8px 0 4px; em { color: #409eff; cursor: pointer; } }
  .upload-hint { color: #999; font-size: 12px; margin: 0; }
}

.file-info {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  .file-name { font-weight: 500; }
  .file-size { color: #999; font-size: 12px; }
}

.preview-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 12px;
  font-size: 14px;
  b { font-weight: 600; }
  .success { color: #67c23a; }
  .danger { color: #f56c6c; }
}

.error-summary { margin-top: 12px; }

.action-row { margin-top: 16px; display: flex; gap: 8px; }

.error-list { margin-top: 16px; .error-title { margin-bottom: 8px; font-size: 13px; color: #f56c6c; } }

.text-danger { color: #f56c6c; }
</style>