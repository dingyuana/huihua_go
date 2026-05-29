# 汇华财务 Go 版前端 SPEC

> 项目：huihua-finance (Go) 前端界面
> 版本：v2.0.0（Naive UI + TypeScript）
> 日期：2026-05-29
> 状态：**设计稿**，待老丁审核后实施

---

## 变更记录

| 日期 | 版本 | 变更内容 | 作者 |
|------|------|---------|------|
| 2026-05-29 | v1.0.0 | 初稿（Vue3 + Element Plus） | Hermes |
| 2026-05-29 | v2.0.0 | 改用 Naive UI + TypeScript（AI友好） | Hermes |

---

## 1. 技术选型

### 1.1 技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 框架 | Vue 3 | 3.4+ | Composition API + `<script setup>` |
| 类型 | TypeScript | 5.x | AI生成代码质量高，类型即文档 |
| 构建 | Vite | 5.x | 快速HMR，开发体验好 |
| UI库 | Naive UI | 2.x | TypeScript-first，轻量，AI友好 |
| 状态 | Pinia | 2.x | 替代Vuex，store按业务域分 |
| 路由 | Vue Router | 4.x | 路由守卫做权限控制 |
| 请求 | Axios | 1.x | 拦截器统一处理JWT |
| 后端 | Go Fiber | - | localhost:8080 |

### 1.2 AI友好性设计原则

1. **TypeScript-first**：所有接口定义放 `types/index.ts`，AI 生成代码时类型完整
2. **组件单一职责**：一个 `.vue` 文件只做一个功能，AI 修改范围明确
3. **API模式统一**：每个业务域一个 `api/xx.ts`，扁平不嵌套
4. **数据模型集中**：`types/` 目录统一存放所有 TS 接口
5. **组件API一致性**：props/emits/expose 模式统一，AI 可预测
6. **RESTful封装**：按业务域分文件，AI 调用路径清晰

### 1.3 为什么选 Naive UI

- TypeScript-first，类型完整，AI 生成时不丢类型
- 组件设计现代（DatePicker、Table、Tree 比 Element Plus 好用）
- 按需导入，体积可控
- 文档中文友好，国人维护
- 比 Arco Design 更轻量，比 Vuetify 更贴合业务后台

---

## 2. 项目结构

```
frontend/
├── src/
│   ├── api/                    # 按业务域分文件
│   │   ├── index.ts            # axios实例 + 拦截器
│   │   ├── adapter/            # 🔑 Go后端适配层
│   │   │   ├── client.ts       # Go后端专用请求实例
│   │   │   └── fieldMapper.ts  # 字段映射规则
│   │   ├── auth.ts            # 登录/token
│   │   ├── voucher.ts         # 凭证
│   │   ├── invoice.ts         # 发票
│   │   ├── bank.ts            # 银行流水/对账
│   │   ├── party.ts           # 往来单位
│   │   └── report.ts          # 报表
│   ├── types/                 # 所有 TS 接口定义
│   │   └── index.ts           # 统一导出
│   ├── stores/                # Pinia 状态管理
│   │   ├── auth.ts            # 用户/角色状态
│   │   └── app.ts             # 全局状态
│   ├── router/
│   │   └── index.ts           # 路由 + 守卫
│   ├── views/                 # 页面（按业务域分目录）
│   │   ├── auth/
│   │   │   └── Login.vue
│   │   ├── dashboard/
│   │   │   └── Dashboard.vue
│   │   ├── voucher/
│   │   │   ├── VoucherList.vue
│   │   │   ├── VoucherForm.vue
│   │   │   └── VoucherDetail.vue
│   │   ├── bank/
│   │   │   ├── BankTxnList.vue
│   │   │   └── Reconciliation.vue
│   │   ├── invoice/
│   │   │   └── InvoiceList.vue
│   │   ├── party/
│   │   │   └── PartyList.vue
│   │   ├── report/
│   │   │   └── TrialBalance.vue
│   │   └── setup/
│   │       └── SetupWizard.vue
│   ├── components/            # 公共组件
│   │   ├── Layout.vue         # 布局（侧边栏+顶栏）
│   │   ├── StatusTag.vue      # 状态标签
│   │   ├── VoucherForm.vue   # 凭证录入表单
│   │   ├── VoucherList.vue   # 凭证列表
│   │   └── BankTxnList.vue   # 流水列表
│   └── main.ts
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

---

## 3. API 适配层设计（核心）

### 3.1 Go 后端与 Python 后端差异

| 对比项 | Go 后端 | Python 后端 |
|--------|--------|-------------|
| 基础路径 | `/api/v1/` | `/api/v1/` |
| 凭证路径 | `/vouchers` | `/accounting/vouchers` |
| 发票路径 | `/invoices` | `/invoice` |
| 字段名 | `posting_date` | `voucher_date` |
| 借贷方向 | `debit`, `credit` (bool/float) | `dc` ('D'/'C'), `amount` |
| 凭证状态 | int `0,1,2,3,4,5` | string `'draft'` 等 |
| 响应格式 | 直接返回或 `{data}` | `{code, msg, data}` |

### 3.2 适配层结构

```
src/api/adapter/
├── client.ts      # axios实例，指向 http://localhost:8080
└── fieldMapper.ts # 字段映射（Go API → 前端统一格式）
```

**字段映射示例：**

```typescript
// fieldMapper.ts
export const statusMap = {
  0: { label: '待录入', type: 'default' },
  1: { label: '已提交', type: 'warning' },
  2: { label: '已核准', type: 'success' },
  3: { label: '已驳回', type: 'error' },
  4: { label: '已过账', type: 'success' },
  5: { label: '已红冲', type: 'warning' },
}

export const fieldToFrontend = {
  posting_date: 'voucher_date',
  debit: 'debit_amount',
  credit: 'credit_amount',
}

export const fieldToGo = {
  voucher_date: 'posting_date',
  debit_amount: 'debit',
  credit_amount: 'credit',
}
```

---

## 4. 页面清单与优先级

### 4.1 MVP 页面清单

| # | 页面 | 路由 | 优先级 | 说明 |
|---|------|------|--------|------|
| 1 | 登录页 | `/login` | P0 | JWT登录 |
| 2 | 工作台 | `/dashboard` | P0 | 概览统计 |
| 3 | 凭证列表 | `/vouchers` | P0 | 列表+筛选 |
| 4 | 凭证录入 | `/vouchers/new` | P0 | 创建凭证 |
| 5 | 凭证详情 | `/vouchers/:id` | P0 | 查看+状态操作 |
| 6 | 银行流水 | `/bank-transactions` | P0 | 列表+分类 |
| 7 | 银行对账 | `/bank-reconciliation` | P1 | 对账报告+匹配 |
| 8 | 发票管理 | `/invoices` | P1 | 列表+状态 |
| 9 | 往来单位 | `/parties` | P1 | 列表+CRUD |
| 10 | 试算平衡表 | `/reports/trial-balance` | P1 | 报表查看 |
| 11 | 账套初始化 | `/setup` | P2 | Setup向导 |

**MVP目标：凭证+银行流水两个核心链路可演示。**

### 4.2 凭证状态机

```
[录入中(0)] → [已提交(1)] → [已核准(2)] → [已过账(4)]
                    ↓              ↓
              [已驳回(3)]      [已红冲(5)]
```

### 4.3 银行流水状态

```
[未分类] → [已分类] → [已匹配] → [已对账]
```

---

## 5. 开发计划

### 5.1 分阶段

| 阶段 | 内容 | 时间 | 交付 |
|------|------|------|------|
| **Phase 1** | 项目搭建 + API适配层 + 登录 | 1天 | 可登录 |
| **Phase 2** | 凭证管理（列表+录入+详情+状态流转） | 1.5天 | 核心功能闭环 |
| **Phase 3** | 银行流水 + 对账 | 1天 | 银行对账闭环 |
| **Phase 4** | 发票 + 往来单位 + 报表 | 1天 | 完善功能 |
| **Phase 5** | 账套初始化 + Dashboard | 0.5天 | 完整可用 |
| **总计** | | **5天** | MVP完成 |

### 5.2 Phase 1 详细任务

1. 初始化 Vue 3 + Vite + TypeScript 项目
2. 安装 Naive UI、Pinia、Vue Router、Axios
3. 配置 Vite 代理（`/api` → `http://localhost:8080`）
4. 创建 `types/index.ts` — 所有数据模型 TS 接口
5. 创建 `api/adapter/` — 适配层（client + fieldMapper）
6. 创建 axios 实例，统一拦截器处理 JWT
7. 创建 `api/auth.ts` — 登录接口封装
8. 创建路由 `router/index.ts` — 路由守卫
9. 创建 `views/auth/Login.vue` — 登录页
10. 创建 `components/Layout.vue` — 布局组件
11. 创建 `views/dashboard/Dashboard.vue` — 工作台

### 5.3 Phase 2 详细任务

1. 创建 `api/voucher.ts` — 凭证接口封装
2. 创建 `components/VoucherForm.vue` — 凭证录入（多行分录）
3. 创建 `components/VoucherList.vue` — 凭证列表（分页+筛选）
4. 创建 `views/voucher/VoucherList.vue` — 列表页
5. 创建 `views/voucher/VoucherForm.vue` — 录入页
6. 创建 `views/voucher/VoucherDetail.vue` — 详情页（含状态操作按钮）
7. 实现状态流转：提交/核准/驳回/红冲

---

## 6. 技术注意点

1. **Token存储**：`localStorage('go_token')`，结构兼容 Go 后端
2. **状态值转换**：通过 `fieldMapper.statusMap` 转换，不直接用 Go 返回值
3. **金额精度**：前端用 string 或 number.toFixed(2)，不丢失精度
4. **日期格式**：前后端统一 `YYYY-MM-DD`
5. **Naive UI组件选择**：
   - Table → `n-data-table`
   - Form → `n-form` + `n-form-item`
   - Dialog → `n-modal`
   - Tree → `n-tree`
   - DatePicker → `n-date-picker`

---

## 7. 不做的事情

- 不做移动端适配
- 不做权限细粒度控制（按角色简单显示/隐藏菜单）
- 不做打印、导出PDF
- 不做微信/飞书集成
- 不做Excel导入（Go后端可做但前端MVP阶段不接）

---

## 8. 成功标准

MVP完成当且仅当：
1. ✅ 用户可以通过界面登录 Go 后端
2. ✅ 可以创建、查看、流转凭证（完整链路）
3. ✅ 可以查看银行流水并进行对账（完整链路）
4. ✅ 可以查看财务报表（试算平衡表）
5. ✅ 所有操作有错误提示
6. ✅ 无明显前端bug

---

## 9. 参考资料

- API对比报告：`docs/api-comparison.md`
- Go后端MEMORY：`MEMORY.md`
- 设计文档：`docs/02-需求分析/`、`docs/03-技术方案/`
- Go后端路由：`cmd/api/main.go`
- Go后端Schema：`internal/model/`