# 银行流水驱动业财一体化智能财务平台 — 前端架构总览

**版本**：V1.0
**日期**：2026-05-27
**状态**：设计稿（待评审）
**技术栈**：Vue 3 + TypeScript + Element Plus + Pinia + Vite

---

## 1. 设计原则

### 1.1 核心原则

| 原则 | 含义 |
|:---|:---|
| **模块独立** | 每个功能模块独立路由、独立 Store、可插拔（与后端 F0-F8 模块一一对应） |
| **角色驱动** | 路由、菜单、按钮、数据字段均按 RBAC 角色动态控制，不渲染无权限元素 |
| **多租户透明** | tenant_id 在 JWT 中自动携带，前端不做显式处理（切换时重置 Store） |
| **离线可用** | 核心表单（流水核对、凭证录入）在网络抖动时自动保存草稿到 localStorage |
| **IM 先行** | Web 端页面设计时考虑"IM 端查询也可以访问该页面"，有移动端适配预留 |
| **响应式表格** | 财务系统以表格为核心交互，所有列表页支持列自定义、固定列、排序、筛选 |

---

## 2. 项目目录结构

```
huihua-finance-web/
├── index.html
├── vite.config.ts
├── tsconfig.json
├── .env.development
├── .env.production
│
├── public/
│   ├── favicon.ico
│   └── tenant-logos/              # 租户 Logo（代账会计切换时动态加载）
│
├── src/
│   ├── main.ts                    # 入口：创建 App、注册 Router/Pinia/全局组件
│   ├── App.vue                    # 根组件：<router-view /> + GlobalErrorBoundary
│   ├── env.d.ts                   # 环境变量类型扩展
│   │
│   ├── types/                     # 📁 TypeScript 类型定义（与后端 API 对齐）
│   │   ├── api/                   #    API 请求/响应体类型
│   │   │   ├── account.ts         #    科目表相关
│   │   │   ├── bank.ts            #    银行流水相关
│   │   │   ├── invoice.ts         #    发票相关
│   │   │   ├── payment.ts         #    收付款相关
│   │   │   ├── voucher.ts         #    凭证相关
│   │   │   ├── reconciliation.ts  #    核销相关
│   │   │   ├── period.ts          #    期间/结账相关
│   │   │   └── report.ts          #    报表相关
│   │   ├── models/                #    领域模型
│   │   │   ├── account.ts         #    Account, AccountType enum
│   │   │   ├── journal.ts         #    JournalEntry, JournalEntryLine
│   │   │   ├── bank.ts            #    BankTransaction, BankAccount
│   │   │   ├── invoice.ts         #    SalesInvoice, InvoiceStatus
│   │   │   ├── payment.ts         #    PaymentEntry, PaymentType
│   │   │   ├── party.ts           #    Customer, Supplier, Employee
│   │   │   ├── tenant.ts          #    Tenant, Company
│   │   │   └── user.ts            #    User, Role, Permission
│   │   ├── store/                 #    Pinia Store 类型
│   │   │   └── index.ts
│   │   ├── enums.ts               #    全局枚举（DocStatus, AccountType 等）
│   │   ├── api.ts                 #    通用 API 类型（PageResult, ApiResponse）
│   │   └── router.ts              #    路由元信息类型（RouteMeta, Permission）
│   │
│   ├── config/                    # 📁 应用配置
│   │   ├── app.config.ts          #    应用级常量（API_BASE_URL, 分页默认值）
│   │   ├── menu.config.ts         #    菜单配置（按角色分组）
│   │   ├── dict.config.ts         #    字典映射（单据类型、流水分类、凭证类型）
│   │   └── theme.config.ts        #    主题配置（品牌色、字体、圆角）
│   │
│   ├── router/                    # 📁 路由
│   │   ├── index.ts               #    路由实例 + 全局守卫
│   │   ├── routes/                #    路由定义（按模块拆分）
│   │   │   ├── base.ts            #    基础路由（登录、首页、404）
│   │   │   ├── setup.ts           #    F1 基础设置
│   │   │   ├── bank.ts            #    F2 票据采集
│   │   │   ├── reconciliation.ts  #    F3 核销
│   │   │   ├── voucher.ts         #    F5 凭证
│   │   │   ├── period.ts          #    F6 期末处理
│   │   │   └── report.ts          #    F7 经营分析
│   │   └── guards.ts              #    守卫：认证守卫 + 权限守卫 + 租户守卫
│   │
│   ├── stores/                    # 📁 Pinia 状态管理
│   │   ├── app.store.ts           #    全局应用状态（loading、sidebar、theme）
│   │   ├── auth.store.ts          #    认证状态（token、user、role、permissions）
│   │   ├── tenant.store.ts        #    多租户状态（当前 tenant、公司列表）
│   │   ├── modules/               #    业务模块 Store
│   │   │   ├── account.store.ts   #    科目表缓存
│   │   │   ├── bank.store.ts      #    银行流水
│   │   │   ├── invoice.store.ts   #    发票
│   │   │   ├── voucher.store.ts   #    凭证
│   │   │   ├── reconciliation.store.ts # 核销
│   │   │   ├── period.store.ts    #    期间/结账
│   │   │   └── report.store.ts    #    报表
│   │   └── plugins/               #    Pinia 插件
│   │       ├── tenant-reset.ts    #    切换租户时重置所有业务 Store
│   │       └── auto-persist.ts    #    关键表单自动保存到 localStorage
│   │
│   ├── api/                       # 📁 API 请求层
│   │   ├── request.ts             #    axios 实例（拦截器、Token 注入、错误处理）
│   │   ├── modules/               #    模块 API（每个文件导出函数）
│   │   │   ├── account.ts         #    /api/v1/accounts/*
│   │   │   ├── bank.ts            #    /api/v1/bank/*
│   │   │   ├── invoice.ts         #    /api/v1/invoices/*
│   │   │   ├── payment.ts         #    /api/v1/payments/*
│   │   │   ├── voucher.ts         #    /api/v1/vouchers/*
│   │   │   ├── reconciliation.ts  #    /api/v1/reconciliation/*
│   │   │   ├── period.ts          #    /api/v1/periods/*
│   │   │   ├── report.ts          #    /api/v1/reports/*
│   │   │   └── tenant.ts          #    /api/v1/tenants/*
│   │   └── mock/                  #    Mock 数据（开发阶段使用）
│   │       ├── accounts.ts        #    科目表 Mock
│   │       ├── bank-transactions.ts
│   │       ├── invoices.ts
│   │       └── vouchers.ts
│   │
│   ├── hooks/                     # 📁 组合式函数（Composables）
│   │   ├── usePage.ts             #    分页查询通用逻辑
│   │   ├── useDict.ts             #    字典解析（枚举→标签）
│   │   ├── usePermission.ts       #    权限判断（v-permission 指令底层）
│   │   ├── useTenant.ts           #    租户上下文
│   │   ├── useCurrency.ts         #    金额格式化（千分位、货币符号）
│   │   ├── useDebounce.ts         #    防抖（搜索框等场景）
│   │   ├── useConfirm.ts          #    操作确认弹窗（带备注输入框）
│   │   └── useFormDraft.ts        #    表单自动草稿恢复
│   │
│   ├── components/                # 📁 公共组件
│   │   ├── app/                   #    应用级组件
│   │   │   ├── AppLayout.vue      #    主布局（Sidebar + Header + Content）
│   │   │   ├── AppSidebar.vue     #    侧边栏菜单（按角色过滤）
│   │   │   ├── AppHeader.vue      #    顶部导航（租户选择器 + 用户信息 + 通知）
│   │   │   ├── AppTabs.vue        #    多标签页（Tab 式页面切换）
│   │   │   └── AppFooter.vue      #    底部状态栏
│   │   ├── business/              #    业务组件
│   │   │   ├── AccountTree.vue    #    科目树选择器（弹窗模式）
│   │   │   ├── AccountSelector.vue#    科目选择器（搜索+树形）
│   │   │   ├── PartySelector.vue  #    客商选择器（模糊搜索+税号）
│   │   │   ├── BankAccountSelect.vue# 银行账户选择器
│   │   │   ├── PeriodPicker.vue   #    会计期间选择器
│   │   │   ├── AmountInput.vue    #    金额输入（千分位+双币种）
│   │   │   ├── DocStatusTag.vue   #    单据状态标签（Draft/Submitted/Cancelled）
│   │   │   ├── VoucherNoDisplay.vue# 凭证编号展示+跳转
│   │   │   └── UploadFile.vue     #    文件上传（银行流水/发票）
│   │   ├── common/                #    通用组件
│   │   │   ├── PageHeader.vue     #    页面标题+操作按钮区
│   │   │   ├── PageTable.vue      #    高级表格（分页、排序、列自定义）
│   │   │   ├── SearchForm.vue     #    搜索表单面板
│   │   │   ├── ConfirmDialog.vue  #    确认弹窗（带备注输入）
│   │   │   ├── EmptyState.vue     #    空状态展示
│   │   │   └── LoadingSkeleton.vue#    加载骨架屏
│   │   └── charts/                #    图表组件（报表模块）
│   │       ├── BarChart.vue
│   │       ├── LineChart.vue
│   │       └── PieChart.vue
│   │
│   ├── directives/                # 📁 自定义指令
│   │   ├── permission.ts          #    v-permission="['cashier']" 按钮级权限
│   │   ├── role.ts                #    v-role="'admin'" 角色判断
│   │   ├── number.ts              #    v-number 金额自动格式化
│   │   └── focus.ts               #    v-focus 自动聚焦
│   │
│   ├── utils/                     # 📁 工具函数
│   │   ├── format.ts              #    日期/金额/手机号格式化
│   │   ├── validate.ts            #    表单校验规则（金额、税号、账号）
│   │   ├── tree.ts                #    树形数据处理（列表→树、展开/折叠）
│   │   ├── excel.ts               #    Excel 导入/导出工具
│   │   ├── crypto.ts              #    前端加密（敏感字段脱敏）
│   │   └── watermark.ts           #    水印工具（代账会计场景）
│   │
│   ├── styles/                    # 📁 样式
│   │   ├── variables.scss         #    SCSS 变量（覆盖 Element Plus 主题）
│   │   ├── reset.scss             #    浏览器默认样式重置
│   │   ├── transitions.scss       #    过渡动画
│   │   ├── print.scss             #    打印样式（凭证、报表）
│   │   └── business.scss          #    业务专用样式（科目树、日记账格式）
│   │
│   └── views/                     # 📁 页面（按模块分组）
│       ├── login/
│       │   └── LoginView.vue      #    登录页
│       ├── dashboard/
│       │   └── DashboardView.vue  #    首页看板
│       ├── setup/                 #    F1 基础设置
│       │   ├── CompanySetup.vue   #    账套创建向导
│       │   ├── AccountChart.vue   #    科目表管理（树 + CRUD）
│       │   ├── BankAccountList.vue#    资金账户管理
│       │   ├── PartyList.vue      #    客商档案
│       │   ├── RuleLibrary.vue    #    智能分类规则库
│       │   └── MappingRules.vue   #    科目映射规则
│       ├── bank/                  #    F2 票据采集
│       │   ├── ImportView.vue     #    流水导入上传
│       │   ├── FieldMapping.vue   #    字段映射配置
│       │   ├── CashierWorkbench.vue# 出纳核对工作台
│       │   ├── InvoiceList.vue    #    发票列表
│       │   └── InvoiceDetail.vue  #    发票详情（OCR 结果）
│       ├── reconciliation/        #    F3 核销
│       │   ├── PreCheckView.vue   #    核销预检看板
│       │   ├── MatchView.vue      #    匹配推荐列表
│       │   └── ManualMatch.vue    #    手工核销
│       ├── voucher/               #    F5 凭证
│       │   ├── TemplateList.vue   #    凭证模板
│       │   ├── VoucherList.vue    #    凭证列表
│       │   ├── VoucherEdit.vue    #    凭证编辑/查看
│       │   ├── ReviewWorkbench.vue#    审核工作台
│       │   └── BatchGenerate.vue  #    批量生成
│       ├── reconciliation-bank/   #    F4 银企对账
│       │   ├── MatchingView.vue   #    打分匹配看板
│       │   └── BalanceSheet.vue   #    余额调节表
│       ├── period/                #    F6 期末处理
│       │   ├── HealthCheck.vue    #    结账体检报告
│       │   ├── DepreciationRun.vue#    折旧执行
│       │   └── FinancialReports.vue# 财务报表（BS/P&L/CF）
│       ├── analytics/             #    F7 经营分析
│       │   └── AnalyticsView.vue  #    经营看板
│       └── error/
│           ├── 403.vue            #    无权限
│           └── 404.vue            #    未找到
│
├── tests/
│   ├── unit/                      #    单元测试
│   │   ├── stores/
│   │   ├── components/
│   │   └── utils/
│   └── e2e/                       #    E2E 测试
│
├── mock/                          # Mock Service Worker 配置
├── .eslintrc.cjs
├── .prettierrc
├── commitlint.config.js
└── README.md
```

---

## 3. 布局系统

### 3.1 布局层次

```
┌─────────────────────────────────────────────────────┐
│  AppHeader                                           │
│  ┌───────┬─────────────────────────────────────────┐ │
│  │       │  PageHeader (标题 + 操作按钮)            │ │
│  │ App   ├─────────────────────────────────────────┤ │
│  │ Side  │  SearchForm (搜索条件)                   │ │
│  │ bar   ├─────────────────────────────────────────┤ │
│  │       │                                         │ │
│  │ 菜    │  PageTable / 表单 / 详情区              │ │
│  │ 单    │                                         │ │
│  │       │                                         │ │
│  └───────┴─────────────────────────────────────────┘ │
│  AppFooter                                           │
└─────────────────────────────────────────────────────┘
```

### 3.2 布局变体

| 布局模式 | 适用场景 | 说明 |
|:---|:---|:---|
| `default` | 列表页、工作台 | 侧边栏展开（200px）+ 顶部栏 + 内容区 |
| `collapsed` | 详情页、编辑页 | 侧边栏折叠（64px），内容区最大化 |
| `fullscreen` | 审核工作台、向导式 | 隐藏侧边栏，仅顶部栏 |
| `blank` | 登录页 | 无侧边栏、无顶部栏 |

通过 `layout` 字段在路由 `meta` 中指定：

```typescript
const routeMeta: RouteMeta = {
  title: '科目表管理',
  layout: 'default',
  roles: ['admin', 'accountant'],
  permissions: ['account:read'],
  keepAlive: true,
}
```

---

## 4. 路由设计

### 4.1 路由树（完整）

```
/login                        → LoginView              [所有人，无布局]
/                             → DashboardView           [所有角色]

/setup                        → SetupLayout
  /setup/company              → CompanySetup            [admin, 代账会计]
  /setup/accounts             → AccountChart            [admin, 代账会计]
  /setup/bank-accounts        → BankAccountList         [admin, 出纳, 代账会计]
  /setup/parties              → PartyList               [admin, 应收/应付会计]
  /setup/rules                → RuleLibrary             [admin]
  /setup/mapping-rules        → MappingRules            [admin]

/bank                         → BankLayout
  /bank/import                → ImportView              [出纳, 代账会计]
  /bank/field-mapping         → FieldMapping            [出纳, admin]
  /bank/workbench             → CashierWorkbench        [出纳, 代账会计]

/invoices                     → InvoiceLayout
  /invoices                   → InvoiceList             [应收/应付会计, 代账会计]
  /invoices/:id               → InvoiceDetail           [应收/应付会计, 代账会计]

/reconciliation               → ReconciliationLayout
  /reconciliation/precheck    → PreCheckView            [应收/应付会计]
  /reconciliation/match       → MatchView               [应收/应付会计]
  /reconciliation/manual      → ManualMatch             [应收/应付会计]

/bank-reconciliation          → BankRecLayout
  /bank-reconciliation/match  → MatchingView            [出纳, 代账会计]
  /bank-reconciliation/balance→ BalanceSheet            [出纳, admin]

/vouchers                     → VoucherLayout
  /vouchers/templates         → TemplateList            [admin]
  /vouchers                   → VoucherList             [admin, 代账会计]
  /vouchers/create            → VoucherEdit             [admin, 代账会计]
  /vouchers/:id               → VoucherEdit             [admin, 代账会计]
  /vouchers/review            → ReviewWorkbench         [admin]
  /vouchers/batch-generate    → BatchGenerate           [admin]

/period                       → PeriodLayout
  /period/health-check        → HealthCheck             [admin]
  /period/depreciation        → DepreciationRun         [admin]
  /period/reports             → FinancialReports        [admin, 老板]

/analytics                    → AnalyticsView           [老板, admin]
```

### 4.2 路由守卫链

```
NavigationGuard Chain:
  ① AuthGuard       → 检查 token 是否存在、是否过期
  ② TenantGuard     → 检查当前是否已选择租户（代账会计场景）
  ③ PermissionGuard → 检查角色是否有该路由权限，无则跳转 403
  ④ KeepAliveGuard  → 路由切换时缓存/销毁页面组件
  ⑤ TitleGuard      → 动态设置 document.title
```

```typescript
// guards.ts — 核心守卫逻辑
router.beforeEach(async (to, from, next) => {
  // ① 认证
  const authStore = useAuthStore()
  if (!authStore.isLoggedIn && to.path !== '/login') {
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }

  // ② 租户选择（代账会计必须显式选择客户账套）
  const tenantStore = useTenantStore()
  if (authStore.user?.role === 'agent' && !tenantStore.currentTenantId) {
    return next({ path: '/tenant-switch' })
  }

  // ③ 路由权限
  const allowedRoles = to.meta.roles as string[]
  if (allowedRoles && !allowedRoles.includes(authStore.user?.role ?? '')) {
    return next({ path: '/403' })
  }

  next()
})
```

---

## 5. 状态管理（Pinia）

### 5.1 Store 层次

```
├── app.store.ts          # 全局 UI 状态
│   ├── sidebarCollapsed  # 侧边栏折叠
│   ├── currentLayout     # 当前布局模式
│   ├── globalLoading     # 全局加载
│   └── theme             # 主题配置
│
├── auth.store.ts         # 🔑 认证（核心）
│   ├── token             # JWT
│   ├── user              # User { id, name, email, role }
│   ├── permissions       # string[] 如 ['account:read', 'voucher:write']
│   ├── login()           # 登录 → 保存 token + user
│   ├── logout()          # 登出 → 清除 + 跳转
│   └── fetchPermissions() # 从后端加载权限列表
│
├── tenant.store.ts       # 🏢 多租户（代账会计专用）
│   ├── currentTenantId   # 当前操作客户账套
│   ├── currentCompany    # 当前公司信息
│   ├── tenantList        # 代账会计管辖的客户列表
│   ├── switchTenant()    # 切换 → 重置所有业务 Store
│   └── getWatermark()    # 返回水印文本「当前客户: XXX」
│
├── modules/
│   ├── account.store.ts  # 科目表
│   │   ├── tree          # 科目树（扁平+树形双向缓存）
│   │   ├── ledgerOnly    # 仅可记账科目（is_group=false）
│   │   ├── fetchTree()   # 加载整棵树
│   │   └── getPath(code) # 返回科目编码的全路径
│   │
│   ├── bank.store.ts     # 银行流水
│   │   ├── transactionList # 当前批次流水
│   │   ├── importResult  # 上次导入结果
│   │   ├── classifications # 分类结果
│   │   ├── importFile()  # 上传+解析
│   │   └── classify()    # 执行智能分类
│   │
│   ├── voucher.store.ts  # 凭证
│   │   ├── currentVoucher # 当前编辑中的凭证
│   │   ├── draftCache    # 草稿自动保存
│   │   ├── submit()      # 提交审核
│   │   └── reverse()     # 红字冲销
│   │
│   └── ... 其他模块 Store
```

### 5.2 多租户切换时的 Store 重置

```typescript
// plugins/tenant-reset.ts
export const tenantResetPlugin: PiniaPlugin = (context) => {
  const tenantStore = useTenantStore()

  // 当 currentTenantId 变化时，重置所有非 auth/tenant 的 Store
  watch(() => tenantStore.currentTenantId, () => {
    const nonSystemStores = ['account', 'bank', 'invoice', 'voucher', 'reconciliation', 'period', 'report']
    nonSystemStores.forEach(name => {
      const store = useStore(name)
      store.$reset()
    })
  })
}
```

---

## 6. API 请求层

### 6.1 axios 封装

```typescript
// api/request.ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth.store'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截器：注入 JWT
request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

// 响应拦截器：统一错误处理
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore().logout()
      return Promise.reject(error)
    }
    if (error.response?.status === 403) {
      ElMessage.error('权限不足，请联系管理员')
    }
    // 业务错误统一显示
    const msg = error.response?.data?.message || '请求失败'
    ElMessage.error(msg)
    return Promise.reject(error)
  }
)

export default request
```

### 6.2 通用 API 响应类型

```typescript
// types/api.ts
/** 后端统一响应格式 */
export interface ApiResponse<T = unknown> {
  code: number        // 0=成功, 非0=业务错误
  message: string
  data: T
}

/** 分页请求 */
export interface PageQuery {
  page: number
  pageSize: number
  sort?: string       // "posting_date desc"
}

/** 分页响应 */
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
```

### 6.3 模块 API 示例

```typescript
// api/modules/account.ts
import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { Account } from '@/types/models/account'

/** 获取科目树（所有节点） */
export function fetchAccountTree(): Promise<ApiResponse<Account[]>> {
  return request.get('/accounts/tree')
}

/** 查询科目列表（分页） */
export function fetchAccountList(params: PageQuery & { keyword?: string }): Promise<ApiResponse<PageResult<Account>>> {
  return request.get('/accounts', { params })
}

/** 创建科目 */
export function createAccount(data: Partial<Account>): Promise<ApiResponse<Account>> {
  return request.post('/accounts', data)
}

/** 更新科目 */
export function updateAccount(id: string, data: Partial<Account>): Promise<ApiResponse<Account>> {
  return request.put(`/accounts/${id}`, data)
}

/** 删除科目 */
export function deleteAccount(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/accounts/${id}`)
}
```

---

## 7. 权限系统

### 7.1 权限映射

每个 API 端点 + 页面 + 操作按钮通过 `role` 和 `permission` 两层控制：

```typescript
// 角色 → 路由菜单映射
const roleMenuMap: Record<Role, MenuItem[]> = {
  cashier: [
    { path: '/bank/import', title: '流水导入', icon: 'Upload' },
    { path: '/bank/workbench', title: '核对工作台', icon: 'List' },
    { path: '/bank-reconciliation/match', title: '银企对账', icon: 'BalanceTwo' },
  ],
  accountant_ar: [
    { path: '/invoices', title: '发票管理', icon: 'Document' },
    { path: '/reconciliation/precheck', title: '核销中心', icon: 'Link' },
  ],
  admin: [
    { path: '/setup/accounts', title: '科目表', icon: 'Collection' },
    { path: '/vouchers', title: '凭证管理', icon: 'Notebook' },
    { path: '/vouchers/review', title: '审核工作台', icon: 'Select' },
    { path: '/period/health-check', title: '结账', icon: 'Timer' },
    { path: '/period/reports', title: '财务报表', icon: 'DataAnalysis' },
  ],
  boss: [
    { path: '/analytics', title: '经营分析', icon: 'TrendCharts' },
  ],
  employee: [
    { path: '/expense/reimbursement', title: '我的报销', icon: 'Money' },
  ],
  agent: [
    // 继承 admin 的菜单，但数据范围受限 + 水印
    ...adminMenuItems,
  ],
}
```

### 7.2 按钮级权限

```vue
<!-- 使用 v-permission 指令 -->
<el-button v-permission="['voucher:submit']" type="primary" @click="submit">
  提交审核
</el-button>
<el-button v-permission="['voucher:reverse']" type="warning" @click="reverse">
  红字冲销
</el-button>

<!-- 无权限时按钮被 v-if 移除（不渲染） -->
```

```typescript
// directives/permission.ts
const permissionDirective: Directive = {
  mounted(el, binding) {
    const authStore = useAuthStore()
    const requiredPermissions = binding.value as string[]
    const hasPermission = requiredPermissions.some(p => authStore.permissions.includes(p))
    if (!hasPermission) {
      el.parentNode?.removeChild(el)
    }
  }
}
```

---

## 8. 多租户前端处理

### 8.1 代账会计的工作模式

```
登录 → 选择一家客户公司 → 进入该公司账套
                          ↓
                   界面显示水印：「当前操作：XX 公司」
                          ↓
                   所有 API 请求自动携带该租户的 JWT
                          ↓
                   切换公司 → 重置 Store → 重新加载数据
```

### 8.2 多租户安全约束

| 约束 | 实现方式 |
|:---|:---|
| 禁止跨租户数据泄露 | tenant_id 在后端 RLS 强制过滤，前端不关心 |
| 切换留痕 | 每次切换调用后端 API 记录 audit_log |
| 操作可视化隔离 | 界面水印 + 顶部栏显示当前公司名 + 高亮边框 |
| 缓存隔离 | 切换租户时 Pinia 业务 Store 全部 $reset |

---

## 9. 公共业务组件

### 9.1 科目选择器 `AccountSelector.vue`

```
┌─────────────────────────────────────────┐
│ 🔍 搜索科目名称或编码                    │
├─────────────────────────────────────────┤
│ 1001 银行存款                            │
│ ├─ 1001-01 银行存款-工行                 │
│ │  └─ 1001-01-01 银行存款-工行-人民币    │  ← 🔵 Ledger（可记账）
│ ├─ 1001-02 银行存款-建行                 │  ← 🟡 Group（不可记账）
│ 1002 应收账款                             │
│ 2001 应付账款                             │
└─────────────────────────────────────────┘
```

- **搜索过滤**：输入编码或名称即时过滤树节点
- **颜色区分**：Group 科目浅黄色+斜体、Ledger 科目白色+正常
- **禁用选择**：Group 科目点击无效并提示"汇总科目不可记账"
- **支持** `v-model` 双向绑定

### 9.2 金额输入 `AmountInput.vue`

```vue
<template>
  <el-input
    :model-value="formattedValue"
    @input="onInput"
    placeholder="请输入金额"
  >
    <template #prefix>
      <el-tag size="small">{{ currencySymbol }}</el-tag>
    </template>
    <template #suffix>
      <span v-if="showBalance" class="balance-text">
        余额: {{ formattedBalance }}
      </span>
    </template>
  </el-input>
</template>
```

- 输入时实时千分位格式化
- 支持双币种（原币金额 + 本位币金额同步展示）
- 支持 `max` 限制（核销时金额不可超过未结清金额）
- 负数显示红色

### 9.3 客商选择器 `PartySelector.vue`

- 输入 2 个字符后触发远程搜索
- 结果展示：名称 | 税号 | 开户行（三列）
- 已选后显示标签，可清除
- 支持客商类型过滤：`customer` / `supplier` / `both`

---

## 10. 关键页面流程

### 10.1 银行流水导入 → 凭证生成（全链路）

```
[ImportView]             → 上传 CSV/Excel → 后端解析
    ↓
[FieldMapping]           → 自动映射字段 / 手动修正
    ↓
[CashierWorkbench]       → 智能分类展示 → 逐笔确认
    ↓                              ↓ 修正分类
  6 类单据确认                       待处理 → 手工指派
    ↓
[系统自动]               → 生成单据草稿
    ↓
[系统自动]               → 核销（满足 L1/L2 条件时）
    ↓
[系统自动]               → 凭证自动生成（Draft）
    ↓
[ReviewWorkbench]        → 审核 → Submitted
```

### 10.2 结账体检流程

```
[HealthCheck.vue]
┌──────────────────────────────────────────────────────────┐
│ 🔴 结账前体检报告 — 2026年05月                          │
├──┬────────────────────────────┬──────────┬───────────────┤
│ #│ 检查项                     │ 状态     │ 操作          │
├──┼────────────────────────────┼──────────┼───────────────┤
│ 1│ 凭证借贷平衡               │ ✅ 通过  │               │
│ 2│ 凭证完整性（无未审核）     │ ⚠️ 3张待审│ 点击查看 →   │
│ 3│ 凭证编号连续性             │ ✅ 通过  │               │
│ 4│ 固定资产折旧               │ ❌ 1笔未过账│ 立即处理→   │
│ 5│ 银行日记账一致性           │ ✅ 通过  │               │
│ 6│ 现金账实一致               │ 🔵 未盘点│ 录入盘点→   │
│ 7│ 往来核销完成度             │ ❌ 2笔超30天│ 查看清单→  │
│ 8│ 进项发票到期               │ ⚠️ 3张快到期│ 查看→      │
│ 9│ 损益结转                   │ ❌ 未结转 │ 生成结转凭证→│
│10│ 期间锁定状态               │ ✅ 正常  │               │
├──┴────────────────────────────┴──────────┴───────────────┤
│ 总评：🔴 3项阻断项，需全部修复后才能结账                   │
│              [ 导出报告 ]          [ 开始结账(禁用) ]       │
└──────────────────────────────────────────────────────────┘
```

---

## 11. 与后端 API 契约规范

| 规范项 | 约定 |
|:---|:---|
| Base URL | `/api/v1` |
| 响应格式 | `{ code: 0, message: "ok", data: {...} }` |
| 分页 | `{ list: [...], total: number, page: number, pageSize: number }` |
| 认证方式 | `Authorization: Bearer <JWT>` |
| 时间格式 | ISO 8601，UTC（前端按用户时区转换显示） |
| 金额格式 | `Decimal(18,2)` 字符串（避免浮点精度丢失） |
| 枚举值 | 后端传 code，前端通过 `dict.config.ts` 映射为中文标签 |
| 错误码 | `0` 成功，`10xxx` 认证，`20xxx` 业务校验，`30xxx` 权限 |
| 文件上传 | `multipart/form-data`，返回 `{ fileId, fileName, url }` |

---

## 12. 开发阶段建议

### 第一阶段（对齐后端 P0 — 基础支撑）

| 前端任务 | 对应后端 | 估算 |
|:---|:---|:---:|
| Vue 3 项目脚手架（Vite + TS + ESLint + Prettier） | F0.1 | 8h |
| Axios 封装 + 拦截器 + 统一错误处理 | F0.1 | 4h |
| 登录页 + JWT 认证流程 + 路由守卫 | F0.1 | 8h |
| AppLayout（侧边栏 + Header + 内容区） | — | 8h |
| 多租户切换逻辑 + Store 重置 | F0.1 | 4h |
| 公共业务组件（AccountSelector + AmountInput + DocStatusTag） | F0.2 | 12h |
| 权限指令 `v-permission` + 菜单过滤 | — | 4h |
| **小计** | | **48h** |

### 第二阶段（P1 — 科目表 + 流水导入）

| 前端任务 | 对应后端 | 估算 |
|:---|:---|:---:|
| 科目表页面（树状展示 + CRUD 弹窗） | F1.1 | 16h |
| 账套创建向导（Step 表单） | F1.1 | 12h |
| 银行流水上传 + 字段映射页面 | F2.1 | 12h |
| 出纳核对工作台（分类列表 + 批量操作） | F2.2 | 16h |
| 客商档案页面（列表 + 导入） | F1.1 | 8h |
| 规则库配置页面 | F1.3 | 8h |
| **小计** | | **72h** |

### 第三阶段（P1 — 核销 + 凭证）

| 前端任务 | 对应后端 | 估算 |
|:---|:---|:---:|
| 核销预检看板（6 项清单） | F3.1 | 8h |
| 匹配推荐列表（Top 3） | F3.1 | 8h |
| 手工核销页面（勾选+金额分配） | F3.1 | 12h |
| 凭证列表 + 编辑页 | F5.3 | 16h |
| 审核工作台（批量审核/驳回） | F5.4 | 12h |
| 凭证模板配置页 | F5.1 | 8h |
| **小计** | | **64h** |

### 第四阶段（P1 — 期末 + 报表）

| 前端任务 | 对应后端 | 估算 |
|:---|:---|:---:|
| 结账体检报告页（10 项清单） | F6.3 | 12h |
| 财务报表 BS/P&L/CF | F6.5 | 16h |
| 银企对账页面（打分匹配 + 调节表） | F4.1 | 16h |
| **小计** | | **44h** |

**前端总估算**：约 **228h**（最小）~ **350h**（最大），约 **5.7~8.7 人月**

> 注：IM 端（微信/钉钉）的单独前端项目在前端架构设计范围外，需单独评估。

---

## 13. 附录：技术选型理由

| 选型 | 选择 | 理由 |
|:---|:---|:---|
| 构建工具 | Vite 5 | 开发冷启动 < 1 秒，HMR 即时，ESBuild 编译极快 |
| 框架 | Vue 3.4+ Composition API | 现有团队熟悉，`<script setup>` 模式比 Options API 更简洁 |
| UI 库 | Element Plus | 财务系统需要成熟的表格/表单/弹窗组件，EP 在中文场景最成熟 |
| 状态管理 | Pinia | 官方推荐，TypeScript 友好，支持 DevTools |
| 路由 | Vue Router 4 | 官方路由，支持动态路由、路由元信息、导航守卫 |
| HTTP | Axios | 拦截器机制成熟，请求/响应统一处理 |
| 图表 | ECharts 5 | 财务报表需要复杂的图表类型（柱状+折线混合、堆叠图） |
| 样式 | SCSS + CSS Variables | Element Plus 主题定制需要 SCSS，CSS Variables 实现运行时主题 |
| 测试 | Vitest (UT) + Playwright (E2E) | Vitest 与 Vite 共享配置，Playwright 多浏览器支持 |
| 代码规范 | ESLint + Prettier + Husky + lint-staged | 提交前自动格式化，保证代码风格一致 |
| 包管理 | pnpm | 速度快，依赖隔离好（避免幽灵依赖） |
