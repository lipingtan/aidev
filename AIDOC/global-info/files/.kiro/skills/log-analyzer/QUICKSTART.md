# 日志分析器 - 快速开始指南

## 📚 文档结构

```
log-analyzer/
├── SKILL.md              # 完整技能文档（中文）
├── README.md             # 技能概述
├── examples.md           # 实战案例集（5个详细案例）
├── QUICKSTART.md         # 本文件 - 快速开始指南
└── samples/              # 日志样本目录
    ├── README.md         # 样本说明文档
    ├── case1-db-pool-exhaustion.log
    ├── case3-memory-leak.log
    └── case4-api-performance.log
```

---

## 🚀 5分钟快速上手

### 步骤1：分析第一个日志

```bash
# 使用提供的样本
/log-analyzer samples/case1-db-pool-exhaustion.log
```

### 步骤2：查看分析报告

你会得到类似这样的报告：

```markdown
## 根因分析报告

### 执行摘要
**分析时段：** 2025-01-29 10:00:00 - 10:30:00 UTC
**置信度：** 高（95%）

### 根因
**数据库连接池耗尽**（置信度：95%）

### 证据链
- 10:00:12 - 首次连接等待警告
- 10:00:23 - 连接池耗尽
- 10:00:30 - 熔断器打开

### 建议
- 立即：增加连接池大小到100
- 短期：添加监控告警
```

### 步骤3：尝试更多功能

```bash
# 分析特定时间范围
/log-analyzer samples/case1-db-pool-exhaustion.log \
  --since "2025-01-29 10:00" \
  --until "2025-01-29 10:30"

# 聚焦性能问题
/log-analyzer samples/case3-memory-leak.log --focus performance

# 导出报告
/log-analyzer samples/case4-api-performance.log \
  --output performance-report.md
```

---

## 📖 学习路径

### 初级（新手）

1. **阅读README.md** - 了解基本概念
2. **运行第一个分析** - 使用samples中的样本
3. **查看输出报告** - 理解报告结构

**时间：** 15分钟

### 中级（日常使用）

1. **阅读SKILL.md** - 学习所有功能
2. **查看examples.md** - 学习5个实战案例
3. **分析自己的日志** - 开始实际使用

**时间：** 1小时

### 高级（深度定制）

1. **研究样本日志** - samples/README.md
2. **创建自定义规则** - 针对你的应用
3. **集成到CI/CD** - 自动化分析

**时间：** 4小时

---

## 🎯 常见使用场景

### 场景1：生产事故调查

```bash
# 分析事故时段
/log-analyzer /var/log/app.log \
  --since "2025-01-29 14:00" \
  --until "2025-01-29 15:00"

# 多服务关联分析
/log-analyzer api.log service.log db.log \
  --since "2025-01-29 14:00" \
  --correlate-with
```

### 场景2：性能优化

```bash
# 对比部署前后
/log-analyzer app.log \
  --baseline "2025-01-29 10:00:00 to 2025-01-29 11:00:00" \
  --since "2025-01-29 14:00:00" \
  --focus performance

# 导出性能报告
/log-analyzer app.log \
  --focus performance \
  --output performance-analysis.md
```

### 场景3：定期监控

```bash
# 每日检查
/log-analyzer /var/log/app-$(date +%Y%m%d).log \
  --level ERROR,WARN \
  --output daily-report-$(date +%Y%m%d).md

# 设置cron任务
0 8 * * * /log-analyzer /var/log/app.log --level ERROR --output /reports/daily-$(date +\%Y\%m\%d).md
```

---

## 💡 核心功能速查

### 命令选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `--since` | 起始时间 | `--since "2025-01-29 10:00"` |
| `--until` | 结束时间 | `--until "2025-01-29 11:00"` |
| `--level` | 日志级别过滤 | `--level ERROR` |
| `--focus` | 分析维度 | `--focus performance` |
| `--filter` | 关键词过滤 | `--filter "database"` |
| `--baseline` | 基线对比 | `--baseline normal.log` |
| `--output` | 输出文件 | `--output report.md` |

### 分析维度

| 维度 | 检测内容 | 选项 |
|------|----------|------|
| 异常分析 | 错误、异常、堆栈跟踪 | `--focus exceptions` |
| 性能分析 | 响应时间、慢查询 | `--focus performance` |
| 依赖分析 | 外部服务、级联故障 | `--focus dependencies` |
| 业务逻辑 | 工作流、数据流 | `--focus business` |

---

## 🔍 实战案例快速索引

### 案例1：数据库连接池耗尽
- **文件：** `examples.md` 第1节
- **样本：** `samples/case1-db-pool-exhaustion.log`
- **主题：** 资源配置、容量规划

### 案例2：微服务级联故障
- **文件：** `examples.md` 第2节
- **主题：** 服务依赖、熔断器、容错

### 案例3：内存泄漏导致OOM
- **文件：** `examples.md` 第3节
- **样本：** `samples/case3-memory-leak.log`
- **主题：** 内存管理、GC监控、堆转储分析

### 案例4：API性能突然下降
- **文件：** `examples.md` 第4节
- **样本：** `samples/case4-api-performance.log`
- **主题：** 性能回归、N+1查询、缓存优化

### 案例5：缓存穿透
- **文件：** `examples.md` 第5节
- **主题：** 安全防护、布隆过滤器、限流

---

## 📊 输出报告结构

### 标准报告模板

```markdown
## 根因分析报告

### 1. 执行摘要
- 分析时段
- 日志来源
- 置信度

### 2. 问题描述
- 一句话总结

### 3. 根因
- 最可能的原因
- 置信度

### 4. 详细发现
- 主要根因
- 贡献因素
- 级联效应

### 5. 时间线
- 事件序列表

### 6. 建议
- 立即行动（今天）
- 短期改进（本周）
- 长期优化（本季度）
```

---

## 🛠️ 故障排除

### 问题1："无法检测日志格式"

**解决方法：**
```bash
# 手动指定格式
/log-analyzer app.log --format json
/log-analyzer app.log --format text
/log-analyzer app.log --format spring
```

### 问题2："未检测到异常"

**可能原因：**
- 时间范围不包含事故时段
- 日志级别过滤太严格
- 日志格式不兼容

**解决方法：**
```bash
# 扩大时间范围
/log-analyzer app.log --since "2025-01-29 00:00"

# 降低日志级别过滤
/log-analyzer app.log --level INFO,WARN,ERROR

# 检查日志格式
head -20 app.log
```

### 问题3："分析太慢"

**优化方法：**
```bash
# 缩小时间范围
/log-analyzer app.log --since "2025-01-29 14:00" --until "2025-01-29 14:30"

# 过滤关键词
/log-analyzer app.log --filter "ERROR,Exception"

# 只分析ERROR日志
/log-analyzer app.log --level ERROR
```

---

## 📈 最佳实践

### ✅ DO（推荐做法）

1. **提供充足上下文**
   - 包含事故前30分钟和事故后15分钟
   - 更多数据 = 更准确的分析

2. **包含所有相关服务**
   - API网关、应用服务器、数据库
   - 消息队列、缓存

3. **保留所有日志级别**
   - DEBUG日志提供上下文
   - INFO日志显示正常流程

4. **使用关联ID**
   - 启用请求跟踪
   - 跨服务追踪问题

### ❌ DON'T（避免做法）

1. ❌ 只分析ERROR日志（会丢失上下文）
2. ❌ 忽略时间戳（无法关联事件）
3. ❌ 过早过滤DEBUG日志
4. ❌ 孤立分析单个服务

---

## 🔗 相关资源

### 内部资源
- `SKILL.md` - 完整技能文档
- `examples.md` - 5个详细实战案例
- `samples/README.md` - 日志样本说明

### 外部资源
- [日志格式最佳实践](https://www.example.com/logging)
- [根因分析方法论](https://www.example.com/rca)
- [性能优化指南](https://www.example.com/performance)

### 相关技能
- `/pdf` - 从PDF提取日志
- `/project-analyze` - 理解代码库
- `/docx` - 生成Incident报告

---

## 📞 获取帮助

### 内置帮助

```bash
# 查看完整帮助
/log-analyzer --help

# 查看特定选项
/log-analyzer --help --output
```

### 社区支持
- GitHub Issues: https://github.com/example/log-analyzer/issues
- Stack Overflow: https://stackoverflow.com/questions/tagged/log-analyzer
- 邮件列表: log-analyzer@example.com

---

## 🎓 认证

完成以下任务以认证你的技能：

### 初级认证
- [ ] 成功分析至少3个样本日志
- [ ] 理解报告的所有部分
- [ ] 能够解释根因和证据

### 中级认证
- [ ] 分析自己的生产日志
- [ ] 识别并解决一个实际问题
- [ ] 创建自定义分析规则

### 高级认证
- [ ] 贡献一个新的日志样本
- [ ] 优化技能性能
- [ ] 帮助他人解决问题

---

## 📝 更新日志

### v1.0.0 (2025-01-29)
- ✨ 初始版本
- ✅ 支持4种分析维度
- ✅ 5个实战案例
- ✅ 3个日志样本
- 📖 完整中文文档

---

**准备好开始了吗？**

```bash
# 分析你的第一个日志
/log-analyzer samples/case1-db-pool-exhaustion.log

# 查看更多案例
cat examples.md

# 开始使用自己的日志
/log-analyzer /path/to/your/application.log
```

**祝你分析愉快！🎉**
