import type { RouteRecordRaw } from 'vue-router'
import AppLayout from '@/components/app/AppLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录', layout: 'blank' },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '无权限', layout: 'blank' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '页面不存在', layout: 'blank' },
  },
  {
    path: '/',
    component: AppLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: '首页', keepAlive: true },
      },
      // ==== F1 基础设置 ====
      {
        path: 'setup/company',
        name: 'SetupCompany',
        component: () => import('@/views/setup/SetupWizard.vue'),
        meta: { title: '创建账套', roles: ['admin'] },
      },
      {
        path: 'setup/accounts',
        name: 'AccountChart',
        component: () => import('@/views/setup/AccountChart.vue'),
        meta: { title: '科目表', roles: ['admin', 'agent'], keepAlive: true },
      },
      {
        path: 'setup/bank-accounts',
        name: 'BankAccountList',
        component: () => import('@/views/setup/BankAccountList.vue'),
        meta: { title: '资金账户', roles: ['admin', 'cashier', 'agent'] },
      },
      {
        path: 'setup/parties',
        name: 'PartyList',
        component: () => import('@/views/setup/PartyList.vue'),
        meta: { title: '客商档案', roles: ['admin', 'accountant_ar', 'agent'], keepAlive: true },
      },
      {
        path: 'setup/rules',
        name: 'RuleLibrary',
        component: () => import('@/views/setup/RuleLibrary.vue'),
        meta: { title: '规则库', roles: ['admin'] },
      },
      // ==== F2 票据采集 ====
      {
        path: 'bank/import',
        name: 'BankImport',
        component: () => import('@/views/bank/ImportView.vue'),
        meta: { title: '流水导入', roles: ['cashier', 'admin', 'agent'] },
      },
      {
        path: 'bank/workbench',
        name: 'CashierWorkbench',
        component: () => import('@/views/bank/CashierWorkbench.vue'),
        meta: { title: '核对工作台', roles: ['cashier', 'admin', 'agent'] },
      },
      {
        path: 'invoices',
        name: 'InvoiceList',
        component: () => import('@/views/invoices/InvoiceList.vue'),
        meta: { title: '发票管理', roles: ['accountant_ar', 'admin', 'agent'], keepAlive: true },
      },
      // ==== F3 核销 ====
      {
        path: 'reconciliation/precheck',
        name: 'ReconciliationPrecheck',
        component: () => import('@/views/reconciliation/PreCheckView.vue'),
        meta: { title: '核销预检', roles: ['accountant_ar', 'admin'] },
      },
      {
        path: 'reconciliation/match',
        name: 'ReconciliationMatch',
        component: () => import('@/views/reconciliation/MatchView.vue'),
        meta: { title: '匹配推荐', roles: ['accountant_ar', 'admin'] },
      },
      {
        path: 'reconciliation/manual',
        name: 'ManualMatch',
        component: () => import('@/views/reconciliation/ManualMatch.vue'),
        meta: { title: '手工核销', roles: ['accountant_ar', 'admin'] },
      },
      // ==== F5 凭证 ====
      {
        path: 'vouchers',
        name: 'VoucherList',
        component: () => import('@/views/voucher/VoucherList.vue'),
        meta: { title: '凭证列表', roles: ['admin', 'agent'], keepAlive: true },
      },
      {
        path: 'vouchers/create',
        name: 'VoucherCreate',
        component: () => import('@/views/voucher/VoucherEdit.vue'),
        meta: { title: '新增凭证', roles: ['admin', 'agent'] },
      },
      {
        path: 'vouchers/review',
        name: 'VoucherReview',
        component: () => import('@/views/voucher/ReviewWorkbench.vue'),
        meta: { title: '审核工作台', roles: ['admin'] },
      },
      // ==== F4 银企对账 ====
      {
        path: 'bank-reconciliation/match',
        name: 'BankReconciliation',
        component: () => import('@/views/reconciliation-bank/MatchingView.vue'),
        meta: { title: '银企对账', roles: ['cashier', 'admin'] },
      },
      {
        path: 'bank-reconciliation/balance',
        name: 'BalanceSheet',
        component: () => import('@/views/reconciliation-bank/BalanceSheet.vue'),
        meta: { title: '余额调节表', roles: ['cashier', 'admin'] },
      },
      // ==== F6 期末处理 ====
      {
        path: 'period/health-check',
        name: 'PeriodHealthCheck',
        component: () => import('@/views/period/HealthCheck.vue'),
        meta: { title: '结账体检', roles: ['admin'] },
      },
      {
        path: 'period/reports',
        name: 'FinancialReports',
        component: () => import('@/views/period/FinancialReports.vue'),
        meta: { title: '财务报表', roles: ['admin', 'boss'] },
      },
      // 后续模块路由在此嵌套
    ],
  },
]

export default routes
