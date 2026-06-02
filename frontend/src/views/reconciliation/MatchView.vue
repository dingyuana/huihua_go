<template>
  <div class="match-view">
    <div class="page-header">
      <h3>智能匹配推荐</h3>
    </div>
    <el-card>
      <div class="payment-info">
        <el-descriptions
          :column="3"
          size="small"
          border
        >
          <el-descriptions-item label="收款单">
            {{ paymentInfo.no || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="对方">
            {{ paymentInfo.counterparty || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="金额">
            {{ paymentInfo.amount || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- L1: 精确匹配 -->
      <div class="match-level level-1">
        <h4>⭐ L1 精确匹配 — 自动核销</h4>
        <el-table
          :data="l1Matches"
          size="small"
          border
        >
          <el-table-column
            prop="invoice_no"
            label="发票号"
            width="140"
          />
          <el-table-column
            prop="customer"
            label="对方"
            width="140"
          />
          <el-table-column
            prop="amount"
            label="金额"
            width="120"
          />
          <el-table-column
            label="核销后"
            width="120"
          >
            <template #default="{ row }">
              {{ row.after }}
            </template>
          </el-table-column>
          <el-table-column
            label="状态"
            width="80"
          >
            <template #default>
              <el-tag
                size="small"
                type="success"
              >
                自动 ✅
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- L2: FIFO 匹配 -->
      <div class="match-level level-2">
        <h4>📋 L2 FIFO 先进先出 — 自动分配</h4>
        <el-table
          :data="l2Matches"
          size="small"
          border
        >
          <el-table-column
            prop="invoice_no"
            label="发票号"
            width="140"
          />
          <el-table-column
            prop="date"
            label="日期"
            width="80"
          />
          <el-table-column
            prop="outstanding"
            label="未结清"
            width="100"
            align="right"
          />
          <el-table-column
            prop="allocated"
            label="本次分配"
            width="100"
            align="right"
          />
          <el-table-column
            label="分配后"
            width="100"
            align="right"
          >
            <template #default="{ row }">
              {{ row.after }}
            </template>
          </el-table-column>
          <el-table-column
            label="状态"
            width="80"
          >
            <template #default>
              <el-tag
                size="small"
                type="success"
              >
                自动 ✅
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- L3: 模糊匹配 -->
      <div class="match-level level-3">
        <h4>🔍 L3 模糊匹配 — 需确认</h4>
        <el-table
          :data="l3Matches"
          size="small"
          border
        >
          <el-table-column
            prop="invoice_no"
            label="发票号"
            width="140"
          />
          <el-table-column
            label="相似度"
            width="100"
          >
            <template #default="{ row }">
              {{ row.similarity }}%
            </template>
          </el-table-column>
          <el-table-column
            prop="amount"
            label="金额"
            width="120"
          />
          <el-table-column
            label="操作"
            width="140"
          >
            <template #default>
              <el-button
                size="small"
                type="primary"
                @click="ElMessage.success('已确认')"
              >
                确认
              </el-button>
              <el-button
                size="small"
                @click="ElMessage.info('已忽略')"
              >
                忽略
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- L4: 小额尾差 -->
      <div class="match-level level-4">
        <h4>🪙 L4 小额尾差处理 — 自动核销</h4>
        <el-tag
          type="info"
          size="small"
        >
          阈值: ¥1.00（可配置）
        </el-tag>
        <el-table
          :data="l4Matches"
          size="small"
          border
          style="margin-top:8px"
        >
          <el-table-column
            prop="invoice_no"
            label="发票号"
            width="140"
          />
          <el-table-column
            label="尾差金额"
            width="120"
            align="right"
          >
            <template #default="{ row }">
              {{ row.amount }}
            </template>
          </el-table-column>
          <el-table-column
            label="状态"
            width="100"
          >
            <template #default>
              <el-tag
                size="small"
                type="warning"
              >
                尾差 🪙
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-button
        type="primary"
        class="confirm-all"
        @click="ElMessage.success('匹配完成')"
      >
        确认全部匹配
      </el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const paymentInfo = ref<{ no?: string; counterparty?: string; amount?: string }>({})

const l1Matches = ref<any[]>([])
const l2Matches = ref<any[]>([])
const l3Matches = ref<any[]>([])
const l4Matches = ref<any[]>([])
</script>
<style scoped>
.page-header h3 { font-size: 18px; margin-bottom: 16px; }
.payment-info { margin-bottom: 20px; }
.match-level { margin-bottom: 20px; padding: 16px; border: 1px solid #e8e8e8; border-radius: 6px; }
.match-level h4 { margin-bottom: 8px; font-size: 14px; }
.level-1 { border-left: 3px solid #52c41a; }
.level-1 h4 { color: #52c41a; }
.level-2 { border-left: 3px solid #1890ff; }
.level-2 h4 { color: #1890ff; }
.level-3 { border-left: 3px solid #faad14; }
.level-3 h4 { color: #faad14; }
.level-4 { border-left: 3px solid #909399; }
.level-4 h4 { color: #909399; }
.confirm-all { margin-top: 8px; }
</style>
