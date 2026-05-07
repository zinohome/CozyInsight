CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) DEFAULT '',
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS menus (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    path VARCHAR(128) NOT NULL,
    component VARCHAR(128) DEFAULT '',
    icon VARCHAR(64) DEFAULT '',
    sort INT DEFAULT 0,
    status TINYINT DEFAULT 1 COMMENT '0=禁用, 1=启用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

CREATE TABLE IF NOT EXISTS role_menus (
    role_id BIGINT UNSIGNED NOT NULL,
    menu_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 初始化菜单数据
INSERT INTO menus (id, parent_id, name, path, component, icon, sort) VALUES
(1, 0, '工作台', '/', '', 'DashboardOutlined', 1),
(2, 0, '数据源', '/datasource', '', 'DatabaseOutlined', 2),
(3, 0, '数据集', '/dataset', '', 'TableOutlined', 3),
(4, 0, '图表', '/chart', '', 'BarChartOutlined', 4),
(5, 0, '仪表板', '/dashboard', '', 'LayoutOutlined', 5),
(6, 0, '系统管理', '/system', '', 'SettingOutlined', 6),
(7, 6, '用户管理', '/system/user', '', 'UserOutlined', 1),
(8, 6, '角色管理', '/system/role', '', 'TeamOutlined', 2),
(9, 6, '操作日志', '/system/log', '', 'FileTextOutlined', 3)
ON DUPLICATE KEY UPDATE name = VALUES(name);
