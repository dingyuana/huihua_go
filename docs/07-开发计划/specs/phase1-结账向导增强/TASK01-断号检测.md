# SPEC: TASK01 — 结账检查：凭证断号检测

## 基本信息

- **任务 ID**: phase1-close-001
- **类型**: feature
- **优先级**: high
- **依赖**: 无
- **执行者**: OpenCode
- **估算**: 后端 120 行 + 测试 80 行

## 背景

当前 PeriodClose 页面的"凭证统计"只显示各状态的凭证张数。用户看不到凭证号是否连续。SOP 第 1 步明确要求检查"凭证号是否断号，断号是否合理（作废/遗漏）"。

财务合规要求：凭证号必须连续，断号必须有合理解释（如作废）。

## 目标

在 `pre_close_check` 返回体中追加 `voucher_gaps` 字段，列出所有断号项。前端在"基础检查"区域展示断号告警。

## 技术约束

- 不修改数据库模型
- 不新增路由（复用 GET `/api/v1/period/{year}/{month}/pre-close-check`）
- 不修改现有返回字段
- 断号检测逻辑放在 `period_pre_close_check` 函数中（或抽成独立函数放在 accounting.py 中供复用）

## 详细设计

### 返回值结构（追加字段）

```json
{
  "voucher_gaps": [
    {
      "expected_no": 35,
      "is_filled": false,
      "gap_type": "missing",
      "message": "第 35 号凭证缺失"
    },
    {
      "expected_no": 42,
      "is_filled": true,
      "fill_voucher_id": "uuid-xxx",
      "fill_voucher_no": "记-42",
      "gap_type": "voided",
      "void_reason": "作废",
      "message": "第 42 号凭证已作废"
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|:---|:---|:---|
| `expected_no` | int | 预期存在的凭证号 |
| `is_filled` | bool | 是否已填补（作废也算已填补） |
| `gap_type` | string | `"missing"`(缺失) / `"voided"`(作废) / `"skipped"`(跳号) |
| `fill_voucher_id` | string? | 填补该号的凭证 ID |
| `fill_voucher_no` | string? | 填补该号的凭证编号（前台展示） |
| `void_reason` | string? | 作废原因 |
| `message` | string | 用户可读的提示 |

### 函数：scan_voucher_gaps(db, year, month)

```python
def scan_voucher_gaps(db: Session, year: int, month: int) -> list[dict]:
    """
    检测指定期间的凭证断号。

    逻辑：
    1. 查询该月所有凭证，按 voucher_no 排序
    2. 解析凭证号中的数字部分（"记-35" → 35，"转-12" → 12）
    3. 收集所有出现的数字，检查 1..max 之间的缺失
    4. 对缺失的号码，检查是否有作废/冲销凭证占位
    5. 无占位的标记为 gap_type="missing"（需关注）
    6. 有作废占位的标记为 gap_type="voided"（需确认合理性）

    凭证号格式（从现有数据推断）：
    - "记-35" / "记-36" 等 → 数字在 "-" 之后
    - 也可能是纯数字（需兼容）

    边界情况：
    - 空期间（无凭证）→ 返回空列表
    - 凭证号非数字（如 "调-1"）→ 跳过该字号的检测
    - 跨字号（"记-1" 到 "记-5" 和 "转-1" 到 "转-3"）→ 各自独立检测
    - 已作废凭证（status=rejected 或 is_reversed=true）→ 计入但标记为 voided

    返回: list[dict]，按 expected_no 升序排列
    """
```

### 凭证号解析策略

凭证号字段 `voucher_no` 的常见格式：
- `记-1`, `记-35` → 分词：word="记", num=35
- `转-1`, `转-12` → 分词：word="转", num=12
- `35`（纯数字）→ 分词：word=null, num=35

解析函数：

```python
import re

def _parse_voucher_no(voucher_no: str) -> tuple[str | None, int]:
    """
    解析凭证号，返回 (字头, 数字)。
    例: "记-35" → ("记", 35), "35" → (None, 35)
    """
    m = re.match(r'^(\D+)?-?(\d+)$', voucher_no.strip())
    if not m:
        return (None, 0)
    word = m.group(1) or None
    num = int(m.group(2))
    return (word, num)
```

### 入口修改

在 `period_pre_close_check` 函数中，return 前追加：

```python
voucher_gaps = scan_voucher_gaps(db, year, month)

return {
    **base_result,  # 现有返回键，不包括 risk_warnings 等（已单独列出）
    "voucher_gaps": voucher_gaps,
}
```

### 前端展示

在 PeriodClose.vue 的"基础检查"区域，追加断号检查项：

```html
<div v-if="voucherGaps.length > 0" class="check-item failed">
  <el-icon><CircleClose /></el-icon>
  <div class="check-item-content">
    <span class="check-item-title">凭证号连续</span>
    <span class="check-item-hint">
      {{ missingCount }} 个断号待处理
    </span>
  </div>
</div>
<div v-else class="check-item passed">
  <el-icon><CircleCheck /></el-icon>
  <span class="check-item-title">凭证号连续</span>
</div>

<!-- 断号详情（可展开） -->
<el-collapse v-if="voucherGaps.length > 0">
  <el-collapse-item title="查看断号明细">
    <div v-for="gap in voucherGaps" :key="gap.expected_no"
         class="gap-item" :class="gap.gap_type">
      <span>{{ gap.message }}</span>
      <el-tag v-if="gap.gap_type === 'voided'" size="mini" type="info">已作废</el-tag>
      <el-tag v-else size="mini" type="danger">缺失</el-tag>
    </div>
  </el-collapse-item>
</el-collapse>
```

**交互逻辑**：
- `missing` 类型的断号 → 红色，计入"未通过检查项"，阻断结账
- `voided` 类型的断号 → 橙色，计入"需确认"但不阻断结账
- 无断号 → 绿色"凭证号连续"，不展示更多内容

**可阻断条件**：只有当 `voucher_gaps` 中存在 `gap_type === "missing"` 的项时，`can_close` 才返回 false。仅有 `voided` 类型时不阻断。

## 测试策略

### 单元测试：scan_voucher_gaps

在 `test_period.py` 中新增测试类 `TestScanVoucherGaps`：

| 场景 | 测试数据 | 期望 |
|:---|:---|:---|
| 连续号 | 1,2,3,4,5 | 空列表 |
| 中间缺号 | 1,2,4,5 | [{expected_no:3, gap_type:"missing"}] |
| 多段断号 | 1,4,5,8 | [{expected_no:2}, {expected_no:3}, {expected_no:6}, {expected_no:7}] |
| 空期间 | 无凭证 | 空列表 |
| 作废占位 | 1,2,3(作废),4,5 | [{expected_no:3, gap_type:"voided"}] |
| 多字头 | 记-1,记-2,转-1,转-2 | 空列表（各自连续） |
| 非数字凭证号 | 调-1 | 跳过该字头 |
| 首号缺失 | 2,3,4 | [{expected_no:1}] |

### 集成测试

在 `TestPeriodAPI` 中追加：
- `test_pre_close_check_voucher_gaps`：调用 pre_close_check，验证返回体中 voucher_gaps 存在且为 list

## 验收标准

- [ ] 有断号时 pre_close_check 返回 voucher_gaps 非空列表
- [ ] 无断号时返回空列表
- [ ] 已作废凭证占位的断号标记 gap_type="voided"
- [ ] 缺失断号标记 gap_type="missing"
- [ ] 多字头凭证各自独立检测
- [ ] 空期间返回空列表
- [ ] 前端展示断号告警，缺失类型阻断结账
- [ ] 所有现有测试通过

## 不变清单

- 不修改 AccVoucher 模型
- 不修改 AccPeriod 模型
- 不修改凭证状态机（draft→pending→posted 流程）
- 不修改损益结转逻辑
- 不修改前端人工确认清单检查项
