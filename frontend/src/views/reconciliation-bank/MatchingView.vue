<template>
  <div class="bank-rec">
    <div class="page-header">
      <h3>银企对账</h3>
      <div>
        <el-select v-model="bankAccount" style="width: 240px; margin-right: 8px">
          <el-option label="工商银行-基本户 (1102****4567)" value="ba-1" />
          <el-option label="建设银行-一般户 (4302****4321)" value="ba-2" />
        </el-select>
        <el-button type="primary" @click="runMatch">执行对账</el-button>
      </div>
    </div>

    <!-- 统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6"><el-card><p class="stat-val">{{ stats.total }}</p><p class="stat-lbl">总笔数</p></el-card></el-col>
      <el-col :span="6"><el-card><p class="stat-val success">{{ stats.autoMatched }}</p><p class="stat-lbl">自动勾兑</p></el-card></el-col>
      <el-col :span="6"><el-card><p class="stat-val warning">{{ stats.needConfirm }}</p><p class="stat-lbl">待确认</p></el-card></el-col>
      <el-col :span="6"><el-card><p class="stat-val danger">{{ stats.unmatched }}</p><p class="stat-lbl">未匹配</p></el-card></el-col>
    </el-row>

    <!-- 自动勾兑率 -->
    <el-card class="rate-card">
      <span>自动勾兑率：</span>
      <el-progress :percentage="autoMatchRate" :color="autoMatchRate > 90 ? '#52c41a' : autoMatchRate > 80 ? '#faad14' : '#ff4d4f'" />
    </el-card>

    <!-- Tab -->
    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="`自动匹配 (${stats.autoMatched})`" name="auto" />
        <el-tab-pane :label="`待确认 (${stats.needConfirm})`" name="confirm" />
        <el-tab-pane :label="`未匹配 (${stats.unmatched})`" name="unmatched" />
      </el-tabs>

      <el-table :data="matchList" border stripe size="small">
        <el-table-column label="匹配得分" width="100">
          <template #default="{ row }">
            <el-tag :type="row.score >= 85 ? 'success' : row.score >= 60 ? 'warning' : 'danger'" size="small">{{ row.score }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bank_txn" label="银行流水" min-width="200" />
        <el-table-column prop="gl_entry" label="GL 条目" min-width="200" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button v-if="row.needConfirm" size="small" type="primary" @click="confirmMatch">确认</el-button>
            <el-tag v-else size="small" type="success">已匹配</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="rec-actions">
      <el-button @click="$router.push('/bank-reconciliation/balance')">查看余额调节表</el-button>
      <el-button type="primary" @click="lockReconciliation">锁定对账结果</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const bankAccount = ref('ba-1')
const activeTab = ref('auto')

const stats = computed(() => ({
  total: 150, autoMatched: 130, needConfirm: 12, unmatched: 8,
}))
const autoMatchRate = computed(() => Math.round(stats.value.autoMatched / stats.value.total * 100))

const matchList = ref([
  { score: 95, bank_txn: '05-20 收款 +12,000 上海XX', gl_entry: '日记账 +12,000', needConfirm: false },
  { score: 88, bank_txn: '05-21 付款 -5,000 北京YY', gl_entry: '日记账 -5,000', needConfirm: false },
  { score: 72, bank_txn: '05-22 收款 +3,000', gl_entry: '日记账 +2,900', needConfirm: true },
  { score: 35, bank_txn: '05-23 付款 -200', gl_entry: '无匹配', needConfirm: true },
])

function runMatch() { ElMessage.success('对账完成，自动勾兑率 86.7%') }
function confirmMatch() { ElMessage.success('匹配已确认') }
function lockReconciliation() { ElMessage.success('对账结果已锁定') }
</script>
<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.stat-row { margin-bottom: 16px; }
.stat-val { font-size: 24px; font-weight: 700; margin-bottom: 4px; &.success { color: #52c41a; } &.warning { color: #faad14; } &.danger { color: #ff4d4f; } }
.stat-lbl { font-size: 13px; color: #999; }
.rate-card { margin-bottom: 16px; display: flex; align-items: center; gap: 16px; }
.rec-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
</style>
