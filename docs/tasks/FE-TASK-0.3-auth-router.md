# FE-TASK-0.3 | 认证流程与路由框架

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：10-14h
**前置**：FE-TASK-0.2
**状态**：待开发

---

## 任务描述

实现登录页、JWT 认证流程、路由框架和导航守卫。

### 具体步骤

1. **认证 Store**（`src/stores/auth.store.ts`）
   - `token: string | null` — 从 localStorage 恢复
   - `user: User | null` — 用户信息
   - `permissions: string[]` — 权限列表
   - `login(account, password)` — 调用 API → 存 token → 存 user
   - `logout()` — 清除 token → 跳转登录页
   - `fetchMe()` — 从 `/auth/me` 加载用户信息并刷新权限
   - `isLoggedIn` — computed，根据 token 是否存在

2. **多租户 Store**（`src/stores/tenant.store.ts`）
   - `currentTenantId: string | null`
   - `currentCompany: Company | null`
   - `tenantList: Tenant[]` — 代账会计专用
   - `switchTenant(id)` — 调用 refreshToken → 重置业务 Store
   - `watermark` — computed，返回水印文本

3. **登录页**（`src/views/login/LoginView.vue`）
   - 账号 + 密码输入框
   - 登录按钮 → 调用 `authStore.login()`
   - 登录成功 → 跳转 redirect 或 dashboard
   - 登录失败 → 显示错误信息
   - 记住我（7 天 token 保存）

4. **路由定义**（`src/router/routes/`）
   - `base.ts` — `login`, `dashboard`, `403`, `404`
   - 其余路由使用 `()` 动态导入，按模块文件留空供后续阶段填充

5. **路由守卫**（`src/router/guards.ts`）
   - `AuthGuard` — 无 token 跳 `/login`
   - `TenantGuard` — 代账会计未选择租户时跳租户选择页
   - `PermissionGuard` — 角色不匹配跳 `/403`
   - 登录后 `fetchMe()` 加载权限（首次或 token 变化时）

6. **全局错误页面**
   - `403.vue` — 「抱歉，您没有权限访问此页面」
   - `404.vue` — 「页面不存在」

7. **Dashboard 首页占位**（`src/views/dashboard/DashboardView.vue`）
   - 当前租户名 + 欢迎语
   - 快速入口卡片（待后续填充）

---

## 验收标准

- [ ] 输入错误账号密码登录，显示「账号或密码错误」
- [ ] 登录成功后跳转到 dashboard，侧边栏按角色展示菜单
- [ ] 手动清除 localStorage 后刷新页面，跳转到登录页
- [ ] token 过期后 API 返回 401，自动跳登录页
- [ ] 访问无权限路由（如出纳访问审核页），显示 403 页面
- [ ] Dashboard 显示当前公司名称和欢迎信息

---

## 参考

- 架构文档：第 4 章路由设计、第 5 章状态管理
- API 契约：`api-contracts/v1/auth-health-tenant.md`
