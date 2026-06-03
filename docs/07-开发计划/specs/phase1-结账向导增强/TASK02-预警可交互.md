# SPEC: TASK02 — 风险预警可交互：确认/忽略机制

## 基本信息

- **任务 ID**: phase1-close-002
- **类型**: feature
- **优先级**: high
- **依赖**: phase0-close-001（pre_close_check 已有 risk_warnings）
- **执行者**: OpenCode
- **估算**: 后端 80 行模型+API + 前端 120 行交互 + 测试 60 行

## 背景

当前 risk_warnings 在页面上只读展示，用户无法操作。SOP 要求"异常必处理"——用户需要能对每条预警做出回应（确认/忽略并备注），且确认状态需要持久化，刷新页面后不丢失。

## 目标

1. 新建 `WarningAcknowledgement` 模型，持久化预警处理状态
2. 新建 POST 接口，接收预警确认/忽略操作
3. pre_close_check 返回体中追加每个预警的 `acknowledged` 状态
4. 前端预警卡片增加 [确认] [忽略并备注] 按钮

## 技术约束

- 只需新建一个模型（`Base.metadata.create_all` 自动建表，无需 Alembic）
- 不修改现有 pre_close_check 返回字段（仅追加 `acknowledged` 字段到每个 warning 对象）
- 不修改凭证状态机、结账锁定逻辑
- 忽略操作需要填写原因（必填）

## 详细设计

### 1. 新增模型：WarningAcknowledgement

在 `backend/app/models/accounting.py` 中追加：

```python
class WarningAcknowledgement(Base):
    """结账预警确认记录"""
    __tablename__ = "warning_acknowledgements"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    period_year = Column(Integer, nullable=False, index=True)
    period_month = Column(Integer, nullable=False, index=True)

    # 预警标识（用于匹配同一条预警）
    warning_type = Column(String(50), nullable=False)      # "reclassification" / "negative_inventory" 等
    warning_key = Column(String(100), nullable=False)      # 唯一标识: "1122-XX客户" 或 "AR-XX客户"

    # 处理状态
    action = Column(String(20), nullable=False, default="pending")
    # "acknowledged" = 已确认（用户确认已处理）
    # "ignored" = 已忽略（用户认为不构成问题）
    # "pending" = 待处理（默认）

    # 操作信息
    remark = Column(Text, nullable=True)                   # 忽略原因 / 处理备注
    acknowledged_by = Column(String(36), nullable=True)    # 操作人 ID
    acknowledged_by_name = Column(String(50), nullable=True)
    acknowledged_at = Column(DateTime, nullable=True)

    # 审计
    created_at = Column(DateTime, default=datetime.now)

    __table_args__ = (
        # 同一期间+同类型+同key 唯一（防止重复确认）
        UniqueConstraint('period_year', 'period_month', 'warning_type', 'warning_key',
                         name='uq_warning_ack'),
    )
```

### 2. 新增 API

在 `period.py` 中新增路由：

```
POST /api/v1/period/warnings/acknowledge
```

```python
class WarningAckRequest(BaseModel):
    period_year: int
    period_month: int
    warning_type: str
    warning_key: str
    action: Literal["acknowledged", "ignored"]
    remark: str | None = None

@router.post("/warnings/acknowledge")
async def acknowledge_warning(
    body: WarningAckRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(require_roles("accounting", "finance")),
):
    """
    确认或忽略一条结账预警。

    幂等：同一 (period_year, period_month, warning_type, warning_key)
    多次调用后以最后一次为准（upsert 语义）。
    """
```

### 3. 修改 pre_close_check 返回值

为每个 risk_warning 对象追加 `acknowledged` 字段：

```python
# 在 period_pre_close_check 中，return 前追加：
acknowledged_warnings = db.query(WarningAcknowledgement).filter(
    WarningAcknowledgement.period_year == year,
    WarningAcknowledgement.period_month == month,
).all()
ack_map = {(w.warning_type, w.warning_key): w for w in acknowledged_warnings}

for warning in risk_warnings:
    ack = ack_map.get((warning["type"], warning["key"]))
    warning["acknowledged"] = ack.action if ack else "pending"
    if ack and ack.action == "ignored":
        warning["acknowledged_reason"] = ack.remark
```

### 4. 前端交互设计

当前 risk_warnings 区域已经展示预警列表。改造后：

```html
<el-card class="risk-section" shadow="never">
  <template #header>
    <div class="section-header">
      <span class="section-step">风险预警</span>
      <el-tag v-if="pendingWarningCount === 0" type="success">已全部处理</el-tag>
      <el-tag v-else type="danger">{{ pendingWarningCount }} 项待处理</el-tag>
    </div>
  </template>
  <div v-for="warn in riskWarnings" :key="warn.key"
       class="risk-item" :class="'risk-' + warn.severity">

    <el-icon><WarningFilled /></el-icon>
    <div class="risk-content">
      <span class="risk-message">{{ warn.message }}</span>
    </div>

    <!-- 操作按钮 -->
    <div class="risk-actions" v-if="warn.acknowledged === 'pending'">
      <el-button size="small" type="primary" plain
                 @click="handleAck(warn, 'acknowledged')">
        确认已处理
      </el-button>
      <el-button size="small" type="info" plain
                 @click="handleIgnore(warn)">
        忽略
      </el-button>
    </div>

    <!-- 已处理状态 -->
    <el-tag v-else-if="warn.acknowledged === 'acknowledged'"
            type="success" size="small">已确认</el-tag>
    <el-tooltip v-else-if="warn.acknowledged === 'ignored'"
                :content="'忽略原因: ' + (warn.acknowledged_reason || '未填写')">
      <el-tag type="info" size="small">已忽略</el-tag>
    </el-tooltip>
  </div>
</el-card>
```

**忽略弹窗**：

```javascript
const handleIgnore = async (warn) => {
  try {
    const { value: remark } = await ElMessageBox.prompt(
      '请填写忽略原因（财务经理审批时需要此说明）',
      '忽略预警',
      { confirmButtonText: '确定忽略', cancelButtonText: '取消',
        inputPattern: /.+/, inputErrorMessage: '忽略原因不能为空' }
    )
    await acknowledgeWarning({
      period_year: year, period_month: month,
      warning_type: warn.type, warning_key: warn.key,
      action: 'ignored', remark
    })
    warn.acknowledged = 'ignored'
    warn.acknowledged_reason = remark
  } catch {}
}
```

### 5. 前端 API 函数

在 `frontend/src/api/report.js` 中追加：

```javascript
// ============ 结账预警确认 ============
export const acknowledgeWarning = (data) => {
  return request.post('/v1/period/warnings/acknowledge', data)
}
```

## 测试策略

### 单元测试

| 场景 | 步骤 | 期望 |
|:---|:---|:---|
| 确认预警 | POST /warnings/acknowledge 设置 action=acknowledged | 200，下次 pre_close_check 返回 acknowledged=true |
| 忽略预警 | POST /warnings/acknowledge 设置 action=ignored+remark | 200，下次返回 acknowledged=ignored |
| 忽略无原因 | POST /warnings/acknowledge 设置 action=ignored, remark="" | 400（必填校验） |
| 重复确认 | 两次 POST 相同 warning_type+key | 200 幂等 |
| 跨期独立 | 期间1 确认的预警不影响 期间2 | 各自独立 |
| 未认证 | 不传 auth_headers | 401 |

### 集成测试

在 `TestPeriodAPI` 中追加：
- `test_acknowledge_warning`：确认/忽略后，pre_close_check 返回对应状态
- `test_acknowledge_unauthorized`：未认证用户无法操作
- `test_acknowledge_ignore_requires_remark`：忽略时 remark 必填

## 验收标准

- [ ] `WarningAcknowledgement` 模型正确定义，启动时自动建表
- [ ] POST /warnings/acknowledge 幂等，可重复确认
- [ ] 忽略操作 remark 必填，空值返回 400
- [ ] pre_close_check 返回中每个 warning 有 `acknowledged` 字段
- [ ] 未确认的预警在前端显示操作按钮
- [ ] 已确认的预警显示"已确认"标签
- [ ] 已忽略的预警显示"已忽略"标签，hover 显示原因
- [ ] 所有预警处理后，风险预警区域显示"已全部处理"
- [ ] 56 个现有测试全部通过

## 不变清单

- 不修改 AccVoucher / AccPeriod / AccAccount 等现有模型
- 不修改现有 pre_close_check 返回字段（只追加 `acknowledged` 到每个 warning 对象）
- 不修改凭证状态机
- 不修改结账锁定逻辑
- 前端不改风险预警区域以外的部分
