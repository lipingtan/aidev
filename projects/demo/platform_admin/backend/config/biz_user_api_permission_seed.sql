-- biz_user 管理端 API 权限注册
INSERT INTO admin_api_permission (type, http_method, url_pattern, permission_code, display_name, auth_required, app_code) VALUES
('ENDPOINT', 'GET', '/api/v1/admin/biz-users', 'biz_user:list', 'C端用户列表', 1, 'platform'),
('ENDPOINT', 'GET', '/api/v1/admin/biz-users/:id', 'biz_user:detail', 'C端用户详情', 1, 'platform'),
('ENDPOINT', 'POST', '/api/v1/admin/biz-users', 'biz_user:create', '创建C端用户', 1, 'platform'),
('ENDPOINT', 'PUT', '/api/v1/admin/biz-users/:id', 'biz_user:update', '更新C端用户', 1, 'platform'),
('ENDPOINT', 'DELETE', '/api/v1/admin/biz-users/:id', 'biz_user:delete', '删除C端用户', 1, 'platform'),
('ENDPOINT', 'POST', '/api/v1/admin/biz-users/:id/reset-password', 'biz_user:reset_password', '重置C端用户密码', 1, 'platform'),
('ENDPOINT', 'POST', '/api/v1/admin/biz-users/:id/force-logout', 'biz_user:force_logout', '强制登出C端用户', 1, 'platform'),
('ENDPOINT', 'POST', '/api/v1/admin/biz-users/:id/toggle-status', 'biz_user:toggle_status', '启用禁用C端用户', 1, 'platform');
