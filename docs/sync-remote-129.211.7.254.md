# 远端 129.211.7.254 同步步骤

本机所有改动都在 `feat/frontend` 分支。要让远端 `129.211.7.254:3002` 看到新功能，按以下步骤在**远端服务器**操作。

## 1. 拉取最新代码

```bash
cd /root/data/disk/huihua-finance   # 或你部署的实际路径
git fetch origin
git checkout feat/frontend
git pull origin feat/frontend
```

## 2. 重启后端 (Go API)

```bash
# 编译新二进制
go build -o api ./cmd/api

# 停掉旧进程
pkill -f "^./api$"  # 或你的 supervisor 命令

# 启动新进程
nohup ./api > api.log 2>&1 &
```

## 3. 重启前端 (Vite dev)

```bash
cd frontend
# 如果依赖没变，无需重装
# pnpm install     # 仅当 package.json 有更新

# 停掉旧进程
pkill -f "vite"  # 或 pkill -f "node.*vite"

# 启动新进程
nohup pnpm dev > frontend.log 2>&1 &
```

## 4. 跑历史数据修复脚本（一次性，可选）

`feat/frontend` 上有 commit `7c39a82c` 带的清理脚本，**修复旧 10086 污染**：

```bash
PGPASSWORD=你的密码 psql -h 127.0.0.1 -U huihua -d huihua_finance \
  -f scripts/cleanup_10086_counterparty.sql
```

修复前：18 条 bank_transactions 的 counterparty='10086'
修复后：4 条 tax_payment → '国家税务总局平度市税务局'，14 条 → NULL

## 5. 验证

```bash
# 后端健康
curl -s http://localhost:8080/health    # → {"status":"ok"}

# 前端可访问
curl -s -I http://localhost:3002/        # → 200 OK

# 凭证列表
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)
curl -s "http://localhost:8080/api/v1/vouchers?limit=3" \
  -H "Authorization: Bearer $TOKEN" | jq '.vouchers[0]'
# 应能看到 counterparty_name 字段
```

## 6. 关键 commit 列表

```
e546e001 fix: /accounts/tree 返回真正的嵌套树结构（修 /vouchers/create 科目下拉空 bug）
96746a47 feat: 导入银行流水后默认直接生成草稿凭证
4255fd70 fix: 银行流水对方解析按方向感知选列
7c39a82c fix: 修复 import 对方解析错把银行号当名称；清理历史 10086 污染数据
1961c5b2 feat: 银行流水导入时自动创建对方档案
acbb12bf feat: 凭证列表显示对方名称
ab3b7c1d feat: 导入流水后自动生成凭证改为 opt-in，并仅处理本次批次
```

## 7. 注意事项

- **远端 `129.211.7.254` 不可达**（ping 100% 丢包，端口 3002 connection refused）。先在那台机器上检查防火墙/安全组/服务是否启动。
- **生产环境不要用 Vite dev server** (`:3002`)，代码泄漏 + 性能差。建议 `pnpm build` 出 `dist/` 用 nginx 托管。
- **远端如果曾经打开过 `VITE_ENABLE_MOCK=true`**，即使本地删了 mock 文件，远端工作副本里还有。检查 `frontend/.env*`。
