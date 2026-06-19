CREATE TABLE schedule_tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '任务名称',
    type VARCHAR(64) NOT NULL COMMENT '任务类型:email_report/snapshot/data_sync',
    cron_expr VARCHAR(128) NOT NULL COMMENT 'cron 表达式',
    config TEXT COMMENT '任务配置(JSON 字符串)',
    enabled TINYINT DEFAULT 0 COMMENT '是否启用',
    status VARCHAR(32) DEFAULT 'inactive' COMMENT '状态:inactive/running/error',
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者用户ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_enabled (enabled),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;