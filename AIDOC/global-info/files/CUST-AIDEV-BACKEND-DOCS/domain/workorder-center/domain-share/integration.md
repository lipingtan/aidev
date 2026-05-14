# 工单管理系统集成

## 内部依赖

| 依赖模块 | 接口 | 用途 |
|---------|------|------|
| user-center | UserApi | 查询维修人员列表、用户信息 |
| user-center | PermissionApi | 权限校验 |
| owner-center | OwnerApi | 查询业主信息、房产绑定 |
| base-center | BaseApi | 查询小区/楼栋/单元/房屋层级数据 |
| common | FileService | 文件上传/下载 |

## 被依赖（对外 API）

| 调用方 | 方法 | 用途 |
|--------|------|------|
| notice-center | getWorkOrderBrief | 工单通知时获取摘要信息 |
| billing-center | getWorkOrderBrief | 缴费关联工单时获取摘要 |
| base-center | countPendingByBuildingId | 楼栋统计展示 |
| 首页仪表板 | countByCommunityId | 小区工单概览 |

## 外部系统

| 系统 | 集成方式 | 用途 |
|------|---------|------|
| 短信平台 | Spring Event 异步 | 工单通知短信（预留） |
| Redis | Spring Data Redis | 编号生成自增 + 列表缓存 |
