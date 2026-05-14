# 工单中心域

## 域职责
报修工单、投诉建议、工单派发与流转、满意度评价。

## 对应代码模块
`spmp-workorder-center`

## 目录结构

```
workorder-center/
├── domain-share/        # 域共享文档
│   ├── api.md           # API 文档
│   ├── database.md      # 数据库设计
│   ├── flow.md          # 业务流程
│   └── integration.md   # 系统集成
├── spec/                # 需求规范
│   ├── normalCR/        # 常规 CR
│   │   └── CR04-工单管理/
│   │       ├── requirements.md
│   │       ├── design.md
│   │       ├── tasks.md
│   │       ├── requirements_plan.md
│   │       ├── design_plan.md
│   │       └── change-log/    # CR 级变更日志
│   └── minorCR/         # 小型 CR
├── change-log/          # 域级变更日志
│   ├── README.md
│   └── 2026-04/
├── code-review/         # 代码审查
└── vibe/                # Vibe 开发
```

## 当前进展

| CR | 状态 | 说明 |
|----|------|------|
| CR04-工单管理 | 开发中 | 后端+前端主体已完成，Excel 导出和联调待完成 |
