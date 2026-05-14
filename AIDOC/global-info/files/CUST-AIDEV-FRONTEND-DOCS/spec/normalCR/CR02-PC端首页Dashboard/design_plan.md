# CR02 - PC端首页Dashboard Design Plan（已确认）

| 项目 | 内容 |
|------|------|
| 需求编号 | FE-CR02 |
| 文档名称 | design_plan |
| 版本 | 1.0 |
| 创建日期 | 2026-04-22 |
| 状态 | 已确认 |

---

## 一、目标

本文件用于在 Design 阶段先确认技术方案、接口边界与交付拆分。  
你确认后，我再生成正式 `design.md`。

---

## 二、设计前提（来自已确认 Requirement）

- 首页 Dashboard 采用“后端新增聚合接口 + 前端消费聚合结果”
- 指标口径按已确认方案不变
- 时间范围固定为：`MONTH` / `QUARTER` / `YEAR`
- 快捷入口按权限过滤显示
- 异常处理采用模块级降级，不阻断整页

---

## 三、待确认设计问题

### [Question-D01] 后端聚合接口协议是否采用“单接口全量返回”？

当前建议（默认）：
- 新增单接口：`GET /dashboard/overview`
- 请求参数：`timeRange`（MONTH/QUARTER/YEAR）
- 响应包含：KPI、工单趋势、收缴趋势、公告统计、模块状态（success/error）

[Answer-D01（已确认）]
- B. 改为多接口聚合，按所提的维度每个单独给接口，这样也便于做权限控制

---

### [Question-D02] 前端页面组件化粒度是否按“页面+2个子组件”？

当前建议（默认）：
- `HomeView.vue`：页面编排与状态管理
- `KpiCard.vue`：指标卡
- `QuickEntry.vue`：快捷入口

[Answer-D02（已确认）]
- A. 同意该粒度  


---

### [Question-D03] 图表实现是否统一沿用 ECharts 现有风格？

当前建议（默认）：
- 保持与 `workorder/statistics`、`billing/statistics` 同风格
- 深浅主题共用配置基线，仅调整颜色变量映射

[Answer-D03（已确认）]
- A. 同意复用现有图表风格  

---

### [Question-D04] 聚合接口异常语义是否需要标准化字段？

当前建议（默认）：
- 每个模块返回：`status`（SUCCESS/ERROR/NO_PERMISSION）+ `message`
- 前端按模块状态渲染内容、占位或错误

[Answer-D04（已确认）]
- A. 同意标准化模块状态字段  


---

### [Question-D05] 本 CR 的交付边界是否仅限前端文档与前端实现计划？

当前建议（默认）：
- 本仓只落前端 spec（design + tasks）
- 后端接口实现作为“外部依赖”在设计中标注

[Answer-D05（已确认）]
- B. 需要同时输出后端接口设计草案（可附录在 design）

---

## 四、确认后下一步

你确认上述 D01~D05 后，我将：
1. 生成正式 `design.md`（仅 Design 阶段产物）
2. 明确接口契约、前端架构、状态流、异常降级、验收映射
3. 等你确认 Design 阶段通过后，再进入 `tasks.md`

