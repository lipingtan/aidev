# 设计：platform_admin 前端页面补全

## 技术方案

### 文件结构

```
platform_admin/platform_admin/frontend/src/views/game/
├── index.vue       # 已有：游戏管理
├── dlc.vue         # 新增：DLC管理（FR-1）
├── player.vue      # 新增：玩家管理（FR-2）
├── order.vue       # 新增：订单管理（FR-3）
├── payment.vue     # 新增：支付配置（FR-4）
└── h5.vue          # 新增：H5页面管理（FR-5）

platform_admin/backend/config/
└── menu_init.sql   # 新增：菜单初始化SQL（FR-6）
```

### 页面统一结构规范

所有页面复用 `index.vue` 的代码风格：

```vue
<template>
  <div class="main">
    <!-- 搜索栏 -->
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      ...
    </el-form>
    <!-- 操作栏（仅有增删权限的页面显示） -->
    <div class="flex justify-between mb-3">...</div>
    <!-- 数据表格 -->
    <el-table v-loading="loading" :data="list" border stripe>...</el-table>
    <!-- 分页 -->
    <el-pagination ... />
    <!-- 弹窗 -->
    <el-dialog ...>...</el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance } from "element-plus";
defineOptions({ name: "XxxManage" });
// query / form / rules / loadData / onSearch / onReset / openDialog / onSubmit
</script>
```

### API 设计（已有，直接复用）

| 方法 | 路径 | 描述 | 对应函数 |
|------|------|------|----------|
| GET | /api/v1/game-dlc | DLC列表 | getDlcList |
| POST | /api/v1/game-dlc | 新增DLC | createDlc |
| PUT | /api/v1/game-dlc/:id | 编辑DLC | updateDlc |
| DELETE | /api/v1/game-dlc/:id | 删除DLC | deleteDlc |
| GET | /api/v1/game-player | 玩家列表 | getPlayerList |
| GET | /api/v1/game-player/:id | 玩家详情 | getPlayerById |
| PUT | /api/v1/game-player/:id/ban | 封禁/解封 | banPlayer |
| GET | /api/v1/game-order | 订单列表 | getOrderList |
| POST | /api/v1/game-order/:id/refund | 退款 | refundOrder |
| GET | /api/v1/game-payment-config | 支付配置列表 | getPaymentConfigList |
| POST | /api/v1/game-payment-config | 保存支付配置 | savePaymentConfig |
| GET | /api/v1/game-h5 | H5页面列表 | getH5PageList |
| POST | /api/v1/game-h5 | 新增H5页面 | createH5Page |
| PUT | /api/v1/game-h5/:id | 编辑H5页面 | updateH5Page |
| DELETE | /api/v1/game-h5/:id | 删除H5页面 | deleteH5Page |

### 各页面核心逻辑

#### dlc.vue
- query: `{ gameId: "", name: "", pageIndex: 1, pageSize: 10 }`
- form: `{ id, gameId, dlcKey, name, version, description, price(元), isFree, status, minGameVersion }`
- 提交时：`data.price = Math.round(form.price * 100)`
- 列表显示：`(row.price / 100).toFixed(2)` 元

#### player.vue
- query: `{ gameId: "", uid: "", nickname: "", status: "", pageIndex: 1, pageSize: 10 }`
- 详情弹窗（只读）：展示全部字段含 platformId、deviceId、banReason
- 封禁弹窗：`banForm = { banReason: "" }`，调用 `banPlayer(id, { status: 2, banReason })`
- 解封：直接调用 `banPlayer(id, { status: 1, banReason: "" })`，无需弹窗

#### order.vue
- query: `{ gameId: "", orderNo: "", status: "", pageIndex: 1, pageSize: 10 }`
- 金额显示：`(row.amount / 100).toFixed(2) + ' ' + row.currency`
- status 颜色映射：`pending→warning, paid→success, refunded→info, failed→danger`
- 导出 CSV：前端将当前列表数据转为 CSV 并触发下载，文件名 `orders_YYYYMMDD.csv`

#### payment.vue
- query: `{ gameId: "", pageIndex: 1, pageSize: 10 }`
- form: `{ id, gameId, channel, channelName, enabled, env, ...渠道字段 }`
- channel 切换逻辑：`computed showStripe = form.channel === 'stripe'` 等
- 列表不展示 `*Key / *Secret / *PrivateKey` 字段

#### h5.vue
- query: `{ gameId: "", name: "", pageIndex: 1, pageSize: 10 }`
- form: `{ id, gameId, pageKey, name, pageType, useExternal, content, externalUrl, status, remark }`
- pageType 选项：`custom / 商城页 / 活动页 / 公告页`
- useExternal 切换：`v-if="form.useExternal === 1"` 显示 externalUrl，否则显示 content textarea

### 菜单初始化 SQL 设计

```sql
-- 查询游戏管理父菜单ID后插入子菜单
-- 假设游戏管理父菜单 menu_id 需运行时确认，脚本使用子查询获取
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, menu_type, visible, status, perms, icon, create_by, update_by, remark, create_time, update_time)
SELECT '游戏管理', 0, 1, 'game', '', 'M', '0', '0', '', 'ep:game-pad', 1, 1, '', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE path = 'game' AND menu_type = 'M');

-- 子菜单基于父菜单ID插入（通过子查询）
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, menu_type, visible, status, perms, icon, create_by, update_by, remark, create_time, update_time)
VALUES
('DLC管理',  (SELECT menu_id FROM sys_menu WHERE path='game' AND menu_type='M' LIMIT 1), 2, 'dlc',     'game/dlc',     'C', '0', '0', 'game:dlc:list',     'ep:box',          1, 1, '', NOW(), NOW()),
('玩家管理', (SELECT menu_id FROM sys_menu WHERE path='game' AND menu_type='M' LIMIT 1), 3, 'player',  'game/player',  'C', '0', '0', 'game:player:list',  'ep:user',         1, 1, '', NOW(), NOW()),
('订单管理', (SELECT menu_id FROM sys_menu WHERE path='game' AND menu_type='M' LIMIT 1), 4, 'order',   'game/order',   'C', '0', '0', 'game:order:list',   'ep:shopping-cart',1, 1, '', NOW(), NOW()),
('支付配置', (SELECT menu_id FROM sys_menu WHERE path='game' AND menu_type='M' LIMIT 1), 5, 'payment', 'game/payment', 'C', '0', '0', 'game:payment:list', 'ep:credit-card',  1, 1, '', NOW(), NOW()),
('H5页面',   (SELECT menu_id FROM sys_menu WHERE path='game' AND menu_type='M' LIMIT 1), 6, 'h5',      'game/h5',      'C', '0', '0', 'game:h5:list',      'ep:document',     1, 1, '', NOW(), NOW());
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 游戏管理页（index.vue）功能不受影响 | 游戏列表正常加载、新增/编辑/删除正常 |
| RG-2 | 登录流程不受影响 | /login 接口正常返回 token |
| RG-3 | 现有菜单路由不受影响 | 已有菜单项正常跳转 |
| RG-4 | api/game.ts 现有函数签名不变 | 游戏管理页调用无报错 |

## 正确性属性

- 所有金额字段：前端输入"元"，提交时×100转"分"，展示时÷100
- 敏感字段（*Key/*Secret/*PrivateKey）：列表不展示，仅在编辑表单中可填写
- 封禁操作：必须携带 banReason，解封时 banReason 置空
- 导出 CSV：包含当前搜索条件下的全部数据（非仅当前页）
- 多租户：列表展示 tenantId 字段，查询时由后端按 tenant_id 自动隔离
