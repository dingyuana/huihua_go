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
      {
        path: 'setup/voucher-templates',
        name: 'VoucherTemplateList',
        component: () => import('@/views/setup/VoucherTemplateList.vue'),
        meta: { title: '凭证模板', roles: ['admin'] },
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
        path: 'payments',
        name: 'PaymentList',
        component: () => import('@/views/payments/PaymentList.vue'),
        meta: { title: '收付款单', roles: ['cashier', 'admin', 'agent'] },
      },
      {
        path: 'invoices',
        name: 'InvoiceList',
        component: () => import('@/views/invoices/InvoiceList.vue'),
        meta: { title: '发票管理', roles: ['accountant_ar', 'admin', 'agent'], keepAlive: true },
      },
      {
        path: 'receivables-payables',
        name: 'ReceivablesPayables',
        component: () => import('@/views/receivables-payables/UnifiedView.vue'),
        meta: { title: '应收应付汇总', roles: ['admin', 'agent'] },
      },
      {
        path: 'ar-invoices',
        name: 'ArInvoiceList',
        component: () => import('@/views/ar-invoices/ArInvoiceList.vue'),
        meta: { title: '应收款单', roles: ['admin', 'agent'] },
      },
      {
        path: 'ap-invoices',
        name: 'ApInvoiceList',
        component: () => import('@/views/ap-invoices/ApInvoiceList.vue'),
        meta: { title: '应付款单', roles: ['admin', 'agent'] },
      },
      {
        path: 'advance-receipts',
        name: 'AdvanceReceiptList',
        component: () => import('@/views/advance-receipts/AdvanceReceiptList.vue'),
        meta: { title: '预收款单', roles: ['admin', 'agent'] },
      },
      {
        path: 'advance-payments',
        name: 'AdvancePaymentList',
        component: () => import('@/views/advance-payments/AdvancePaymentList.vue'),
        meta: { title: '预付款单', roles: ['admin', 'agent'] },
      },
      {
        path: 'reports/aging',
        name: 'AgingAnalysis',
        component: () => import('@/views/reports/AgingAnalysis.vue'),
        meta: { title: '账龄分析', roles: ['admin', 'agent'] },
      },
      {
        path: 'reports/credit',
        name: 'CreditControl',
        component: () => import('@/views/reports/CreditControl.vue'),
        meta: { title: '信用管控', roles: ['admin', 'agent'] },
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
      {
        path: 'reconciliation/review',
        name: 'ReconciliationReview',
        component: () => import('@/views/reconciliation/ReviewView.vue'),
        meta: { title: '核销审批', roles: ['accountant_ar', 'admin'] },
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
        path: 'vouchers/:id',
        name: 'VoucherDetail',
        component: () => import('@/views/voucher/VoucherEdit.vue'),
        meta: { title: '凭证详情', roles: ['admin', 'agent'] },
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
      {
        path: 'bank-reconciliation/pending-confirm',
        name: 'BankReconciliationPendingConfirm',
        component: () => import('@/views/reconciliation-bank/PendingConfirmView.vue'),
        meta: { title: '银企对账待确认', roles: ['cashier', 'admin'] },
      },
      {
        path: 'bank-reconciliation/diff-report',
        name: 'DiffReport',
        component: () => import('@/views/reconciliation-bank/DiffReport.vue'),
        meta: { title: '对账差异报告', roles: ['cashier', 'admin'] },
      },
      // ==== F6 期末处理 ====
      {
        path: 'period/health-check',
        name: 'PeriodHealthCheck',
        component: () => import('@/views/period/HealthCheck.vue'),
        meta: { title: '结账体检', roles: ['admin'] },
      },
      {
        path: 'period/voucher-gaps',
        name: 'VoucherGaps',
        component: () => import('@/views/period/VoucherGapView.vue'),
        meta: { title: '断号检测', roles: ['admin'] },
      },
      {
        path: 'period/reports',
        name: 'FinancialReports',
        component: () => import('@/views/period/FinancialReports.vue'),
        meta: { title: '财务报表', roles: ['admin', 'boss'] },
      },
      // ==== 资产折旧 ====
      {
        path: 'period/depreciation',
        name: 'Depreciation',
        redirect: '/depreciation/list',
        meta: { title: '资产折旧', roles: ['admin', 'agent'] },
      },
      {
        path: 'depreciation/list',
        name: 'DepreciationList',
        component: () => import('@/views/depreciation/DepreciationList.vue'),
        meta: { title: '折旧记录', roles: ['admin', 'agent'] },
      },
      {
        path: 'depreciation/run',
        name: 'DepreciationRun',
        component: () => import('@/views/depreciation/DepreciationRun.vue'),
        meta: { title: '执行折旧', roles: ['admin', 'agent'] },
      },
      {
        path: 'asset/list',
        name: 'AssetList',
        component: () => import('@/views/asset/AssetList.vue'),
        meta: { title: '固定资产', roles: ['admin', 'agent'] },
      },
      {
        path: 'asset/new',
        name: 'AssetNew',
        component: () => import('@/views/asset/AssetDetail.vue'),
        meta: { title: '新增资产', roles: ['admin', 'agent'] },
      },
      {
        path: 'asset/:id',
        name: 'AssetDetail',
        component: () => import('@/views/asset/AssetDetail.vue'),
        meta: { title: '资产详情', roles: ['admin', 'agent'] },
      },
      // ==== 工资单 ====
      {
        path: 'payroll',
        name: 'PayrollList',
        component: () => import('@/views/payroll/PayrollList.vue'),
        meta: { title: '工资单', roles: ['admin', 'agent'] },
      },
      {
        path: 'payroll/new',
        name: 'PayrollNew',
        component: () => import('@/views/payroll/PayrollForm.vue'),
        meta: { title: '新建工资单', roles: ['admin', 'agent'] },
      },
      {
        path: 'payroll/:id',
        name: 'PayrollDetail',
        component: () => import('@/views/payroll/PayrollDetail.vue'),
        meta: { title: '工资单详情', roles: ['admin', 'agent'] },
      },
      {
        path: 'payroll/:id/edit',
        name: 'PayrollEdit',
        component: () => import('@/views/payroll/PayrollForm.vue'),
        meta: { title: '编辑工资单', roles: ['admin', 'agent'] },
      },
      // ==== 报销单 ====
      {
        path: 'expense/reimbursement',
        name: 'ReimbursementList',
        component: () => import('@/views/reimbursement/ReimbursementList.vue'),
        meta: { title: '报销单', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense/reimbursement/new',
        name: 'ReimbursementNew',
        component: () => import('@/views/reimbursement/ReimbursementForm.vue'),
        meta: { title: '新建报销单', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense/reimbursement/:id',
        name: 'ReimbursementDetail',
        component: () => import('@/views/reimbursement/ReimbursementDetail.vue'),
        meta: { title: '报销单详情', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense/reimbursement/:id/edit',
        name: 'ReimbursementEdit',
        component: () => import('@/views/reimbursement/ReimbursementForm.vue'),
        meta: { title: '编辑报销单', roles: ['admin', 'agent'] },
      },
      // ==== 期初余额 ====
      {
        path: 'opening-balance',
        name: 'OpeningBalance',
        component: () => import('@/views/opening-balance/OpeningBalanceList.vue'),
        meta: { title: '期初余额', roles: ['admin', 'agent'] },
      },
      {
        path: 'opening-balance/import',
        name: 'OpeningBalanceImport',
        component: () => import('@/views/opening-balance/OpeningBalanceImport.vue'),
        meta: { title: '期初余额导入', roles: ['admin', 'agent'] },
      },
      // ==== 进项发票 ====
      {
        path: 'expense-invoices',
        redirect: '/expense-invoices/list',
        meta: { title: '进项发票', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense-invoices/list',
        name: 'ExpenseInvoiceList',
        component: () => import('@/views/expense-invoices/ExpenseInvoiceList.vue'),
        meta: { title: '进项发票', roles: ['admin', 'agent'], keepAlive: true },
      },
      {
        path: 'expense-invoices/create',
        name: 'ExpenseInvoiceCreate',
        component: () => import('@/views/expense-invoices/ExpenseInvoiceForm.vue'),
        meta: { title: '新增进项发票', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense-invoices/edit/:id',
        name: 'ExpenseInvoiceEdit',
        component: () => import('@/views/expense-invoices/ExpenseInvoiceForm.vue'),
        meta: { title: '编辑进项发票', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense-invoices/detail/:id',
        name: 'ExpenseInvoiceDetail',
        component: () => import('@/views/expense-invoices/ExpenseInvoiceDetail.vue'),
        meta: { title: '进项发票详情', roles: ['admin', 'agent'] },
      },
      {
        path: 'expense-invoices/import',
        name: 'ExpenseInvoiceImport',
        component: () => import('@/views/expense-invoices/ExpenseInvoiceImport.vue'),
        meta: { title: '进项发票导入', roles: ['admin', 'agent'] },
      },
      // 后续模块路由在此嵌套
    ],
  },
]

export default routes
