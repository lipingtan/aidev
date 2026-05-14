# 域规范目录说明

本目录用于存放 SPMP 智慧物业管理平台后端开发相关的规范文档。

## 目录结构

```
domain/
├── README.md                  # 本文件
├── domain-index.md            # 域索引（快速定位）
├── naming-convention.md       # 命名规范
├── templates/                 # 标准模板
├── user-center/               # 用户中心域
├── owner-center/              # 业主中心域
├── workorder-center/          # 工单中心域
├── billing-center/            # 缴费中心域
├── notice-center/             # 公告中心域
├── access-center/             # 门禁中心域
└── base-center/               # 基础中心域
```

## 各域目录结构

每个业务域目录结构一致：

```
{domain}-center/
├── domain-share/              # 域共享文档（API、数据库、流程）
├── spec/                      # 需求规范
│   ├── normalCR/              # 常规 CR
│   └── minorCR/               # 小型 CR
├── change-log/                # 变更日志
├── code-review/               # 代码审查
├── vibe/                      # Vibe 开发
└── README.md                  # 域文档索引
```

## 使用指南

1. 确定需求归属的域
2. 在对应域的 `spec/` 目录下创建需求目录
3. 参考 `templates/` 中的模板编写文档
