CREATE TABLE IF NOT EXISTS row_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    dataset_id BIGINT UNSIGNED NOT NULL,
    field_name VARCHAR(128) NOT NULL COMMENT '数据集字段名',
    operator VARCHAR(16) NOT NULL DEFAULT '=' COMMENT '比较运算符: =, !=, >, <, >=, <=, IN',
    value VARCHAR(255) NOT NULL COMMENT '对比值或用户属性占位符',
    user_attr VARCHAR(64) DEFAULT '' COMMENT '用户属性字段名，如 dept_id',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dataset_id (dataset_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行级数据权限表';
