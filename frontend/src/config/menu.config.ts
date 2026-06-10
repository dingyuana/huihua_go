import type { Role } from '@/types/enums'
import type { Component } from 'vue'

export interface MenuItem {
  path: string
  title: string
  icon?: string | Component
  children?: MenuItem[]
  permissions?: string[]
}

/**
 * 按角色的菜单配置
 */
export const roleMenuMap: Record<Role, MenuItem[]> = {
  cashier: [
    { path: '/bank/import', title: '流水导入', icon: 'Upload' },
    { path: '/bank/workbench', title: '核对工作台', icon: 'List' },
    { path: '/payments', title: '收付款单', icon: 'Wallet' },
    { path: '/bank-reconciliation/match', title: '银企对账', icon: 'BalanceTwo' },
    { path: '/bank-reconciliation/diff-report', title: '对账差异报告', icon: 'DataAnalysis' },
  ],
  accountant_ar: [
    { path: '/invoices', title: '发票管理', icon: 'Document' },
    {
      path: '/reconciliation', title: '核销中心', icon: 'Link',
      children: [
        { path: '/reconciliation/precheck', title: '预检' },
        { path: '/reconciliation/match', title: '匹配推荐' },
        { path: '/reconciliation/manual', title: '手工核销' },
      ],
    },
  ],
  admin: [
    {
      path: '/setup', title: '基础设置', icon: 'Setting',
      children: [
        { path: '/setup/company', title: '公司信息' },
        { path: '/setup/accounts', title: '科目表' },
        { path: '/setup/bank-accounts', title: '资金账户' },
        { path: '/setup/parties', title: '客商档案' },
        { path: '/setup/rules', title: '规则库' },
        { path: '/setup/voucher-templates', title: '凭证模板' },
      ],
    },
    {
      path: '/bank', title: '银行流水', icon: 'Money',
      children: [
        { path: '/bank/import', title: '流水导入' },
        { path: '/bank/workbench', title: '核对工作台' },
        { path: '/payments', title: '收付款单' },
      ],
    },
    {
      path: '/receivables-payables', title: '往来管理', icon: 'Coin',
      children: [
        { path: '/invoices', title: '发票管理', icon: 'Document' },
        { path: '/expense-invoices/list', title: '进项发票', icon: 'Document' },
        { path: '/receivables-payables', title: '应收应付汇总' },
        { path: '/ar-invoices', title: '应收款单' },
        { path: '/ap-invoices', title: '应付款单' },
        { path: '/advance-receipts', title: '预收款单' },
        { path: '/advance-payments', title: '预付款单' },
        { path: '/expense/reimbursement', title: '费用报销单' },
      ],
    },
    {
      path: '/payroll', title: '工资管理', icon: 'Money',
      children: [
        { path: '/payroll', title: '工资单列表' },
        { path: '/payroll/new', title: '新增工资单' },
      ],
    },
    { path: '/reports/aging', title: '账龄分析', icon: 'Histogram' },
    { path: '/reports/credit', title: '信用管控', icon: 'CreditCardFilled' },
    {
      path: '/reconciliation', title: '核销中心', icon: 'Link',
      children: [
        { path: '/reconciliation/precheck', title: '预检' },
        { path: '/reconciliation/match', title: '匹配推荐' },
        { path: '/reconciliation/manual', title: '手工核销' },
      ],
    },
    {
      path: '/vouchers', title: '凭证管理', icon: 'Notebook',
      children: [
        { path: '/vouchers', title: '凭证列表' },
        { path: '/vouchers/create', title: '新增凭证' },
        { path: '/vouchers/review', title: '审核工作台' },
        { path: '/vouchers/batch-generate', title: '批量生成' },
      ],
    },
      {
        path: '/period', title: '期末处理', icon: 'Timer',
        children: [
          { path: '/period/health-check', title: '结账体检' },
          { path: '/period/voucher-gaps', title: '断号检测' },
          { path: '/period/depreciation', title: '折旧处理' },
          { path: '/period/reports', title: '财务报表' },
        ],
      },
    { path: '/bank-reconciliation/match', title: '银企对账', icon: 'BalanceTwo' },
    { path: '/bank-reconciliation/diff-report', title: '对账差异报告', icon: 'DataAnalysis' },
  ],
  boss: [
    { path: '/analytics', title: '经营分析', icon: 'TrendCharts' },
    { path: '/period/reports', title: '财务报表', icon: 'DataAnalysis' },
  ],
  employee: [
    { path: '/expense/reimbursement', title: '我的报销', icon: 'Money' },
  ],
  agent: [], // 继承 admin 菜单，运行时动态注入
}
