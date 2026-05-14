# 部署指南 — _[需求名称]_

| 项目         | 内容                                      |
|--------------|-------------------------------------------|
| 文档名称     | _[需求名称]_ 部署指南                     |
| CR 编号      | _[CRxx]_                                  |
| 版本         | v1.0                                      |
| 部署负责人   | _[部署负责人姓名]_                        |
| 审核人       | _[审核人姓名 / DBA 姓名]_                 |
| 创建日期     | _[YYYY-MM-DD]_                            |
| 最后更新     | _[YYYY-MM-DD]_                            |
| 计划部署时间 | _[YYYY-MM-DD HH:mm]_                      |
| 部署环境     | DEV → UAT → PROD                          |

---

## 部署概述

### 本次 CR 部署内容

本次 CR（_[CRxx]_）主要包含以下变更内容：

1. **数据库变更**：新增 `pay_order`（付款申请表），并初始化演示数据。
2. **应用服务变更**：部署 `payment-service` 服务新版本（_[v1.x.x]_），新增付款申请相关接口。
3. **配置变更**：新增付款模块相关配置项，各环境配置值不同（详见环境配置差异表）。

### 影响范围

| 影响项         | 说明                                                                 |
|----------------|----------------------------------------------------------------------|
| 影响服务       | `payment-service`（主服务）、`notify-service`（通知服务，间接影响） |
| 影响接口       | `/api/v1/pay-order/**`（全部为新增接口，不影响现有接口）            |
| 影响数据库     | 新增表 `pay_order`，不修改现有表结构                                 |
| 影响用户       | 财务部门用户、审批人员                                               |
| 下游依赖       | 无下游系统依赖本次新增接口                                           |

### 预计停机时间

| 环境 | 预计停机时间 | 说明                                           |
|------|-------------|------------------------------------------------|
| DEV  | 约 5 分钟   | 滚动重启，开发环境可接受短暂中断               |
| UAT  | 约 10 分钟  | 蓝绿部署，切换前通知测试团队暂停测试           |
| PROD | **约 15 分钟** | 维护窗口：_[YYYY-MM-DD 02:00–04:00]_，需提前公告 |

> **注意**：若部署过程中出现异常，最长回滚时间约 30 分钟，请确保维护窗口留有足够余量。

---

## 环境配置差异表

> 以下配置项需在各环境的配置中心（Nacos / Apollo）或 `application-{env}.yml` 中分别设置。

| 配置项                                  | DEV 值                          | UAT 值                          | PROD 值                              |
|-----------------------------------------|---------------------------------|---------------------------------|--------------------------------------|
| `payment.service.url`                   | `http://dev-payment:8080`       | `http://uat-payment:8080`       | `http://prod-payment:8080`           |
| `payment.approve.timeout-days`          | `3`                             | `3`                             | `7`                                  |
| `payment.notify.email-enabled`          | `false`                         | `true`                          | `true`                               |
| `payment.notify.email-recipients`       | _（不启用）_                    | `uat-finance@example.com`       | `finance@example.com`                |
| `payment.amount.single-limit`           | `999999`                        | `999999`                        | `500000`                             |
| `spring.datasource.url`                 | `jdbc:mysql://dev-db:3306/app`  | `jdbc:mysql://uat-db:3306/app`  | `jdbc:mysql://prod-db:3306/app`      |
| `spring.datasource.username`            | `dev_user`                      | `uat_user`                      | `prod_user`                          |
| `logging.level.com.example.payment`     | `DEBUG`                         | `INFO`                          | `WARN`                               |

> **安全提示**：数据库密码、密钥等敏感配置不得写入本文档，应通过密钥管理系统（Vault / KMS）注入。

---

## SQL 脚本执行顺序

> 所有 SQL 脚本位于 `spec/normalCR/CRxx-示例需求名/sql/` 目录下，执行前请仔细阅读各脚本的注意事项。

| 执行顺序 | 脚本文件                          | 执行环境       | 说明                                                   |
|----------|-----------------------------------|----------------|--------------------------------------------------------|
| 1        | `sql-001-示例表名-ddl.md`         | DEV / UAT / PROD | 创建 `pay_order` 表，使用 `IF NOT EXISTS`，可重复执行 |
| 2        | `sql-001-示例表名-dml.md`         | DEV / UAT       | 初始化演示数据，**仅在 DEV 和 UAT 执行，PROD 不执行** |

> **重要**：DDL 脚本必须在 DML 脚本之前执行，且每个脚本执行完成后需人工确认结果再执行下一个。

---

## 部署步骤

> 执行前请确认：已完成 UAT 验证、已获得上线审批、已通知相关团队。

### 步骤 1：执行 DDL 脚本

**执行目标**：在目标环境创建 `pay_order` 表。

**执行方式**：使用 DBA 账号连接目标数据库，执行 `sql-001-示例表名-ddl.md` 中的 DDL 脚本。

```bash
# 连接目标数据库（以 PROD 为例）
mysql -h prod-db-host -P 3306 -u prod_dba_user -p app_db

# 或使用命令行直接执行脚本文件
mysql -h prod-db-host -P 3306 -u prod_dba_user -p app_db < sql-001-pay_order-ddl.sql
```

**执行后验证**：

```sql
-- 确认表已创建
SHOW TABLES LIKE 'pay_order';

-- 确认表结构正确
DESCRIBE pay_order;

-- 确认索引已创建
SHOW INDEX FROM pay_order;
```

**预期结果**：`SHOW TABLES` 返回 `pay_order`，`SHOW INDEX` 返回 6 条索引记录。

---

### 步骤 2：执行 DML 脚本

> **注意**：DML 脚本仅在 DEV 和 UAT 环境执行，PROD 环境跳过此步骤。

**执行目标**：向 `pay_order` 表插入初始化演示数据（共 3 条）。

```bash
# 连接 UAT 数据库并执行 DML 脚本
mysql -h uat-db-host -P 3306 -u uat_user -p app_db < sql-001-pay_order-dml.sql
```

**执行后验证**：

```sql
-- 确认数据已插入（应返回 3 行）
SELECT COUNT(*) AS total FROM pay_order;

-- 查看插入的数据
SELECT order_no, status, amount FROM pay_order ORDER BY id;
```

**预期结果**：`COUNT(*)` 返回 `3`，三条记录的 `order_no` 分别为 `PAY20240101000001`、`PAY20240101000002`、`PAY20240101000003`。

---

### 步骤 3：部署应用服务

**执行目标**：将 `payment-service` 升级至新版本 _[v1.x.x]_。

#### 3.1 拉取新版本镜像

```bash
# 登录镜像仓库
docker login registry.example.com -u deploy_user -p ${REGISTRY_PASSWORD}

# 拉取新版本镜像
docker pull registry.example.com/payment-service:v1.x.x

# 确认镜像拉取成功
docker images | grep payment-service
```

#### 3.2 更新配置

```bash
# 确认 Nacos 配置中心已更新对应环境的配置项
# 登录 Nacos 控制台：http://nacos-host:8848/nacos
# 检查命名空间：prod
# 检查 Data ID：payment-service.yaml
# 确认以下配置项已正确设置：
#   payment.service.url
#   payment.approve.timeout-days
#   payment.notify.email-enabled
```

#### 3.3 滚动重启服务（Kubernetes 环境）

```bash
# 更新 Deployment 镜像版本
kubectl set image deployment/payment-service \
  payment-service=registry.example.com/payment-service:v1.x.x \
  -n production

# 监控滚动更新进度
kubectl rollout status deployment/payment-service -n production

# 确认所有 Pod 已更新
kubectl get pods -n production -l app=payment-service
```

#### 3.4 滚动重启服务（传统部署环境）

```bash
# 备份当前版本 JAR 包
cp /app/payment-service/payment-service.jar \
   /app/payment-service/payment-service.jar.bak_$(date +%Y%m%d%H%M%S)

# 上传新版本 JAR 包
scp payment-service-v1.x.x.jar deploy@prod-server:/app/payment-service/payment-service.jar

# 重启服务
systemctl restart payment-service

# 查看启动日志
journalctl -u payment-service -f --since "1 minute ago"
```

---

### 步骤 4：验证部署结果

#### 4.1 健康检查

```bash
# 检查服务健康状态（Spring Boot Actuator）
curl -s http://prod-payment:8080/actuator/health | python3 -m json.tool

# 预期响应：
# {
#   "status": "UP",
#   "components": {
#     "db": { "status": "UP" },
#     "diskSpace": { "status": "UP" }
#   }
# }
```

```bash
# 检查服务版本信息
curl -s http://prod-payment:8080/actuator/info | python3 -m json.tool

# 预期响应中包含：
# {
#   "app": {
#     "version": "v1.x.x",
#     "name": "payment-service"
#   }
# }
```

#### 4.2 冒烟测试

```bash
# 获取认证 Token（替换为实际测试账号）
TOKEN=$(curl -s -X POST http://prod-gateway:9000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"smoke_test_user","password":"${SMOKE_TEST_PASSWORD}"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

# 冒烟测试 1：创建付款申请
curl -s -X POST http://prod-gateway:9000/api/v1/pay-order \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00,
    "currency": "CNY",
    "payType": 0,
    "remark": "冒烟测试-部署验证"
  }' | python3 -m json.tool

# 预期响应：HTTP 200，body 中 code=200，data 包含 orderNo 字段

# 冒烟测试 2：查询付款申请列表
curl -s -X GET "http://prod-gateway:9000/api/v1/pay-order?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer ${TOKEN}" | python3 -m json.tool

# 预期响应：HTTP 200，body 中 code=200，data.total >= 1
```

#### 4.3 日志检查

```bash
# 检查服务启动日志，确认无 ERROR 级别日志
kubectl logs -n production -l app=payment-service --since=5m | grep -E "ERROR|WARN"

# 检查数据库连接是否正常
kubectl logs -n production -l app=payment-service --since=5m | grep "HikariPool"
```

---

## 回滚方案

> 当部署出现以下情况时，立即启动回滚：健康检查失败、冒烟测试失败、核心业务报错率超过 1%。

### 回滚决策标准

| 异常情况                         | 处理方式         | 负责人         |
|----------------------------------|------------------|----------------|
| 服务启动失败                     | 立即回滚应用     | 部署负责人     |
| 健康检查持续失败超过 5 分钟      | 立即回滚应用     | 部署负责人     |
| 冒烟测试失败                     | 评估后决定是否回滚 | 部署负责人 + 开发 |
| 数据库脚本执行失败               | 执行 SQL 回滚脚本 | DBA            |
| 业务错误率超过 1%（监控告警）    | 立即回滚应用     | 部署负责人     |

### 应用服务回滚

```bash
# Kubernetes 环境：回滚到上一个版本
kubectl rollout undo deployment/payment-service -n production

# 确认回滚完成
kubectl rollout status deployment/payment-service -n production

# 验证回滚后版本
kubectl get pods -n production -l app=payment-service -o jsonpath='{.items[0].spec.containers[0].image}'
```

```bash
# 传统部署环境：恢复备份的 JAR 包
cp /app/payment-service/payment-service.jar.bak_YYYYMMDDHHMMSS \
   /app/payment-service/payment-service.jar

# 重启服务
systemctl restart payment-service

# 确认服务恢复
curl -s http://prod-payment:8080/actuator/health
```

### SQL 回滚

```bash
# 若 DML 脚本执行失败，执行 DML 回滚脚本
mysql -h prod-db-host -P 3306 -u prod_dba_user -p app_db <<'EOF'
DELETE FROM `pay_order`
WHERE `order_no` IN (
  'PAY20240101000001',
  'PAY20240101000002',
  'PAY20240101000003'
);
EOF

# 若 DDL 脚本执行失败（表结构异常），执行 DDL 回滚脚本
# 警告：此操作将删除 pay_order 表及其所有数据，请谨慎执行！
mysql -h prod-db-host -P 3306 -u prod_dba_user -p app_db <<'EOF'
DROP TABLE IF EXISTS `pay_order`;
EOF
```

### 回滚后验证

```bash
# 确认服务健康
curl -s http://prod-payment:8080/actuator/health

# 确认回滚版本正确
curl -s http://prod-payment:8080/actuator/info

# 通知相关团队回滚已完成，并记录回滚原因
```

---

## 发布检查清单

### 发布前检查

- [ ] UAT 环境所有测试用例已通过，测试报告已归档（`test/test-report.md`）
- [ ] DDL 脚本已在 UAT 环境验证执行成功
- [ ] DML 脚本已在 UAT 环境验证执行成功
- [ ] 代码已通过 Code Review，PR 已合并至主干分支
- [ ] 新版本镜像已构建并推送至镜像仓库（`registry.example.com/payment-service:v1.x.x`）
- [ ] 各环境配置项已在配置中心（Nacos）更新完毕
- [ ] 已获得项目负责人上线审批（审批单号：_[审批单号]_）
- [ ] 已获得 DBA 对 SQL 脚本的审核确认
- [ ] 已通知财务部门和测试团队本次上线计划及维护窗口
- [ ] 已确认维护窗口期间无其他并行上线计划
- [ ] 回滚方案已评审，回滚所需资源（备份 JAR、回滚脚本）已就绪
- [ ] 监控告警已配置（错误率、响应时间、健康检查）

### 发布后检查

- [ ] 服务健康检查接口返回 `{"status":"UP"}`
- [ ] 服务版本信息确认为新版本 `v1.x.x`
- [ ] 冒烟测试全部通过（创建付款申请、查询列表）
- [ ] 数据库表 `pay_order` 结构与 DDL 脚本一致
- [ ] 监控大盘无异常告警（错误率、响应时间 P95 在正常范围内）
- [ ] 应用日志无 `ERROR` 级别异常输出
- [ ] 已通知财务部门和测试团队上线完成，可开始验收
- [ ] 已在部署记录中填写本次部署结果（执行时间、执行人、结果）
- [ ] 若有异常已记录并创建跟进 Issue

---

## 部署执行记录

> 每次部署完成后，请填写以下记录并提交至版本库。

| 环境 | 计划时间             | 实际开始时间         | 实际结束时间         | 执行人       | 结果         | 备注                   |
|------|----------------------|----------------------|----------------------|--------------|--------------|------------------------|
| DEV  | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[姓名]_     | _[成功/失败]_ | _[备注信息]_           |
| UAT  | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[姓名]_     | _[成功/失败]_ | _[备注信息]_           |
| PROD | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[YYYY-MM-DD HH:mm]_ | _[姓名]_     | _[成功/失败]_ | _[备注信息]_           |
