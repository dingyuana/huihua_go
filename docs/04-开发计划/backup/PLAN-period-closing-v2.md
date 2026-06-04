# 期末结转模块 - 开发计划

> 项目：慧财智能财务平台
> 周期：2026-06-03 ~ 2026-06-17（2周）
> 版本：v2.0（人工终审强内控版）
> 状态：**执行中**

---

## 目录
1. 项目概述
2. 需求分析与优先级
3. 分阶段开发计划
4. 关键任务清单
5. 依赖与风险
6. 验收标准
7. 代码提交规范

---

## 1. 项目概述

### 1.1 目标
基于《期末结转模块 需求规格说明书 V1.2》，实现符合财政部会计软件规范的期末结转模块，核心特性：
- 人工发起结转，禁止自动触发
- 生成凭证草稿（DocStatus=0），不计入正式账务
- 人工逐张审核 + 人工确认过账，两步分离
- 已过账凭证不可直接修改，仅支持红字冲销
- 完整的操作日志留存，满足审计要求

### 1.2 范围界定

| 包含 | 不包含 |
|------|--------|
| 期间损益结转 | 成本结转（商业企业不适用） |
| 税金结转（增值税/附加税费） | 制造费用结转 |
| 自定义结转模板 | 存货成本核算 |
| 结转草稿审核/过账流程 | 生产成本归集 |
| 权限互斥校验 | 完工产品入库结转 |

### 1.3 技术栈
- Go 1.24 + Fiber 2.52
- PostgreSQL 15（RLS多租户）
- Vue 3 + Element Plus（前端）

---

## 2. 需求分析与优先级

### 2.1 需求分解

| 需求编号 | 需求描述 | 来源 | 优先级 |
|---------|---------|------|--------|
| REQ-001 | 所有自动生成的结转凭证默认草稿状态 | RC-001 | P0 |
| REQ-002 | 草稿凭证不计入正式账务，不更新科目余额 | RC-002 | P0 |
| REQ-003 | 禁止自动过账、自动审核、自动确认结转 | RC-003 | P0 |
| REQ-004 | 人工逐张审核、人工确认过账，两步分离 | RC-004 | P0 |
| REQ-005 | 已过账凭证不可直接删除/修改，仅允许红字冲销 | RC-005 | P0 |
| REQ-006 | 结转前置检查（凭证过账、银行对账、折旧计提等） | 5.1节 | P0 |
| REQ-007 | 期间损益结转功能 | 4.1节 | P0 |
| REQ-008 | 税金结转功能（增值税、附加税费） | 4.3节 | P1 |
| REQ-009 | 自定义结转模板管理 | 4.4节 | P2 |
| REQ-010 | 权限互斥校验（制单/审核/过账互斥） | 8.1节 | P0 |
| REQ-011 | 全流程操作留痕，日志留存≥30年 | 8.3节 | P0 |
| REQ-012 | 异常处理与事务原子性保障 | 第7章 | P0 |

### 2.2 功能模块划分

```
期末结转模块
├── 核心流程层
│   ├── 结转前置检查服务
│   ├── 结转草稿生成服务
│   ├── 草稿审核服务
│   └── 凭证过账服务
├── 数据层
│   ├── 结转草稿模型
│   ├── 结转模板模型
│   └── 操作日志模型
├── 权限层
│   ├── 权限互斥校验
│   └── 操作权限控制
└── 接口层
    ├── 结转相关API
    └── 模板管理API
```

---

## 3. 分阶段开发计划

### 3.1 整体时间安排

| 阶段 | 周期 | 天数 | 核心任务 |
|------|------|------|---------|
| Phase 1 | 06-03 ~ 06-05 | 3天 | 数据库设计与Migration |
| Phase 2 | 06-06 ~ 06-09 | 4天 | 核心Service层开发 |
| Phase 3 | 06-10 ~ 06-12 | 3天 | Handler层与API开发 |
| Phase 4 | 06-13 ~ 06-14 | 2天 | 前端页面开发 |
| Phase 5 | 06-15 ~ 06-17 | 3天 | 测试与集成 |

### 3.2 Phase 1：数据库设计与Migration（3天）

**目标**：完成数据库Schema设计和Migration脚本

| 日期 | 任务 | 责任人 | 交付物 |
|------|------|--------|--------|
| 06-03 | 创建结转草稿表（closing_drafts） | Backend | migration_023.sql |
| 06-03 | 创建结转草稿分录表（closing_draft_lines） | Backend | migration_023.sql |
| 06-04 | 创建结转模板表（closing_templates） | Backend | migration_023.sql |
| 06-04 | 创建结转模板行表（closing_template_lines） | Backend | migration_023.sql |
| 06-05 | 编写种子数据（预置期间损益结转模板） | Backend | seed_data.sql |
| 06-05 | 执行Migration验证 | Backend | DB schema验证通过 |

### 3.3 Phase 2：核心Service层开发（4天）

**目标**：完成核心业务逻辑开发

| 日期 | 任务 | 责任人 | 交付物 |
|------|------|--------|--------|
| 06-06 | 创建ClosingDraft模型 | Backend | `internal/model/closing_draft.go` |
| 06-06 | 创建ClosingTemplate模型 | Backend | `internal/model/closing_template.go` |
| 06-07 | 创建ClosingDraftRepository | Backend | `internal/repository/closing_draft_repo.go` |
| 06-07 | 创建ClosingTemplateRepository | Backend | `internal/repository/closing_template_repo.go` |
| 06-08 | 创建ClosingService（核心逻辑） | Backend | `internal/service/closing_service.go` |
| 06-08 | 实现结转草稿生成逻辑 | Backend | `GenerateClosingDraft()` |
| 06-09 | 实现草稿审核/过账逻辑 | Backend | `ReviewDraft()` / `PostDraft()` |
| 06-09 | 实现税金结转逻辑 | Backend | `GenerateTaxClosing()` |

### 3.4 Phase 3：Handler层与API开发（3天）

**目标**：完成RESTful API开发

| 日期 | 任务 | 责任人 | 交付物 |
|------|------|--------|--------|
| 06-10 | 创建ClosingHandler | Backend | `internal/handler/closing_handler.go` |
| 06-10 | 实现前置检查API | Backend | `GET /api/v1/periods/{periodNo}/pre-close-check` |
| 06-11 | 实现草稿CRUD API | Backend | 草稿列表/详情/创建/修改/删除 |
| 06-11 | 实现审核/过账API | Backend | `POST /review` / `POST /post` |
| 06-12 | 实现模板管理API | Backend | 模板CRUD接口 |
| 06-12 | 注册路由并测试 | Backend | 所有API连通测试通过 |

### 3.5 Phase 4：前端页面开发（2天）

**目标**：完成期末结转前端页面

| 日期 | 任务 | 责任人 | 交付物 |
|------|------|--------|--------|
| 06-13 | 创建结转工作台页面 | Frontend | `ClosingWorkbench.vue` |
| 06-13 | 创建草稿列表组件 | Frontend | `ClosingDraftList.vue` |
| 06-14 | 创建草稿详情/审核弹窗 | Frontend | `ClosingDraftDetail.vue` |
| 06-14 | 创建模板管理页面 | Frontend | `ClosingTemplate.vue` |

### 3.6 Phase 5：测试与集成（3天）

**目标**：完成测试验证和集成

| 日期 | 任务 | 责任人 | 交付物 |
|------|------|--------|--------|
| 06-15 | 单元测试编写与执行 | Backend | 覆盖率≥80% |
| 06-15 | 集成测试验证 | Backend | API接口测试通过 |
| 06-16 | 前端联调测试 | Fullstack | 端到端流程验证 |
| 06-16 | 权限互斥测试 | Fullstack | 权限隔离验证 |
| 06-17 | 验收标准验证 | Fullstack | 所有AC项通过 |
| 06-17 | 代码Review与提交 | Team | 代码合并至主分支 |

---

## 4. 关键任务清单

### 4.1 后端任务

| 任务ID | 任务描述 | 所属文件 | 状态 |
|--------|---------|---------|------|
| BE-001 | 创建closing_drafts表Migration | `migrations/023_closing.sql` | pending |
| BE-002 | 创建closing_draft_lines表Migration | `migrations/023_closing.sql` | pending |
| BE-003 | 创建closing_templates表Migration | `migrations/023_closing.sql` | pending |
| BE-004 | 创建closing_template_lines表Migration | `migrations/023_closing.sql` | pending |
| BE-005 | 创建ClosingDraft模型 | `internal/model/closing_draft.go` | pending |
| BE-006 | 创建ClosingTemplate模型 | `internal/model/closing_template.go` | pending |
| BE-007 | 创建ClosingDraftRepository | `internal/repository/closing_draft_repo.go` | pending |
| BE-008 | 创建ClosingTemplateRepository | `internal/repository/closing_template_repo.go` | pending |
| BE-009 | 创建ClosingService | `internal/service/closing_service.go` | pending |
| BE-010 | 实现GenerateClosingDraft方法 | `internal/service/closing_service.go` | pending |
| BE-011 | 实现ReviewDraft方法 | `internal/service/closing_service.go` | pending |
| BE-012 | 实现PostDraft方法 | `internal/service/closing_service.go` | pending |
| BE-013 | 实现GenerateTaxClosing方法 | `internal/service/closing_service.go` | pending |
| BE-014 | 创建ClosingHandler | `internal/handler/closing_handler.go` | pending |
| BE-015 | 注册路由 | `cmd/api/main.go` | pending |
| BE-016 | 编写单元测试 | `internal/service/closing_service_test.go` | pending |

### 4.2 前端任务

| 任务ID | 任务描述 | 所属文件 | 状态 |
|--------|---------|---------|------|
| FE-001 | 创建API模块 | `frontend/src/api/modules/closing.ts` | pending |
| FE-002 | 创建结转工作台页面 | `frontend/src/views/period/ClosingWorkbench.vue` | pending |
| FE-003 | 创建草稿列表组件 | `frontend/src/components/period/ClosingDraftList.vue` | pending |
| FE-004 | 创建草稿详情弹窗 | `frontend/src/components/period/ClosingDraftDetail.vue` | pending |
| FE-005 | 创建模板管理页面 | `frontend/src/views/period/ClosingTemplate.vue` | pending |
| FE-006 | 添加路由配置 | `frontend/src/router/index.ts` | pending |

---

## 5. 依赖与风险

### 5.1 依赖关系

| 依赖项 | 说明 | 状态 |
|--------|------|------|
| 科目数据 | 需要完整的会计科目数据（收入/费用/税金科目） | ✅ 已就绪 |
| 凭证状态机 | 需复用现有凭证状态机逻辑 | ✅ 已就绪 |
| 审核流程 | 需复用现有审批流程框架 | ✅ 已就绪 |
| 权限系统 | 需集成现有RBAC权限系统 | ✅ 已就绪 |
| GL过账服务 | 需调用现有GL过账逻辑 | ✅ 已就绪 |

### 5.2 风险识别

| 风险编号 | 风险描述 | 影响 | 缓解措施 |
|---------|---------|------|---------|
| R-001 | 结转模板配置复杂，用户难以理解 | 降低用户体验 | 预置标准模板，提供可视化配置 |
| R-002 | 科目余额计算错误导致结转错误 | 账务数据错误 | 增加校验逻辑，生成前核对余额 |
| R-003 | 并发操作导致数据不一致 | 数据错乱 | 事务锁 + 乐观锁机制 |
| R-004 | 期末结转数据量大，性能问题 | 系统响应慢 | 分批处理 + 异步任务 |
| R-005 | 权限互斥校验复杂 | 权限漏洞 | 底层固化权限规则 |

---

## 6. 验收标准

### 6.1 功能验收

| AC编号 | 验收项 | 验证方法 | 优先级 |
|--------|--------|---------|--------|
| AC-001 | 结转凭证默认草稿状态（DocStatus=0） | 检查数据库字段 | P0 |
| AC-002 | 草稿凭证不计入正式账务 | 检查科目余额未更新 | P0 |
| AC-003 | 无自动结转机制 | 检查无定时任务配置 | P0 |
| AC-004 | 审核与过账两步分离 | 验证审核后需手动过账 | P0 |
| AC-005 | 已过账凭证不可修改 | 验证修改接口返回错误 | P0 |
| AC-006 | 结转前置检查完整 | 验证7项检查全部通过 | P0 |
| AC-007 | 期间损益结转正确 | 验证收入/费用科目余额归零 | P0 |
| AC-008 | 税金结转正确 | 验证增值税科目结转至未交增值税 | P1 |
| AC-009 | 权限互斥校验 | 验证同一用户无法同时拥有制单+审核权限 | P0 |
| AC-010 | 操作日志完整 | 验证所有操作记录入库 | P0 |
| AC-011 | 事务原子性保障 | 验证部分失败时整体回滚 | P0 |
| AC-012 | 接口安全合规 | 验证Token失效后接口拒绝访问 | P0 |

### 6.2 技术验收

| 指标 | 要求 |
|------|------|
| 单元测试覆盖率 | ≥80% |
| API响应时间 | < 500ms |
| 代码规范 | 符合Go Code Review Comments |
| 编译 | `go build ./...` 通过 |
| 数据库Migration | 执行无错误 |

---

## 7. 代码提交规范

### 7.1 Commit Message格式

```
<类型>(<模块>): <描述>

<详细说明>

<关联任务ID>
```

**类型**：
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `refactor`: 代码重构
- `test`: 测试代码
- `chore`: 构建/工具更新

**示例**：
```
feat(closing): 实现结转草稿生成逻辑

- 创建ClosingService
- 实现GenerateClosingDraft方法
- 支持期间损益结转和税金结转

TASK-CLOSING-001
```

### 7.2 分支策略

| 分支 | 用途 |
|------|------|
| `main` | 主分支，稳定版本 |
| `dev` | 开发分支，集成所有功能 |
| `feature/closing-v2` | 期末结转v2功能分支 |

### 7.3 代码Review流程

1. 完成功能开发后，提交PR至`dev`分支
2. 至少1位开发者Review通过
3. 所有测试通过
4. 合并至`dev`分支

---

**计划版本**：v1.0  
**生成日期**：2026-06-03  
**项目版本**：慧财智能财务平台 v1.0.0