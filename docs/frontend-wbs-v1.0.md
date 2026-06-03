# 前端 WBS 任务清单总览

**版本**：V1.0
**日期**：2026-05-27
**技术栈**：Vue 3 + TypeScript + Element Plus + Pinia + Vite 5

---

## 总览

| 阶段 | 任务数 | 最低工时 | 最高工时 | 对应后端 |
|:---|:---:|:---:|:---:|:---|
| Phase 0 基础支撑 | 6 | 48h | 72h | P0 (F0.1-F0.2) |
| Phase 1 基础设置 | 4 | 44h | 64h | P1 (F1.1-F1.3) |
| Phase 2 票据采集 | 4 | 48h | 72h | P1 (F2.1-F2.3) |
| Phase 3 核销+凭证 | 4 | 48h | 72h | P1 (F3.1+F5.1-F5.4) |
| Phase 4 期末+对账 | 3 | 44h | 60h | P1+P2 (F6+F4.1) |
| **总计** | **21** | **232h** | **340h** | |

---

## 依赖关系图

```
Phase 0 (基础支撑)
├── FE-TASK-0.1 项目脚手架
├── FE-TASK-0.2 API 请求层 ← 0.1
├── FE-TASK-0.3 认证路由 ← 0.2
├── FE-TASK-0.4 布局权限 ← 0.3
├── FE-TASK-0.5 公共组件 ← 0.2
└── FE-TASK-0.6 Mock 数据层 ← 0.2
         │
         ↓
Phase 1 (基础设置) ← 全部依赖 Phase 0
├── FE-TASK-1.1 账套向导 ← 0.4
├── FE-TASK-1.2 科目表页 ← 0.4 + 0.5
├── FE-TASK-1.3 资金账户 ← 0.4
└── FE-TASK-1.4 客商档案 ← 0.4 + 0.5
         │
         ↓
Phase 2 (票据采集) ← 依赖 Phase 1
├── FE-TASK-2.1 流水导入 ← 1.1 + 1.3
├── FE-TASK-2.2 核对工作台 ← 2.1
├── FE-TASK-2.3 发票管理 ← 1.4
└── FE-TASK-2.4 规则库配置 ← 1.2
         │
         ↓
Phase 3 (核销+凭证)
├── FE-TASK-3.1 核销预检匹配 ← 2.2 + 2.3
├── FE-TASK-3.2 手工核销 ← 3.1
├── FE-TASK-3.3 凭证管理 ← 1.2 + 0.5
└── FE-TASK-3.4 审核工作台 ← 3.3
         │
         ↓
Phase 4 (期末+对账) ← 依赖 Phase 3 + Phase 2
├── FE-TASK-4.1 结账体检 ← 3.4
├── FE-TASK-4.2 财务报表 ← 4.1
└── FE-TASK-4.3 银企对账 ← 2.2 + 3.4
```

---

## 任务清单

### Phase 0 — 基础支撑（48-72h）

| ID | 任务名称 | 工时 | 前置 | 核心产出 |
|:---|:---|:---:|:---:|:---|
| FE-TASK-0.1 | 项目脚手架 | 8-12h | — | Vite 项目初始化、ESLint/Prettier、目录结构、pnpm |
| FE-TASK-0.2 | API 请求层 | 6-10h | 0.1 | axios 实例、拦截器、通用类型、模块 API |
| FE-TASK-0.3 | 认证流程+路由 | 10-14h | 0.2 | 登录页、JWT 存储、路由守卫、Store |
| FE-TASK-0.4 | 布局+权限 | 10-14h | 0.3 | AppLayout、侧边栏/Header、权限指令 |
| FE-TASK-0.5 | 公共业务组件 | 8-12h | 0.2 | 科目选择器、金额输入、状态标签 |
| FE-TASK-0.6 | Mock 数据层 | 6-10h | 0.2 | MSW 配置、各模块 Mock 数据 |

### Phase 1 — 基础设置（44-64h）

| ID | 任务名称 | 工时 | 前置 | 核心产出 |
|:---|:---|:---:|:---:|:---|
| FE-TASK-1.1 | 账套创建向导 | 10-14h | 0.4 | Step 表单、期间设置、初始化科目表 |
| FE-TASK-1.2 | 科目表管理页 | 14-20h | 0.4+0.5 | 科目树、CRUD 弹窗、编码自动生成 |
| FE-TASK-1.3 | 资金账户管理 | 8-12h | 0.4 | 银行账户表格、关联 GL 科目选择 |
| FE-TASK-1.4 | 客商档案 | 12-18h | 0.4+0.5 | 客商列表、导入 Excel、搜索选择器 |

### Phase 2 — 票据采集（48-72h）

| ID | 任务名称 | 工时 | 前置 | 核心产出 |
|:---|:---|:---:|:---:|:---|
| FE-TASK-2.1 | 银行流水导入 | 12-18h | 1.1+1.3 | 文件上传、字段映射、解析预览 |
| FE-TASK-2.2 | 出纳核对工作台 | 16-24h | 2.1 | 分类列表、批量操作、确认流程 |
| FE-TASK-2.3 | 发票管理 | 12-18h | 1.4 | 发票列表、OCR 结果、状态流转 |
| FE-TASK-2.4 | 规则库配置 | 8-12h | 1.2 | 规则列表、正则编辑、优先级拖拽 |

### Phase 3 — 核销+凭证（48-72h）

| ID | 任务名称 | 工时 | 前置 | 核心产出 |
|:---|:---|:---:|:---:|:---|
| FE-TASK-3.1 | 核销预检+匹配 | 12-18h | 2.2+2.3 | 预检清单、Top 3 推荐列表 |
| FE-TASK-3.2 | 手工核销 | 8-14h | 3.1 | 勾选表格、金额分配、回退操作 |
| FE-TASK-3.3 | 凭证管理页 | 16-22h | 1.2+0.5 | 凭证列表、编辑页、分录操作 |
| FE-TASK-3.4 | 审核工作台 | 12-18h | 3.3 | 批量审核/驳回、风险卡片 |

### Phase 4 — 期末+对账（44-60h）

| ID | 任务名称 | 工时 | 前置 | 核心产出 |
|:---|:---|:---:|:---:|:---|
| FE-TASK-4.1 | 结账体检报告 | 12-16h | 3.4 | 10 项检查清单、结账/反结账 |
| FE-TASK-4.2 | 财务报表 | 14-22h | 4.1 | BS/P&L/CF 三表、ECharts 图表、导出 |
| FE-TASK-4.3 | 银企对账 | 18-22h | 2.2+3.4 | 打分看板、余额调节表、对账锁定 |

---

## 文件清单

```
huihua-finance-web/
└── tasks/
    ├── FE-TASK-0.1-project-scaffold.md
    ├── FE-TASK-0.2-api-layer.md
    ├── FE-TASK-0.3-auth-router.md
    ├── FE-TASK-0.4-layout-permission.md
    ├── FE-TASK-0.5-common-components.md
    ├── FE-TASK-0.6-mock-data.md
    ├── FE-TASK-1.1-company-setup.md
    ├── FE-TASK-1.2-account-chart.md
    ├── FE-TASK-1.3-bank-account.md
    ├── FE-TASK-1.4-party-management.md
    ├── FE-TASK-2.1-bank-import.md
    ├── FE-TASK-2.2-cashier-workbench.md
    ├── FE-TASK-2.3-invoice-management.md
    ├── FE-TASK-2.4-rule-library.md
    ├── FE-TASK-3.1-reconciliation-precheck.md
    ├── FE-TASK-3.2-manual-reconciliation.md
    ├── FE-TASK-3.3-voucher-management.md
    ├── FE-TASK-3.4-review-workbench.md
    ├── FE-TASK-4.1-period-close.md
    ├── FE-TASK-4.2-financial-reports.md
    └── FE-TASK-4.3-bank-reconciliation.md
```
