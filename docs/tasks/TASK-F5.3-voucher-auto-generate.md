# TASK-F5.3 | F5 | 凭证自动生成引擎

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P1（MVP 核心）
**状态**：待开发

---

## 任务描述

实现凭证自动生成引擎，是"银行流水 → 凭证"全链路闭环的核心。

### 5.3.1 触发规则配置

定义触发规则表（`voucher_generation_rules`）：

| 单据类型 | 是否自动 | 触发时机 | 允许操作人 |
|:---|:---:|:---|:---|
| 银行费用单 | 否 | 一键生成 | 出纳 |
| 利息收入单 | 否 | 一键生成 | 出纳 |
| 银行转账单 | 否 | 一键生成 | 出纳 |
| 收款单 | 是 | 核销完成时 | 系统 |
| 付款单 | 是 | 核销完成时 | 系统 |
| 提现单 | 否 | 一键生成 | 出纳 |
| 折旧凭证 | 是 | 折旧计划到期（定时） | 系统 |
| 期末结转凭证 | 是 | 结账时 | 系统 |

### 5.3.2 生成引擎

**自动生成流程**：
1. 单据状态变更为 Submitted
2. 检查触发规则：`is_auto = true`？
3. 查询科目映射规则（单据类型 → 借方科目 + 贷方科目 + 辅助核算）
4. 构建 JournalEntry + JournalEntryLines
5. **借贷平衡校验**：`SUM(debit) == SUM(credit)`，不平衡则挂起至"待修复"列表
6. 凭证进入 Draft 状态，等待审核（或直接提交，取决于配置）

**重试机制**：
- 生成失败时：首次失败后 5 分钟自动重试，最多重试 3 次
- 3 次后仍未成功：冻结该单据，通知财务主管，标记 `voucher_status = 'failed'`

**批量生成**：
- 会计可选择多个单据，触发批量生成
- 批量处理有进度条，逐个单据处理
- 失败停在当前单据，已生成的不回退

### 5.3.3 科目映射规则执行

```go
// 根据单据类型获取科目映射
func (s *VoucherService) GetAccountMapping(ctx context.Context, docType string, doc *Document) (*AccountMapping, error) {
    rule, _ := s.ruleRepo.GetByDocType(ctx, docType)
    
    // 借方科目
    debitAccount := s.resolveAccount(rule.DebitAccountID, doc)
    // 贷方科目
    creditAccount := s.resolveAccount(rule.CreditAccountID, doc)
    
    // 支持条件判断（如金额 > 0 时走不同科目）
    if rule.ConditionExpr != "" {
        if !evalCondition(rule.ConditionExpr, doc) {
            // 使用备选科目
            debitAccount = rule.DebitAccountAlt
            creditAccount = rule.CreditAccountAlt
        }
    }
    
    return &AccountMapping{
        DebitAccount:   debitAccount,
        CreditAccount:  creditAccount,
        CostCenter:     doc.CostCenterID,
        Project:        doc.ProjectID,
    }, nil
}
```

### 5.3.4 凭证编号生成

```go
func (s *VoucherService) GenerateVoucherNo(ctx context.Context, postingDate time.Time) (string, error) {
    year := postingDate.Year()
    month := int(postingDate.Month())
    
    // 查询当月最大序号
    maxNo, _ := s.repo.GetMaxVoucherNo(ctx, year, month)
    nextSeq := 1
    if maxNo != "" {
        // 格式：记-YYYY-MM-NNNN，提取 NNNN
        seq, _ := strconv.Atoi(maxNo[13:])
        nextSeq = seq + 1
    }
    
    return fmt.Sprintf("记-%04d-%02d-%04d", year, month, nextSeq), nil
}
```

---

## 验收标准

- [ ] 收款单核销完成后自动生成凭证草稿，无需人工触发
- [ ] 生成失败（借贷不平衡）时挂起至"待修复"列表，通知财务主管
- [ ] 重试 3 次后仍未成功，标记 `voucher_status = 'failed'`，不再重试
- [ ] 凭证编号连续：`记-2026-05-0001` → `记-2026-05-0002`，跳号需备注原因
- [ ] 作废凭证编号保留，不可复用
- [ ] 批量生成：100 张单据批量生成 < 60 秒完成

---

## 前置依赖

TASK-F3.1（核销引擎）、TASK-F5.2（凭证状态机）、TASK-F1.5（科目映射规则）

---

## 预计工时

- 最小：32h
- 最大：56h

---

## 技术提示

### 借贷平衡校验

```go
func ValidateBalance(lines []JournalEntryLine) error {
    var totalDebit, totalCredit decimal.Decimal
    for _, l := range lines {
        totalDebit = totalDebit.Add(l.Debit)
        totalCredit = totalCredit.Add(l.Credit)
    }
    if !totalDebit.Equal(totalCredit) {
        return fmt.Errorf("凭证借贷不平衡: 借方 %.2f, 贷方 %.2f, 差额 %.2f",
            totalDebit, totalCredit, totalDebit.Sub(totalCredit))
    }
    return nil
}
```

### 异步生成（避免阻塞 API）

```go
func (s *VoucherService) EnqueueGeneration(ctx context.Context, docID uuid.UUID, docType string) {
    s.taskQueue <- GenerateTask{DocID: docID, DocType: docType}
}

func (s *VoucherService) StartWorker(ctx context.Context) {
    for {
        select {
        case task := <-s.taskQueue:
            for attempt := 1; attempt <= 3; attempt++ {
                err := s.generate(ctx, task.DocID, task.DocType)
                if err == nil {
                    break
                }
                if attempt < 3 {
                    time.Sleep(5 * time.Minute)  // 5 分钟后重试
                } else {
                    s.notifySupervisor(task.DocID, err)  // 冻结并通知
                }
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### 参考资料

- ERPNext：`erpnext/accounts/utils.py` 的 `make_gl_entries()`
- ERPNext：`accounts/doctype/journal_entry/journal_entry.py` 的 `validate` 方法

---

## 上下文信息（架构师决策记录）

- **决策**：凭证生成后进入 Draft 状态，需要经过审核才 Submit，这是参考 ERPNext 的成熟工作流；可直接提交取决于公司配置
- **决策**：重试间隔 5 分钟是参考实际业务——如果是银行接口暂时不可用，5 分钟后通常恢复
- **风险**：批量生成时如果借贷平衡校验失败，是否跳过继续处理后续单据？建议：是的，跳过失败的单据并记录，最后汇总报告失败列表