# 后端技术栈组件说明

| 字段 | 内容 |
|------|------|
| 文档名称 | 后端技术栈组件说明 |
| 版本 | 2.0 |
| 创建日期 | 2026-04-17 |
| 负责人 | 技术团队 |
| 审核人 | 技术负责人 |

---

## 架构说明

本项目采用单体应用 + DDD 模块化架构，一个 Spring Boot 应用包含所有业务模块。

---

## 构建系统

| 属性 | 值 |
|------|----|
| 构建工具 | Maven 3.x |
| Java 版本 | JDK 1.8 |
| 编译目标版本 | Java 1.8 |
| 源码编码 | UTF-8 |
| 项目打包方式 | JAR（Spring Boot 可执行 JAR） |

---

## 核心框架和库

### 基础框架

| 依赖名称 | GroupId | 版本 | 用途 |
|----------|---------|------|------|
| Spring Boot | `org.springframework.boot` | `2.4.4` | 核心框架 |

### 数据访问层

| 依赖名称 | GroupId | 版本 | 用途 |
|----------|---------|------|------|
| MyBatis Plus | `com.baomidou` | `3.4.x` | ORM 框架 |
| MySQL Connector | `com.mysql` | `8.0.x` | MySQL JDBC 驱动 |
| Druid | `com.alibaba` | `1.2.x` | 数据库连接池 |
| Redis（Lettuce） | `org.springframework.boot` | 随 Spring Boot | Redis 客户端 |

### 安全框架

| 依赖名称 | GroupId | 版本 | 用途 |
|----------|---------|------|------|
| Spring Security | `org.springframework.boot` | 随 Spring Boot | 认证授权 |
| JWT | `io.jsonwebtoken` | `0.11.x` | Token 生成与验证 |

### 工具库

| 依赖名称 | GroupId | 版本 | 用途 |
|----------|---------|------|------|
| Lombok | `org.projectlombok` | `1.18.x` | 减少样板代码 |
| Hutool | `cn.hutool` | `5.8.x` | Java 工具类库 |
| Jackson | `com.fasterxml.jackson.core` | 随 Spring Boot | JSON 序列化/反序列化 |
| MapStruct | `org.mapstruct` | `1.5.x` | 对象映射 |

### 测试框架

| 依赖名称 | GroupId | 版本 | 用途 |
|----------|---------|------|------|
| JUnit 5 | `org.junit.jupiter` | 随 Spring Boot | 单元测试框架 |
| Mockito | `org.mockito` | 随 Spring Boot | Mock 框架 |
| Spring Boot Test | `org.springframework.boot` | 随 Spring Boot | 集成测试支持 |

---

## 数据库信息

| 属性 | 值 |
|------|------|
| 数据库类型 | MySQL |
| 数据库版本 | 8.0 |
| 数据库名 | spmp |
| 是否分库分表 | 否（共用一个 schema，按表前缀区分模块） |
| Redis 版本 | 7.2.1 |

---

## 常用命令

```bash
# 清理并编译（跳过测试）
mvn clean install -DskipTests

# 清理并编译（执行测试）
mvn clean install

# 运行单元测试
mvn test

# 运行指定测试类
mvn test -Dtest=WorkOrderServiceTest

# 本地运行
mvn spring-boot:run -Dspring-boot.run.profiles=local
```

---

## 部署配置

| 属性 | 值 |
|------|------|
| 部署方式 | 单体 JAR 部署 |
| 容器化 | Docker（可选） |
| CI/CD 工具 | Jenkins |
| 环境划分 | dev / test / uat / prod |
