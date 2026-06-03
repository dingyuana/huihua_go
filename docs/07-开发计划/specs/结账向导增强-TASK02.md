# SPEC: TASK02 — 前端重构 PeriodClose 页面（结账向导）

## 基本信息

- **任务 ID**: close-wizard-002
- **类型**: feature
- **优先级**: high
- **依赖**: TASK01（后端返回增强后，前端才能展示新字段）
- **执行者**: OpenCode

## 背景

当前 PeriodClose 页面布局为4张并列卡片（凭证统计、结账检查、损益结转、结账操作），只展示了3个检查项的绿色/红色状态勾。缺少：

1. **风险预警区域** — 应收重分类、计提缺失等高亮展示
2. **关键指标区域** — 毛利率、费用率环比变化
3. **人工确认清单** — 结账前会计需逐项确认的事项
4. **视觉层次** — 从"并列卡片"改为"自上而下的结构化工序"

## 目标

将 PeriodClose 页面从4卡片布局重构为6区域上下布局的**结账向导**，展示后端返回的所有新字段。使用 Element Plus 现有组件，不引入新UI库。

## 技术约束

- 只修改 `frontend/src/pages/accounting/PeriodClose.vue`
- 不修改 Element Plus 版本或引入新依赖
- 不修改路由
- 所有 API 调用在 `frontend/src/api/report.js` 中只新增类型注释（不修改现有函数）
- 风险预警不自动修改数据，只展示提示

## 详细设计

### 页面结构（从上到下6个区域）

```
┌─ 期间选择 ──────────────────────────────┐
│  [2026年5月  ▼]  [查询状态]             │
└──────────────────────────────────────────┘

┌─ 一、基础检查 ─────────────────────────┐
│  ☑ 全部凭证已记账 (58张已记账 / 0张未记)│
│  ☑ 资产负债表平衡 (差额: 0.00)         │
│  ☑ 损益已结转                         │
│  ☑ 折旧已计提 ← 新增                   │
└──────────────────────────────────────────┘

┌─ 二、风险预警 ────（带颜色区分）───────┐
│  ⚠️ 应收账款-XX客户 贷方余额15,000元   │
│     → 系统建议重分类至预收账款          │
│                                         │
│  ❌ 折旧未计提                         │
│     → 存在2项使用中资产未计提本月折旧   │
└──────────────────────────────────────────┘

┌─ 三、关键指标 ───────────────────────┐
│  指标            本期    上月    预警  │
│  ──────────────────────────────────  │
│  毛利率          15%     22%    🔴   │
│  期间费用率       18%    17.5%   🟢   │
│  营业利润率        5%      8%    🟡   │
└──────────────────────────────────────────┘

┌─ 四、人工确认 ──────────────────────┐
│  ☐ 已核对银行余额调节表              │
│  ☐ 已确认往来款余额                  │
│  ☐ 已确认税金计提完整                │
│  ☐ 已核实关键指标异常原因            │
└──────────────────────────────────────────┘

┌─ 五、损益结转 ──────────────────────┐
│  [预览结转分录]  [结转损益]          │
│  收入: ¥XX  费用: ¥XX  利润: ¥XX   │
└──────────────────────────────────────────┘

┌─ 六、结账操作 ──────────────────────┐
│  [月末结账]  [反结账]                │
│  状态：未结账 / 已结账 (2026-05-31) │
└──────────────────────────────────────────┘
```

### 区域1：期间选择（保持现有）

使用现有的 `<el-date-picker type="month">` + 查询按钮。

### 区域2：基础检查（增强）

在现有3项的基础上，追加第4项"折旧已计提"：

```html
<el-card>
  <div class="check-item" :class="item.passed ? 'passed' : 'failed'">
    <el-icon><CircleCheck v-if="item.passed" /><CircleClose v-else /></el-icon>
    <span>{{ item.name }}</span>
    <span v-if="item.hint" class="check-item-hint">{{ item.hint }}</span>
  </div>
  <!-- 追加折旧项：根据 pending_accruals 中 type=depreciation 的 missing 值判断 -->
  <div class="check-item" :class="depreciationPassed ? 'passed' : 'failed'">
    ...
  </div>
</el-card>
```

### 区域3：风险预警（新增）

根据 `risk_warnings` 和 `pending_accruals` 渲染预警列表：

- severity=cirtical → 红色底色 + ❌ 图标
- severity=warning → 黄色底色 + ⚠️ 图标
- severity=info → 灰色底色 + ℹ️ 图标
- 列表为空时显示 ✅ 无风险预警

**此区域只在有预警（risk_warnings非空 或 pending_accruals有missing=true）时渲染**。

### 区域4：关键指标（新增）

使用 `<el-table>` 展示指标对比：

| 列 | 数据来源 |
|----|---------|
| 指标 | key_indicators[].name |
| 本期 | key_indicators[].current_value + unit |
| 上月 | key_indicators[].last_value + unit（null时显示-） |
| 预警 | alert=true → 🔴 / false → 🟢 |

无需图表，表格足够。

### 区域5：人工确认清单（新增）

4个勾选框，全部勾选后 `canClose` 计算属性才返回 true：

```js
const confirmItems = ref([
  { key: 'bank_reconciled', label: '已核对银行余额调节表', checked: false },
  { key: 'receivable_confirmed', label: '已确认往来款余额', checked: false },
  { key: 'tax_accrual_confirmed', label: '已确认税金计提完整', checked: false },
  { key: 'indicator_reviewed', label: '已核实关键指标异常原因', checked: false },
])

const canClose = computed(() => 
  confirmItems.value.every(i => i.checked) &&
  checkItemsPassed() &&
  profitLossDone
)
```

### 区域6：损益结转（调整样式）

从卡片改为行内样式，精简宽度。保持现有逻辑。

### 区域7：结账操作（调整样式）

从卡片改为行内，保持现有逻辑。

### 关键交互规则

1. **人工确认清单全部勾选后**，月末结账按钮才可用（与基础检查通过叠加）
2. **风险预警不阻断结账**（只展示提示，不修改数据）
3. **pending_accruals 中 missing=true 的项**加入基础检查项，缺失时标记红色，但不强制阻断
4. key_indicators 中 alert=true 的项，在人工确认清单中对应第4项"已核实关键指标异常原因"
5. 空风险/无缺失时，区域3直接显示 ✅ 无风险预警

### 状态管理

在 `loadPeriodStatus()` 中，从后端获取扩展后的数据后，拆解为前端响应式变量：

```js
// 现有
const status = ref({ checkItems: [], profitLossStatus: null })
const currentPeriod = ref(null)

// 新增
const riskWarnings = ref([])       // risk_warnings[]
const keyIndicators = ref([])      // key_indicators[]
const pendingAccruals = ref([])    // pending_accruals[]
const depreciationPassed = ref(true)  // 从 pending_accruals 计算
```

## 验收标准

- [ ] 页面按6区上下布局展示，不重叠、滚动正常
- [ ] 有风险预警时，预警区域以颜色区分展示，无预警时显示 ✅ 无风险预警
- [ ] 关键指标区域展示表格，有环比数据，上月为空时显示"-"
- [ ] 人工确认4项全部勾选后，结账按钮可用
- [ ] 损益结转预览/执行/反结账功能正常
- [ ] 后端返回空数据（新账套）时，所有区域正常显示无报错
- [ ] 移动端/窄屏下布局自适应

## OpenCode 指令

**目标**：重构 `/root/huihua-financial-master/frontend/src/pages/accounting/PeriodClose.vue`，从4卡片布局改为6区域上下布局，展示后端 `pre_close_check` 新返回的 `risk_warnings`、`key_indicators`、`pending_accruals` 字段。

**约束**：
- 只修改 PeriodClose.vue
- 不新增 npm 依赖
- 使用 Element Plus 现有组件（el-card, el-table, el-checkbox, el-button, el-icon 等）
- 风险预警区域只读展示，不触发数据修改
- 现有功能（损益结转预览/执行/反结账）不受影响

**上下文**：
- repo: `/root/huihua-financial-master`
- 需修改的文件：`frontend/src/pages/accounting/PeriodClose.vue`
- 参考现有文件：`frontend/src/api/report.js`（API函数路径）
- API 返回值增强通过 TASK01 实现，本任务假设后端已返回新字段

**验收**：
- 页面按6区域布局展示
- risk_warnings 有数据时显示预警区域
- 人工确认4项勾选后可结账
- 损益结转和反结账功能正常
- 浏览器控制台无报错
