<template>
  <div class="balance-sheet-page">
    <div class="page-header">
      <h3>余额调节表</h3>
      <div>
        <el-select v-model="bankAccount" style="width: 240px; margin-right: 8px">
          <el-option label="工商银行-基本户 (1102****4567)" value="ba-1" />
          <el-option label="建设银行-一般户 (4302****4321)" value="ba-2" />
        </el-select>
        <el-button @click="exportPdf">导出 PDF</el-button>
      </div>
    </div>

    <!-- 余额概要 -->
    <el-row :gutter="16" class="balance-summary">
      <el-col :span="6">
        <el-card><p class="b-label">银行对账单余额</p><p class="b-value">¥1,250,000.00</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card><p class="b-label">企业日记账余额</p><p class="b-value">¥1,245,000.00</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card><p class="b-label">差额</p><p class="b-value diff">¥5,000.00</p></el-card>
      </el-col>
      <el-col :span="6">
        <el-card :class="isBalanced ? 'balanced' : 'unbalanced'">
          <p class="b-label">调节后余额</p>
          <p class="b-value">¥1,247,000.00</p>
          <p class="b-status">{{ isBalanced ? '✅ 平衡' : '❌ 不平衡' }}</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 四类差异 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card class="diff-card">
          <template #header><span class="diff-title bank-receipt">🏦 银行已收企业未达</span></template>
          <el-table :data="bankReceiptNotInGL" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="diff-card">
          <template #header><span class="diff-title bank-payment">🏦 银行已付企业未达</span></template>
          <el-table :data="bankPaymentNotInGL" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card class="diff-card">
          <template #header><span class="diff-title gl-receipt">📒 企业已收银行未达</span></template>
          <el-table :data="glReceiptNotInBank" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="diff-card">
          <template #header><span class="diff-title gl-payment">📒 企业已付银行未达</span></template>
          <el-table :data="glPaymentNotInBank" size="small" border>
            <el-table-column prop="date" label="日期" width="80" />
            <el-table-column prop="desc" label="摘要" min-width="140" />
            <el-table-column prop="amount" label="金额" width="100" align="right" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 调节计算 -->
    <el-card class="calc-card">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="银行对账单余额">¥1,250,000.00</el-descriptions-item>
        <el-descriptions-item label="企业日记账余额">¥1,245,000.00</el-descriptions-item>
        <el-descriptions-item label="+ 银行已收企业未达">¥5,000.00</el-descriptions-item>
        <el-descriptions-item label="+ 企业已收银行未达">¥0.00</el-descriptions-item>
        <el-descriptions-item label="- 银行已付企业未达">¥0.00</el-descriptions-item>
        <el-descriptions-item label="- 企业已付银行未达">¥2,000.00</el-descriptions-item>
        <el-descriptions-item label="调整后银行余额" :span="2">
          <b class="balanced-text">¥1,247,000.00</b>
        </el-descriptions-item>
        <el-descriptions-item label="调整后企业余额" :span="2">
          <b class="balanced-text">¥1,247,000.00</b>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <div class="actions">
      <el-button type="primary" @click="confirmAndLock">确认并锁定对账</el-button>
      <el-button @click="$router.push('/bank-reconciliation/match')">返回对账</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const bankAccount = ref('ba-1')
const isBalanced = ref(true)

const bankReceiptNotInGL = ref([
  { date: '2026-05-30', desc: '银行利息收入', amount: '5,000.00' },
])

const bankPaymentNotInGL = ref([] as any[])

const glReceiptNotInBank = ref([] as any[])

const glPaymentNotInBank = ref([
  { date: '2026-05-28', desc: '在途付款-供应商', amount: '2,000.00' },
])

function exportPdf() { ElMessage.success('余额调节表已导出') }

function confirmAndLock() {
  ElMessage.success('对账结果已确认并锁定')
}
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
.balance-summary { margin-bottom: 16px; }
.b-label { font-size: 12px; color: #999; margin-bottom: 4px; }
.b-value { font-size: 22px; font-weight: 700; &.diff { color: #faad14; } }
.b-status { font-size: 13px; margin-top: 4px; }
.balanced .b-status { color: #52c41a; }
.unbalanced .b-status { color: #ff4d4f; }
.diff-card { margin-bottom: 0; }
.diff-title { font-weight: 600; font-size: 13px; }
.diff-title.bank-receipt { color: #52c41a; }
.diff-title.bank-payment { color: #ff4d4f; }
.diff-title.gl-receipt { color: #1890ff; }
.diff-title.gl-payment { color: #faad14; }
.calc-card { margin-top: 16px; }
.balanced-text { color: #52c41a; }
.actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
</style>
