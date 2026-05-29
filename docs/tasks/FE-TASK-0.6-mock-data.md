# FE-TASK-0.6 | Mock 数据层

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：6-10h
**前置**：FE-TASK-0.2
**状态**：待开发

---

## 任务描述

搭建 Mock 数据层，使前端可在后端未完成时独立开发和调试。

### 具体步骤

1. **安装 MSW**（Mock Service Worker）
   - `pnpm add msw@latest -D`
   - 初始化 Service Worker：`npx msw init public/`

2. **Mock 数据处理函数**（`src/api/mock/`）
   - `utils.ts` — 通用：分页封装、延迟模拟、UUID 生成
   - `accounts.ts` — 科目表：50 条内置《小企业会计准则》科目树数据
   - `bank-transactions.ts` — 银行流水：100 条模拟流水
   - `invoices.ts` — 发票：30 条模拟发票
   - `vouchers.ts` — 凭证：50 条模拟凭证（含分录）
   - `auth.ts` — 认证：3 个角色的模拟用户

3. **Mock 数据覆盖范围**
   - 认证：`POST /auth/login` → 返回 3 个角色的 token
   - 科目表：`GET /accounts/tree` → 完整科目树
   - 银行流水：`GET /bank-transactions` → 分页 + 筛选
   - 发票：`GET /invoices` → 分页 + 状态筛选
   - 凭证：`GET /vouchers` → 分页 + `POST /vouchers/submit` → 成功
   - 核销：`POST /reconciliation/precheck` → 返回预检结果
   - 结账：`POST /periods/health-check` → 返回 10 项检查
   - 报表：`GET /reports/balance-sheet` → 返回资产负债表

4. **Mock 启用控制**
   - 通过环境变量 `VITE_ENABLE_MOCK=true` 控制
   - 开发环境默认启用，生产环境禁用
   - 在 `main.ts` 中条件性启动 MSW

5. **数据保真要求**
   - 金额数据使用字符串 `"1234.56"` 格式
   - UUID 格式正确
   - 分页响应格式与 `PageResult` 一致
   - 枚举值与后端一致

---

## 验收标准

- [ ] `VITE_ENABLE_MOCK=true` 时，Mock Service Worker 启动成功
- [ ] 登录 mock 返回 3 个角色的 JWT，各自菜单正确
- [ ] 科目树 mock 返回完整的内置科目表（50+ 条）
- [ ] 银行流水 mock 支持 `?classification=pending&page=1` 筛选
- [ ] 所有 CRUD 操作的 mock 返回正确的响应格式
- [ ] 切换 `VITE_ENABLE_MOCK=false`，请求正常转发到后端

---

## 参考

- MSW 文档：https://mswjs.io/docs/
- API 契约：`api-contracts/v1/` 各文件定义
