-- 业务菜单初始化 SQL
-- 注意：业务功能菜单由各插件在启动时自行注册到"扩展功能"目录下

-- 预置"扩展功能"父菜单（所有插件菜单的容器）
INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at)
SELECT 'PluginExtensions', '扩展功能', 'ep:menu', '/extensions', '/0/', 'M', '无', '', 0, false, '', 'Layout', 100, '0', '0', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE menu_name = 'PluginExtensions' AND deleted_at IS NULL
);

-- 预置"插件管理"子菜单（管理员管理插件的入口）
INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at)
SELECT 'PluginManage', '插件管理', 'ep:box', '/extensions/plugins', CONCAT('/0/', (SELECT menu_id FROM sys_menu WHERE menu_name = 'PluginExtensions' AND deleted_at IS NULL LIMIT 1), '/'), 'C', '无', 'plugin:list', (SELECT menu_id FROM sys_menu WHERE menu_name = 'PluginExtensions' AND deleted_at IS NULL LIMIT 1), false, '', 'admin/plugin/index', 0, '0', '0', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE menu_name = 'PluginManage' AND deleted_at IS NULL
);
