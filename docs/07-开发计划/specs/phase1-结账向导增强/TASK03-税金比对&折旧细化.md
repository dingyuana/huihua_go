# SPEC: TASK03 — 税金计提比对 + 折旧金额明细

## 基本信息

- **任务 ID**: phase1-close-003
- **类型**: feature
- **优先级**: medium
- **依赖**: phase0-close-001（check_pending_accruals 已存在）
- **执行者**: OpenCode
- **估算**: 后端 90 行 + 前端 60 行 + 测试 40 行

## 背景

当前 `check_pending_accruals` 返回的折旧/税金项只有布尔值 `missing: true/false`：
- 对折旧：只能判断"有没有计提"，不能看到"应该提多少、已经提了多少"
- 对税金：只能判断"有没有申报"，不能看到"实缴 vs 应计提差额"

SOP 第 3 步（资产与存货）和第 4 步（税金复核）要求能看到具体金额差异，而不是只有"有/无"。

## 目标

1. 增强折旧检查：返回应计提金额、已计提金额、差额
2. 增强税金检查：比较 tax_declarations 实缴额与科目余额应计提额
3. 前端 pending_accruals 区域展示金额明细

## 技术约束

- 不新增数据库模型（复用现有的 AccAccount / FixedAsset / AssetDepreciation / TxDeclaration）
- 不新增路由（扩展 pre_close_check 返回中的 pending_accruals 结构）
- 不修改 pending_accruals 现有字段（仅追加新字段）
- 上月无数据时金额显示 null

## 详细设计

### 1. 增强后的 pending_accruals 结构

```json
{
  "pending_accruals": [
    {
      "type": "depreciation",
      "item": "本月固定资产折旧",
      "missing": false,
      "details": "存在2项使用中资产未计提本月折旧",
      // 新增字段 ↓
      "expected_amount": 12500.00,
      "actual_amount": 8500.00,
      "difference": 4000.00,
      "items": [
        { "asset_name": "服务器", "expected": 5000.00, "actual": 5000.00 },
        { "asset_name": "办公桌椅", "expected": 2500.00, "actual": 0 },
        { "asset_name": "车辆", "expected": 5000.00, "actual": 3500.00 }
      ]
    },
    {
      "type": "tax",
      "item": "城市维护建设税及附加",
      "missing": false,
      // 新增字段 ↓
      "expected_amount": 1500.00,
      "actual_amount": 1500.00,
      "difference": 0,
      "tax_details": {
        "tax_base": 50000.00,
        "tax_rate": "7%+3%+2%",
        "declared_amount": 1500.00,
        "accrued_amount": 1500.00
      }
    },
    {
      "type": "income_tax",
      "item": "企业所得税",
      "missing": true,
      "expected_amount": 8500.00,
      "actual_amount": 0,
      "difference": 8500.00,
      "details": "本期有利润56,666.67元，预计应计提所得税8,500元但尚未计提"
    }
  ]
}
```

### 2. 折旧检查增强

**当前逻辑**（`check_pending_accruals` 函数中的折旧部分）：

```python
# 当前：只检查 "有没有"
assets_in_use = db.query(FixedAsset).filter(
    FixedAsset.depreciation_status == "using"
).all()
if assets_in_use:
    depreciation_this_month = db.query(AssetDepreciation).filter(
        AssetDepreciation.period_year == year,
        AssetDepreciation.period_month == month,
    ).first()
    missing = depreciation_this_month is None
```

**增强后逻辑**：

```python
assets_in_use = db.query(FixedAsset).filter(
    FixedAsset.depreciation_status == "using"
).all()

if not assets_in_use:
    # 无使用中资产 → 跳过
    continue

# 计算应计提总额
total_expected = Decimal("0")
items = []
for asset in assets_in_use:
    # 月折旧额 = 原值 / (预计年限 * 12)
    if asset.expected_years and asset.expected_years > 0:
        monthly = Decimal(str(asset.purchase_amount or "0")) / Decimal(asset.expected_years * 12)
    else:
        continue
    total_expected += monthly

    # 查询当月是否已计提
    dep = db.query(AssetDepreciation).filter(
        AssetDepreciation.asset_id == asset.id,
        AssetDepreciation.period_year == year,
        AssetDepreciation.period_month == month,
    ).first()

    items.append({
        "asset_name": asset.asset_name,
        "expected": round(float(monthly), 2),
        "actual": round(float(dep.depreciation_amount), 2) if dep else 0,
    })

# 已计提总额
total_actual = sum(i["actual"] for i in items)

result_item = {
    "type": "depreciation",
    "item": "本月固定资产折旧",
    "missing": total_actual < total_expected,
    "expected_amount": round(float(total_expected), 2),
    "actual_amount": round(float(total_actual), 2),
    "difference": round(float(total_expected - total_actual), 2),
    "items": items,
    "details": f"应计提 {total_expected:.2f} 元，已计提 {total_actual:.2f} 元"
}
```

### 3. 税金检查增强

**当前逻辑**：检查 tax_declarations 表是否有该月的申报记录。
**增强后逻辑**：比较申报表实缴额与科目余额应计提额。

```python
def _check_tax_accrual_detail(db, year, month):
    """
    比较增值税及附加的应计提 vs 实际申报。

    应计提口径：
    - 增值税：取 AccAccount 中 2221(应交税费) 的期末贷方余额
      → 当月应交增值税 = closing_balance
    - 城建税: 应交增值税 * 7%（企业所在城市税率）
    - 教育费附加: 应交增值税 * 3%
    - 地方教育附加: 应交增值税 * 2%

    实际申报口径：
    - 查询 tax_declarations 表该月的申报记录，sum(tax_amount)

    边界：
    - 应交增值税为 0 或负数 → 不计提附加税，标记为无需计提
    - 无上月数据时 last_value=null
    """
    # 取应交增值税期末余额
    vat_account = db.query(AccAccount).join(AccSubject).filter(
        AccAccount.year == year,
        AccAccount.month == month,
        AccSubject.subject_code == "2221",
    ).first()

    if not vat_account or Decimal(str(vat_account.closing_balance or "0")) <= 0:
        return  # 无需计提

    vat_amount = Decimal(str(vat_account.closing_balance))
    expected_surtax = vat_amount * Decimal("0.12")  # 7%+3%+2%

    # 查询实际已申报的附加税
    actual = db.query(func.sum(TxDeclaration.tax_amount)).filter(
        TxDeclaration.period_year == year,
        TxDeclaration.period_month == month,
    ).scalar() or Decimal("0")

    # 企业所得税检查（如果利润为正）
    profit = _get_period_profit(db, year, month)
    income_tax_items = []
    if profit > 0:
        expected_income_tax = profit * Decimal("0.25")  # 25% 税率
        actual_income_tax = ...  # 查询所得税申报
        income_tax_items = [...]
```

### 4. 前端展示

PeriodClose.vue 的 "待办事项" 区域增强：

```html
<el-card v-if="hasAccrualDetails" class="accrual-section">
  <template #header>
    <span class="section-title">计提检查明细</span>
    <el-tag v-if="allAccrualsDone" type="success">全部完成</el-tag>
    <el-tag v-else type="warning">{{ pendingCount }} 项待处理</el-tag>
  </template>

  <!-- 折旧明细 -->
  <div v-for="item in accrualDetails" :key="item.type" class="accrual-group">
    <div class="accrual-header">
      <span>{{ item.item }}</span>
      <span v-if="item.missing" class="accrual-warn">未完成</span>
      <span v-else class="accrual-ok">已完成</span>
    </div>

    <!-- 金额对比行 -->
    <div class="accrual-amount-row" v-if="item.expected_amount != null">
      <span>应计提：¥{{ formatAmount(item.expected_amount) }}</span>
      <span>已计提：¥{{ formatAmount(item.actual_amount) }}</span>
      <span v-if="item.difference > 0" class="accrual-diff">
        差额：¥{{ formatAmount(item.difference) }}
      </span>
    </div>

    <!-- 折旧明细表 -->
    <el-table v-if="item.items?.length" :data="item.items" size="mini">
      <el-table-column prop="asset_name" label="资产" />
      <el-table-column prop="expected" label="应提" align="right" />
      <el-table-column prop="actual" label="已提" align="right" />
    </el-table>
  </div>
</el-card>
```

## 测试策略

### 单元测试

| 场景 | 数据 | 期望 |
|:---|:---|:---|
| 所有资产已计提 | 2项资产都提了折旧 | missing=false, difference=0 |
| 部分资产未计提 | 2项资产只有1项提了 | missing=true, difference=应提差额 |
| 无使用中资产 | 所有资产已报废/未使用 | 不返回折旧项 |
| 增值税0 | 无应交增值税 | 不返回附加税项 |
| 利润为正但未计提所得税 | 有利润但无所得税申报 | income_tax missing=true |
| 利润为负 | 亏损 | 不返回所得税项 |

### 集成测试

- `test_pending_accruals_depreciation_detail`：验证折旧金额明细
- `test_pending_accruals_tax_comparison`：验证税金比对

## 验收标准

- [ ] depreciation 项返回 expected_amount / actual_amount / difference
- [ ] 折旧明细含每项资产的应提和已提金额
- [ ] 有增值税时返回附加税应计提 vs 已申报对比
- [ ] 利润为正时返回所得税检查项
- [ ] 所有检查项通过时 pending_accruals 中 missing=false
- [ ] 前端展示金额明细表格
- [ ] 无待处理项时显示"全部完成"
- [ ] 56 个现有测试全部通过

## 不变清单

- 不修改 FixedAsset / AssetDepreciation 模型
- 不修改 TxDeclaration 模型
- 不修改 AccAccount 模型
- 不修改 pre_close_check 的 risk_warnings / key_indicators 字段
- 不修改凭证状态机
