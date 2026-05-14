# 模块间通信说明

| 字段 | 内容 |
|------|------|
| 文档名称 | 模块间通信说明 |
| 版本 | 2.0 |
| 创建日期 | 2026-04-17 |
| 负责人 | 技术团队 |
| 审核人 | 技术负责人 |

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        前端层 (Frontend)                        │
├─────────────────────────────────────────────────────────────────┤
│  spmp-web-pc (PC 管理端)        │  spmp-web-h5 (H5 业主端)     │
└──────────────────────┬──────────────────────────┬───────────────┘
                       │ HTTP                     │ HTTP
                       ▼                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   spmp-backend (单体后端)                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ ┌──────────────┐  │
│  │  user    │ │  owner   │ │  workorder   │ │  billing     │  │
│  │  用户权限│ │  业主管理│ │  工单管理    │ │  缴费管理    │  │
│  └──────────┘ └──────────┘ └──────────────┘ └──────────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ ┌──────────────┐  │
│  │  notice  │ │  access  │ │  base        │ │  common      │  │
│  │  公告管理│ │  门禁管理│ │  基础数据    │ │  公共组件    │  │
│  └──────────┘ └──────────┘ └──────────────┘ └──────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                MySQL 8.0  │  Redis 7.2                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 模块间通信方式

### 同步调用（通过 api 包接口）

模块间通过 `api/` 包暴露的接口进行松耦合调用，不直接依赖对方的 Service 实现：

```java
// workorder 模块暴露的 API 接口
public interface WorkOrderApi {
    WorkOrderBriefDTO getWorkOrderBrief(Long workOrderId);
    int countPendingByBuildingId(Long buildingId);
}

// notice 模块调用 workorder 模块
@Service
public class NoticeServiceImpl implements NoticeService {
    @Autowired
    private WorkOrderApi workOrderApi;  // 注入接口，不是实现类
}
```

### 异步通信（通过 Spring Event）

模块间异步通信使用 Spring ApplicationEvent，不引入外部 MQ：

```java
// 工单模块发布事件
@Service
public class WorkOrderServiceImpl {
    @Autowired
    private ApplicationEventPublisher eventPublisher;

    public void completeWorkOrder(Long id) {
        // 业务逻辑...
        eventPublisher.publishEvent(new WorkOrderCompletedEvent(id));
    }
}

// 公告模块监听事件
@Component
public class WorkOrderEventListener {
    @EventListener
    public void onWorkOrderCompleted(WorkOrderCompletedEvent event) {
        // 发送通知...
    }
}
```

---

## 模块依赖关系

| 模块 | 依赖的其他模块 API |
|------|-------------------|
| user | base（查询小区/楼栋信息用于数据权限） |
| owner | user（验证用户）、base（查询房屋信息） |
| workorder | user（查询维修人员）、owner（查询业主）、base（查询楼栋） |
| billing | owner（查询业主房产）、base（查询房屋） |
| notice | user（查询推送目标）、base（查询小区/楼栋） |
| access | owner（验证业主）、base（查询门禁设备） |
| base | 无（基础模块，不依赖其他业务模块） |
| common | 无（被所有模块依赖） |

依赖原则：
- `common` 和 `base` 是底层模块，被其他模块依赖
- 业务模块间只通过 `api/` 接口通信
- 禁止循环依赖

---

## 外部系统集成

| 外部系统 | 集成方式 | 说明 |
|----------|----------|------|
| 微信支付 | HTTP API | 在线缴费支付（billing 模块） |
| 支付宝 | HTTP API | 在线缴费支付（billing 模块） |
| 短信平台 | HTTP API | 通知短信发送（notice 模块） |
| 门禁硬件 | MQTT/HTTP | 门禁控制（access 模块） |

---

## API 路径规范

所有接口统一前缀 `/api/v1/{模块}/{资源}`：

| 模块 | 路径前缀 | 示例 |
|------|----------|------|
| user | `/api/v1/user/` | `/api/v1/user/roles` |
| owner | `/api/v1/owner/` | `/api/v1/owner/properties` |
| workorder | `/api/v1/workorder/` | `/api/v1/workorder/orders` |
| billing | `/api/v1/billing/` | `/api/v1/billing/bills` |
| notice | `/api/v1/notice/` | `/api/v1/notice/announcements` |
| access | `/api/v1/access/` | `/api/v1/access/visitors` |
| base | `/api/v1/base/` | `/api/v1/base/communities` |
