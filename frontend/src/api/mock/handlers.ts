import { http, HttpResponse } from 'msw'
import { authHandlers } from './auth'
import { accountHandlers } from './accounts'
import { bankHandlers } from './bank-transactions'
import { invoiceHandlers } from './invoices'
import { voucherHandlers } from './vouchers'
import { periodHandlers } from './periods'

/** 所有 Mock handlers */
export const handlers = [
  ...authHandlers,
  ...accountHandlers,
  ...bankHandlers,
  ...invoiceHandlers,
  ...voucherHandlers,
  ...periodHandlers,
]
