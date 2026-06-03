# SPEC: phase2-balance-reconciliation — 资产负债表不平衡诊断

## 基本信息

- **任务 ID**: phase2-balance-reconciliation
- **类型**: feature
- **优先级**: high
- **阶段**: Phase 2（紧接 Phase1 预警可交互之后）
- **执行者**: OpenCode
- **依赖**: phase0-close-001（pre_close_check 已有 report_balance_ok）
- **估算**: 后端 150 行 + 前端 180 行 + 测试 60 行

## 背景

当前 `pre_close_check` 的报表平衡检查只返回布尔值 `report_balance_ok: false`，没有任何诊断信息。用户看到"报表不平衡"提示后，无法知道是哪个科目、哪张单据出了问题。

**实际场景**：
- 资产 ¥-39,200 ≠ 负债 ¥21,000 + 权益 ¥-1,400
- 用户需要知道：哪个科目方向反了、哪张凭证导致异常

## 目标

当 `report_balance_ok=false` 时，在 `pre_close_check` 返回体中新增 `balance_reconciliation` 字段，包含：
1. 每个科目的余额贡献量（资产/负债/权益各三大类各自列出）
2. 异常科目检测（方向反了、数值异常）
3. 关联凭证列表（导致异常的可疑单据）

---

## 技术设计

### 1. 后端：`balance_reconciliation` 返回结构

在 `pre_close_check` 返回体中新增字段（当 `report_balance_ok=false` 时填充，平衡时为 `null`）：

```json
{
  "report_balance_ok": false,
  "balance_reconciliation": {
    "status": "unbalanced",
    "assets_total": -39200.00,
    "liabilities_total": 21000.00,
    "equity_total": -1400.00,
    "imbalance_amount": -40000.00,
    "asset_items": [...],
    "liability_items": [...],
    "equity_items": [...],
    "anomalies": [...]
  }
}
```

当 `report_balance_ok=true` 时：
```json
{
  "report_balance_ok": true,
  "balance_reconciliation": null
}
```

### 2. 分类计算逻辑（沿用现有逻辑，新增明细）

| 类别 | 科目前缀 | 计算规则 |
|---|---|---|
| 资产 | 1xx / 5xx | `total += closing`（借方余额正常，负值为异常） |
| 负债 | 2xx | `total += abs(closing)`（贷方余额正常，借方余额为异常） |
| 权益 | 3xx / 4xx / 6xx | `total += abs(closing)`（贷方余额正常，借方余额为异常） |

### 3. 异常检测规则（`anomalies` 列表）

每个异常科目返回结构：
```json
{
  "code": "1122",
  "name": "应收账款",
  "balance": -34800.00,
  "direction": "credit (normal: debit)",
  "impact": "资产虚减 34,800，权益虚减 34,800",
  "suggestion": "请检查是否有收款凭证方向填反，或应收已收款但未做对应科目对冲",
  "severity": "error",
  "related_vouchers": ["202605-001", "202605-003"]
}
```

**触发条件**：
1. **方向异常**：
   - 资产类科目（1xx）：`closing < 0` → 借方余额为负
   - 负债类科目（2xx）：`closing > 0` → 贷方余额为负（即借方正余额）
   - 权益类科目（3/4/6xx）：`closing < 0` → 贷方余额为负
2. **数值异常**：
   - 某科目 `|closing| > 0` 且 `< 1.00`，但涉及凭证 >= 3 张 → 可疑（小额异常）
3. **零星凭证**：某科目余额不为零，但只有 1 张已记账凭证且金额与余额接近 → 可能是该凭证方向填反

### 4. 关联凭证查询

对于每个 `anomalies` 中的科目，查询最近 10 张涉及该科目的凭证（按日期倒序）：
```sql
SELECT av.voucher_no, av.voucher_date, av.status, av.created_by_name,
       accd.debit_amount, accd.credit_amount
FROM acc_voucher_detail accd
JOIN acc_voucher av ON av.id = accd.voucher_id
WHERE accd.subject_code = :code
  AND av.period_year = :year AND av.period_month = :month
  AND av.status = 'posted'
ORDER BY av.voucher_date DESC
LIMIT 10
```

### 5. 新增辅助函数

在 `accounting.py` 中新建 `_balance_reconciliation()` 函数：

```python
def _balance_reconciliation(db: Session, year: int, month: int) -> dict | None:
    """
    当报表不平衡时，返回详细的对账诊断信息。
    报表平衡时返回 None（前端直接展示 ✅ 通过）。
    """
```

### 6. `pre_close_check` 修改点

```python
# 替换简单的3项汇总计算（lines 1300-1322）
# 旧代码：只算 total_assets/liabilities/equity
# 新代码：
reconciliation = _balance_reconciliation(db, year, month)
report_balance_ok = reconciliation is None  # None 表示平衡
```

---

## 前端设计

### PeriodClose.vue 改造

当 `status.reportBalanceOk === false` 时，新增"资产负债明细"区块：

```
┌─ ⚠️ 资产负债表不平衡 ──────────────────────┐
│  资产: ¥-39,200  ≠  负债: ¥21,000 + 权益: ¥-1,400   │
│  差额: ¥-40,000                                    │
├─ 资产明细 ──────────────────────────────────────────┤
│  1001 库存现金          ¥5,000     ✅ 正常           │
│  1122 应收账款         ¥-34,800   ❌ 方向异常       │
│  ...                                              │
├─ 负债明细 ──────────────────────────────────────────┤
│  2202 应付账款          ¥21,000    ❌ 方向异常       │
├─ 权益明细 ──────────────────────────────────────────┤
│  4001 实收资本         ¥-1,400    ❌ 方向异常       │
├─ 异常科目 ──────────────────────────────────────────┤
│  ❌ 1122 应收账款: 贷方余额 34,800              │
│     影响: 资产虚减 34,800                        │
│     建议: 检查收款凭证方向是否填反               │
│     关联凭证: [202605-001] [202605-003]           │
│                                               │
│  ❌ 2202 应付账款: 借方余额 21,000              │
│     影响: 负债方向反了                           │
│     建议: 检查付款凭证方向是否填反               │
│     关联凭证: [202605-005]                       │
└──────────────────────────────────────────────────┘
```

**交互**：
- 点击科目代码 → 弹出该科目近期待记账凭证列表
- 点击凭证号 → 跳转凭证详情页

---

## 数据库 / 模型变更

- **无新表**（复用现有 `acc_account`、`acc_voucher`、`acc_voucher_detail` 表）
- **无数据迁移**

---

## API 变更

| 接口 | 方法 | 变更 |
|---|---|---|
| `/accounting/period/{year}/{month}/pre-close-check` | GET | 返回体新增 `balance_reconciliation` 字段（不平衡时填充，平衡时为 `null`）|

**完全向后兼容**：平衡时行为不变，不平衡时才展开诊断信息。

---

## 测试用例

| 用例 | 场景 | 预期 |
|---|---|---|
| TC-BR-001 | 报表平衡（所有科目方向正常） | `report_balance_ok=true`，`balance_reconciliation=null` |
| TC-BR-002 | 应收账款出现贷方余额 | `report_balance_ok=false`，`anomalies` 包含 1122，关联凭证列表非空 |
| TC-BR-003 | 应付账款出现借方余额 | 同上，code=2202 |
| TC-BR-004 | 多个科目同时异常 | `anomalies` 包含所有异常科目各自独立条目 |
| TC-BR-005 | 某科目方向异常但有未记账凭证 | 关联凭证中标注哪些是已记账、哪些是未记账 |
| TC-BR-006 | 报表导出时附上不平衡诊断 | 导出 Excel 时，诊断信息写入说明 Sheet |

---

## 与现有功能的关系

- **继承 `_scan_risk_warnings`**：异常科目检测复用已有方向检测逻辑（1122/1221/2202/2241），但泛化到所有资产/负债/权益科目
- **继承 `risk_warnings`**：已确认/忽略的预警不受影响，继续在现有风险预警区块展示
- **不修改凭证状态机**：仅展示信息，不修改任何数据
- **不阻断结账**：`can_close` 仍由 `report_balance_ok` 决定，但用户可以清楚看到为什么不平衡

---

## 实施步骤

1. **后端** `_balance_reconciliation()` 函数（独立函数，便于测试）
2. **后端** `pre_close_check` 集成，替换简单汇总逻辑
3. **前端** `PeriodClose.vue` 新增不平衡诊断展示区块
4. **测试** 单元测试 + 冒烟测试
5. **更新** `docs/permission-matrix.md`（如需）

---

## 文件清单

| 文件 | 操作 |
|---|---|
| `backend/app/api/v1/accounting.py` | 修改 `_balance_reconciliation` + 集成 |
| `frontend/src/pages/accounting/PeriodClose.vue` | 修改 展示不平衡诊断区块 |
| `docs/specs/phase2-balance-reconciliation.md` | 新建本文档 |