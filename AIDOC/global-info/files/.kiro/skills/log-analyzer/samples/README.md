# 日志样本说明

本目录包含用于测试和演示日志分析器的真实日志样本。

## 可用样本

### 1. 数据库连接池耗尽 (case1-db-pool-exhaustion.log)

**场景：** 高峰期数据库连接池耗尽导致大规模请求失败

**关键指标：**
- 连接池配置：最大50个连接
- 峰值需求：177个连接
- 错误率：85%
- 平均响应时间：从200ms恶化至30秒

**测试命令：**
```bash
/log-analyzer case1-db-pool-exhaustion.log
```

**预期发现：**
- ⚠️ 连接池使用率持续上升
- 🔴 10:00:23 - 连接池耗尽
- 🔴 10:00:30 - 熔断器打开
- 📊 恢复时间：15分钟

---

### 2. 微服务级联故障 (未提供完整样本)

**场景：** Redis故障导致商品服务不可用，级联影响订单、用户服务

**关键指标：**
- 影响：4个服务（商品、订单、用户、网关）
- 故障时间：5分钟
- 根因：Redis内存溢出

**示例日志结构：**
```
Redis OOM → 商品服务故障 → 订单服务失败 → 用户服务失败 → 系统不可用
```

---

### 3. 内存泄漏导致OOM (case3-memory-leak.log)

**场景：** Java应用运行4小时后因内存泄漏崩溃

**关键指标：**
- 内存增长率：~124MB/小时
- 崩溃时间：11:30:03
- GC频率：从125次/小时增至1111次/小时
- 主要泄漏源：ReportCache (450MB) + SessionManager (250MB)

**测试命令：**
```bash
/log-analyzer case3-memory-leak.log --focus performance
```

**预期发现：**
- 📈 堆使用率线性增长：50% → 95%
- 🔴 GC频率恶化：125 → 1111次/小时
- 🔴 GC时间恶化：2.3s → 89.2s/小时
- 💥 11:30:03 - OutOfMemoryError崩溃

---

### 4. API性能突然下降 (case4-api-performance.log)

**场景：** 部署v2.4.0后支付API响应时间从150ms恶化至2500ms

**关键指标：**
- 响应时间恶化：+1616%（慢了16倍）
- 部署版本：v2.4.0（15:58部署）
- 根因：新增的欺诈检测功能每次都查询数据库

**测试命令：**
```bash
/log-analyzer case4-api-performance.log \
  --since "2025-01-29 15:50" \
  --until "2025-01-29 16:10" \
  --focus performance
```

**预期发现：**
- ⏱️ 部署前平均响应：112-128ms
- ⏱️ 部署后平均响应：2567ms
- 🔴 新增数据库查询：每次1111ms
- 📊 性能恶化：+2194%

---

### 5. 缓存穿透 (未提供样本)

**场景：** 恶意攻击使用大量无效ID穿透缓存

**关键指标：**
- 数据库QPS：从1000增至10000
- 缓存命中率：0%
- 所有查询都是无效ID

**示例日志结构：**
```log
Cache miss: key=product:invalid-id
Querying database: id=invalid-id
Product not found: invalid-id
...（重复数千次）
```

---

## 使用指南

### 基本测试

```bash
# 分析单个日志文件
/log-analyzer samples/case1-db-pool-exhaustion.log

# 指定时间范围
/log-analyzer samples/case3-memory-leak.log \
  --since "2025-01-29 08:00" \
  --until "2025-01-29 11:30"

# 聚焦特定维度
/log-analyzer samples/case4-api-performance.log --focus performance
```

### 对比分析

```bash
# 对比部署前后的性能
/log-analyzer samples/case4-api-performance.log \
  --baseline "2025-01-29 15:50:00 to 2025-01-29 15:55:00" \
  --since "2025-01-29 16:00:00" \
  --until "2025-01-29 16:10:00"
```

### 多维度分析

```bash
# 同时分析异常、性能、依赖
/log-analyzer samples/case1-db-pool-exhaustion.log \
  --focus exceptions,performance,dependencies
```

### 导出报告

```bash
# 生成Markdown报告
/log-analyzer samples/case3-memory-leak.log \
  --output memory-leak-report.md

# 导出为JSON
/log-analyzer samples/case4-api-performance.log \
  --output performance-analysis.json \
  --format json
```

---

## 预期输出示例

### 执行摘要

```markdown
## 根因分析报告

### 执行摘要

**分析时段：** 2025-01-29 10:00:00 - 10:30:00 UTC
**日志条目：** 8,234条
**置信度：** 高（95%）

### 问题陈述
生产环境在10:00-10:15期间出现大规模请求失败，错误率达到85%，API平均响应时间从200ms恶化至30秒。

### 根因
**数据库连接池耗尽**（置信度：95%）
```

### 详细发现

```markdown
#### 1. 主要根因：连接池配置不足

**置信度：** 高（95%）

**证据链：**
- 10:00:12.890 - 首次连接等待警告（2500ms）
- 10:00:15.123 - 首次连接超时错误
- 10:00:18.456 - 连接池使用率85%（43/50）
- 10:00:23.789 - 连接池耗尽

**支持指标：**
- 连接池容量：50（最大值）
- 峰值需求：177个连接
- 缺口：127个连接（255%超负荷）
```

---

## 自定义日志样本

如果你想创建自己的测试样本：

### 1. 基本格式

```log
TIMESTAMP LEVEL [THREAD] CLASS_NAME : MESSAGE
EXCEPTION_STACK_TRACE (如果适用)
```

### 2. 包含关键元素

- ✅ 时间戳（精确到毫秒）
- ✅ 日志级别（INFO、WARN、ERROR）
- ✅ 线程信息
- ✅ 类/组件名称
- ✅ 上下文信息（请求ID、用户ID等）
- ✅ 异常堆栈跟踪（如有错误）

### 3. 模拟真实场景

```log
# 正常开始
2025-01-29 10:00:00.123 INFO  [main] Application : Starting

# 警告出现
2025-01-29 10:00:15.234 WARN  [thread-1] Service : Slow operation

# 错误发生
2025-01-29 10:00:30.456 ERROR [thread-2] Controller : Request failed
java.lang.Exception: Something went wrong
  at com.example.Controller.process(Controller.java:123)

# 恢复正常
2025-01-29 10:05:00.789 INFO  [thread-3] Application : Recovered
```

---

## 贡献新样本

欢迎贡献更多真实场景的日志样本！

### 样本命名规范

```
case{number}-{short-description}.log

示例：
case1-db-pool-exhaustion.log
case2-microservice-cascade.log
case3-memory-leak.log
case4-api-performance.log
```

### 包含内容

每个样本应该包含：
1. ✅ 清晰的场景描述
2. ✅ 真实的日志格式
3. ✅ 完整的事件序列（正常 → 异常 → 恢复）
4. ✅ 明确的根因线索
5. ✅ 元数据文件（README.md）说明

---

## 故障排除

### 问题：日志格式无法识别

**解决方案：**
```bash
# 手动指定格式
/log-analyzer sample.log --format text
```

### 问题：时间范围不正确

**解决方案：**
```bash
# 检查日志时间戳格式
head -10 sample.log

# 使用正确的时区
/log-analyzer sample.log --timezone "Asia/Shanghai"
```

### 问题：分析结果不明显

**可能原因：**
- 样本太短（建议至少100行）
- 缺少异常事件
- 时间范围选择不当

---

## 联系方式

如有问题或建议，请：
- 提交Issue到项目仓库
- 发送邮件到：support@example.com
- 查看在线文档：https://docs.example.com/log-analyzer
