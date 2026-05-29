# SPEC: TASK06 — 前端：科目映射配置页面

## 基本信息

- **任务 ID**: phase1-bill-006
- **类型**: feature
- **优先级**: medium
- **依赖**: TASK03（后端映射配置API可用）
- **执行者**: OpenCode

## 背景

TASK03 实现了后端的映射配置接口。需要前端提供一个管理界面，让财务主管可以查看和修改各类单据的默认科目映射。

## 目标

创建一个科目映射配置页面，表格展示所有映射，支持修改借贷科目。

## 技术约束

- 新建文件：`src/pages/system/BusinessMapping.vue`
- 注册路由，仅管理员(finance_manager/company_admin)可见
- 使用 Element Plus 的 el-table 和 el-cascader 选择科目

## 详细设计

### 页面布局

```
┌──────────────────────────────────────────────────────────┐
│  单据科目映射配置                [仅管理员可见]          │
├──────────────────────────────────────────────────────────┤
│  单据类型: [全部 ▼]                                     │
├──────────────────────────────────────────────────────────┤
│  单据类型 | 费用类型   | 借方科目         | 贷方科目    │
│  ────────|───────────|─────────────────|──────────────│
│  费用报销 | 差旅费    | 管理费用-差旅费  | 银行存款     │ [编辑]│
│  费用报销 | 办公费    | 管理费用-办公费  | 银行存款     │ [编辑]│
│  收款单   | 默认      | 银行存款        | 应收账款     │ [编辑]│
│  付款单   | 默认      | 应付账款        | 银行存款     │ [编辑]│
├──────────────────────────────────────────────────────────┤
│  共N条记录                                               │
└──────────────────────────────────────────────────────────┘
```

### 编辑弹窗

点击"编辑"弹出科目选择器：
```
┌──────────────────────────────────┐
│  编辑科目映射                      │
├──────────────────────────────────┤
│  单据类型: 费用报销单              │
│  费用类型: 差旅费                  │
│                                   │
│  借方科目: [管理费用 - 差旅费  ▼]  │
│  贷方科目: [银行存款  ▼]          │
│                                   │
│  [取消]  [保存]                    │
└──────────────────────────────────┘
```

### 科目选择器

使用 el-tree-select 或 el-cascader 从科目树中选择。调用现有 API `getAccountTree()` 获取科目树数据。

### API 函数

```js
// 追加到 src/api/business.js
export const getMappings = (params) => request.get('/v1/business/mappings', { params })
export const getMappingById = (id) => request.get(`/v1/business/mappings/${id}`)
export const updateMapping = (id, data) => request.put(`/v1/business/mappings/${id}`, data)
```

### 路由注册

```js
{
  path: 'system/business-mapping',
  name: 'BusinessMapping',
  component: () => import('../pages/system/BusinessMapping.vue'),
  meta: { title: '单据科目映射', roles: ['company_admin', 'finance_manager'] }
}
```

## 验收标准

- [ ] 映射列表展示所有单据类型的映射关系
- [ ] 可按单据类型筛选
- [ ] 编辑弹窗中的科目选择器可搜索并选择科目
- [ ] 修改保存后，映射表数据更新
- [ ] 修改后生成的新单据按新映射走（需要开一个新单据验证）

## OpenCode 指令

**目标**：创建一个科目映射配置页面，管理员可查看和修改单据→科目的映射关系。

**约束**：
- 新建 `src/pages/system/BusinessMapping.vue`
- 注册路由，仅管理员权限
- 科目选择使用 el-cascader 从 `getAccountTree` 获取数据
- 编辑使用 el-dialog 弹窗

**上下文**：
- repo: `/root/huihua-financial-master`
- 科目树 API: `src/api/accounting.js` 中的 `getAccountTree()`
- 参考现有系统设置页面的风格：`src/pages/system/company/CompanySettings.vue`

**验收**：
- 列表正常展示
- 编辑弹窗可修改并保存
- 权限控制正常
