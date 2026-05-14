# 正常需求模板

---

## 需求信息

| 字段 | 内容 |
|------|------|
| 需求编号 | CR{域编号}{序号} |
| 需求名称 | {需求名称} |
| 所属域 | {域名称} |
| 创建日期 | {日期} |
| 负责人 | {负责人} |

---

## 目录结构

```
CR{域编号}{序号}-{需求名称}/
├── api/                              # API 变更
│   └── api-{序号}-{模块名}.md
├── change-log/                       # 变更日志
│   └── Task-{序号}-{任务名}-change-log.md
├── design/                           # 设计文档
│   └── design.md
├── requirements/                     # 需求文档
│   ├── prd.md
│   ├── requirements.md
│   └── requirements_plan.md
├── sql/                              # 数据库变更
│   ├── sql-{序号}-{表名}-ddl.md
│   └── sql-{序号}-{表名}-dml.md
├── task/                             # 任务列表
│   ├── Task-{序号}-{任务名}.md
│   └── tasks.md
└── test/                             # 测试文档
    ├── test-cases.md
    ├── test-report.md
    └── test-strategy.md
```

---

## 需求文档

详见：[requirements/](./requirements/)

---

## 设计文档

详见：[design/design.md](./design/design.md)

---

## 任务列表

详见：[task/tasks.md](./task/tasks.md)

---

## API 变更

详见：[api/](./api/)

---

## 数据库变更

详见：[sql/](./sql/)

---

## 测试文档

详见：[test/](./test/)

---

## 变更日志

详见：[change-log/](./change-log/)
