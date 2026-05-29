<template>
  <div class="precheck">
    <div class="page-header"><h3>核销预检</h3></div>
    <el-row :gutter="16">
      <el-col :span="8">
        <el-card><template #header>选择收款单</template>
          <el-select v-model="selectedPayment" placeholder="搜索收款单" filterable style="width: 100%">
            <el-option label="SK-2026-05-0001 上海XX ¥12,000" value="p1" />
            <el-option label="SK-2026-05-0002 深圳AA ¥56,500" value="p2" />
          </el-select>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card><template #header>选择发票</template>
          <el-select v-model="selectedInvoice" placeholder="搜索发票" filterable style="width: 100%">
            <el-option label="12345678 上海XX ¥12,000" value="i1" />
            <el-option label="87654321 深圳AA ¥56,500" value="i2" />
          </el-select>
        </el-card>
      </el-col>
      <el-col :span="8" class="action-col">
        <el-button type="primary" size="large" :disabled="!selectedPayment || !selectedInvoice" @click="runPrecheck">执行预检</el-button>
      </el-col>
    </el-row>

    <el-card v-if="precheckDone" class="result-card">
      <template #header>核销预检结果</template>
      <el-table :data="checkResults" border stripe size="small">
        <el-table-column label="#" width="50"><template #default="{ $index }">{{ $index + 1 }}</template></el-table-column>
        <el-table-column prop="name" label="检查项" min-width="160" />
        <el-table-column label="结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'passed' ? 'success' : row.status === 'warning' ? 'warning' : 'danger'" size="small">
              {{ row.status === 'passed' ? '通过' : row.status === 'warning' ? '警告' : '阻断' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="详情" min-width="200" />
      </el-table>
      <div class="precheck-actions">
        <el-button type="primary" :disabled="blockerCount > 0" @click="showForcePassDialog = true">强制通过并核销</el-button>
        <span v-if="blockerCount > 0" class="blocker-msg">尚有 {{ blockerCount }} 项阻断，请先处理</span>
      </div>
    </el-card>

    <!-- 强制通过弹窗 -->
    <el-dialog v-model="showForcePassDialog" title="强制通过核销" width="450px">
      <el-alert type="warning" :closable="false" show-icon>
        <p>以下检查项未通过，强制通过需备注原因：</p>
        <ul>
          <li v-for="c in blockedChecks" :key="c.name" style="margin:4px 0">{{ c.name }}: {{ c.message }}</li>
        </ul>
      </el-alert>
      <el-form class="force-form">
        <el-form-item label="备注原因" required>
          <el-input v-model="forcePassReason" type="textarea" :rows="3" placeholder="请说明强制通过的原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showForcePassDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!forcePassReason" @click="executeForcePass">确认强制通过</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const selectedPayment = ref('')
const selectedInvoice = ref('')
const precheckDone = ref(false)
const showForcePassDialog = ref(false)
const forcePassReason = ref('')

const blockedChecks = computed(() => checks.value.filter(c => c.status === 'blocked'))

function executeForcePass() {
  ElMessage.success(`已强制通过核销（原因：${forcePassReason.value}）`)
  showForcePassDialog.value = false
  forcePassReason.value = ''
  router.push('/reconciliation/match')
}

const checks = ref([
  { name: '对方单位匹配', status: 'passed', message: '上海XX ↔ 发票购方税号一致' },
  { name: '金额超限检查', status: 'passed', message: '¥12,000 ≤ ¥12,000' },
  { name: '重复核销检查', status: 'passed', message: '未重复核销' },
  { name: '跨账套检查', status: 'passed', message: '同一租户' },
  { name: '业务类型一致性', status: 'passed', message: '收款单匹配销项发票' },
  { name: '到期日检查', status: 'blocked', message: '发票已过期（2026-04-20），需人工确认' },
])

const checkResults = computed(() => checks.value)
const blockerCount = computed(() => checks.value.filter(c => c.status === 'blocked').length)

function runPrecheck() {
  precheckDone.value = true
  ElMessage.success('预检完成')
}
</script>
<style scoped>
.page-header h3 { font-size: 18px; margin-bottom: 16px; }
.action-col { display: flex; align-items: flex-end; padding-bottom: 20px; }
.result-card { margin-top: 16px; }
.precheck-actions { margin-top: 16px; display: flex; align-items: center; gap: 12px; }
.blocker-msg { color: #ff4d4f; font-size: 13px; }
</style>
