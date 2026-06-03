# SPEC: TASK01 — 结账后自动归档

## 基本信息

- **任务 ID**: phase4-archive-001
- **类型**: feature
- **优先级**: medium
- **依赖**: 无
- **执行者**: OpenCode

## 背景

当前归档(ArchiveList)需要手动操作。结账完成后，系统应自动将当月数据打包归档。

## 目标

在结账完成事件中触发自动归档：将当月凭证、科目余额表、报表、申报表打包。

## 详细设计

### 归档内容

```
归档包/2026/05/
├── 凭证/          所有已过账凭证的PDF导出
├── 账簿/
│   ├── 科目余额表.xlsx
│   └── 总账.xlsx
├── 报表/
│   ├── 资产负债表.xlsx
│   ├── 利润表.xlsx
│   └── 现金流量表.xlsx
├── 申报表/        （当月的税务申报记录）
└── 结账报告.md     结账检查报告
```

### 实现逻辑

在 `close_period_core()` 中触发：

```python
def close_period_core(...):
    # ... 现有结账逻辑 ...
    
    # 新增：触发归档
    archive_service.archive_period(db, year, month)
    
    # ... 返回 ...
```

`archive_service.archive_period()` 负责：
1. 查询当月所有已过账凭证
2. 导出科目余额表
3. 导出三大报表
4. 写入 ArchiveList
5. （可选）将导出文件保存到存储目录

### 归档记录表

复用现有 `ArchiveList` 模型（或扩展字段）：

```python
# ArchiveList 已有字段：id, title, type, file_path, created_at
# 扩展字段：period_year, period_month, archive_content (JSON)
period_year = Column(Integer)
period_month = Column(Integer)
archive_content = Column(Text)  # 归档内容清单JSON
```

## 验收标准

- [ ] 结账成功 → 自动生成一条归档记录
- [ ] 归档记录显示该月的凭证、报表清单
- [ ] 归档记录可查看和下载

## OpenCode 指令

**目标**：在结账完成后自动触发当月数据归档。

**约束**：
- 在 `close_period_core` 中触发
- 归档逻辑写在 `app/services/archive_service.py`（新建）
- 写入 ArchiveList 表

**上下文**：
- repo: `/root/huihua-financial-master`
- 参考现有 ArchiveList 模型

**验收**：
- 结账后自动产生归档记录
- 归档内容完整
