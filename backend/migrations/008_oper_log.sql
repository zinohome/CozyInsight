CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    username VARCHAR(64) DEFAULT '',
    method VARCHAR(16) NOT NULL,
    path VARCHAR(255) NOT NULL,
    query VARCHAR(1024) DEFAULT '',
    body TEXT,
    ip VARCHAR(64) DEFAULT '',
    user_agent VARCHAR(255) DEFAULT '',
    status_code INT DEFAULT 200,
    duration BIGINT DEFAULT 0 COMMENT '耗时毫秒',
    error_message VARCHAR(512) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_path (path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
