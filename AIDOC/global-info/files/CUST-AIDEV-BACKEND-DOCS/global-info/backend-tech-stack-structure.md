# 后端技术栈目录结构说明

| 字段 | 内容 |
|------|------|
| 文档名称 | 后端技术栈目录结构说明 |
| 版本 | 2.0 |
| 创建日期 | 2026-04-17 |
| 负责人 | 技术团队 |
| 审核人 | 技术负责人 |

---

## 架构风格

采用单体应用 + DDD 模块化架构。后端为一个 Spring Boot 应用，按业务域划分为独立模块（package），模块间通过 `api/` 包暴露接口进行松耦合通信，便于后续按需拆分为微服务。

---

## 顶层目录结构

```
src/
├── spmp-backend/           # 后端单体应用（Spring Boot）
├── spmp-web-pc/            # PC 端前端（Vue 3 + Element Plus）
└── spmp-web-h5/            # H5 端前端（Vue 3 + Vant）
```

---

## 后端模块结构

```
spmp-backend/
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── com/
│   │   │       └── spmp/
│   │   │           ├── SpmpApplication.java       # 启动类
│   │   │           ├── user/                      # 用户与权限模块
│   │   │           ├── owner/                     # 业主管理模块
│   │   │           ├── workorder/                 # 工单管理模块
│   │   │           ├── billing/                   # 缴费管理模块
│   │   │           ├── notice/                    # 公告管理模块
│   │   │           ├── access/                    # 门禁管理模块
│   │   │           ├── base/                      # 基础数据模块
│   │   │           └── common/                    # 公共组件
│   │   └── resources/
│   │       ├── application.yml                    # 主配置文件
│   │       ├── application-local.yml              # 本地开发配置
│   │       ├── application-dev.yml                # 开发环境配置
│   │       └── mapper/                            # MyBatis XML 映射文件
│   │           ├── user/
│   │           ├── owner/
│   │           ├── workorder/
│   │           ├── billing/
│   │           ├── notice/
│   │           ├── access/
│   │           └── base/
│   └── test/
│       └── java/
│           └── com/
│               └── spmp/
│                   ├── user/
│                   ├── owner/
│                   ├── workorder/
│                   ├── billing/
│                   ├── notice/
│                   ├── access/
│                   └── base/
├── pom.xml
└── Dockerfile
```

---

## 业务模块内部结构

每个业务模块采用 DDD 分层，以 `workorder`（工单模块）为例：

```
workorder/
├── api/                        # 模块对外接口（供其他模块调用）
│   ├── WorkOrderApi.java       # 接口定义
│   └── dto/                    # 跨模块传输的 DTO
│       └── WorkOrderBriefDTO.java
├── controller/                 # HTTP 接口层（面向前端）
│   └── WorkOrderController.java
├── service/                    # 业务逻辑层
│   ├── WorkOrderService.java   # 接口
│   └── impl/
│       └── WorkOrderServiceImpl.java
├── domain/                     # 领域模型
│   ├── entity/                 # 数据库实体
│   │   └── WorkOrder.java
│   ├── dto/                    # 模块内部 DTO
│   │   └── WorkOrderCreateDTO.java
│   └── vo/                     # 返回给前端的 VO
│       └── WorkOrderVO.java
├── repository/                 # 数据访问层
│   └── WorkOrderMapper.java    # MyBatis Mapper 接口
├── event/                      # 领域事件（可选）
│   └── WorkOrderCreatedEvent.java
└── constant/                   # 模块常量和枚举
    ├── WorkOrderStatus.java
    └── WorkOrderConstants.java
```

---

## 公共组件结构

```
common/
├── config/                     # 全局配置
│   ├── MybatisPlusConfig.java
│   ├── RedisConfig.java
│   └── WebMvcConfig.java
├── exception/                  # 统一异常处理
│   ├── BusinessException.java
│   ├── ErrorCode.java
│   └── GlobalExceptionHandler.java
├── result/                     # 统一返回格式
│   ├── Result.java
│   └── PageResult.java
├── security/                   # 安全与权限
│   ├── SecurityConfig.java
│   ├── JwtTokenProvider.java
│   └── DataPermissionInterceptor.java
└── util/                       # 工具类
    └── DateUtils.java
```

---

## 模块职责说明

| 模块 | 包名 | 职责 |
|------|------|------|
| user | `com.spmp.user` | 用户账号、角色管理、菜单权限、数据权限（全部→片区→小区→楼栋四级） |
| owner | `com.spmp.owner` | 业主信息、房产绑定、家庭成员、业主认证 |
| workorder | `com.spmp.workorder` | 报修工单、投诉建议、工单派发与流转、满意度评价 |
| billing | `com.spmp.billing` | 账单生成、在线支付、缴费催收、收费统计 |
| notice | `com.spmp.notice` | 社区公告发布、通知推送、已读管理 |
| access | `com.spmp.access` | 访客预约、门禁记录、临时通行证 |
| base | `com.spmp.base` | 小区、楼栋、单元、房屋等基础数据维护 |
| common | `com.spmp.common` | 公共配置、异常处理、统一返回、安全组件、工具类 |

---

## 模块间通信规范

模块间通过 `api/` 包进行松耦合通信：

```java
// workorder 模块的 api 包 —— 对外暴露接口
public interface WorkOrderApi {
    WorkOrderBriefDTO getWorkOrderBrief(Long workOrderId);
    int countPendingByBuildingId(Long buildingId);
}

// workorder 模块的 service 实现这个接口
@Service
public class WorkOrderServiceImpl implements WorkOrderService, WorkOrderApi {
    // ...
}

// notice 模块需要查工单信息时，注入 WorkOrderApi（不是 WorkOrderService）
@Service
public class NoticeServiceImpl implements NoticeService {
    @Autowired
    private WorkOrderApi workOrderApi;
}
```

禁止事项：
- 禁止模块间直接注入对方的 Service 实现类
- 禁止模块间共享 entity 类
- 禁止模块间直接访问对方的 Mapper

---

## 数据库表命名规范

共用一个数据库 schema（`spmp`），按模块前缀区分：

| 模块 | 表前缀 | 示例 |
|------|--------|------|
| user | `sys_` | `sys_user`、`sys_role`、`sys_menu` |
| owner | `ow_` | `ow_owner`、`ow_property`、`ow_family_member` |
| workorder | `wo_` | `wo_work_order`、`wo_dispatch_record` |
| billing | `bl_` | `bl_bill`、`bl_payment_record` |
| notice | `nt_` | `nt_notice`、`nt_read_record` |
| access | `ac_` | `ac_visitor`、`ac_access_record` |
| base | `bs_` | `bs_community`、`bs_building`、`bs_unit`、`bs_house` |

通用字段：
- `id` — 主键（BIGINT，自增）
- `create_time` — 创建时间
- `update_time` — 更新时间
- `create_by` — 创建人
- `update_by` — 更新人
- `del_flag` — 逻辑删除标记（0-正常，1-删除）

---

## 资源文件结构

### application.yml 主配置

```yaml
spring:
  application:
    name: spmp-backend
  profiles:
    active: local
  datasource:
    url: ${DB_URL:jdbc:mysql://localhost:3306/spmp?useUnicode=true&characterEncoding=utf8&serverTimezone=Asia/Shanghai}
    username: ${DB_USERNAME:root}
    password: ${DB_PASSWORD:}
    driver-class-name: com.mysql.cj.jdbc.Driver
    type: com.alibaba.druid.pool.DruidDataSource
  data:
    redis:
      host: ${REDIS_HOST:localhost}
      port: ${REDIS_PORT:6379}
      password: ${REDIS_PASSWORD:}
      database: 0

server:
  port: 8080

mybatis-plus:
  mapper-locations: classpath*:mapper/**/*.xml
  configuration:
    map-underscore-to-camel-case: true
  global-config:
    db-config:
      logic-delete-field: delFlag
      logic-delete-value: 1
      logic-not-delete-value: 0
```
