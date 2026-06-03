# TASK-F4.1 | F4 | 资金日记账与银企对账

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P2（增强功能）
**状态**：待开发

---

## 任务描述

### 4.1.1 日记账自动登记

- **银行日记账**：银行类单据（收付款/费用/利息/转账）确认提交时，自动登记银行日记账（按 Bank Account 分账页）
- **现金日记账**：现金类单据审核后自动登记；支持手工补录（出纳操作）
- **现金盘点**：出纳录入实盘库存现金，系统自动与账面余额比对，差异 > ±100 元时触发主管审批

### 4.1.2 多维度打分匹配算法

对账匹配采用加权打分制：

| 维度 | 权重 | 满分 | 计算规则 |
|:---|:---|:---|:---|
| 金额匹配 | 50% | 50 | 金额完全一致得 50；差异 1% 以内按比例衰减 |
| 日期匹配 | 20% | 20 | 日期相同得 20；±1 天按比例衰减；±3 天以上 0 |
| 对方户名相似度 | 15% | 15 | Levenshtein ≥ 90% 得 15；80%-90% 得 10；70%-80% 得 5 |
| 摘要关键词匹配 | 10% | 10 | 命中预设关键词得 10；部分命中得 5 |
| 银行流水号精确匹配 | 5% | 5 | 流水号完全匹配直接加 5 |

**得分阈值**：
- ≥ 85 分：自动勾兑，无需人工确认
- 60-85 分：推送至待确认列表，会计确认后勾兑
- < 60 分：视为不匹配，进入余额调节表

**一对多/多对一支持**：
- POS 汇总入账：多笔银行收款合并为一张发票（按比例分配）
- 分批付款：一笔发票分多笔银行转账支付（聚合匹配）

```go
type MatchScore struct {
    TotalScore      float64
    AmountScore     float64
    DateScore       float64
    NameScore       float64
    DescScore       float64
    RefNoScore      float64
    IsAutoMatched   bool  // ≥85 分
    NeedConfirm     bool  // 60-85 分
}
```

### 4.1.3 余额调节表

未匹配项归集为四类：
- 银行已收企业未达
- 银行已付企业未达
- 企业已收银行未达
- 企业已付银行未达

生成余额调节表草案，会计逐项确认（创建调整单或标记"在途"）。

### 4.1.4 对账锁定

对账结果确认后，该银行账户在该期间的记录被锁定。解锁须经财务主管审批。

---

## 验收标准

- [ ] 小样本测试（50 条银行流水 + 50 条 GL 条目），自动勾兑率 ≥ 90%
- [ ] 对账锁定后修改匹配状态须主管审批，审计日志记录
- [ ] 余额调节表四类差异自动归类，计算正确
- [ ] 一对多/多对一匹配结果正确（金额按比例分配）
- [ ] 银行流水号精确匹配时，额外 5 分计入总分

---

## 前置依赖

TASK-F2.2（出纳核对工作台），需要银行流水数据

---

## 预计工时

- 最小：32h
- 最大：56h

---

## 技术提示

### 批量打分查询

```sql
-- 获取未匹配的银行流水和 GL 条目
SELECT bt.id, bt.txn_date, bt.amount, bt.counterparty_name,
       gl.id, gl.posting_date, gl.debit, gl.credit, gl.party_name
FROM bank_transactions bt
LEFT JOIN bank_reconciliation_details brd ON bt.id = brd.bank_transaction_id AND brd.reconciled_at IS NOT NULL
JOIN gl_entries gl ON bt.company_id = gl.company_id
WHERE bt.matched = false
  AND gl.voucher_type = 'payment_entry'
  AND bt.txn_date BETWEEN $period_start AND $period_end
  AND brd.id IS NULL;
```

### 一对多处理（POS 汇总入账）

```go
// 场景：多笔银行小额收款（每笔几十元）对应一张大额发票
// 识别方式：同一天同一客户，多笔小额 sum ≈ 发票金额
func groupPOSSummary(txns []BankTxn) []MatchGroup {
    // 按 date + counterparty 分组
    // 组内金额 sum，接近发票金额时聚合
}
```

---

## 上下文信息（架构师决策记录）

- **决策**：阈值 85/60 是参考行业经验值，建议在上线后用实际数据调优；如果自动勾兑率 < 80%，需要降低阈值或优化相似度算法
- **风险**：摘要关键词匹配（检查项 4）依赖预设关键词词库，需要在 F1.3 规则库中预先配置，建议上线时内置 50 个常用关键词
- **风险**：MT940 格式的银行流水号在不同银行可能格式不同，需要处理前后空格和大小写