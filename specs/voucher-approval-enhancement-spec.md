# 凭证审批流程增强 Spec

## 背景
当前凭证审批流程存在以下不足：
1. 已提交的凭证无法撤回（场景B）
2. 驳回后状态变回 draft，而非独立的 "已驳回" 状态（场景C）
3. 制单人与审核人可以是同一人，缺乏职责分离控制
4. 缺少会计期间结账后的跨期保护

## 目标
实现完整的凭证审批生命周期管理，满足财务合规要求。

---

## 详细设计

### 1. 新增 `rejected` 状态

**文件**: `internal/model/journal.go`

在 VoucherStatus 常量中新增：
```go
VoucherStatusRejected VoucherStatus = "rejected" // 已驳回（需修改后重新提交）
```

状态映射：docstatus = 4

### 2. 新增 `revoke` (撤回) action

**文件**: `internal/model/journal.go`

在 VoucherAction 常量中新增：
```go
VoucherActionRevoke VoucherAction = "revoke" // 撤回（仅 posted 状态可用，且无人审批时）
```

### 3. 修改状态转换规则

**文件**: `internal/service/voucher_state_machine.go`

**修改前**的状态转换：
| 当前状态 | 允许的操作 |
|----------|------------|
| draft (0) | submit, cancel |
| posted (1) | approve, reject, reverse, cancel |
| verified (2) | reverse |
| cancelled (3) | 无 |

**修改后**的状态转换：
| 当前状态 | 允许的操作 |
|----------|------------|
| draft (0) | submit, cancel |
| posted (1) | approve, reject, reverse, **revoke** |
| rejected (4) | submit, cancel |
| verified (2) | reverse |
| cancelled (3) | 无 |

关键变化：
- **新增 revoke**: `posted (1) → draft (0)` — 制单人撤回自己的凭证（仅当无 pending 审批任务时）
- **修改 reject**: `posted (1) → rejected (4)` — 审核人驳回（不再回到 draft）
- **新增 rejected 状态 submit**: `rejected (4) → posted (1)` — 修改后重新提交
- **移除 posted cancel**: posted 状态不再允许 cancel，只能用 revoke（制单人）或 reverse（审核人）

### 4. 职责分离校验

**文件**: `internal/service/approval_service.go`

在 `ApproveTask` 方法中，执行审批前校验：
```go
// 获取凭证信息
journal, err := s.journalRepo.GetByID(ctx, tenantID, task.JournalEntryID)
if err != nil {
    return fmt.Errorf("get journal entry: %w", err)
}

// 职责分离：制单人不能审核自己的凭证
if journal.CreatedBy == userID {
    return errors.New("制单人不能审核自己的凭证")
}
```

**文件**: `internal/handler/approval_handler.go`

在 Handler 层也做同样的校验，返回 403 错误码。

### 5. 操作留痕

当前 `AuditLog` 系统已记录：
- `action` — 操作类型（voucher_status_change）
- `actor_id` / `actor_name` — 操作人
- `changed_fields` — 变更字段
- `metadata` — 包含 reason 等元数据

**新增留痕内容**：
- revoke 操作：记录 `{"action": "revoke", "reason": "..."}` 
- reject 操作：记录驳回原因
- 职责分离违规：记录审计告警

### 6. 会计期间结账控制（本期暂缓）

**状态**: 列入未来开发计划，本期不实施。

原因：需要新增完整的会计期间管理模块，工作量较大。

---

## API 变更

### 新增 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/vouchers/:id/revoke` | 撤回已提交的凭证 |
| POST | `/vouchers/:id/resubmit` | 重新提交被驳回的凭证 |

### 现有 API 行为变更

| 方法 | 路径 | 变更 |
|------|------|------|
| POST | `/approvals/:id/approve` | 新增制单人≠审核人校验 |
| POST | `/approvals/:id/reject` | reject 后凭证状态变为 rejected (4) 而非 draft (0) |

---

## 前端变更

### VoucherList.vue
- 列表新增"撤回"按钮（仅对 posted 状态且当前用户是制单人的凭证显示）
- 列表新增"重新提交"按钮（仅对 rejected 状态且当前用户是制单人的凭证显示）
- 状态筛选新增"已驳回"选项

### VoucherEdit.vue
- rejected 状态的凭证允许编辑
- verified 状态的凭证禁止编辑（已有）
- 编辑时显示驳回原因

### ReviewWorkbench.vue
- 驳回操作后，前端提示凭证已进入"已驳回"状态

---

## 数据迁移

需要在数据库中新增状态常量（如果 docstatus 使用 check constraint）：
```sql
-- 确保 docstatus = 4 被允许
-- 如果存在 check constraint，需要更新
```

---

## 测试要点

1. 撤回：posted → draft，清除审批任务
2. 驳回：posted → rejected，保留审批任务记录
3. 重新提交：rejected → posted，创建新审批任务
4. 职责分离：制单人审批自己的凭证应返回 403
5. 操作留痕：所有操作在 audit_log 中有记录
