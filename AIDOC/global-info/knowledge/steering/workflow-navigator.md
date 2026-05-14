# 工作流导航图

> 三套规范体系的快速入门指南。根据你要做的事情，找到正确的起点和阅读顺序。

---

## 一、我要做什么？→ 从哪里开始

| 我要做的事 | 规范体系 | 起点文件 |
|-----------|----------|----------|
| 从零开始一个新短剧系列 | 短剧 | `video/series-bootstrap-workflow.md` |
| 为已有系列制作新一集 | 短剧 | `video/shot-production-workflow.md` |
| 优化已生成的视频效果 | 短剧 | `video/iteration-protocol.md` |
| 从零开始一个新游戏项目 | 游戏 | `godot/project-bootstrap.md`（先读 `pre-development-defaults.md`） |
| 为游戏开发一个新功能 | 游戏 | `godot/feature-development-flow.md` |
| 修复游戏 Bug / 优化性能 | 游戏 | `godot/iteration-workflow.md` |
| 从零开始一个新 Go 后端项目 | Go | `go/go-project-bootstrap.md` |
| 为 Go 项目开发新功能 | Go | `go/go-development-workflow.md` |
| 修复 Go 项目 Bug | Go | `go/go-development-workflow.md`（第5节 Bugfix） |
| 部署 Go 项目 | Go | `go/go-deployment.md` |

---

## 二、短剧规范 · 阅读顺序

### 首次接触（理解全貌）

```
1. product.md                    ← 这是什么工作台
2. project-overview.md           ← 目录结构概览
3. structure.md                  ← 详细目录规范
4. series-bootstrap-workflow.md  ← 完整制作流程（Phase 0~5）
```

### 开始制作（按需读取）

```
series-bootstrap-workflow.md     ← 系列启动（Phase 0~5）
    ├── narrative-quality-review.md  ← Phase 3 后的质量门控
    ├── climax-design-philosophy.md  ← 高潮章设计
    └── unpredictability-design.md   ← 不可预测性设计

shot-production-workflow.md      ← 逐镜头制作（bootstrap 完成后）
    ├── cinematography.md            ← 镜头语言 + 场景→镜头决策表
    ├── pacing.md                    ← 节奏控制
    ├── prompt-engineering.md        ← 提示词编写 + 组装流程
    ├── character-consistency.md     ← 角色一致性
    ├── negative-prompts.md          ← 负面提示词库
    └── style-keywords.md            ← 风格关键词库

iteration-protocol.md            ← 迭代优化（视频不满意时）
```

### 规范间关系图

```
product.md ──→ project-overview.md ──→ structure.md
                                           │
series-bootstrap-workflow.md ◄─────────────┘
    │ Phase 0~5 完成
    ▼
shot-production-workflow.md
    │ 生成视频
    ▼
iteration-protocol.md
    │ 发现引擎限制
    ▼
engines/kling/limitations.md
engines/wan/limitations.md
```

---

## 三、游戏开发规范 · 阅读顺序

### 首次接触（理解全貌）

```
1. pre-development-defaults.md   ← 写代码前必读（画质/性能定稿）
2. development-workflow.md       ← 双轨并行工作流（Phase 0~5）
3. execution-protocol.md         ← PRE/POST-CHECK 协议
4. feature-development-flow.md   ← 单功能开发 7 步流程
```

### 开始开发（按需读取）

```
project-bootstrap.md             ← 新项目启动（Phase 0）
    ├── game-type-blueprints.md      ← 功能清单蓝图
    └── performance-budget.md        ← 性能预算

development-workflow.md          ← Phase 0~5 全流程
    ├── execution-protocol.md        ← 每步的 PRE/POST-CHECK
    ├── task-representation.md       ← 任务三要素格式
    └── code-generation.md           ← 代码生成规范

feature-development-flow.md      ← 单功能开发
    ├── experience-benchmarks.md     ← 手感参数参考
    ├── templates/rpg/*.md           ← RPG 系统模板
    ├── templates/action/*.md        ← 动作系统模板
    └── patterns/*.md                ← 设计模式

iteration-workflow.md            ← Bug/性能/调优
asset-pipeline.md                ← 资产导入管线
branch-routing.md                ← 2D/3D 分支路由
```

### 规范间关系图

```
pre-development-defaults.md
    │ 定稿画质/性能
    ▼
project-bootstrap.md ──→ development-workflow.md
    │ Phase 0              │ Phase 1~5
    │                      │
    │                      ├── feature-development-flow.md（Phase 2~4 每个功能）
    │                      │       └── task-representation.md
    │                      │       └── code-generation.md
    │                      │       └── execution-protocol.md
    │                      │
    │                      ├── asset-pipeline.md（资产集成）
    │                      │
    │                      └── iteration-workflow.md（Bug/优化/调优）
    │
    └── game-type-blueprints.md（功能清单来源）
```

---

## 四、Go 软件规范 · 阅读顺序

### 首次接触（理解全貌）

```
1. go-project-bootstrap.md      ← 新项目启动
2. go-development-workflow.md   ← 5 阶段开发流程
3. go-conventions.md            ← 编码规范
4. go-project-structure.md      ← 目录结构
```

### 开始开发（按需读取）

```
go-project-bootstrap.md          ← 新项目启动
    └── go-project-structure.md      ← 目录结构

go-development-workflow.md       ← 需求→设计→任务→执行→Bugfix
    ├── go-task-representation.md    ← 任务三要素
    ├── go-regression-guard.md       ← 回归防护
    └── go-api-design.md             ← API 设计规范

go-conventions.md                ← 编码规范（随时参考）
go-security.md                   ← 安全编码（随时参考）
go-testing.md                    ← 测试规范
go-code-review.md                ← 代码审查
go-frontend-integration.md      ← 前后端联调
go-deployment.md                 ← 部署发布
go-git-workflow.md               ← Git 工作流
go-debugging.md                  ← 排查方法论（遇到问题时）
```

### 规范间关系图

```
go-project-bootstrap.md
    │ 项目初始化
    ▼
go-development-workflow.md
    │
    ├── Requirement 阶段
    │       └── go-api-design.md（接口设计）
    │
    ├── Design 阶段
    │       ├── go-project-structure.md（架构）
    │       └── go-regression-guard.md（回归防护）
    │
    ├── Task 阶段
    │       └── go-task-representation.md（三要素）
    │
    ├── 执行阶段
    │       ├── go-conventions.md（编码规范）
    │       ├── go-security.md（安全）
    │       └── go-testing.md（测试）
    │
    ├── 联调阶段
    │       └── go-frontend-integration.md
    │
    ├── 审查阶段
    │       └── go-code-review.md
    │
    ├── 发布阶段
    │       ├── go-deployment.md
    │       └── go-git-workflow.md
    │
    └── Bugfix
            └── go-debugging.md
```

---

## 五、跨体系共用规范

| 规范 | 适用体系 | 位置 |
|------|----------|------|
| core.md（知识库索引） | 全部 | `.kiro/steering/core.md` |
| 目录分离原则 | 全部 | core.md 中定义 |
| 文件写入规则 | 全部 | core.md 中定义 |
| 语言规范（中文沟通/英文代码） | 全部 | core.md 中定义 |
