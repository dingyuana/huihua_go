# SPEC: TASK01 — 后端扩展 pre_close_check 返回值

## 基本信息

- **任务 ID**: close-wizard-001
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **执行者**: OpenCode

## 背景

当前 `/api/v1/accounting/period/{year}/{month}/pre-close-check` 只返回3项布尔值检查（凭证已记账、报表平衡、损益已结转），缺少：

1. 风险预警：往来科目异常方向的自动检测（应收贷方→预收、应付借方→预付）
2. 关键指标环比：毛利率、费用率、应交增值税与上月对比
3. 计提检查：折旧是否已计提、税金计提是否完整

## 目标

在 `pre_close_check` 的返回体中追加 `risk_warnings`、`key_indicators`、`pending_accruals` 三个字段，不破坏现有返回格式。

## 技术约束

- 不修改数据库模型（AccVoucher、AccPeriod、FixedAsset 等）
- 不新增路由，复用现有 GET `/period/{year}/{month}/pre-close-check`
- 不修改现有核销逻辑、凭证状态机、结账锁定逻辑
- 函数在 `backend/app/api/v1/accounting.py` 中实现
- 使用 Decimal 做金额计算，保留2位小数

## 详细设计

### 返回值结构（追加字段）

```json
{
  "risk_warnings": [
    {
      "type": "reclassification",
      "severity": "warning",
      "subject_code": "1122",
      "subject_name": "应收账款",
      "auxiliary_name": "XX客户",
      "balance": -15000.00,
      "message": "应收账款-XX客户为贷方余额15,000元，建议重分类至预收账款"
    }
  ],
  "key_indicators": [
    {
      "name": "毛利率",
      "current_value": 15.0,
      "last_value": 22.0,
      "unit": "%",
      "alert": true,
      "message": "毛利率从22.0%下降至15.0%，降幅超过5个百分点"
    }
  ],
  "pending_accruals": [
    {
      "type": "depreciation",
      "item": "本月固定资产折旧",
      "missing": true,
      "details": "存在2项使用中资产未计提本月折旧"
    },
    {
      "type": "tax",
      "item": "城市维护建设税及附加",
      "missing": false
    }
  ]
}
```

### 函数1：scan_reclassification_risks(db, year, month)

```python
def scan_reclassification_risks(db, year, month):
    """
    扫描应收/应付科目的异常余额方向。
    
    逻辑：
    - 查询 acc_account 表，筛选 subject_id 对应的科目编码
    - 应收账款(1122)：closing_balance < 0 → 贷方余额 → 需重分类至预收
    - 应付账款(2202)：closing_balance > 0（按借贷方向判断）
    - 其他应收(1221)、其他应付(2241) 同理
    
    返回: list[dict]
    """
```

### 函数2：compute_key_indicators(db, year, month)

```python
def compute_key_indicators(db, year, month):
    """
    计算关键财务指标并与上月作环比。
    
    指标清单：
    1. 毛利率 = (营业收入 - 营业成本) / 营业收入
       - 取科目6001(主营业务收入) + 6051(其他业务收入)
       - 取科目6401(主营业务成本) + 6402(其他业务成本)
    2. 期间费用率 = (销售费用+管理费用+财务费用) / 营业收入
    3. 营业利润率 = 营业利润 / 营业收入
    
    数据来源：acc_account 表该月的期末余额
    
    环比逻辑：
    - 当月 - 上月 = 变化绝对值（毛利率用百分点，费用率用百分点）
    - 变化超过5个百分点 → alert=true
    
    边界：
    - 上月数据不存在（如1月或新建账套）→ last_value=null, alert=false
    - 收入为0 → 毛利率/费用率返回 null, 标记 alert=true
    
    返回: list[dict]
    """
```

### 函数3：check_pending_accruals(db, year, month)

```python
def check_pending_accruals(db, year, month):
    """
    检查该月折旧计提和税金计提是否完成。
    
    折旧检查：
    - 查询 fixed_assets 表，筛选 depreciation_status='using' 的资产
    - 如果有使用中资产，查询 AssetDepreciation 表该月是否有折旧记录
    - 有资产但无记录 → missing=true
    
    税金检查（简化版）：
    - 查询 tax_declarations 表该月的申报记录
    - 需覆盖：城市维护建设税及附加、企业所得税（如果有利润）
    - 无申报记录且存在应纳税额 → missing=true
    - 无申报记录但无应纳税额 → missing=false（无需计提）
    
    返回: list[dict]
    """
```

### 入口修改

在 `pre_close_check` 函数体尾部，return 前追加：

```python
risk_warnings = scan_reclassification_risks(db, year, month)
key_indicators = compute_key_indicators(db, year, month)
pending_accruals = check_pending_accruals(db, year, month)

return {
    **base_result,  # 现有返回键
    "risk_warnings": risk_warnings,
    "key_indicators": key_indicators,
    "pending_accruals": pending_accruals,
}
```

## 验收标准

- [ ] 存在应收贷方余额时，`risk_warnings` 返回对应的重分类提示
- [ ] 无异常余额时，`risk_warnings` 返回空列表
- [ ] 毛利率环比下降超过5%时，`key_indicators` 中该条 alert=true
- [ ] 上月数据不存在时，last_value 返回 null
- [ ] 有使用中资产但无折旧记录时，`pending_accruals` 标记 missing=true
- [ ] 无使用中资产时，折旧计提检查返回 missing=false
- [ ] 所有现有 pre_close_check 字段不变（period_status, unposted_vouchers, report_balance_ok 等）
- [ ] 现有测试用例全部通过

## OpenCode 指令

**目标**：在 `/root/huihua-financial-master/backend/app/api/v1/accounting.py` 中，扩展 `pre_close_check` 返回值，追加 `risk_warnings`、`key_indicators`、`pending_accruals` 三个字段。

**约束**：
- 不修改数据库模型
- 不新增路由
- 不修改现有返回字段
- 使用 Decimal 计算金额
- 所有新增函数放在 `pre_close_check` 函数之前（在 `# ==================== 结账前预检查（US-4.2） ====================` 段落后）

**上下文**：
- repo: `/root/huihua-financial-master`
- 需修改的文件：`backend/app/api/v1/accounting.py`
- 参考模型文件：`backend/app/models/accounting.py`（AccVoucher/AccPeriod）、`backend/app/models/fixed_asset.py`（FixedAsset/AssetDepreciation）
- 现有测试：`backend/app/tests/test_mvp_comprehensive.py`

**验收**：
- risk_warnings 在应收贷方余额时返回结构正确的预警
- key_indicators 能在有数据时计算环比
- pending_accruals 能判断折旧是否已计提
- 现有测试全部通过
