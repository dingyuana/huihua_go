# 期末结转模块 - SPEC 开发文档

> 项目：慧财智能财务平台
> 版本：v2.0（人工终审强内控版）
> 日期：2026-06-03
> 状态：**设计稿**

---

## 目录
1. 核心刚性约束
2. 模块定位与设计原则
3. 用户角色与权限互斥
4. 结转类型与模板定义
5. 标准操作流程
6. 状态流转定义
7. 异常处理机制
8. 权限与内控安全
9. 接口设计规范
10. 数据库Schema设计
11. 验收标准

---

## 1. 核心刚性约束（最高优先级）

| 约束编号 | 约束内容 | 说明 |
|---------|---------|------|
| RC-001 | 所有系统自动生成的结转凭证，默认固定为【草稿】状态 | DocStatus = 0 |
| RC-002 | 草稿凭证不计入正式账务、不更新科目余额、不参与报表取数、不允许期末结账 | 账证分离 |
| RC-003 | 系统禁止一切自动过账、自动审核、自动确认结转行为 | 无定时触发、无后台静默执行 |
| RC-004 | 所有结转草稿必须人工逐张审核、人工确认过账，两步人工操作完全分离 | 不可合并 |
| RC-005 | 已过账正式结转凭证不可逆，禁止直接删除/修改，调整仅允许红字冲销 | 数据不可篡改 |

---

## 2. 模块定位与设计原则

### 2.1 模块定位
期末结转是慧话财务整体账务链路的核心闭环环节：
```
业务单据 → 日常凭证 → 审核记账 → 期末结转 → 财务分析 → 期末关账
```

**核心运行模式**：人工主动发起 → 系统规则计算 → 生成非正式草稿 → 人工逐笔终审 → 人工确认过账入账

### 2.2 设计原则

| 原则 | 详细说明 |
|------|---------|
| 人工发起 | 所有结转操作必须财务人员手动触发，无自动触发机制 |
| 模板驱动 | 所有结转分录、取数规则、科目映射由模板定义 |
| 草稿隔离 | 草稿阶段完全隔离正式账务数据 |
| 人工终审 | 生成草稿、审核草稿、确认过账为独立三步 |
| 原子事务 | 单类结转草稿生成整体事务回滚 |
| 权限互斥 | 制单、审核、过账、模板配置权限严格隔离 |
| 可配置扩展 | 支持集团统一管控+单体灵活配置 |

---

## 3. 用户角色与权限互斥

### 3.1 角色职责

| 角色 | 核心职责 | 权限约束 |
|------|---------|---------|
| 普通财务人员（制单） | 执行结转前置检查、触发生成草稿、查看/修改/删除草稿 | 无审核/过账/模板配置权限 |
| 财务审核人员（终审） | 逐笔审核结转草稿、标记审核状态、驳回异常草稿 | 不可生成草稿、不可修改模板 |
| 财务主管/负责人 | 大额审批、审批流配置、结转模板终审、月末关账 | 不可直接过账 |
| 系统管理员 | 权限分配、日志查询、参数维护 | 无业务操作权限 |

### 3.2 权限互斥规则
- 制单权限 ↔ 审核权限：互斥
- 审核权限 ↔ 过账权限：互斥  
- 制单权限 ↔ 模板配置权限：互斥

---

## 4. 结转类型与模板定义

### 4.1 结转类型

| 类型 | 编码 | 说明 | 优先级 |
|------|------|------|--------|
| 期间损益结转 | `income_expense` | 归集收入/费用科目余额至本年利润 | P0 |
| 税金结转 | `tax` | 归集应交增值税、附加税费至未交增值税 | P1 |
| 自定义结转 | `custom` | 用户自主配置取数来源、借贷科目、结转比例 | P2 |

> **注意**：成本结转已从范围中移除（商业企业不适用）

### 4.2 模板数据结构

```go
type ClosingTemplate struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    Name            string
    TemplateType    string          // income_expense / tax / custom
    Description     string
    Lines           []TemplateLine
    IsActive        bool
    Priority        int
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type TemplateLine struct {
    ID              uuid.UUID
    TemplateID      uuid.UUID
    SeqNo           int
    SourceAccountID *uuid.UUID      // 取数来源科目（可为空表示全部）
    DebitAccountID  uuid.UUID
    CreditAccountID uuid.UUID
    Formula         string          // 取数公式
    IsEnabled       bool
}
```

---

## 5. 标准操作流程

### 5.1 执行顺序（系统刚性控制）

```
期末检查 → 生成结转草稿 → 人工逐张审核 → 人工确认过账
     ↓              ↓              ↓              ↓
  校验通过        草稿生成       审核通过/驳回     正式入账
```

### 5.2 第一步：结转前置检查

| 检查项 | 失败提示 | 处理逻辑 |
|--------|---------|---------|
| 当期所有凭证全部过账 | 存在N张未审核未过账凭证 | 阻断后续操作 |
| 当期银行流水全部对账完成 | 存在N笔银行流水未对账 | 阻断后续操作 |
| 结转模板配置完整有效 | 【XX结转】模板未配置 | 阻断对应类型结转 |
| 折旧/摊销已计提 | 当期折旧/摊销未计提 | 阻断后续所有结转 |
| 往来科目无异常余额方向 | 存在往来科目余额方向异常 | 预警提示，不阻断 |

### 5.3 第二步：生成结转草稿

```go
// 核心逻辑
func GenerateClosingDraft(ctx, tenantID, periodNo, userID) error {
    // 1. 读取期间信息
    // 2. 读取结转模板
    // 3. 计算科目余额
    // 4. 生成凭证草稿 (DocStatus = 0)
    // 5. 记录操作日志
}
```

### 5.4 第三步：人工逐张审核

**强制审核内容**：
- 科目选用是否准确
- 结转金额与科目余额表是否一致
- 借贷方向、辅助核算是否正确
- 是否存在多余/漏结转

**审核操作**：
| 操作 | 说明 |
|------|------|
| 查看详情 | 查看完整分录、取数来源、科目余额 |
| 修改草稿 | 修改内容全程留痕 |
| 删除草稿 | 错误草稿可删除，删除行为日志留存 |
| 审核通过 | 标记为「已审核待过账」 |
| 审核驳回 | 填写驳回原因，退回草稿状态 |

### 5.5 第四步：人工确认过账

**过账前置校验**：
- 凭证借贷平衡
- 草稿自审核后无数据变更
- 当前会计期间未关账
- 无并发修改冲突

**过账成功后行为**：
1. 凭证状态变更为「已过账」
2. 更新对应科目期末余额
3. 凭证纳入报表取数
4. 锁定凭证数据，禁止直接修改

---

## 6. 状态流转定义

### 6.1 凭证状态机

```
[未执行] → 人工触发生成 → [草稿已生成] → 人工审核 → [已审核待过账] → 人工确认 → [已过账]
                                 ↓                              ↓
                            [审核驳回]                      [过账失败]
                                 ↓
                          [退回草稿状态]
```

### 6.2 禁止行为

| 禁止行为 | 原因 |
|---------|------|
| 系统定时任务自动生成结转草稿 | 违反人工发起原则 |
| 月末自动触发结转流程 | 违反人工发起原则 |
| 审核通过后自动过账 | 违反两步分离原则 |
| 后台静默修改结转草稿 | 违反操作留痕原则 |

---

## 7. 异常处理机制

| 异常场景 | 系统处理 | 人工处置 |
|---------|---------|---------|
| 生成草稿时科目余额为零 | 自动跳过该分录，标注"余额为零" | 核对科目余额 |
| 草稿生成借贷不平衡 | 整体事务回滚 | 核查结转模板 |
| 过账时凭证存在并发修改 | 阻断过账，提示"凭证已变更" | 重新审核后过账 |
| 流程中断（断电/网络） | 断点续传，自动恢复 | 继续执行流程 |
| 大额结转金额异常 | 自动标红预警，触发主管审批 | 人工专项核查 |

---

## 8. 权限与内控安全

### 8.1 数据不可篡改机制

| 阶段 | 状态 | 可修改 | 可删除 | 日志记录 |
|------|------|--------|--------|---------|
| 草稿阶段 | draft | ✅ | ✅ | 全程留存修改前后对比 |
| 审核通过后 | approved | ❌ | ❌ | 操作日志留存 |
| 过账正式后 | posted | ❌ (仅红字冲销) | ❌ | 永久锁定 |

### 8.2 操作日志规范

所有操作记录包含：
- 操作人、账号、IP、设备指纹
- 操作时间、操作类型、变更内容
- 审核意见、驳回原因
- 日志留存时长：不低于30年

---

## 9. 接口设计规范

### 9.1 接口列表

| 接口名称 | HTTP方法 | 路径 | 核心用途 |
|---------|---------|------|---------|
| 期末结转前置检查 | GET | `/api/v1/periods/{periodNo}/pre-close-check` | 执行全量前置校验 |
| 结转草稿生成 | POST | `/api/v1/periods/{periodNo}/closing-drafts` | 人工触发生成结转草稿 |
| 草稿列表查询 | GET | `/api/v1/closing-drafts` | 查询草稿/待审核/待过账列表 |
| 草稿详情查询 | GET | `/api/v1/closing-drafts/{id}` | 查询草稿详情 |
| 草稿修改 | PUT | `/api/v1/closing-drafts/{id}` | 人工修改草稿分录 |
| 草稿删除 | DELETE | `/api/v1/closing-drafts/{id}` | 删除异常草稿 |
| 草稿审核 | POST | `/api/v1/closing-drafts/{id}/review` | 审核通过/驳回 |
| 结转凭证过账 | POST | `/api/v1/closing-drafts/{id}/post` | 人工确认过账 |
| 结转模板列表 | GET | `/api/v1/closing-templates` | 查询结转模板 |
| 结转模板创建 | POST | `/api/v1/closing-templates` | 创建结转模板 |
| 结转模板更新 | PUT | `/api/v1/closing-templates/{id}` | 更新结转模板 |
| 结转模板删除 | DELETE | `/api/v1/closing-templates/{id}` | 删除结转模板 |

### 9.2 接口通用安全规则

- Token时效校验
- IP白名单校验
- 请求参数校验
- 数据签名防篡改
- 接口调用日志全留存
- 支持接口版本兼容

---

## 10. 数据库Schema设计

### 10.1 结转草稿表（closing_drafts）

```sql
CREATE TABLE closing_drafts (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    period_no int NOT NULL,
    template_id uuid REFERENCES closing_templates(id),
    draft_type varchar(50) NOT NULL, -- income_expense / tax / custom
    voucher_no varchar(50),
    doc_status int DEFAULT 0, -- 0=draft, 1=approved, 2=posted
    total_debit decimal(20,4),
    total_credit decimal(20,4),
    created_by uuid NOT NULL REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    reviewed_at timestamp,
    posted_by uuid REFERENCES users(id),
    posted_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_closing_drafts_tenant_period ON closing_drafts(tenant_id, period_no);
CREATE INDEX idx_closing_drafts_status ON closing_drafts(doc_status);
```

### 10.2 结转草稿分录表（closing_draft_lines）

```sql
CREATE TABLE closing_draft_lines (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    draft_id uuid NOT NULL REFERENCES closing_drafts(id) ON DELETE CASCADE,
    seq_no int NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts(id),
    debit decimal(20,4) DEFAULT 0,
    credit decimal(20,4) DEFAULT 0,
    summary text,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_closing_draft_lines_draft ON closing_draft_lines(draft_id);
```

### 10.3 结转模板表（closing_templates）

```sql
CREATE TABLE closing_templates (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    name varchar(100) NOT NULL,
    template_type varchar(50) NOT NULL, -- income_expense / tax / custom
    description text,
    is_active boolean DEFAULT true,
    priority int DEFAULT 1,
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_closing_templates_tenant ON closing_templates(tenant_id);
```

### 10.4 结转模板行表（closing_template_lines）

```sql
CREATE TABLE closing_template_lines (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    template_id uuid NOT NULL REFERENCES closing_templates(id) ON DELETE CASCADE,
    seq_no int NOT NULL,
    source_account_id uuid REFERENCES accounts(id),
    debit_account_id uuid NOT NULL REFERENCES accounts(id),
    credit_account_id uuid NOT NULL REFERENCES accounts(id),
    formula text,
    is_enabled boolean DEFAULT true,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_closing_template_lines_template ON closing_template_lines(template_id);
```

---

## 11. 验收标准

| 编号 | 验收项 | 验证方法 |
|------|--------|---------|
| AC-001 | 系统生成的结转凭证默认草稿状态 | 检查`doc_status=0` |
| AC-002 | 草稿凭证不计入正式账务 | 检查科目余额未更新 |
| AC-003 | 无自动结转/审核/过账机制 | 检查无定时任务 |
| AC-004 | 权限严格互斥 | 验证同一用户无法同时拥有制单+审核权限 |
| AC-005 | 流程顺序刚性控制 | 前置校验失败时阻断后续操作 |
| AC-006 | 所有操作全程留痕 | 检查操作日志记录完整 |
| AC-007 | 正式凭证不可直接修改 | 验证过账后修改接口返回错误 |
| AC-008 | 异常处理完整 | 验证事务原子性，无部分入账 |
| AC-009 | 接口安全合规 | 验证Token失效后接口拒绝访问 |
| AC-010 | 税金结转功能可用 | 验证增值税/附加税费正确结转 |

---

**文档版本**：v2.0  
**生成日期**：2026-06-03  
**项目版本**：慧财智能财务平台 v1.0.0