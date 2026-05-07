-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(128) NOT NULL UNIQUE,
    nick_name VARCHAR(128) DEFAULT '',
    avatar VARCHAR(255) DEFAULT '',
    phone VARCHAR(32) DEFAULT '',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    is_admin TINYINT DEFAULT 0 COMMENT '0=普通用户, 1=管理员',
    last_login_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 初始化管理员用户 (密码: admin123，bcrypt hash)
INSERT INTO users (username, password_hash, email, nick_name, is_admin, status)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mqrq3QsQJk9LJ6Vy1LQZ.1L6Vy1LQZ.', 'admin@cozyinsight.local', 'Administrator', 1, 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
