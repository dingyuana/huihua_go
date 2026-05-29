# TASK-F0.1 工程架构评审报告

**任务**：TASK-F0.1 技术选型与项目脚手架
**评审状态**：CLEAN — Ready to implement
**评审工具**：gstack /plan-eng-review
**日期**：2026-05-27

---

## 评审结论

| 维度 | 结果 |
|:---|:---|
| 范围 | ✅ 接受（最小可行脚手架） |
| 架构 | ✅ 2 个问题已标注（RLS 连接处理、错误传播） |
| 代码质量 | ✅ 1 个问题（DRY tenant 提取） |
| 测试 | ✅ 4 条路径，2 个关键测试已标注 |
| 性能 | ✅ 1 个建议（连接池大小配置） |

---

## 核心架构决策

### RLS 实现模式（关键）

```sql
-- 所有业务表含 tenant_id 列
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON accounts
    USING (tenant_id = current_setting('app.current_tenant')::uuid);
```

### 连接池策略

每个请求需先 `SET app.current_tenant`，推荐使用 `pgxpool`，连接预配置 RLS。

### 中间件链

```
Request → JWT Middleware → 提取 tenant_id → 
  → 从连接池获取连接 → SET app.current_tenant = $tenant_id → 
  → 查询执行（RLS 强制）→ 释放连接
```

### Docker 部署结构

- `docker-compose.yml`：PostgreSQL 15 + Redis + API
- 一键启动：`docker-compose up -d`
- 迁移脚本挂载：`./migrations:/docker-entrypoint-initdb.d`

---

## 步骤并行化

| 步骤 | 内容 | 依赖 |
|:---|:---|:---:|
| Step 1 | 项目骨架 + go.mod | — |
| Step 2 | PostgreSQL RLS + migrations | Step 1 |
| Step 3 | JWT + Tenant middleware | Step 2 |
| Step 4 | Health endpoint + Docker | Step 3 |
| Step 5 | 单元测试 | Step 4 |

Step 1-3 可并行开发，Step 4-5 串行。

---

## 两个必须完成的关键测试

1. **RLS 跨租户隔离验证**：Tenant A token 无法查询 Tenant B 数据（接受标准 #2）
2. **JWT tenant_id 提取**：token 解码后含 tenant_id，middleware 正确注入

---

**下一步**：进入 superpowers Phase 2，生成 `docs/plans/` 实现计划。