-- 游戏管理菜单初始化 SQL
-- 幂等设计：可重复执行，不会产生重复数据
-- 执行顺序：先插入父菜单，再插入子菜单

-- -------------------------------------------------------
-- 1. 插入游戏管理父菜单（M 类型，parent_id=0）
-- -------------------------------------------------------
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GameManage', '游戏管理', 'ep:game-pad', '/game', '/0/', 'M', '无', '', 0, false, '', 'Layout', 90, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL
);

-- -------------------------------------------------------
-- 2. 插入五个子菜单（C 类型，parent_id 通过子查询获取）
-- -------------------------------------------------------

-- DLC管理
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GameDlc', 'DLC管理', 'ep:box', '/game/dlc', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'game:dlc:list', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), false, '', 'game/dlc', 1, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game/dlc' AND menu_type = 'C' AND deleted_at IS NULL
);

-- 玩家管理
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GamePlayer', '玩家管理', 'ep:user', '/game/player', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'game:player:list', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), false, '', 'game/player', 2, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game/player' AND menu_type = 'C' AND deleted_at IS NULL
);

-- 订单管理
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GameOrder', '订单管理', 'ep:shopping-cart', '/game/order', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'game:order:list', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), false, '', 'game/order', 3, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game/order' AND menu_type = 'C' AND deleted_at IS NULL
);

-- 支付配置
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GamePayment', '支付配置', 'ep:credit-card', '/game/payment', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'game:payment:list', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), false, '', 'game/payment', 4, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game/payment' AND menu_type = 'C' AND deleted_at IS NULL
);

-- H5页面
INSERT INTO sys_menu (name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at, deleted_at)
SELECT 'GameH5', 'H5页面', 'ep:document', '/game/h5', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'game:h5:list', (SELECT menu_id FROM sys_menu WHERE path = '/game' AND menu_type = 'M' AND deleted_at IS NULL LIMIT 1), false, '', 'game/h5', 5, '0', '0', 1, 1, NOW(), NOW(), NULL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE path = '/game/h5' AND menu_type = 'C' AND deleted_at IS NULL
);
