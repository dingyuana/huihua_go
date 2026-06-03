# FE-TASK-0.2 | API 请求层与类型定义

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：6-10h
**前置**：FE-TASK-0.1
**状态**：待开发

---

## 任务描述

封装 axios 实例、定义通用类型、实现按模块组织的 API 函数。

### 具体步骤

1. **通用类型定义**（`src/types/`）
   - `api.ts` — `ApiResponse<T>`, `PageResult<T>`, `PageQuery`
   - `enums.ts` — `DocStatus`, `AccountType`, `RootType`, `PartyType`, `PaymentType`, `InvoiceStatus`, `BankTxnDirection`, `BankTxnClassification`, `VoucherType`, `ImportFormat`, `ReconciliationMatchLevel`
   - `router.ts` — `RouteMeta`（含 `title`, `layout`, `roles`, `permissions`, `keepAlive`）

2. **领域模型类型**（`src/types/models/`）
   - `account.ts` — `Account`, `AccountTree`
   - `journal.ts` — `JournalEntry`, `JournalEntryLine`
   - `bank.ts` — `BankAccount`, `BankTransaction`, `ImportResult`
   - `invoice.ts` — `SalesInvoice`
   - `payment.ts` — `PaymentEntry`, `PaymentAllocation`
   - `party.ts` — `Party`
   - `tenant.ts` — `Tenant`, `Company`
   - `user.ts` — `User`, `Role`, `Permission`

3. **axios 实例**（`src/api/request.ts`）
   - 请求拦截器：注入 `Authorization: Bearer <token>`
   - 响应拦截器：401 跳登录、403 提示权限不足、统一错误 Message
   - 类型化 `request.get<T>()` 返回 `Promise<ApiResponse<T>>`

4. **模块 API**（`src/api/modules/`，每文件 4-8 个函数）
   - `account.ts` — `fetchAccountTree`, `fetchAccountList`, `createAccount`, `updateAccount`, `deleteAccount`
   - `bank.ts` — `importFile`, `classify`, `fetchTransactions`, `confirmTransaction`
   - `invoice.ts` — `uploadInvoice`, `fetchInvoices`, `fetchInvoiceDetail`
   - `payment.ts` — `fetchPayments`, `createPayment`
   - `voucher.ts` — `createVoucher`, `submitVoucher`, `cancelVoucher`, `reverseVoucher`
   - `reconciliation.ts` — `preCheck`, `match`, `manualAllocate`, `execute`, `reverse`
   - `period.ts` — `healthCheck`, `closePeriod`, `reopenPeriod`
   - `report.ts` — `fetchBalanceSheet`, `fetchProfitLoss`, `fetchCashFlow`
   - `tenant.ts` — `fetchTenants`, `switchTenant`
   - `auth.ts` — `login`, `logout`, `fetchMe`, `refreshToken`

---

## 验收标准

- [ ] `ApiResponse<T>` 类型能正确推导响应数据类型
- [ ] 401 时自动跳转 `/login` 并清除 token
- [ ] 403 时弹出 Element Plus Message 提示
- [ ] 所有 API 函数使用泛型，调用时 `const res = await fetchAccountTree()` 有完整类型提示
- [ ] 金额按字符串 `"1234.56"` 传输，非数值

---

## 参考

- API 契约：`api-contracts/v1/` 各模块请求/响应体定义
- 架构文档：第 6 章 API 请求层
