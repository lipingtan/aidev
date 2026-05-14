---
inclusion: always
---

# Tech Stack

## Backend

| Category | Technology |
|---|---|
| Language | Java 1.8 |
| Framework | Spring Boot 2.4.4 + Spring Cloud 2020.0.1 + Spring Cloud Alibaba |
| Internal Framework | yh-framework-dependencies 1.2.0 |
| ORM | MyBatis Plus (随内部框架) |
| Database | MySQL 8.0 |
| Cache | Redis 7.2.1 (Spring Data Redis / Lettuce) |
| Connection Pool | Druid |
| Service Discovery | Nacos (Discovery + Config) |
| HTTP Client | OpenFeign |
| Security | Spring Security + JWT (RBAC) |
| Message Queue | RocketMQ 5.0.4 |
| Build | Maven 3.x |
| Utilities | Lombok, Hutool, Jackson |
| Testing | JUnit, Mockito, Spring Boot Test |
| Logging | SLF4J + Logback |

### Common Backend Commands

```bash
# Build (skip tests)
mvn clean install -DskipTests

# Build (with tests)
mvn clean install

# Run tests
mvn test

# Run specific test class
mvn test -Dtest=PaymentServiceTest

# Run locally with profile
mvn spring-boot:run -Dspring-boot.run.profiles=local
```

### Key Backend Conventions

- All API paths versioned: `/api/v1/{module}/{resource}`
- Unified response wrapper: `Result<T>` with `code`, `message`, `data`, `traceId`
- Always check `result.isSuccess()` before calling `result.getData()`
- Feign Clients in `infrastructure/feign/`, DTOs in `infrastructure/feign/dto/`, fallbacks in `infrastructure/feign/fallback/`
- All write Feign calls must carry `bizSerialNo` for idempotency
- Logical delete via `del_flag` field + `@TableLogic`
- Auto-fill `create_time`, `update_time` via `MetaObjectHandler`
- Unit test coverage target: ≥ 90% for core business logic

---

## Frontend

| Category | Technology |
|---|---|
| Framework | Vue 3.4.27 (Composition API, `<script setup>`) |
| Language | TypeScript 5.4.5 |
| UI Library | Element Plus 2.7.4 |
| State Management | Pinia |
| Router | Vue Router 4.x (history mode) |
| HTTP Client | Axios (unified instance in `src/utils/request.js`) |
| Build Tool | Vite 5.2.12 |

### Key Frontend Conventions

- All route components use lazy loading: `() => import('@/views/...')`
- Route names: kebab-case `{module}-{action}`, e.g. `payment-list`, `payment-detail`
- API base URL from env var `VITE_API_BASE_URL`; all requests go through unified axios instance
- Axios interceptor adds `Authorization: Bearer {token}` header; handles 401 by redirecting to login
- Composables prefixed with `use`, stored in `src/composables/`
- Props typed with `defineProps<{}>()` + `withDefaults`; events typed with `defineEmits<{}>()`
- No direct prop mutation — emit events to parent
- Global state in Pinia only when truly cross-component; local state uses `ref`/`reactive`
- Nginx must have `try_files $uri $uri/ /index.html` for history mode routing
