# CR02 - PC端首页Dashboard 技术设计文档

| 项目 | 内容 |
|------|------|
| 需求编号 | FE-CR02 |
| 需求名称 | PC端首页 Dashboard 建设 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-22 |
| 状态 | 草稿 |

---

## 一、设计概述

基于已确认 Requirement 与 Design Plan，本方案采用“多接口聚合”实现首页 Dashboard：后端按业务维度拆分接口以便权限控制，前端按模块并发请求并做模块级降级渲染。

设计确认结论：
- D01：采用多接口拆分，不做单接口全量返回
- D02：前端采用 `HomeView + KpiCard + QuickEntry` 组件粒度
- D03：图表复用现有 ECharts 风格
- D04：接口统一返回模块状态字段，前端按状态渲染
- D05：design 文档附带后端接口设计草案

---

## 二、架构与模块划分

### 2.1 前端模块

- 页面编排：`src/spmp-web-pc/src/views/home/HomeView.vue`
- 指标卡组件：`src/spmp-web-pc/src/views/home/components/KpiCard.vue`
- 快捷入口组件：`src/spmp-web-pc/src/views/home/components/QuickEntry.vue`
- 图表实现：沿用项目 `ECharts` 使用方式

### 2.2 数据流

1. 进入 `/home`
2. 前端根据时间范围并发请求各模块接口
3. 每个接口返回 `status + message + data`
4. 页面按模块状态分别渲染成功、无权限、错误或空态
5. 切换时间范围后重复步骤 2~4

---

## 三、接口设计草案（后端）

说明：以下为本 CR 约定的后端接口草案，用于前后端联调契约。

### 3.1 KPI 概览接口

- 方法：`GET /dashboard/kpi`
- 参数：
  - `timeRange`: `MONTH | QUARTER | YEAR`
- 返回：

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "status": "SUCCESS",
    "message": "",
    "kpis": {
      "pendingCount": 12,
      "inProgressCount": 8,
      "monthlyCompletedCount": 156,
      "avgRepairDuration": 92,
      "totalReceivable": 320000.50,
      "totalReceived": 285000.00,
      "collectionRate": 89.06,
      "overdueAmount": 35000.50,
      "recentNoticeCount": 6
    }
  }
}
```

### 3.2 工单趋势接口

- 方法：`GET /dashboard/trend/workorder`
- 参数：
  - `timeRange`: `MONTH | QUARTER | YEAR`
- 返回字段：
  - `status`
  - `message`
  - `trend`: `[{ date, value }]`

### 3.3 缴费趋势接口

- 方法：`GET /dashboard/trend/billing`
- 参数：
  - `timeRange`: `MONTH | QUARTER | YEAR`
- 返回字段：
  - `status`
  - `message`
  - `trend`: `[{ period, receivable, received, collectionRate }]`

### 3.4 快捷入口权限接口（可复用现有权限）

- 方法：`GET /dashboard/quick-entries`（可选）
- 参数：无
- 返回字段：
  - `status`
  - `message`
  - `entries`: `[{ key, title, route, permission, visible }]`

说明：若当前系统已有完整 `permissions` 可直接复用，则该接口可不新增，由前端本地映射入口清单并过滤。

---

## 四、状态语义标准化

所有 Dashboard 模块接口返回统一语义：

- `SUCCESS`：有权限且数据可用
- `NO_PERMISSION`：用户无权限查看该模块
- `ERROR`：模块数据拉取失败
- `EMPTY`：有权限但暂无数据（可选）

统一结构建议：

```json
{
  "status": "SUCCESS",
  "message": "",
  "data": {}
}
```

---

## 五、Dashboard 数据权限设计（细化）

### 5.1 权限范围定义

Dashboard 权限控制分为三层：

- 模块级（Module Level）：控制是否可查看某类数据模块
  - 工单模块：KPI 工单项 + 工单趋势图
  - 缴费模块：KPI 缴费项 + 缴费趋势图
  - 公告模块：公告统计项
- 字段级（Field Level）：模块可见时，控制敏感字段是否脱敏/隐藏
  - 示例：金额类字段、收缴率字段、平均时长字段
- 入口级（Entry Level）：快捷入口是否可见、可跳转

### 5.2 建议权限点映射

后端可复用现有权限码，建议最小映射如下（Design 口径）：

| Dashboard 能力 | 建议权限码（示例） | 说明 |
|---|---|---|
| 查看工单概览 | `workorder:statistics` | 控制工单 KPI 与工单趋势 |
| 查看缴费概览 | `billing:statistics` | 控制缴费 KPI 与缴费趋势 |
| 查看公告概览 | `notice:list` | 控制公告统计 |
| 查看工单入口 | `workorder:list` | 控制“工单列表”快捷入口 |
| 查看收费统计入口 | `billing:statistics` | 控制“收费统计”入口 |
| 查看账单入口 | `billing:bill:list` | 控制“账单管理”入口 |

说明：最终权限码以后端权限中心实际配置为准，前后端联调时完成映射校准。

### 5.3 控制链路

1. 用户登录后获取 `roles + permissions`
2. Dashboard 请求时携带用户身份（Token）
3. 后端先鉴权，再按数据权限范围（如小区/项目）聚合数据
4. 每个模块返回标准状态：
   - `SUCCESS`：有权限且返回数据
   - `NO_PERMISSION`：无权限
   - `ERROR`：有权限但执行失败
   - `EMPTY`：有权限但无数据
5. 前端根据状态渲染模块内容或占位

### 5.4 数据范围控制（Data Scope）

除功能权限外，后端需叠加数据范围约束：

- 若用户为“全局数据权限”：返回可见范围内全部数据
- 若用户为“片区/小区/楼栋级权限”：仅聚合其授权范围数据
- 若无对应数据范围权限：返回 `NO_PERMISSION`

约束原则：
- 数据范围判断在后端完成，前端不做数据裁剪判定
- Dashboard 与明细页数据口径一致，避免“看板有数、列表无权限”不一致

### 5.5 前端渲染判定矩阵

| 模块状态 | 前端展示 | 是否展示数值 |
|---|---|---|
| `SUCCESS` | 正常卡片/图表 | 是 |
| `EMPTY` | “暂无数据”空态 | 否 |
| `NO_PERMISSION` | “暂无权限”占位 | 否 |
| `ERROR` | “加载失败，点击重试” | 否 |

### 5.6 验收口径（权限专项）

- 同一账号在 Dashboard 与业务明细页的可见范围一致
- 移除某模块权限后，对应模块与入口同时消失或转“暂无权限”
- 仅有入口权限但无统计权限时：入口可见，统计模块不可见
- 切换不同数据权限账号（全局/小区级）时，指标值按授权范围变化

---

## 六、前端实现设计

### 6.1 页面状态模型

建议状态定义：

```ts
type ModuleStatus = 'SUCCESS' | 'NO_PERMISSION' | 'ERROR' | 'EMPTY'

interface DashboardModuleState<T> {
  loading: boolean
  status: ModuleStatus
  message: string
  data: T
}
```

### 6.2 请求策略

- 首次进入页面：并发请求 KPI、工单趋势、缴费趋势
- 使用 `Promise.allSettled` 避免单点失败影响整页
- 切换时间范围：重新请求全部模块
- 重试机制：模块级按钮触发对应接口重试

### 6.3 快捷入口策略

- 入口候选固定配置在前端常量中
- 根据 `userStore.permissions` 过滤可见入口
- 无可见入口时展示空态占位

---

## 七、图表与主题适配

- 图表风格复用现有统计页配置（折线、柱状、tooltip、legend）
- 颜色映射使用现有 CSS 变量，兼容三主题：
  - 商务经典
  - 暗夜奢华
  - 极光科技
- 响应式布局：
  - `>=1200px`：双列图表
  - `<1200px`：单列图表

---

## 八、异常与降级设计

- 接口失败：仅对应模块进入错误态，展示“加载失败/重试”
- 无权限：展示“暂无权限”占位
- 空数据：展示“暂无数据”
- 页面级兜底：不得白屏，不因单模块异常阻断其他模块

---

## 九、验收映射

- Requirement US-01：KPI 卡片 + 首页替换占位
- Requirement US-02：2 个趋势图 + 时间范围切换
- Requirement US-03：权限控制 + 模块级降级
- 非功能性：三主题可读性、并发与可维护性满足

---

## 十、外部依赖与边界

- 后端需按本草案提供 Dashboard 多接口
- 前端在后端接口未就绪时可使用 mock 数据联调
- 本阶段不进入任务拆解，待你确认 `design.md` 后再产出 `tasks.md`

