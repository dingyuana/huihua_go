# 期间结账 — 设计文档

> 状态: 实施中 (MVP→SaaS 升级)
> 日期: 2026-05-24
> 优先级: P0
> 目标企业: 非制造企业、非代账公司

---

## 1. 设计目标

提供 **SaaS 级7步期末结账向导**，覆盖从凭证完整性 → 资金往来核对 → 资产折旧 → 税金计提 → 损益结转 → 最终审核 → 封账归档的完整闭环。

每个步骤设**阻断门**（`canProceedCurrentStep`），前置检查未通过则不可继续。

---

## 2. 七步向导总览

| Step | 名称 | 阻断门 | 负责人 |
|:----:|------|:------:|--------|
| 0 | 凭证完整性检查 | 无未过账凭证 + 无断号 | 会计 |
| 1 | 资金与往来核对 | 银行全部对账 + 三账一致 | 出纳/会计 |
| 2 | 资产与折旧核查 | 全部已计提（折旧/摊销/利息） | 会计 |
| 3 | 税金计提与复核 | 工作流已核准 | 会计→财务经理 |
| 4 | 一键损益结转 | 损益已结转 | 会计 |
| 5 | 结账最终审核 | 全部检查通过 | 财务经理 |
| 6 | 封账归档 | — | 财务经理 |

---

## 3. 完整检查清单

### 〇、业务子模块结账状态

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 0.1 | 业务单据→凭证完整性 | `_check_submodule_vouchers` | ✅ | 已实现 |
| 0.2 | 工资台账完成状态 | `_check_wage_completeness` | ✅ | 已实现 |

> **不在当前范围内**: 采购/销售/存货子模块（进销存模型不存在，MVP 范围内不包含）

### 一、凭证完整性

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 1.1 | 全部凭证已录入+已记账 | `unposted_vouchers` | ✅ | ✅ 已有 |
| 1.2 | 凭证号连续无断号 | `_check_voucher_continuity` | ✅ | ✅ 已有 |
| 1.3 | 关键凭证附件齐全 | `_check_attachment_completeness` | ⚠️ 建议 | ✅ 已有 |
| 1.4 | 收入金额 ≥ 开票系统开票金额 | `_check_revenue_vs_invoice` | ⚠️ 预警 | ✅ 已实现 |

### 二、资金与往来核对

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 2.1 | 银行日记账与对账单一致 | `_check_bank_reconciliation` | ✅ | ✅ 已实现 |
| 2.2 | 日记账总账余额一致(现金) | `_check_cash_gl_consistency` | ⚠️ 预警 | ✅ 已实现 |
| 2.3 | 应收账款长账龄(>90天) | `_check_ar_ap_aging` | ⚠️ 预警 | ✅ 已实现 |
| 2.4 | 应付账款长账龄(>90天) | `_check_ar_ap_aging` | ⚠️ 预警 | ✅ 已实现 |
| 2.5 | 往来科目余额方向正常 | `_scan_risk_warnings` | ✅ 阻断 | ✅ 已实现 |

### 三、计提与摊销

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 3.1 | 固定资产折旧已计提 | `_check_depreciation_detail` | ✅ | ✅ 已实现 |
| 3.2 | 折旧凭证已过账(1602) | `_check_depreciation_detail`（增强） | ⚠️ | ✅ 已检查余额 |
| 3.3 | 无形资产摊销(1701) | `_check_accruals` | ⚠️ 预警 | ✅ 已实现 |
| 3.4 | 长期待摊费用摊销(1801) | `_check_accruals` | ⚠️ 预警 | ✅ 已实现 |
| 3.5 | 工资及社保已计提 | `_check_wage_completeness` | ✅ | ✅ 已实现 |
| 3.6 | 借款利息已计提(2231) | `_check_accruals` | ⚠️ 预警 | ✅ 已实现 |
| 3.7 | 预提费用已处理(2191) | `_check_accruals` | ⚠️ 预警 | ✅ 已实现 |

### 四、成本核算与进销存

> 当前不做（非制造企业），后续进销存模块上线后扩展

### 五、税金计提

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 5.1 | 增值税与开票系统一致 | `_check_tax_accrual` | ✅ | ✅ 三步工作流 |
| 5.2 | 城建税及附加已计提 | 工作流中包含 | ✅ | ✅ |
| 5.3 | 企业所得税已计提 | 工作流中包含 | ✅ | ✅ |
| 5.4 | 工作流：自动计算→提交→核准 | `tax.py` 4个端点 | ✅ | ✅ |

### 六、最终复核

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 6.1 | 试算平衡(资产=负债+权益) | `report_balance_ok` | ✅ | ✅ 已有 |
| 6.2 | 损益类科目余额方向正常 | `_check_final_review` | ✅ | ✅ 已实现 |
| 6.3 | 毛利率波动预警 | `_check_final_review` | ⚠️ 预警 | ✅ 已实现 |
| 6.4 | 收入×税率≈销项税额 | `_check_final_review` | ⚠️ 预警 | ✅ 已实现 |
| 6.5 | 账税数据相符 | `tax_accrual.all_matched` | ✅ | ✅ 已实现 |
| 6.6 | 三账平衡(总银=日记=对账单) | `_check_final_review` | ⚠️ 预警 | ✅ 已实现 |
| 6.7 | 风险预警已处理 | `_scan_risk_warnings` | ✅ | ✅ 已有 |

### 七、结转后验证（结账完成后自动执行）

| # | 检查项 | 后端函数 | 阻断？ | 状态 |
|---|--------|---------|:------:|:----:|
| 7.1 | 损益类科目期末余额为0 | `_verify_post_close` | ✅ | ✅ 已实现 |
| 7.2 | 本年利润(4104)发生额一致 | `_verify_post_close` | — | ✅ 已实现 |
| 7.3 | 未分配利润变动=净利润本月数 | `_verify_post_close` | — | 待增强（报表级） |

---

## 4. 角色权限分配

| 角色 | Step0 | Step1 | Step2 | Step3 | Step4 | Step5 | Step6 |
|------|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|
| 系统管理员 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 财务经理 | ✅ 核准 | ✅ 核准 | ✅ 核准 | ✅ 核准 | ✅ 核准 | ✅ 决策 | ✅ 执行 |
| 会计 | ✅ 执行 | ✅ 执行 | ✅ 执行 | ✅ 计算+提交 | ✅ 执行 | — | — |
| 出纳 | — | ✅ 对账 | — | — | — | — | — |

- **"继续"按钮** 仅财务经理可见（`isFinanceManager`）
- **操作按钮** 根据角色动态显示（`isAccountant`/`isFinanceManager`）
- Step3 税金工作流：会计只能见"自动计算"/"提交审批"，财务经理只能见"核准通过"

---

## 5. 后端 API

### GET `/api/v1/accounting/period/{year}/{month}/pre-close-check`

**返回结构**（2026-05 增强后）：

```json
{
  "period_status": "open",
  "unposted_vouchers": 0,
  "voucher_counts": {"draft": 0, "pending": 0, "posted": 25, "rejected": 0},
  "report_balance_ok": true,
  "profit_loss_done": true,
  "voucher_continuity": {"has_gaps": false, "items": []},
  "attachment_check": {"checked": true, "missing_count": 0},
  "bank_reconciliation": {"all_reconciled": true, "accounts": [...]},
  "cash_gl_consistency": {"passed": true, "gl_bank_balance": 12345.67, ...},
  "submodule_vouchers": {"passed": true, "total_missing": 0, ...},
  "wage_completeness": {"checked": true, "passed": true, ...},
  "depreciation_check": {"checked": true, "unrecorded_assets": 0, ...},
  "accruals": {"checked": true, "passed": true, "items": {...}},
  "ar_ap_aging": {"checked": true, "passed": true, ...},
  "tax_accrual": {"all_matched": true, "workflow_status": "approved", ...},
  "final_review": {"checked": true, "passed": true, ...},
  "can_close": true,
  "message": "..."
}
```

### GET `/api/v1/accounting/period/{year}/{month}/post-close-verify`

结转后验证：

```json
{
  "pl_accounts_closed": true,
  "pl_nonzero_count": 0,
  "current_year_profit": 12345.67,
  "verified_for_close": true
}
```

---

## 6. 前端步骤定义

### Step 0 — 凭证完整性检查
- 展示: 未过账凭证数 + 断号检测结果 + 附件缺失数
- 阻断: `unposted_vouchers === 0 && !voucher_continuity.has_gaps`
- 按钮: "查看列表"(展开未过账), "查看详情"(断号弹窗)

### Step 1 — 资金与往来核对
- 展示: 各银行对账状态列表 + 余额卡片 + 未调节流水表
- 展示: 应收应付长账龄统计 + 现金日记账一致性
- 阻断: `bank_reconciliation.all_reconciled`

### Step 2 — 资产与折旧核查
- 展示: 折旧计提统计 + 无形资产/长摊/借款利息/预提费用状态
- 展示: 工资台账状态
- 阻断: `depreciation_check.unrecorded_assets === 0`

### Step 3 — 税金计提与复核
- 展示: 工作流进度条(3步) + 税种对比表 + 发票交叉比对
- 阻断: `tax_accrual.all_matched`
- 操作: 自动计算/提交审批(会计) | 核准通过(财务经理)

### Step 4 — 一键损益结转
- 展示: 收入/费用/净利润预览 + 结转分录
- 阻断: `profit_loss_done`

### Step 5 — 结账最终审核
- 展示: 汇总检查项全部通过/失败 + 风险预警列表 + 最终复核项
- 阻断: `allChecksPassed`

### Step 6 — 封账归档
- 展示: 检查通过汇总 + "确认结账"按钮
- 阻断: `allChecksPassed`
- 特殊: 已结账时显示"反结账"按钮

---

## 7. 验收标准

### 后端子模块
- [ ] `_check_submodule_vouchers` — 业务单据→凭证完整性
- [ ] `_check_wage_completeness` — 工资台账状态
- [ ] `_check_accruals` — 无形资产/长摊/利息/预提费用
- [ ] `_check_ar_ap_aging` — 应收应付长账龄
- [ ] `_check_cash_gl_consistency` — 三账余额一致
- [ ] `_check_final_review` — 毛利率/税率交叉/损益方向
- [ ] `_verify_post_close` — 结转后验证
- [ ] 全部集成到 `pre_close_check` 响应

### 前端子模块
- [ ] Step 0 增加收入vs发票比对展示
- [ ] Step 1 增加应收应付+现金一致性汇总
- [ ] Step 2 重命名为"资产与折旧核查" + 增加计提检查项
- [ ] Step 5 重写为完整清单展示
- [ ] 新增结账后验证调用

---

## 8. 后续扩展（非本版本）

- **进销存**：采购/销售/存货子模块结账状态检查
- **代账模式**：收入≥开票、银行流水笔数一致等代账专属检查
- **现金流量表校验**：结转后自动生成
- **自动断号修复**：一键补号
