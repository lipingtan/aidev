CREATE TABLE biz_user (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(128) DEFAULT '',
    nickname VARCHAR(64) DEFAULT '',
    avatar VARCHAR(256) DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    -- FR-21 当前态轻量冗余（完整历史在 biz_user_status_logs 表）
    disabled_at TIMESTAMP NULL,
    disabled_reason_user VARCHAR(500) NOT NULL DEFAULT '',
    token_version INT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    create_by BIGINT DEFAULT 0,
    update_by BIGINT DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    access_token_expire_minutes INT NULL,
    refresh_token_expire_minutes INT NULL,
    UNIQUE KEY uk_tenant_phone (tenant_id, phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_biz_user_tenant ON biz_user(tenant_id);
CREATE INDEX idx_biz_user_phone ON biz_user(phone);
CREATE INDEX idx_biz_user_deleted ON biz_user(deleted_at);

-- FR-21 C端账号启用/禁用操作历史（2026-10-03；启动时 AutoMigrate 亦会自动创建）
CREATE TABLE IF NOT EXISTS biz_user_status_logs (
  id BIGINT NOT NULL PRIMARY KEY,
  tenant_id BIGINT NOT NULL DEFAULT 0,
  user_id BIGINT NOT NULL DEFAULT 0,
  action TINYINT NOT NULL COMMENT '0=禁用 1=启用',
  admin_reason VARCHAR(500) NOT NULL DEFAULT '' COMMENT '管理端内部原因（禁用时）',
  user_reason VARCHAR(500) NOT NULL DEFAULT '' COMMENT '用户展示原因（禁用时）',
  operator_id BIGINT NOT NULL DEFAULT 0,
  operator_name VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_busl_user (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='C端账号启用/禁用操作历史';
