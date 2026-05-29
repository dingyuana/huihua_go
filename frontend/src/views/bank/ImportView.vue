<template>
  <div class="import-view">
    <div class="page-header">
      <h3>银行流水导入</h3>
    </div>

    <el-card class="step-card">
      <div class="step-title">1. 选择银行账户</div>
      <el-select v-model="bankAccountId" placeholder="选择银行账户" style="width: 320px">
        <el-option label="工商银行-基本户 (1102****4567)" value="ba-1" />
        <el-option label="建设银行-一般户 (4302****4321)" value="ba-2" />
      </el-select>
      <el-button v-if="bankAccountId" style="margin-left:8px" @click="fetchOnline">📡 银企直连抓取</el-button>
    </el-card>

    <el-card class="step-card">
      <div class="step-title">2. 上传对账单文件</div>
      <el-upload drag accept=".xlsx,.xls,.csv,.xml" :auto-upload="false" :on-change="handleFileChange" class="upload-area">
        <el-icon :size="48"><UploadFilled /></el-icon>
        <p class="upload-text">拖拽文件到此区域，或 <em>点击上传</em></p>
        <p class="upload-hint">支持 CSV / Excel / CAMT053 / MT940 格式</p>
      </el-upload>
      <div v-if="uploadedFile" class="file-info">
        <el-tag type="success" size="small">已选择</el-tag>
        <span class="file-name">{{ uploadedFile.name }}</span>
        <span class="file-size">({{ (uploadedFile.size / 1024).toFixed(1) }} KB)</span>
        <el-tag :type="formatTagType" size="small" style="margin:0 8px">{{ detectedFormat }}</el-tag>
        <el-button text type="primary" size="small" @click="handleParse">解析文件</el-button>
      </div>
      <!-- 异常记录摘要 -->
      <el-alert v-if="parseErrors.length" :title="`解析完成，发现 ${parseErrors.length} 条异常记录`" type="warning" :closable="false" show-icon style="margin-top:12px" />
    </el-card>

    <!-- 字段映射 -->
    <el-card v-if="showMapping" class="step-card">
      <div class="step-title">3. 字段映射确认</div>
      <p class="step-hint">系统已自动识别以下映射关系，请确认或修正</p>
      <el-table :data="fieldMappings" size="small" border>
        <el-table-column prop="field" label="内部字段" width="120" />
        <el-table-column label="匹配状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.matched ? 'success' : 'danger'" size="small">{{ row.matched ? '已匹配' : '未匹配' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="文件列名">
          <template #default="{ row }">
            <el-select v-model="row.mappedColumn" :placeholder="row.matched ? row.mappedColumn : '请选择'" style="width: 100%">
              <el-option v-for="col in fileColumns" :key="col" :label="col" :value="col" />
            </el-select>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" size="small" style="margin-top:8px" @click="showPreview = true; showMapping = false">确认映射，预览数据</el-button>
    </el-card>

    <!-- 预览 -->
    <el-card v-if="showPreview" class="step-card">
      <div class="step-title">4. 数据预览</div>
      <div class="preview-stats">
        <span>总识别: <b>{{ totalRows }}</b> 条</span>
        <span>正常: <b class="success">{{ normalCount }}</b> 条</span>
        <span>异常: <b class="danger">{{ parseErrors.length }}</b> 条</span>
      </div>
      <el-tabs v-model="previewTab">
        <el-tab-pane label="全部数据" name="all">
          <el-table :data="previewData" size="small" border stripe max-height="300">
            <el-table-column prop="date" label="日期" width="100" />
            <el-table-column prop="amount" label="金额" width="120">
              <template #default="{ row }">
                <span :class="row.direction === 'in' ? 'amount-positive' : 'amount-negative'">
                  {{ row.direction === 'in' ? '+' : '-' }}{{ row.amount }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="direction" label="方向" width="60">
              <template #default="{ row }">{{ row.direction === 'in' ? '收入' : '支出' }}</template>
            </el-table-column>
            <el-table-column prop="counterparty" label="对方户名" width="140" />
            <el-table-column prop="description" label="摘要" min-width="180" show-overflow-tooltip />
            <el-table-column prop="ref" label="流水号" width="130" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="`异常记录 (${parseErrors.length})`" name="errors">
          <el-table :data="parseErrors" size="small" border stripe>
            <el-table-column prop="row" label="行号" width="60" />
            <el-table-column prop="field" label="字段" width="80" />
            <el-table-column prop="value" label="原始值" width="140" />
            <el-table-column prop="issue" label="问题" min-width="200" />
            <el-table-column label="操作" width="100">
              <template #default="{ $index }">
                <el-button size="small" @click="skipRow($index)">跳过</el-button>
                <el-button size="small" type="primary" @click="editRow($index)">补录</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <div v-if="showPreview" class="import-actions">
      <el-button size="large" @click="resetImport">重新选择</el-button>
      <el-button size="large" type="primary" :loading="importing" @click="handleImport">确认导入 ({{ normalCount }} 条)</el-button>
    </div>

    <!-- 导入结果 -->
    <el-result v-if="importDone" icon="success" title="导入完成" class="import-result">
      <template #extra>
        <el-descriptions :column="3" size="small" border>
          <el-descriptions-item label="批次号">BATCH-{{ Date.now().toString(36).toUpperCase() }}</el-descriptions-item>
          <el-descriptions-item label="导入时间">{{ new Date().toLocaleString() }}</el-descriptions-item>
          <el-descriptions-item label="银行账户">{{ bankAccountId === 'ba-1' ? '工商银行-基本户' : '建设银行-一般户' }}</el-descriptions-item>
          <el-descriptions-item label="总条数">{{ importResult.total }}</el-descriptions-item>
          <el-descriptions-item label="成功">{{ importResult.imported }}</el-descriptions-item>
          <el-descriptions-item label="重复">{{ importResult.duplicated }}</el-descriptions-item>
        </el-descriptions>
        <div style="margin-top:16px">
          <el-button type="primary" @click="$router.push('/bank/workbench')">前往核对工作台</el-button>
          <el-button @click="resetImport">继续导入</el-button>
        </div>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const bankAccountId = ref('')
const uploadedFile = ref<File | null>(null)
const showMapping = ref(false)
const showPreview = ref(false)
const previewTab = ref('all')
const importing = ref(false)
const importDone = ref(false)

const importResult = ref({ total: 0, imported: 0, duplicated: 0, failed: 0 })

// 格式自动检测
const detectedFormat = ref('')
const formatTagType = computed(() => {
  const map: Record<string, string> = { 'CSV': '', 'Excel': 'success', 'CAMT053': 'warning', 'MT940': 'info' }
  return map[detectedFormat.value] || ''
})

const fileColumns = ['交易日期', '收入金额', '支出金额', '对方户名', '摘要', '流水号', '余额', '对方账号']

const fieldMappings = ref([
  { field: '交易日期', matched: true, mappedColumn: '交易日期' },
  { field: '金额', matched: true, mappedColumn: '收入金额' },
  { field: '收支方向', matched: false, mappedColumn: '' },
  { field: '对方户名', matched: false, mappedColumn: '' },
  { field: '摘要', matched: true, mappedColumn: '摘要' },
  { field: '流水号', matched: true, mappedColumn: '流水号' },
  { field: '余额', matched: true, mappedColumn: '余额' },
])

const previewData = ref([
  { date: '2026-05-20', amount: '12,000.00', direction: 'in', counterparty: '上海XX贸易公司', description: '网银转账-B2B收款-货款', ref: 'B20260520001', balance: '500,000.00' },
  { date: '2026-05-20', amount: '50.00', direction: 'out', counterparty: '', description: '账户管理费', ref: 'B20260520002', balance: '499,950.00' },
  { date: '2026-05-21', amount: '3,500.00', direction: 'in', counterparty: '北京YY科技', description: '转账收入', ref: 'B20260521001', balance: '503,450.00' },
  { date: '2026-05-21', amount: '5,000.00', direction: 'out', counterparty: '广州ZZ贸易', description: '货款支付', ref: 'B20260521002', balance: '498,450.00' },
  { date: '2026-05-22', amount: '150.00', direction: 'in', counterparty: '', description: '存款利息', ref: 'B20260522001', balance: '498,600.00' },
])

// 异常记录（模拟有字段缺失或不规范的行）
const parseErrors = ref([
  { row: 3, field: '对方户名', value: '(空)', issue: '对方户名为空，无法自动匹配客商' },
  { row: 7, field: '金额', value: '12,0.00', issue: '金额格式异常，可能缺少位数' },
  { row: 12, field: '日期', value: '2026/05/32', issue: '日期不存在' },
])

const totalRows = ref(150)
const normalCount = computed(() => totalRows.value - parseErrors.value.length)

function handleFileChange(file: any) {
  uploadedFile.value = file.raw
  // 模拟格式检测
  const name = file.name.toLowerCase()
  if (name.endsWith('.csv')) detectedFormat.value = 'CSV'
  else if (name.endsWith('.xlsx') || name.endsWith('.xls')) detectedFormat.value = 'Excel'
  else if (name.includes('camt') || name.endsWith('.xml')) detectedFormat.value = 'CAMT053'
  else detectedFormat.value = 'MT940'
}

function handleParse() {
  if (!uploadedFile.value) return
  showMapping.value = true
  ElMessage.success(`文件解析成功，格式：${detectedFormat.value}，共识别 ${totalRows.value} 条记录`)
}

function fetchOnline() {
  ElMessage.success('银企直连抓取成功，获取到 125 条最新流水')
}

function skipRow(idx: number) {
  parseErrors.value.splice(idx, 1)
  totalRows.value--
  ElMessage.info('已跳过该异常记录')
}

function editRow(idx: number) {
  ElMessage.success('请在出纳核对工作台补录该记录')
}

function resetImport() {
  uploadedFile.value = null
  showMapping.value = false
  showPreview.value = false
  importDone.value = false
  bankAccountId.value = ''
  parseErrors.value = []
  detectedFormat.value = ''
}

function detectDuplicates(rows: any[]): { unique: any[]; duplicates: number } {
  const seen = new Set<string>()
  let dup = 0
  const unique = rows.filter(row => {
    const key = `${bankAccountId.value}|${row.ref}|${row.date}|${row.amount}|${row.direction}`
    if (seen.has(key)) { dup++; return false }
    seen.add(key)
    return true
  })
  return { unique, duplicates: dup }
}

async function handleImport() {
  importing.value = true
  await new Promise(r => setTimeout(r, 1500))
  importing.value = false
  const { duplicates } = detectDuplicates(previewData.value)
  const total = totalRows.value
  const imported = total - duplicates - parseErrors.value.length
  importDone.value = true
  importResult.value = { total, imported, duplicated: duplicates, failed: total - imported - duplicates }
}
</script>

<style scoped lang="scss">
.page-header { margin-bottom: 16px; h3 { font-size: 18px; } }
.step-card { margin-bottom: 16px; .step-title { font-weight: 600; margin-bottom: 16px; font-size: 15px; color: #333; } }
.step-hint { color: #999; font-size: 13px; margin-bottom: 12px; }
.upload-area { text-align: center; .upload-text { font-size: 14px; margin: 8px 0; em { color: #1890ff; font-style: normal; } } .upload-hint { color: #999; font-size: 12px; } }
.file-info { margin-top: 12px; display: flex; align-items: center; flex-wrap: wrap; gap: 4px; .file-name { font-weight: 500; } .file-size { color: #999; font-size: 12px; } }
.preview-stats { display: flex; gap: 20px; margin-bottom: 12px; font-size: 13px; .success { color: #52c41a; } .danger { color: #ff4d4f; } }
.import-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
.import-result { margin-top: 24px; p { margin-bottom: 16px; color: #666; } }
</style>
