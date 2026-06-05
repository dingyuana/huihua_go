// 全局枚举

/** 单据状态：0=草稿, 1=已提交, 2=已审核, 3=已作废 */
export enum DocStatus {
  Draft = 0,
  Submitted = 1,
  Verified = 2,
  Cancelled = 3,
}

export const DocStatusLabel: Record<DocStatus, string> = {
  [DocStatus.Draft]: '草稿',
  [DocStatus.Submitted]: '已提交',
  [DocStatus.Verified]: '已审核',
  [DocStatus.Cancelled]: '已作废',
}

/** 科目类型 */
export enum AccountType {
  Asset = 'asset',
  Liability = 'liability',
  Expense = 'expense',
  Income = 'income',
  Equity = 'equity',
}

export const AccountTypeLabel: Record<AccountType, string> = {
  [AccountType.Asset]: '资产',
  [AccountType.Liability]: '负债',
  [AccountType.Expense]: '费用',
  [AccountType.Income]: '收入',
  [AccountType.Equity]: '权益',
}

/** 余额方向 */
export enum RootType {
  Debit = 'debit',
  Credit = 'credit',
}

export const RootTypeLabel: Record<RootType, string> = {
  [RootType.Debit]: '借方',
  [RootType.Credit]: '贷方',
}

/** 客商类型 */
export enum PartyType {
  Customer = 'customer',
  Supplier = 'supplier',
  Employee = 'employee',
  Both = 'both',
}

export const PartyTypeLabel: Record<PartyType, string> = {
  [PartyType.Customer]: '客户',
  [PartyType.Supplier]: '供应商',
  [PartyType.Employee]: '员工',
  [PartyType.Both]: '客户/供应商',
}

/** 收付款类型 */
export enum PaymentType {
  Receive = 'receive',
  Pay = 'pay',
}

export const PaymentTypeLabel: Record<PaymentType, string> = {
  [PaymentType.Receive]: '收款',
  [PaymentType.Pay]: '付款',
}

/** 发票状态 */
export enum InvoiceStatus {
  Unpaid = 'unpaid',
  PartiallyPaid = 'partially_paid',
  Paid = 'paid',
  CreditNote = 'credit_note',
  WrittenOff = 'written_off',
}

export const InvoiceStatusLabel: Record<InvoiceStatus, string> = {
  [InvoiceStatus.Unpaid]: '待核销',
  [InvoiceStatus.PartiallyPaid]: '部分核销',
  [InvoiceStatus.Paid]: '已核销',
  [InvoiceStatus.CreditNote]: '红字发票',
  [InvoiceStatus.WrittenOff]: '已坏账',
}

/** 银行流水方向 */
export enum BankTxnDirection {
  In = 'in',
  Out = 'out',
}

export const BankTxnDirectionLabel: Record<BankTxnDirection, string> = {
  [BankTxnDirection.In]: '收款',
  [BankTxnDirection.Out]: '付款',
}

/** 流水分类 */
export enum BankTxnClassification {
  BusinessReceipt = 'business_receipt',
  BusinessPayment = 'business_payment',
  BankFee = 'bank_fee',
  InterestIncome = 'interest_income',
  InternalTransfer = 'internal_transfer',
  Pending = 'pending',
}

export const BankTxnClassificationLabel: Record<BankTxnClassification, string> = {
  [BankTxnClassification.BusinessReceipt]: '业务收款',
  [BankTxnClassification.BusinessPayment]: '业务付款',
  [BankTxnClassification.BankFee]: '银行费用',
  [BankTxnClassification.InterestIncome]: '利息收入',
  [BankTxnClassification.InternalTransfer]: '内部转账',
  [BankTxnClassification.Pending]: '待处理',
}

/** 导入格式 */
export enum ImportFormat {
  CSV = 'csv',
  Excel = 'excel',
  Camt053 = 'camt053',
  Mt940 = 'mt940',
}

/** 发票类型 */
export enum InvoiceType {
  Sale = 'sale',
  Purchase = 'purchase',
  CreditNote = 'credit_note',
}

export const InvoiceTypeLabel: Record<InvoiceType, string> = {
  [InvoiceType.Sale]: '销项发票',
  [InvoiceType.Purchase]: '进项发票',
  [InvoiceType.CreditNote]: '红字发票',
}

/** 凭证类型 */
export enum VoucherType {
  General = '记',
  Bank = '银',
  Cash = '现',
  Transfer = '转',
  Depreciation = '折旧',
  CarryOver = '结转',
}

/** 用户角色 */
export enum Role {
  Cashier = 'cashier',
  AccountantAR = 'accountant_ar',
  Admin = 'admin',
  Boss = 'boss',
  Employee = 'employee',
  Agent = 'agent',
}

export const RoleLabel: Record<Role, string> = {
  [Role.Cashier]: '出纳',
  [Role.AccountantAR]: '应收/应付会计',
  [Role.Admin]: '财务主管',
  [Role.Boss]: '老板/经理',
  [Role.Employee]: '普通员工',
  [Role.Agent]: '代账会计',
}

/** 布局模式 */
export type LayoutType = 'default' | 'collapsed' | 'fullscreen' | 'blank'
