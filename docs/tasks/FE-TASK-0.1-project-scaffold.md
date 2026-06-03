# FE-TASK-0.1 | 项目脚手架搭建

**版本**：V1.0
**优先级**：P0（基础支撑）
**工时**：8-12h
**状态**：待开发

---

## 任务描述

使用 Vite 5 初始化 Vue 3 + TypeScript 项目，配置完整的工程化工具链。

### 具体步骤

1. **项目初始化**
   - `pnpm create vite huihua-finance-web --template vue-ts`
   - 安装核心依赖：`vue@3.4+`, `vue-router@4`, `pinia`, `element-plus`, `axios`, `echarts@5`, `sass`
   - 安装开发依赖：`eslint`, `prettier`, `husky`, `lint-staged`, `@types/node`, `unplugin-auto-import`, `unplugin-vue-components`

2. **目录结构创建**（按架构文档第 2 章，含所有空目录）

3. **ESLint + Prettier 配置**
   - `.eslintrc.cjs` — 继承 `@vue/typescript/recommended`
   - `.prettierrc` — 单引号、分号、trailing comma、printWidth 100
   - `commitlint.config.js` — conventional commits

4. **Vite 配置**
   - `@` 路径别名 → `src/`
   - 按环境加载 `.env.development` / `.env.production`
   - Element Plus 自动导入（`unplugin-vue-components` + `unplugin-auto-import`）
   - 代理 `/api` → `http://localhost:8080`

5. **全局样式**
   - `src/styles/variables.scss` — 覆盖 Element Plus 主题色（财务蓝：#1890ff）
   - `src/styles/reset.scss`
   - `src/styles/transitions.scss`

6. **环境变量**
   - `.env.development` — `VITE_API_BASE_URL=/api/v1`
   - `.env.production` — `VITE_API_BASE_URL=/api/v1`

7. **Husky + lint-staged**
   - `husky init` → pre-commit hook 运行 lint-staged
   - 提交前自动格式化 `.ts/.vue/.scss` 文件

---

## 验收标准

- [ ] `pnpm dev` 启动成功，浏览器访问显示空白页（后续挂载路由）
- [ ] `pnpm build` 构建成功，输出 `dist/` 目录
- [ ] `pnpm lint` 通过，无 ESLint 错误
- [ ] `@/` 路径别名生效（`import X from '@/components/X'` 可解析）
- [ ] Element Plus 组件无需手动 import（`<el-button>` 自动注册）
- [ ] `/api` 代理到 `localhost:8080`
- [ ] Git commit 时自动触发 lint-staged 格式化

---

## 参考

- 前端架构文档：`frontend-architecture-v1.0.md` 第 2 章（目录结构）& 第 13 章（选型理由）
- Vite 模板：https://github.com/vitejs/vite/tree/main/packages/create-vite
- Element Plus 按需导入：https://element-plus.org/zh-CN/guide/quickstart.html#%E6%8C%89%E9%9C%80%E5%AF%BC%E5%85%A5
