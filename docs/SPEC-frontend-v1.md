# 汇华财务 Go 版前端 SPEC

> 项目：huihua-finance (Go) 前端界面
> 版本：v1.0.0 MVP
> 日期：2026-05-29
> 状态：**设计稿**，待老丁审核后实施

---

## 变更记录

| 日期 | 版本 | 变更内容 | 作者 |
|------|------|---------|------|
| 2026-05-29 | v1.0.0 | 初稿创建 | Hermes |

---

## 1. 项目背景与目标

### 1.1 为什么做这个前端

huihua-finance (Go) 目前是纯后端 API，无界面。项目已达 83% 完成度，但无法对外展示和使用。

目标：**为 Go 后端快速搭建一套管理后台界面，能够演示核心业务流程**。

### 1.2 约束条件

- Go 后端 API 和 Python 版 (`huihua-financial-master`) 的接口**不一致**
- 需要 API 适配层来处理路径差异、字段映射、状态值转换
- 不得修改 Go 后端代码（保持独立演进）

### 1.3 核心差异说明

| 对比项 | Go 后端 | Python 后端 | 影响 |
|--------|--------|-------------|------|
| 基础路径 | `/api/v1/` | `/api/v1/` | 同前缀 |
| 凭证路径 | `/vouchers` | `/accounting/vouchers` | 需映射 |
| 发票路径 | `/invoices` | `/invoice` | 需映射 |
| 字段名 | `posting_date` | `voucher_date` | 需映射 |
| 借贷方向 | `debit`, `credit` (bool/float) | `dc` ('D'/'C'), `amount` | 需映射 |
| 凭证状态 | int `0=draft,1=submitted,2=approved,3=rejected,4=posted` | string `'draft'` `'approved'` `'posted'` | 需映射 |
| 响应格式 | 直接返回对象或 `{data: x}` | 统一 `{code, msg, data}` | 需适配层 |
| 分页格式 | 无统一格式 | `{total, list: []}` | 需适配层 |
| JWT user字段 | `user_id` | `sub` | 需适配层 |

---

## 2. 技术选型

### 2.1 推荐方案

| 选项 | 框架 | 开发速度 | 维护成本 | 适合场景 | 推荐 |
|------|------|---------|---------|---------|------|
| A | Vue 3 + Vite + Element Plus | 快 | 低 | 快速 MVP，展示用 | ✅ |
| B | React + Next.js | 中 | 中 | 长期维护 | ❌ |
| C | 原生 HTML + Alpine.js | 最快 | 高 | 单页面简单 | ❌ |

**推荐选项 A：Vue 3 + Vite + Element Plus**

理由：
1. 与 `huihua-financial-master` 技术栈一致（两个项目维护者相同）
2. Element Plus 组件库完整，适合管理后台
3. Vite 构建快，开发体验好
4. 代码可复用部分可直接拷贝

### 2.2 项目结构

```
frontend/                    # 前端项目根目录
├── src/
│   ├── api/                # API 调用层（包含 Go 后端适配器）
│   │   ├── index.js        # axios 实例 + 拦截器
│   │   ├── adapter/        # 🔑 API 适配层（关键）
│   │   │   ├── go-backend.js   # Go 后端专用请求实例
│   │   │   └── fieldMapper.js # 字段名映射规则
│   │   ├── auth.js
│   │   ├── voucher.js
│   │   ├── invoice.js
│   │   ├── bankAccount.js
│   │   ├── party.js
│   │   └── report.js
│   ├── stores/             # Pinia 状态管理
│   │   └── auth.js
│   ├── router/            # Vue Router
│   ├── views/             # 页面
│   │   ├── Dashboard.vue
│   │   ├── Login.vue
│   │   ├── voucher/       # 凭证管理
│   │   ├── invoice/       # 发票管理
│   │   ├── bank/          # 银行流水
│   │   ├── party/         # 往来单位
│   │   ├── report/        # 财务报表
│   │   └── setup/         # 账套初始化
│   ├── components/        # 公共组件
│   └── styles/
├── vite.config.js
└── package.json
```

### 2.3 API 适配层设计（核心）

由于 Go 后端与 Python 后端字段不兼容，**不在后端改，在前端适配**。

```
┌──────────────┐    ┌─────────────────┐    ┌────────────┐
│   Vue 组件   │ → │  adapter层统一   │ → │  Go API    │
│  (统一字段)  │ ← │  fieldMapper    │ ← │ (原始字段) │
└──────────────┘    └─────────────────┘    └────────────┘
```

**字段映射规则示例：**

| 前端统一字段 | Go 后端字段 | Python 后端字段 |
|-------------|-------------|-----------------|
| `voucher_date` | `posting_date` | `voucher_date` |
| `debit_amount` | `debit` | `amount` (when dc='D') |
| `credit_amount` | `credit` | `amount` (when dc='C') |
| `status` | `0,1,2,3,4` (int) | `'draft'` (string) |
| `status_text` | `'待审核'` 等 | `'待审核'` 等 |

**状态值转换（Go → 前端）：**

| Go 值 (int) | 前端显示 |
|------------|---------|
| 0 | `draft` / `待录入` |
| 1 | `submitted` / `已提交` |
| 2 | `approved` / `已核准` |
| 3 | `rejected` / `已驳回` |
| 4 | `posted` / `已过账` |
| 5 | `reversed` / `已红冲` |

---

## 3. 功能范围（MVP）

### 3.1 MVP 页面清单

| # | 页面 | 路由 | 优先级 | 说明 |
|---|------|------|--------|------|
| 1 | 登录页 | `/login` | P0 | JWT 登录 |
| 2 | 工作台/仪表盘 | `/dashboard` | P0 | 概览：凭证数、发票数、银行账户、未对账流水 |
| 3 | 凭证管理 | `/vouchers` | P0 | 列表 + 状态流转 |
| 4 | 凭证录入 | `/vouchers/new` | P0 | 创建凭证（分录行） |
| 5 | 银行流水 | `/bank-transactions` | P0 | 列表 + 状态 |
| 6 | 银行对账 | `/bank-reconciliation` | P1 | 对账单 + 匹配 |
| 7 | 发票管理 | `/invoices` | P1 | 列表 + 状态 |
| 8 | 往来单位 | `/parties` | P1 | 列表 + CRUD |
| 9 | 试算平衡表 | `/reports/trial-balance` | P1 | 报表 |
| 10 | 账套初始化 | `/setup` | P2 | 向导（公司/科目/期初） |

**MVP 目标：优先让「凭证」和「银行流水」两个核心链路可演示。**

### 3.2 页面优先级说明

- **P0**：必须完成，MVP 核心
- **P1**：MVP 后补充，覆盖主要功能
- **P2**：可选，后期扩展

### 3.3 核心页面详细说明

#### 3.3.1 凭证管理（最重要的页面）

**功能：**
1. 列表页（分页、状态筛选、日期范围）
2. 创建页（多行分录，支持借贷双方科目+金额）
3. 详情页（状态展示 + 操作按钮）
4. 状态操作（提交/核准/驳回/红冲）

**凭证状态机：**

```
[录入中(draft,0)] → [提交(submitted,1)] → [核准(approved,2)] → [过账(posted,4)]
                              ↓                              ↓
                         [驳回(3)]                      [红冲(reversed,5)]
```

**凭证录入字段（前端统一格式 → 映射到 Go API）：**

| 前端字段 | Go API 字段 | 说明 |
|---------|-------------|------|
| `voucher_no` | `voucher_no` | 自动生成 |
| `voucher_date` | `posting_date` | 日期 |
| `company_id` | `company_id` | 必填 |
| `description` | `remark` | 摘要 |
| `lines[].account_id` | `account_id` | 科目ID |
| `lines[].debit_amount` | `debit` | 借方金额 |
| `lines[].credit_amount` | `credit` | 贷方金额 |
| `lines[].party_id` | `party_id` | 往来单位（可选） |
| `lines[].remark` | `remark` | 分录备注 |

#### 3.3.2 银行流水

**功能：**
1. 列表（按银行账户筛选，显示匹配状态）
2. 分类操作（对流水进行分类规则匹配）
3. 匹配查看（查看已匹配凭证）

**状态流转：**

```
[未分类] → [已分类] → [已匹配] → [已对账]
```

---

## 4. 开发计划

### 4.1 分阶段

| 阶段 | 内容 | 预计时间 | 交付 |
|------|------|---------|------|
| **Phase 1：项目搭建** | Vue3 + Vite + Element Plus + 路由 + API适配层 | 1天 | 可运行项目骨架 |
| **Phase 2：登录 + 仪表盘** | JWT登录、Token存储、首页概览 | 0.5天 | 登录可用 |
| **Phase 3：凭证管理** | 列表 + 录入 + 状态流转 | 1.5天 | 核心功能可用 |
| **Phase 4：银行流水** | 列表 + 分类 + 对账 | 1天 | 银行对账闭环 |
| **Phase 5：发票+往来+报表** | 发票、往来单位、试算平衡表 | 1天 | 完善功能 |
| **Phase 6：账套初始化** | Setup向导 | 0.5天 | 可完整初始化 |
| **总计** | | **5.5天** | MVP完成 |

### 4.2 Phase 1 详细任务

1. 初始化 Vue 3 + Vite 项目
2. 安装 Element Plus、Pinia、Vue Router、Axios
3. 配置 Vite 代理（`/api` → `http://localhost:8080`）
4. 创建 API 适配层（`adapter/go-backend.js`）
5. 实现字段映射器（`fieldMapper.js`）
6. 创建 axios 实例，统一拦截器处理 JWT
7. 创建路由守卫，未登录跳转登录页
8. 创建基础布局组件（侧边栏 + 顶栏）

### 4.3 技术注意点

1. **Token 存储**：JWT token 存在 `localStorage`，结构兼容 Go 后端的 `user_id` 字段
2. **响应拦截器**：Go 后端返回格式不统一，需要在适配层统一处理
3. **状态值转换**：所有状态显示值通过 `fieldMapper.status()` 转换，不直接用 Go 返回值
4. **日期格式**：前后端统一用 `YYYY-MM-DD`，时间用 `HH:mm:ss`
5. **金额**：前端统一用字符串，精度不丢失

---

## 5. 风险与决策

### 5.1 已知风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| API字段不完全匹配 | 某些字段可能映射不完整 | Phase 1 结束后做一次完整接口验证 |
| 没有设计稿 | 界面可能不符合预期 | 先做功能，再微调样式 |
| Go 后端可能有 bug | 某些接口实际不可用 | 开发阶段持续用集成测试验证 |
| 两套系统 token 不通用 | 用户需要分别登录 | MVP 阶段接受，分别登录 |

### 5.2 不做的事情

- 不做移动端适配
- 不做权限细粒度控制（按角色简单显示/隐藏菜单即可）
- 不做打印、导出 PDF
- 不做微信/飞书集成
- 不做数据导入（Excel 导入留给 Python 版）

---

## 6. 成功标准

MVP 完成当且仅当：

1. ✅ 用户可以通过界面登录 Go 后端
2. ✅ 可以创建、查看、流转凭证（完整链路）
3. ✅ 可以查看银行流水并进行对账（完整链路）
4. ✅ 可以查看财务报表（试算平衡表）
5. ✅ 所有操作有基本的错误提示
6. ✅ 无明显前端 bug（如样式错乱、死循环、内存泄漏）

---

## 7. 参考资料

- API 对比报告：`docs/api-comparison.md`
- Go 后端 MEMORY：`/root/data/disk/huihua-finance/MEMORY.md`
- Python 版前端参考：`/root/data/disk/huihua-financial-master/frontend/src/`
- Go 后端路由：`cmd/api/main.go`
- Go 后端 Schema：`internal/model/`