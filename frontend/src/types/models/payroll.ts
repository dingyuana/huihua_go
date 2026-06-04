export interface Payroll {
  id: string
  tenant_id: string
  company_id: string
  employee_name: string
  department_name: string
  period_no: number
  gross_salary: string
  individual_tax: string
  social_security: string
  housing_fund: string
  other_deductions: string
  net_salary: string
  payment_date: string
  bank_account_no: string
  status: string
  doc_status: number
  voucher_id?: string
  source: string
  remark: string
  created_by: string
  created_at: string
  updated_at: string
}