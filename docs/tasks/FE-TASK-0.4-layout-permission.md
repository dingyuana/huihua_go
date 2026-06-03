# FE-TASK-0.4 | 布局系统与权限框架

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：10-14h
**前置**：FE-TASK-0.3
**状态**：待开发

---

## 任务描述

实现主布局组件、侧边栏菜单、顶部导航、权限指令。

### 具体步骤

1. **应用 Store**（`src/stores/app.store.ts`）
   - `sidebarCollapsed: boolean`
   - `currentLayout: LayoutType` — `default | collapsed | fullscreen | blank`
   - `globalLoading: boolean`

2. **菜单配置**（`src/config/menu.config.ts`）
   - 按角色分组的数据结构：`Record<Role, MenuItem[]>`
   - 每项：`{ path, title, icon, children?, permissions? }`
   - 6 个角色的菜单树（出纳、应收/应付会计、主管、老板、员工、代账会计）

3. **布局组件**

   **AppLayout.vue** — 主布局容器
   - 根据 `meta.layout` 切换 4 种布局模式
   - 内含 `<router-view>` 并使用 `<keep-alive>` 缓存标记 `keepAlive` 的页面

   **AppSidebar.vue**
   - 使用 `el-menu`，根据当前用户角色过滤菜单
   - 支持折叠（`el-menu collapse`）
   - 递归渲染多级子菜单
   - 当前路由高亮（`:default-active`）

   **AppHeader.vue**
   - 左侧：折叠按钮 + 面包屑
   - 中间：当前租户名称（代账会计可切换）
   - 右侧：通知图标 + 用户头像 + 下拉菜单（个人信息/切换角色/退出）
   - 代账会计模式下显示水印底色

   **AppTabs.vue**（多标签页）
   - 打开新页面时自动添加标签
   - 标签可关闭、拖动排序
   - 刷新后恢复之前打开的标签（localStorage）

4. **权限指令**（`src/directives/`）
   - `permission.ts` — `v-permission="['voucher:submit']"`，无权限移除元素
   - `role.ts` — `v-role="'admin'"`，角色不匹配移除元素
   - `number.ts` — `v-number` 金额输入自动格式化

5. **权限组合式函数**（`src/hooks/usePermission.ts`）
   - `hasPermission(perm: string): boolean`
   - `hasRole(role: Role): boolean`
   - `hasAnyRole(roles: Role[]): boolean`

---

## 验收标准

- [ ] 登录后显示左侧菜单 + 顶部 Header + 内容区布局
- [ ] 出纳登录只看到「流水导入」「核对工作台」「银企对账」
- [ ] 主管登录看到「科目表」「凭证管理」「审核工作台」「结账」「财务报表」
- [ ] 老板登录只看到「经营分析」
- [ ] 代账会计看到完整菜单但带水印
- [ ] 侧边栏折叠/展开动画流畅
- [ ] `v-permission="['voucher:submit']"` 对无权限用户隐藏按钮
- [ ] 多标签页切换正常，刷新后恢复

---

## 参考

- 架构文档：第 3 章布局系统、第 7 章权限系统
- Element Plus Menu：https://element-plus.org/zh-CN/component/menu.html
