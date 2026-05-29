# TASK-F3.1 | F3 | 核销预检与五级匹配策略

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P1（MVP 核心）
**状态**：待开发

---

## 任务描述

实现智能核销引擎，是整个系统的核心业务模块之一。

### 3.1.1 核销预检（6 项前置校验）

核销执行前必须通过以下全部检查，任何一项不通过则**阻止核销并强制提示**：

| # | 检查项 | 检查逻辑 | 不通过时 |
|:---|:---|:---|:---|
| 1 | **对方单位匹配** | 收款方客户名 vs 发票购方税号；付款方供应商名 vs 发票销方税号 | 标记"单位不匹配"，展示差异，会计可选择"强制通过"（须备注原因） |
| 2 | **金额超限检查** | 核销金额 ≤ 单据未结清金额 | 限制输入框最大值 = 未结清金额 |
| 3 | **重复核销检查** | 该发票+该收款单是否已被核销过 | 提示"已核销记录"，展示历史核销明细 |
| 4 | **跨账套检查** | 核销双方是否属于同一客户账套（tenant_id 一致） | 禁止跨账套核销，提示切换账套 |
| 5 | **业务类型一致性** | 收款单匹配销项发票；付款单匹配进项发票 | 不匹配时强制提示，业务类型不符不得核销 |
| 6 | **到期日检查** | 进项发票是否已过认证有效期（当前日期 > 到期日） | 过期发票标记"已过期"，默认不参与核销，需人工确认 |

预检结果展示为清单式列表，每项可查看详情。

### 3.1.2 五级匹配策略

| 层级 | 策略 | 触发条件 | 是否需确认 |
|:---|:---|:---|:---:|
| **L1：精确自动匹配** | 金额相等 + 对方单位一致 → 一对一自动核销 | 自动触发 | 否 |
| **L2：部分自动匹配（FIFO）** | 按发票日期先进先出，自动拆分金额，支持一对多/多对一 | 自动触发 | 否 |
| **L3：模糊匹配推荐** | 名称相似度 > 85% 时推送 Top 3 候选 | 推送至 IM/Web 待确认列表 | 是（会计确认） |
| **L4：小额尾差处理** | 未核销余额 ≤ 预设阈值（默认 1 元，可配置） | 自动触发 | 否 |
| **L5：手工核销中心** | 会计手动勾选发票与收付款单，输入核销金额 | 手动选择触发 | 是 |

**相似度计算**：
```go
// 使用编辑距离（Levenshtein Distance）计算名称相似度
func Similarity(a, b string) float64 {
    distance := levenshteinDistance(a, b)
    maxLen := math.Max(float64(len(a)), float64(len(b)))
    return 1 - float64(distance)/maxLen
}
```

**FIFO 逻辑**：
```go
// 按发票日期排序，优先核销最早的发票
func fifoMatch(invoices []Invoice, paymentAmount decimal.Decimal) []Allocation {
    sort.Slice(invoices, func(i, j int) bool {
        return invoices[i].PostingDate.Before(invoices[j].PostingDate)
    })
    // 先进先出分配
}
```

### 3.1.3 核销执行与状态更新

核销执行后：
1. 更新发票 `outstanding_amount -= 核销金额`
2. 发票状态流转：待核销 → 部分核销 → 已核销
3. 登记 `payment_allocations` 表（invoice_id / payment_id / allocated_amount）
4. 核销双方互相索引（可穿透查询）

### 3.1.4 核销回退

- 核销后 30 天内，原操作人或主管可发起"取消核销"
- 取消后：发票 `outstanding_amount += 核销金额`，状态恢复，`payment_allocations` 标记 `cancelled = true`
- 操作留痕（audit_logs 记录）

---

## 验收标准

- [ ] 核销预检 6 项全部可执行，任意项不通过时弹出详情并阻止执行
- [ ] L1 精确匹配：金额相等 + 对方一致时，系统自动完成核销，无需人工
- [ ] L3 模糊匹配：Top 3 推荐准确率 ≥ 80%，会计确认后才执行
- [ ] L4 尾差处理：≤1 元时自动核销，标记"尾差"
- [ ] 发票状态流转正确：待核销 → 部分核销（余额 > 0）→ 已核销（余额 = 0）
- [ ] 核销回退后余额正确回填，状态恢复
- [ ] 所有操作记录 audit_logs

---

## 前置依赖

TASK-F2.2（出纳核对工作台）、TASK-F2.3（发票采集）、TASK-F1.1（科目表，需客商档案）

---

## 预计工时

- 最小：32h
- 最大：56h

---

## 技术提示

### 核销预检实现

```go
type ReconciliationPreCheck struct {
    invoice   *Invoice
    payment   *PaymentEntry
    checkFunc func(*Invoice, *PaymentEntry) *CheckResult
}

type CheckResult struct {
    Passed   bool
    Message  string
    Severity string  // error/warning/info
}

func (s *ReconciliationService) PreCheck(ctx context.Context, invoiceID, paymentID uuid.UUID) ([]*CheckResult, error) {
    invoice, _ := s.repo.GetInvoice(ctx, invoiceID)
    payment, _ := s.repo.GetPayment(ctx, paymentID)
    
    checks := []func(...) *CheckResult{
        checkPartyMatch,       // 对方单位匹配
        checkAmountLimit,     // 金额超限
        checkNotYetAllocated,  // 重复核销
        checkSameTenant,       // 跨账套
        checkBusinessType,     // 业务类型一致性
        checkExpiry,           // 到期日
    }
    
    var results []*CheckResult
    for _, check := range checks {
        results = append(results, check(invoice, payment))
    }
    return results, nil
}
```

### 一对多 / 多对一处理

```go
// 一笔收款匹配多张发票（FIFO）
func allocateOneToMany(payment *PaymentEntry, invoices []Invoice) []Allocation {
    remaining := payment.PaidAmount
    var allocs []Allocation
    for _, inv := range invoices {
        if remaining.Cmp(inv.OutstandingAmount) <= 0 {
            allocs = append(allocs, Allocation{PaymentID: payment.ID, InvoiceID: inv.ID, Amount: remaining})
            break
        }
        allocs = append(allocs, Allocation{PaymentID: payment.ID, InvoiceID: inv.ID, Amount: inv.OutstandingAmount})
        remaining = remaining.Sub(inv.OutstandingAmount)
    }
    return allocs
}
```

### 参考资料

- ERPNext：`accounts/doctype/payment_entry/payment_entry.js` 中的 `auto_allocate` 逻辑
- Levenshtein Distance Go：`github.com/texttheater/golang-levenshtein`

---

## 上下文信息（架构师决策记录）

- **决策**：核销预检不通过时默认阻止执行，但会计可选择"强制通过"并备注原因（用于边界情况如业务紧急），这是参考 ERPNext 的实际业务需求
- **决策**：L4 尾差阈值可配置（默认 1 元），因为不同企业容差接受度不同
- **风险**：L3 模糊匹配的相似度阈值 85% 是经验值，建议在上线后收集数据调整；若 Top 3 推荐准确率 < 70%，需要优化算法或引入 LLM 判断