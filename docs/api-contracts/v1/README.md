# API 契约目录

**路径**：`api-contracts/v1/`
**用途**：前后端对账与代码生成的唯一数据源
**格式**：Markdown（人类可读）+ 附录 OAS 3.0 YAML（工具可读）

---

## 文件结构

```
api-contracts/v1/
├── openapi.yml              # OAS 3.0 根文件：schemas / enums / components
├── auth-health-tenant.md    # F0 认证 / 健康检查 / 租户切换 / 审计日志
├── setup-f1.md              # F1 账套 / 科目表 / 资金账户 / 客商 / 规则库
├── bank-invoice-f2.md       # F2 银行流水导入 / 字段映射 / 发票 OCR
├── reconciliation-f3.md     # F3 核销预检 / 五级匹配 / 手工核销 / 回退
├── voucher-f5.md            # F5 凭证模板 / 状态机 / 审核 / 批量生成
└── reconciliation-period-analytics-f4-f6-f7.md
                             # F4 银企对账 + F6 结账体检/折旧/报表 + F7 经营分析
```

---

## 模块 → 文件映射

| 模块 | 文件 | 接口数 | 覆盖页面 |
|:---|:---|:---:|:---|
| 认证/auth | auth-health-tenant.md | 4 | 登录页 |
| 健康检查 | auth-health-tenant.md | 1 | — |
| 租户切换 | auth-health-tenant.md | 1 | 租户选择器 |
| 审计日志 | auth-health-tenant.md | 1 | 审计日志页 |
| 账套 | setup-f1.md | 3 | 创建向导 |
| 会计期间 | setup-f1.md | 2 | 期间管理 |
| 科目表 | setup-f1.md | 7 | 科目树 CRUD 页 |
| 资金账户 | setup-f1.md | 5 | 银行账户管理 |
| 客商档案 | setup-f1.md | 6 | 客商列表/导入 |
| 分类规则库 | setup-f1.md | 5 | 规则配置页 |
| 科目映射规则 | setup-f1.md | 2 | 映射配置页 |
| 银行流水导入 | bank-invoice-f2.md | 9 | 上传/核对工作台 |
| 字段映射模板 | bank-invoice-f2.md | 2 | 映射配置页 |
| 发票 | bank-invoice-f2.md | 7 | 发票列表/详情 |
| 核销预检 | reconciliation-f3.md | 2 | 预检看板 |
| 智能匹配 | reconciliation-f3.md | 3 | 匹配推荐页 |
| 手工核销 | reconciliation-f3.md | 3 | 手工核销页 |
| 核销执行/回退 | reconciliation-f3.md | 2 | — |
| 凭证模板 | voucher-f5.md | 3 | 模板配置页 |
| 凭证 CRUD | voucher-f5.md | 6 | 凭证列表/编辑页 |
| 审核工作台 | voucher-f5.md | 4 | 审核页 |
| 银企对账 | r-p-a-f4-f6-f7.md | 5 | 对账看板/调节表 |
| 资金日记账 | r-p-a-f4-f6-f7.md | 3 | 日记账/盘点 |
| 结账体检 | r-p-a-f4-f6-f7.md | 3 | 体检报告页 |
| 固定资产 | r-p-a-f4-f6-f7.md | 3 | 资产/折旧页 |
| 财务报表 | r-p-a-f4-f6-f7.md | 4 | 报表页 |
| 经营分析 | r-p-a-f4-f6-f7.md | 4 | 看板/查询 |
| **总计** | **6 文件** | **~95 接口** | **~25 页面** |

---

## 通用约定

| 规范项 | 约定值 |
|:---|:---|
| Base URL | `/api/v1` |
| Content-Type | `application/json`（文件上传用 `multipart/form-data`） |
| 认证 | `Authorization: Bearer <JWT>` |
| 响应格式 | `{ code: 0, message: "ok", data: {...} }` |
| 分页请求 | `?page=1&pageSize=20&sort=field:desc` |
| 分页响应 | `{ list: [...], total, page, pageSize }` |
| 时间格式 | ISO 8601 UTC |
| 金额格式 | 字符串 `"1234.56"`（避免 JSON 浮点精度丢失） |
| UUID 格式 | 标准 36 字符 `"a1b2c3d4-e5f6-..."` |
| 枚举 | 后端传 code，前端映射中文标签（详见 `dict.config.ts`） |

## 错误码分类

| 范围 | 含义 | 示例 |
|:---|:---|:---|
| `0` | 成功 | |
| `10001-10099` | 认证错误 | 10001=账号/密码错误, 10002=Token过期 |
| `10101-10199` | 多租户 | 10101=租户切换失败 |
| `20001-20099` | 业务校验通用 | 20001=借贷不平衡, 20002=数据不存在 |
| `20101-20199` | 科目表 | 20101=Group科目不可记账 |
| `20201-20299` | 凭证 | 20201=凭证编号已存在 |
| `20301-20399` | 核销 | 20301=预检未通过 |
| `20401-20499` | 导入 | 20401=格式无法解析 |
| `30001-30099` | 权限 | 30001=权限不足 |
| `40001-40099` | 系统 | 40001=内部错误 |

---

## 使用方式

1. **前端开发**：按模块文件查阅请求/响应体结构，生成 TypeScript 类型
2. **后端开发**：按模块文件实现 Handler，确保请求/响应格式一致
3. **测试**：用接口定义编写集成测试断言
4. **Mock**：前端基于接口定义启动 Mock Server 独立开发
