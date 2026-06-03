# TASK-F0.1 | F0 | 技术选型与项目脚手架

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P0（基础支撑）
**状态**：待开发

---

## 任务描述

搭建银行流水驱动业财一体化平台的 Go + PostgreSQL 项目骨架，具体包括：

1. **项目结构**：遵循 Go 项目标准布局（cmd/api/internal/pkg/config/deploy）
2. **依赖管理**：Go Mod，依赖 GORM GEN、pgx、redis、viper、jwt 等
3. **配置管理**：Viper 加载 .env，支持多环境（dev/staging/prod）
4. **数据库**：PostgreSQL 15 连接配置，支持多租户 RLS
5. **Redis**：会话管理、规则缓存、任务队列连接
6. **JWT 认证**：Token 携带 tenant_id， middleware 注入 tenant context
7. **RLS 中间件**：请求入口处执行 `SET app.current_tenant = $tenant_id`
8. **CI/CD**：Dockerfile + docker-compose.yml，可 git clone 后一键启动

---

## 验收标准

- [ ] `docker-compose up -d` 后全部服务启动成功，后端响应 `/health` 返回 `{"status":"ok"}`
- [ ] RLS 验证：使用 Tenant A 的 token 查询，Tenant B 的数据返回空集（跨租户隔离有效）
- [ ] JWT token 解析后包含 `tenant_id` 字段，middleware 正确注入到 `context.Context`
- [ ] 所有 API Handler 通过 `app.TenantFromContext(ctx)` 获取当前租户 ID

---

## 前置依赖

无（项目起点）

---

## 预计工时

- 最小：40h
- 最大：80h

---

## 技术提示

### Go 项目结构建议

```
huihua-finance/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/         # viper 配置加载
│   ├── middleware/     # tenant、auth、log 中间件
│   ├── handler/        # HTTP handlers（按模块组织）
│   ├── service/       # 业务逻辑层
│   ├── repository/    # 数据访问层（GEN 生成）
│   └── model/         # domain models
├── pkg/
│   ├── database/      # postgres 连接 + RLS
│   ├── redis/         # redis 客户端
│   └── jwt/           # jwt 工具
├── migrations/        # SQL 迁移脚本
├── docker-compose.yml
└── Dockerfile
```

### PostgreSQL RLS 核心配置

```sql
-- 开启 RLS
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

-- 创建租户隔离策略
CREATE POLICY tenant_isolation ON accounts
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- 应用层 Go 设置租户
func (db *DB) SetTenant(tenantID uuid.UUID) error {
    _, err := db.Exec(fmt.Sprintf("SET app.current_tenant = '%s'", tenantID))
    return err
}
```

### JWT 中间件伪代码

```go
func TenantMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := c.Locals("user").(*jwt.Token)
        claims := token.claims.(jwt.MapClaims)
        tenantID := claims["tenant_id"].(string)
        
        // 在数据库连接上设置 tenant
        if err := db.SetTenant(tenantID); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "tenant switch failed")
        }
        return c.Next()
    }
}
```

### 参考资料

- PostgreSQL RLS：https://www.postgresql.org/docs/current/ddl-partitioning.html
- GORM GEN：https://gorm.io/gen/
- Go Fiber JWT Middleware：https://github.com/gofiber/jwt

---

## 上下文信息（架构师决策记录）

- **决策**：采用 Go 1.22+ 而非 FastAPI，原因是 Go 强类型适合财务系统高并发，且 PostgreSQL RLS 隔离需要应用层在每个连接上设置 tenant context
- **决策**：不使用 GORM 的 AutoMigrate，全部表通过手写 SQL Migration 管理，确保 RLS 策略和约束与表结构同步
- **风险**：Go 框架选择（Fiber / Gin / Echo）——建议 Fiber（高性能 + 与 Vue 前端同异步模型匹配）；如团队熟悉 Gin 也可接受
- **风险**：Docker Compose 本地开发需要同时启动 PostgreSQL + Redis，网络配置注意端口不冲突