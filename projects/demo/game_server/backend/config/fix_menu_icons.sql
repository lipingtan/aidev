-- 修复旧菜单图标为 ep: 格式
UPDATE sys_menu SET icon = 'ep:setting' WHERE icon = 'api-server' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:user' WHERE icon = 'user' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:menu' WHERE icon = 'tree-table' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:avatar' WHERE icon = 'peoples' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:office-building' WHERE icon = 'tree' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:postcard' WHERE icon = 'pass' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:notebook' WHERE icon = 'education' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:tools' WHERE icon = 'dev-tools' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:document' WHERE icon = 'guide' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:data-analysis' WHERE icon = 'swagger' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:document-copy' WHERE icon = 'log' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:tickets' WHERE icon = 'logininfor' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:edit' WHERE icon = 'skill' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:cpu' WHERE icon = 'druid' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:timer' WHERE icon = 'time-range' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:calendar' WHERE icon = 'job' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:bug' WHERE icon = 'bug' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:code' WHERE icon = 'code' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:edit-pen' WHERE icon = 'build' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:connection' WHERE icon = 'api-doc' AND deleted_at IS NULL;
UPDATE sys_menu SET icon = 'ep:set-up' WHERE icon = 'system-tools' AND deleted_at IS NULL;
-- 按钮级菜单统一设置空图标（F类型不需要图标）
UPDATE sys_menu SET icon = '' WHERE icon = 'app-group-fill' AND deleted_at IS NULL;
