# API 契约 — 认证 & 基础 (F0)

---

## 认证

### POST /auth/login

用户登录，返回 JWT。

**Request Body:**
```json
{
  "account": "admin@example.com",
  "password": "encrypted_password"
}
```

**Response 200:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-06-26T12:00:00Z",
    "user": {
      "id": "uuid",
      "name": "张三",
      "email": "admin@example.com",
      "role": "admin",
      "permissions": ["account:read", "account:write", "voucher:read", "voucher:submit", "voucher:reverse"]
    },
    "tenant_id": "uuid",
    "tenant_name": "北京XX科技有限公司"
  }
}
```

**Error 401:**
```json
{
  "code": 10001,
  "message": "账号或密码错误"
}
```

### POST /auth/logout

登出，使当前 token 失效。

**Headers:** `Authorization: Bearer <token>`

**Response 200:**
```json
{ "code": 0, "message": "ok", "data": null }
```

### GET /auth/me

获取当前用户信息及权限。

**Headers:** `Authorization: Bearer <token>`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "name": "张三",
    "email": "admin@example.com",
    "role": "admin",
    "permissions": ["account:read", "account:write", ...],
    "tenant_id": "uuid",
    "tenant_name": "北京XX科技有限公司"
  }
}
```

### POST /auth/refresh

刷新 token（用于代账会计切换租户等场景）。

**Headers:** `Authorization: Bearer <old_token>`

**Request Body:**
```json
{
  "new_tenant_id": "uuid"   // 切换到的目标租户 ID
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOi...",
    "expires_at": "2026-06-26T12:00:00Z"
  }
}
```

---

## 健康检查

### GET /health

**Response 200:**
```json
{ "status": "ok", "version": "1.0.0", "db": "connected", "redis": "connected" }
```

---

## 租户（代账会计专用）

### GET /tenants

代账会计获取其管辖的客户公司列表。

**Headers:** `Authorization: Bearer <token>`

**Query Params:** `?page=1&pageSize=20&keyword=`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "name": "北京XX科技",
        "company_id": "uuid",
        "fiscal_year_start": "2026-01-01",
        "status": "active",
        "unread_alerts": 3
      }
    ],
    "total": 15,
    "page": 1,
    "pageSize": 20
  }
}
```

---

## 审计日志

### GET /audit-logs

**Headers:** `Authorization: Bearer <token>`

**Query Params:** `?object_type=journal_entry&object_id=uuid&actor_id=uuid&action=submit&page=1&pageSize=20`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "action": "submit",
        "object_type": "journal_entry",
        "object_id": "uuid",
        "actor_id": "uuid",
        "actor_name": "张三",
        "changed_fields": {
          "docstatus": [0, 1]
        },
        "metadata": {
          "ip": "192.168.1.1",
          "user_agent": "Mozilla/5.0..."
        },
        "created_at": "2026-05-27T10:30:00Z"
      }
    ],
    "total": 500,
    "page": 1,
    "pageSize": 20
  }
}
```

> **约束**：审计日志仅有 `GET` 和 `POST`（后端自动写入），无 `PUT`/`DELETE` 接口。
