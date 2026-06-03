# TASK-F6.3 | F6 | 结账前智能体检

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P1（MVP 核心）
**状态**：待开发

---

## 任务描述

实现结账前 10 项健康检查，逐项输出《结账健康度报告》。

### 10 项检查清单

| # | 检查类别 | 检查项 | 不通过时 |
|:---|:---|:---|:---|
| 1 | 凭证平衡 | 本期所有已审核凭证借贷双方合计相等 | 列出不平衡凭证 |
| 2 | 凭证完整性 | 本期是否有未审核的记账凭证 | 列出待审凭证列表 |
| 3 | 凭证编号连续性 | 本期凭证编号是否连续（允许作废产生的空号） | 列出断号位置 |
| 4 | 固定资产折旧 | 本期折旧计划是否全部已过账 | 列出未过账折旧计划 |
| 5 | 银行日记账一致性 | 各银行日记账期末余额与 GL 银行存款科目余额一致 | 列出不一致账户 |
| 6 | 现金账实一致 | 现金日记账余额与实盘库存现金一致（容差 ±100 元） | 列出差异金额 |
| 7 | 往来核销完成度 | 应收/应付中是否有超 30 天未核销且金额 > 1 万元 | 列出清单 |
| 8 | 进项发票到期 | 是否有已过期但未认证的进项发票 | 列出过期发票 |
| 9 | 损益结转 | 损益类科目本期余额是否已结转（或有结转凭证草稿） | 生成结转凭证草稿 |
| 10 | 期间状态 | 当前期间是否已锁定 | 显示锁定状态 |

### 报告输出格式

```go
type HealthCheckReport struct {
    Period        string                    // e.g. "2026-05"
    Company       string
    OverallStatus string                    // green/yellow/red
    Checks        []HealthCheckItem
}

type HealthCheckItem struct {
    ID       int
    Name     string
    Status   string                    // passed/warning/blocked
    Message  string                    // 详细信息
    Action   string                    // 修复建议
    Redirect string                    // 点击跳转 URL（可选）
}
```

### 检查执行逻辑

```go
func (s *PeriodService) RunHealthChecks(ctx context.Context, periodID uuid.UUID) (*HealthCheckReport, error) {
    checks := []HealthCheckFunc{
        checkVoucherBalance,
        checkPendingVouchers,
        checkVoucherNoSequence,
        checkDepreciationPosting,
        checkBankReconciliation,
        checkCashDifference,
        checkOldReceivables,
        checkExpiredInputVAT,
        checkIncomeExpenseClosing,
        checkPeriodLocked,
    }

    report := &HealthCheckReport{PeriodID: periodID}
    for _, check := range checks {
        item := check(ctx, periodID)
        report.Checks = append(report.Checks, *item)
        if item.Status == "blocked" {
            report.OverallStatus = "red"
        } else if item.Status == "warning" && report.OverallStatus != "red" {
            report.OverallStatus = "yellow"
        } else if report.OverallStatus == "" {
            report.OverallStatus = "green"
        }
    }
    return report, nil
}
```

---

## 验收标准

- [ ] 全部 10 项检查均可执行，结果正确
- [ ] 任意阻断项存在时，"结账"按钮禁用并显示红色提示
- [ ] 警告项以黄色展示，不阻止结账但建议修复
- [ ] 每项检查结果可点击跳转至具体处理页面
- [ ] 报告支持导出 PDF（用于存档）

---

## 前置依赖

TASK-F5.4（凭证审核工作台）、TASK-F6.1（固定资产折旧）

---

## 预计工时

- 最小：24h
- 最大：40h

---

## 技术提示

### 检查 1：凭证平衡

```sql
SELECT posting_date, voucher_no,
       SUM(debit) as total_debit,
       SUM(credit) as total_credit,
       SUM(debit) - SUM(credit) as diff
FROM journal_entries je
JOIN journal_entry_lines jel ON je.id = jel.journal_entry_id
WHERE je.docstatus = 1
  AND je.posting_date BETWEEN $period_start AND $period_end
GROUP BY posting_date, voucher_no
HAVING ABS(SUM(debit) - SUM(credit)) > 0.01;
```

### 检查 5：银行日记账一致性

```sql
-- 银行日记账余额
SELECT ba.id, ba.account_number,
       SUM(bt.debit) - SUM(bt.credit) as bank_balance
FROM bank_transactions bt
JOIN bank_accounts ba ON bt.bank_account_id = ba.id
WHERE bt.matched = true
  AND bt.txn_date BETWEEN $period_start AND $period_end
GROUP BY ba.id, ba.account_number;

-- GL 银行存款科目余额
SELECT account_id, SUM(debit) - SUM(credit) as gl_balance
FROM gl_entries
WHERE account_id IN (SELECT clearing_account_id FROM bank_accounts)
  AND posting_date BETWEEN $period_start AND $period_end
  AND is_cancelled = FALSE
GROUP BY account_id;
```

### 参考资料

- ERPNext：`accounts/page/period_list/period_list.js` 中的结账前校验逻辑

---

## 上下文信息（架构师决策记录）

- **决策**：检查结果以清单列表展示而非单个通过/失败，因为实际业务中可能有部分检查通过、部分警告、个别阻断，财务主管需要逐项处理
- **决策**：允许"警告项"不影响结账（因为部分警告项如"小额往来长期未核销"可能是正常的账期安排），但阻断项必须全部修复
- **风险**：检查 7（往来核销完成度）的阈值（30 天 + 1 万元）是可配置的，不同企业标准不同