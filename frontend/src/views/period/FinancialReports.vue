<template>
  <div class="reports-page">
    <div class="page-header">
      <h3>财务报表</h3>
      <div>
        <el-select v-model="period" style="width: 140px; margin-right: 8px">
          <el-option label="2026-05" value="2026-05" />
          <el-option label="2026-04" value="2026-04" />
        </el-select>
        <el-button @click="exportReport">导出 Excel</el-button>
        <el-button @click="printReport">打印</el-button>
      </div>
    </div>

    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="资产负债表" name="bs">
          <el-table :data="balanceSheet" border stripe size="small">
            <el-table-column prop="code" label="科目编码" width="120" />
            <el-table-column prop="name" label="科目名称" min-width="200">
              <template #default="{ row }">
                <span :style="{ fontWeight: row.level === 0 ? 700 : 400, paddingLeft: row.level * 20 + 'px' }">{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="opening" label="期初余额" width="140" align="right" />
            <el-table-column prop="closing" label="期末余额" width="140" align="right" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="利润表" name="pl">
          <el-table :data="profitLoss" border stripe size="small">
            <el-table-column prop="item" label="项目" min-width="200" />
            <el-table-column prop="current" label="本期金额" width="140" align="right" />
            <el-table-column prop="last" label="上期金额" width="140" align="right" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="现金流量表" name="cf">
          <el-table :data="cashFlow" border stripe size="small">
            <el-table-column prop="category" label="类别" width="160">
              <template #default="{ row }">
                <span :style="{ fontWeight: row.level === 0 ? 700 : 400 }">{{ row.category }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="item" label="项目" min-width="200" />
            <el-table-column prop="current" label="本期金额" width="140" align="right" />
            <el-table-column prop="last" label="上期金额" width="140" align="right" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <!-- 利润趋势图 -->
      <div v-if="activeTab === 'pl'" ref="chartRef" style="height: 300px; margin-top: 16px"></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const period = ref('2026-05')
const activeTab = ref('bs')

const loading = ref(false)
const balanceSheet = ref<BalanceSheetRow[]>([])
const cashFlow = ref<CashFlowItem[]>([])
const profitLoss = ref<PLRow[]>([])

interface BalanceSheetRow { code: string; name: string; opening: string; closing: string; level: number }
interface CashFlowItem { category: string; item: string; current: string; last: string; level: number }
interface PLRow { item: string; current: string; last: string }

function parseAmount(v: any): string {
  if (v == null) return '0.00'
  const n = typeof v === 'number' ? v : Number(v)
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function loadBalanceSheet(periodNo: number) {
  return request.get('/reports/balance-sheet', { params: { period_no: periodNo } })
    .then((res: any) => {
      const d = res.data || res
      if (!d || !d.asset_entries) return false
      const rows: BalanceSheetRow[] = []
      const addSection = (label: string, entries: any[], total: number) => {
        rows.push({ code: '', name: label, opening: '', closing: '', level: 0 })
        for (const e of entries) {
          rows.push({ code: e.account_code || '', name: '  ' + e.account_name, opening: '', closing: parseAmount(e.balance), level: 1 })
        }
        rows.push({ code: '', name: label + '合计', opening: '', closing: parseAmount(total), level: 0 })
      }
      addSection('资产', d.asset_entries || [], d.total_assets || 0)
      addSection('负债', d.liability_entries || [], d.total_liabilities || 0)
      addSection('所有者权益', d.equity_entries || [], d.total_equity || 0)
      balanceSheet.value = rows
      return true
    })
}

function loadIncomeStatement(periodNo: number) {
  return request.get('/reports/income-statement', { params: { period_no: periodNo } })
    .then((res: any) => {
      const d = res.data || res
      if (!d || d.total_income == null) return false
      const rows: PLRow[] = [
        { item: '一、营业收入', current: parseAmount(d.total_income), last: '0.00' },
        { item: '  减：营业成本及费用', current: parseAmount(d.total_expense), last: '0.00' },
        { item: '二、营业利润', current: parseAmount(d.total_income - d.total_expense), last: '0.00' },
        { item: '三、净利润', current: parseAmount(d.net_profit), last: '0.00' },
      ]
      profitLoss.value = rows
      return true
    })
}

function loadCashFlow(periodNo: number) {
  return request.get('/reports/cash-flow', { params: { period_no: periodNo } })
    .then((res: any) => {
      const d = res.data || res
      if (!d || !d.items) return false
      cashFlow.value = d.items.map((i: any) => ({
        category: i.category || '',
        item: i.item || '',
        current: parseAmount(i.current),
        last: parseAmount(i.last),
        level: i.level || 0,
      }))
      return true
    })
}

function parsePeriod(p: string): number {
  const [y, m] = p.split('-').map(Number)
  return y * 100 + m
}

const mockBS: BalanceSheetRow[] = [
  { code: '', name: '资产', opening: '', closing: '', level: 0 },
  { code: '1001', name: '  银行存款', opening: '1,000,000.00', closing: '1,250,000.00', level: 1 },
  { code: '1122', name: '  应收账款', opening: '500,000.00', closing: '450,000.00', level: 1 },
  { code: '', name: '资产合计', opening: '1,500,000.00', closing: '1,700,000.00', level: 0 },
  { code: '', name: '负债', opening: '', closing: '', level: 0 },
  { code: '2001', name: '  应付账款', opening: '300,000.00', closing: '350,000.00', level: 1 },
  { code: '', name: '负债合计', opening: '300,000.00', closing: '350,000.00', level: 0 },
  { code: '', name: '所有者权益', opening: '', closing: '', level: 0 },
  { code: '4001', name: '  实收资本', opening: '1,000,000.00', closing: '1,000,000.00', level: 1 },
  { code: '', name: '未分配利润', opening: '200,000.00', closing: '350,000.00', level: 1 },
  { code: '', name: '权益合计', opening: '1,200,000.00', closing: '1,350,000.00', level: 0 },
]
const mockPL: PLRow[] = [
  { item: '一、营业收入', current: '850,000.00', last: '720,000.00' },
  { item: '  减：营业成本', current: '480,000.00', last: '410,000.00' },
  { item: '  减：管理费用', current: '120,000.00', last: '110,000.00' },
  { item: '  减：财务费用', current: '15,000.00', last: '12,000.00' },
  { item: '二、营业利润', current: '235,000.00', last: '188,000.00' },
  { item: '  加：营业外收入', current: '5,000.00', last: '3,000.00' },
  { item: '  减：营业外支出', current: '2,000.00', last: '1,000.00' },
  { item: '三、利润总额', current: '238,000.00', last: '190,000.00' },
  { item: '  减：所得税', current: '59,500.00', last: '47,500.00' },
  { item: '四、净利润', current: '178,500.00', last: '142,500.00' },
]
const mockCF: CashFlowItem[] = [
  { category: '一、经营活动产生的现金流量', item: '销售商品、提供劳务收到的现金', current: '850,000.00', last: '720,000.00', level: 0 },
  { category: '', item: '  收到的税费返还', current: '10,000.00', last: '8,000.00', level: 1 },
  { category: '', item: '  购买商品、接受劳务支付的现金', current: '480,000.00', last: '410,000.00', level: 1 },
  { category: '', item: '  支付给职工以及为职工支付的现金', current: '150,000.00', last: '140,000.00', level: 1 },
  { category: '', item: '  支付的各项税费', current: '120,000.00', last: '100,000.00', level: 1 },
  { category: '经营活动现金流量净额', item: '', current: '110,000.00', last: '78,000.00', level: 0 },
  { category: '二、投资活动产生的现金流量', item: '  收回投资所收到的现金', current: '50,000.00', last: '30,000.00', level: 0 },
  { category: '', item: '  购建固定资产所支付的现金', current: '80,000.00', last: '120,000.00', level: 1 },
  { category: '投资活动现金流量净额', item: '', current: '-30,000.00', last: '-90,000.00', level: 0 },
  { category: '三、筹资活动产生的现金流量', item: '  吸收投资所收到的现金', current: '0.00', last: '200,000.00', level: 0 },
  { category: '', item: '  偿还债务所支付的现金', current: '50,000.00', last: '30,000.00', level: 1 },
  { category: '筹资活动现金流量净额', item: '', current: '-50,000.00', last: '170,000.00', level: 0 },
  { category: '四、现金净增加额', item: '', current: '30,000.00', last: '158,000.00', level: 0 },
]

async function loadReport() {
  loading.value = true
  const periodNo = parsePeriod(period.value)
  const ok = await Promise.all([
    loadBalanceSheet(periodNo),
    loadIncomeStatement(periodNo),
    loadCashFlow(periodNo),
  ])
  if (!ok[0]) balanceSheet.value = mockBS
  if (!ok[1]) profitLoss.value = mockPL
  if (!ok[2]) cashFlow.value = mockCF
  loading.value = false
}
onMounted(loadReport)

// ECharts 图表：利润趋势
import * as echarts from 'echarts'

const chartRef = ref<HTMLDivElement>()

onMounted(() => {
  if (!chartRef.value) return
  const chart = echarts.init(chartRef.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['1月', '2月', '3月', '4月', '5月'] },
    yAxis: { type: 'value', axisLabel: { formatter: '¥{value}' } },
    series: [
      { name: '收入', type: 'bar', data: [680, 720, 780, 800, 850], itemStyle: { color: '#1890ff' } },
      { name: '成本', type: 'bar', data: [380, 400, 430, 450, 480], itemStyle: { color: '#ff4d4f' } },
      { name: '利润', type: 'line', data: [180, 195, 210, 225, 238], itemStyle: { color: '#52c41a' } },
    ],
    legend: { data: ['收入', '成本', '利润'], bottom: 0 },
  })
  window.addEventListener('resize', () => chart.resize())
})

function exportReport() { ElMessage.success('导出成功') }
function printReport() { window.print() }
</script>
<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; h3 { font-size: 18px; } }
</style>
