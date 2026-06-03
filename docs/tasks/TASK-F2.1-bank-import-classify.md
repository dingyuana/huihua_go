# TASK-F2.1 | F2 | 银行流水导入与智能分类

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P1（MVP 核心）
**状态**：待开发

---

## 任务描述

### 2.1.1 多格式导入

支持以下格式的银行对账单导入：
- **CSV/Excel**：用户上传文件，前端解析或后端 POI 解析
- **CAMT053**：ISO 20022 XML 标准格式，欧洲银行通用
- **MT940**：SWIFT 银行对账标准格式

每种格式对应一个 Format Parser，实现公共接口：

```go
type BankStatementParser interface {
    Parse(raw []byte) ([]BankTxnRaw, error)  // 返回原始流水列表
    Validate() error                           // 格式校验
}
```

### 2.1.2 字段提取与清洗

从原始数据中提取标准化字段：
- `txn_date`：交易日期
- `amount`：交易金额（正数=收入/负数=支出）
- `direction`：in（企业收款）/ out（企业付款）
- `counterparty_name`：对方户名
- `description`：摘要
- `reference_no`：银行流水号（唯一键组成部分）

### 2.1.3 重复拦截

唯一键：`bank_account_id + reference_no + txn_date + amount + direction`

- 导入时查询是否已存在，存在则标记为 `is_duplicate = true`
- 重复记录在列表中展示为灰色，不计入对账
- 导入日志记录每批次导入的总数/成功数/重复数

### 2.1.4 智能分类引擎

基于规则库（F1.3），将每条流水分类为六类之一：

| 分类 | 规则命中条件 | 生成单据 |
|:---|:---|:---|
| 业务收款 | direction=in + 对方匹配客户 | 收款单 |
| 业务付款 | direction=out + 对方匹配供应商 | 付款单 |
| 银行费用 | 摘要含"手续费"/"管理费"/"年费" | 银行费用单 |
| 利息收入 | 摘要含"利息" | 利息收入单 |
| 内部转账 | 对方户名含本公司名称关键词 | 银行转账单 |
| 待处理 | 以上均不匹配 | 待处理池 |

---

## 验收标准

- [ ] CSV/Excel/CAMT053/MT940 四种格式均可正确解析（各准备一份标准测试文件）
- [ ] 重复流水拦截：同一唯一键第二次导入时提示"已导入"，记录条数纳入统计
- [ ] 导入日志可查，包含：批次号/导入时间/总条数/成功数/重复数
- [ ] 智能分类准确率 ≥ 85%（规则匹配，不含 AI）
- [ ] 所有 API 查询自动携带 tenant_id 过滤

---

## 前置依赖

TASK-F1.3（银行流水模板配置），需要先有银行对账单格式模板

---

## 预计工时

- 最小：24h
- 最大：40h

---

## 技术提示

### CSV Parser 示例

```go
type CSVParser struct {
    ColumnMapping ColumnMapping  // 字段映射配置
}

type ColumnMapping struct {
    DateCol     int    // 日期列索引
    AmountCol   int    // 金额列索引
    DirectionCol int   // 方向列索引（可选）
    CounterpartyCol int
    DescriptionCol int
    ReferenceCol int
    DebitCol  int      // 借方列（可选）
    CreditCol int      // 贷方列（可选）
}

// 金额方向判断：
// - 有 direction 列：直接读
// - 无 direction 列但有 debit/credit：debit > 0 → in，credit > 0 → out
// - 无 direction 列且无 debit/credit：amount > 0 → in，amount < 0 → out
```

### 重复检测（批量导入时）

```go
func (s *BankTxnService) ImportBatch(ctx context.Context, parser BankStatementParser, bankAccountID uuid.UUID) (*ImportResult, error) {
    rawTxns, _ := parser.Parse()
    
    existingKeys := s.repo.GetExistingKeys(ctx, bankAccountID, extractKeys(rawTxns))
    existingSet := set.New(existingKeys...)
    
    var imported, duplicated int
    for _, txn := range rawTxns {
        key := buildUniqueKey(txn)
        if existingSet.Has(key) {
            duplicated++
            continue
        }
        // 插入并记录
        imported++
    }
    return &ImportResult{Imported: imported, Duplicated: duplicated}, nil
}
```

### CAMT053 结构（参考）

```xml
<Stmt>
  <Ntry>           <!-- 流水条目 -->
    <Amt Ccy="CNY">1234.56</Amt>
    <CdtDbtInd>    <!-- CRDT:收款/DBIT:付款 -->
    < BookgDt/Dt>  <!-- 日期 -->
    < NtryRef>      <!-- 流水号 -->
    < NtryDtls> -> < TxDtls> -> < RmtInf> -> < Ustrd> <!-- 摘要 -->
    < NtryDtls> -> < TxDtls> -> < NtryRef>             <!-- 对方账户 -->
  </Ntry>
</Stmt>
```

### MT940 结构（参考）

```
:20: 报文类型
:25: 账号
:28C: 页码/总页数
:60F: 起始余额
:61: 交易流水（日期/金额/方向/对方账号/摘要）
:86: 交易描述
:62F: 结束余额
```

---

## 上下文信息（架构师决策记录）

- **决策**：四种格式各实现独立 Parser，但输出统一为 `BankTxnRaw` 结构，确保后续处理逻辑统一
- **决策**：重复拦截以唯一键查数据库而非内存集合，因为批量导入可能有几万条流水
- **风险**：CAMT053 有多个变体（camt.052、camt.053、camt.054），需要验证支持到哪个版本；建议从 camt.053 入手，逐步扩展
- **风险**：MT940 格式不统一（不同银行摘要格式差异大），建议优先实现 CSV/Excel，MT940 作为增强