export interface Tenant {
  id: string
  name: string
  company_id: string
  fiscal_year_start: string
  status: string
}

export interface Company {
  id: string
  name: string
  fiscal_year_start: string
  status: string
}
