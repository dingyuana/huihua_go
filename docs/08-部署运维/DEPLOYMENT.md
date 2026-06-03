# 慧财智能财务平台 — 部署运维指南

版本：V1.0
日期：2026-06-01

---

## 一、部署架构

```
                    ┌──────────────────────┐
                    │   Nginx (反向代理)    │
                    │  - 静态文件服务       │
                    │  - /api/* → Go API    │
                    │  - /* → Vue SPA       │
                    └──────────┬───────────┘
                               │
                ┌──────────────┴──────────────┐
                │                             │
       ┌────────▼────────┐           ┌─────────▼────────┐
       │  Go API (:8080) │           │  Vue 3 SPA       │
       │  - Fiber v2     │           │  (dist/ 静态文件)│
       │  - JWT 认证     │           │                   │
       │  - RLS 多租户   │           │                   │
       └────────┬────────┘           └──────────────────┘
                │
       ┌────────┴────────┐
       │                 │
┌──────▼──────┐  ┌──────▼──────┐
│ PostgreSQL  │  │   Redis    │
│   :5432     │  │   :6380    │
│  (含 RLS)   │  │ (缓存/会话)│
└─────────────┘  └─────────────┘
```

---

## 二、环境要求

### 2.1 硬件最低配置

| 环境 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| 开发 | 2 核 | 4 GB | 20 GB |
| 测试 | 4 核 | 8 GB | 50 GB |
| 生产 | 8 核+ | 16 GB+ | 200 GB+ |

### 2.2 软件依赖

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.24+ | 后端编译 |
| Node.js | 22+ | 前端构建 |
| pnpm | 9+ | 前端包管理 |
| PostgreSQL | 15+ | 主数据库 |
| Redis | 7+ | 缓存层 |
| Nginx | 1.24+ | 反向代理 |

---

## 三、环境变量

### 3.1 后端（.env）

```ini
# 服务监听
server.host=0.0.0.0
server.port=8080

# PostgreSQL
database.host=127.0.0.1
database.port=5432
database.user=huihua
database.password=<生产密码>
database.dbname=huihua_finance
database.sslmode=disable

# Redis
redis.host=localhost
redis.port=6380
redis.password=
redis.db=0

# JWT
jwt.secret=<生产随机 64 字符>
jwt.expiry=30m
```

### 3.2 前端（frontend/.env.production）

```ini
VITE_API_BASE=/api/v1
VITE_ENABLE_MOCK=false
```

---

## 四、部署步骤

### 4.1 数据库初始化

```bash
# 1. 启动 PostgreSQL 与 Redis
docker compose up -d postgres redis

# 2. 等待服务就绪
sleep 5
docker compose ps

# 3. 顺序执行迁移脚本（编号 001-028）
cd /opt/huihua-finance
for f in migrations/*.sql; do
    echo "Applying $f"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f "$f"
done
```

### 4.2 后端构建与启动

```bash
# 1. 编译二进制
cd /opt/huihua-finance
go build -o /opt/huihua-api ./cmd/api

# 2. 准备环境变量
cp .env.example .env
vim .env  # 修改生产密码、JWT secret

# 3. 启动服务（用 systemd 或 supervisor）
sudo systemctl enable huihua-api
sudo systemctl start huihua-api
sudo systemctl status huihua-api
```

**systemd 单元文件示例** (`/etc/systemd/system/huihua-api.service`):

```ini
[Unit]
Description=Huihua Finance API
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=huihua
WorkingDirectory=/opt/huihua-finance
EnvironmentFile=/opt/huihua-finance/.env
ExecStart=/opt/huihua-api
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### 4.3 前端构建

```bash
cd /opt/huihua-finance/frontend
pnpm install --frozen-lockfile
pnpm build  # 产物在 dist/

# 部署 dist/ 到 Nginx
sudo rsync -avz dist/ /var/www/huihua-finance/
```

### 4.4 Nginx 配置

```nginx
server {
    listen 80;
    server_name finance.example.com;

    # 静态文件
    root /var/www/huihua-finance;
    index index.html;

    # 前端 SPA 路由
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
        client_max_body_size 50M;  # 支持 Excel 导入
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080/health;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|svg|woff|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 4.5 启动验证

```bash
# 健康检查
curl -s http://localhost:8080/health
# 期望: {"status":"ok"}

# 通过 Nginx 访问
curl -sI http://localhost/
# 期望: 200 OK + text/html

# API 反向代理
curl -s http://localhost/api/v1/health
# 期望: {"status":"ok"}
```

---

## 五、备份与恢复

### 5.1 数据库每日备份

```bash
# /opt/huihua-finance/scripts/backup.sh
#!/bin/bash
BACKUP_DIR=/var/backups/huihua
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
    -Fc -f "$BACKUP_DIR/huihua_${TIMESTAMP}.dump"

# 保留最近 30 天
find $BACKUP_DIR -name "huihua_*.dump" -mtime +30 -delete
```

注册到 crontab：

```cron
0 2 * * * /opt/huihua-finance/scripts/backup.sh
```

### 5.2 恢复

```bash
PGPASSWORD=$DB_PASSWORD pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME \
    --clean --if-exists /var/backups/huihua/huihua_20260601_020000.dump
```

---

## 六、监控告警

### 6.1 关键指标

| 指标 | 阈值 | 告警渠道 |
|------|------|---------|
| HTTP 5xx 错误率 | > 1% / 5min | 飞书/钉钉 |
| API 平均响应时间 | > 500ms | 飞书/钉钉 |
| 数据库连接数 | > 80% max | 飞书/钉钉 |
| 磁盘使用率 | > 85% | 飞书/钉钉 |
| 后端进程存活 | 进程消失 | 飞书/钉钉 |
| 数据库主从延迟 | > 30s | 飞书/钉钉 |

### 6.2 日志位置

| 服务 | 路径 |
|------|------|
| 后端 API | `/var/log/huihua/api.log` |
| Nginx access | `/var/log/nginx/access.log` |
| Nginx error | `/var/log/nginx/error.log` |
| PostgreSQL | `/var/log/postgresql/` |

---

## 七、回滚预案

### 7.1 应用回滚

```bash
# 1. 停止当前服务
sudo systemctl stop huihua-api

# 2. 回滚到上一个版本
cd /opt/huihua-finance
git checkout HEAD~1
go build -o /opt/huihua-api ./cmd/api

# 3. 重启
sudo systemctl start huihua-api
```

### 7.2 数据库回滚

迁移脚本向下兼容设计，回滚只需回退应用代码，**不应**回退数据库 schema。如果必须回退，按反向顺序执行迁移：

```bash
# 假设要回退到 027 之前
# 注意: 028_bank_opening_balance.sql 新增的列和审计表需要保留，
# 因为 027 假设这些表/列已存在
```

---

## 八、常见问题

### Q1: 前端 404 在刷新时
**原因**：SPA 路由不匹配，Nginx 没回退到 index.html
**修复**：检查 `try_files $uri $uri/ /index.html;` 配置

### Q2: API 返回 401 Unauthorized
**原因**：JWT 过期或未携带 token
**修复**：检查前端 `request.ts` 的 interceptor 是否正确添加 `Authorization` 头

### Q3: 导入 Excel 返回 0 成功
**原因**：列名不匹配（中文/英文混排）
**修复**：参考 `bank_transaction_service.go` 的 `findCol()` 函数，确认 Excel 表头包含支持的列名

### Q4: 余额不一致
**原因**：opening_balance 录入有误，或 adjustment 漏操作
**修复**：调用 `GET /api/v1/bank-reconciliation/balance-check` 排查

---

## 九、联系与升级

- 监控仪表盘: <内部 Grafana 链接>
- 日志查询: <内部 Loki 链接>
- 值班群: <飞书群链接>
- 故障升级: <OnCall 升级流程>
