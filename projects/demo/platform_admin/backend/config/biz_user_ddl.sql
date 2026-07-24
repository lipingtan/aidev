CREATE TABLE biz_user (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(128) DEFAULT '',
    nickname VARCHAR(64) DEFAULT '',
    avatar VARCHAR(256) DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    token_version INT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    create_by BIGINT DEFAULT 0,
    update_by BIGINT DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    UNIQUE KEY uk_tenant_phone (tenant_id, phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_biz_user_tenant ON biz_user(tenant_id);
CREATE INDEX idx_biz_user_phone ON biz_user(phone);
CREATE INDEX idx_biz_user_deleted ON biz_user(deleted_at);
