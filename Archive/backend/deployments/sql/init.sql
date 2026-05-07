-- CozyInsight 数据库初始化脚本
-- MySQL 8.0+

CREATE DATABASE IF NOT EXISTS cozy_insight DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cozy_insight;

-- ============================================
-- 认证授权表
-- ============================================

-- 用户表
CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` VARCHAR(50) PRIMARY KEY,
  `username` VARCHAR(100) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255),
  `nick_name` VARCHAR(100),
  `status` INT DEFAULT 1 COMMENT '0=禁用 1=启用',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_username (username),
  INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL UNIQUE,
  `description` VARCHAR(500),
  `type` VARCHAR(50) COMMENT 'system, custom',
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_name (name),
  INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS `sys_permission` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `resource` VARCHAR(100) COMMENT 'datasource, dataset, chart, dashboard',
  `resource_id` VARCHAR(50),
  `action` VARCHAR(50) COMMENT 'read, write, delete, manage',
  `description` VARCHAR(500),
  `create_time` BIGINT,
  INDEX idx_resource (resource, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS `sys_role_permission` (
  `id` VARCHAR(50) PRIMARY KEY,
  `role_id` VARCHAR(50) NOT NULL,
  `permission_id` VARCHAR(50) NOT NULL,
  `create_time` BIGINT,
  INDEX idx_role (role_id),
  INDEX idx_permission (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id` VARCHAR(50) PRIMARY KEY,
  `user_id` VARCHAR(50) NOT NULL,
  `role_id` VARCHAR(50) NOT NULL,
  `create_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_user (user_id),
  INDEX idx_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

-- 资源权限表
CREATE TABLE IF NOT EXISTS `sys_resource_permission` (
  `id` VARCHAR(50) PRIMARY KEY,
  `resource_type` VARCHAR(50) NOT NULL COMMENT 'datasource, dataset, chart, dashboard',
  `resource_id` VARCHAR(50) NOT NULL,
  `target_type` VARCHAR(20) NOT NULL COMMENT 'user, role',
  `target_id` VARCHAR(50) NOT NULL,
  `permission` VARCHAR(20) NOT NULL COMMENT 'read, write, manage',
  `create_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_resource (resource_type, resource_id),
  INDEX idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源权限表';

-- ============================================
-- 数据源表
-- ============================================

CREATE TABLE IF NOT EXISTS `datasource` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `type` VARCHAR(50) NOT NULL COMMENT 'mysql, postgresql, clickhouse, oracle, sqlserver',
  `config` TEXT COMMENT 'JSON配置',
  `status` VARCHAR(20) DEFAULT 'new',
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_type (type),
  INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源表';

-- ============================================
-- 数据集表
-- ============================================

-- 数据集分组
CREATE TABLE IF NOT EXISTS `dataset_group` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `pid` VARCHAR(50) DEFAULT '0',
  `level` INT DEFAULT 0,
  `type` VARCHAR(50) COMMENT 'folder, dataset',
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_pid (pid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集分组';

-- 数据集表
CREATE TABLE IF NOT EXISTS `dataset_table` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `dataset_group_id` VARCHAR(50),
  `datasource_id` VARCHAR(50),
  `db_name` VARCHAR(255),
  `table_name` VARCHAR(255),
  `type` VARCHAR(50) COMMENT 'db, sql, excel, api',
  `mode` INT DEFAULT 0 COMMENT '0=直连 1=抽取',
  `info` TEXT COMMENT 'JSON配置',
  `sql_variable_details` TEXT,
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_group (dataset_group_id),
  INDEX idx_datasource (datasource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集表';

-- 数据集字段
CREATE TABLE IF NOT EXISTS `dataset_table_field` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dataset_table_id` VARCHAR(50) NOT NULL,
  `origin_name` VARCHAR(255),
  `name` VARCHAR(255),
  `type` VARCHAR(50),
  `size` INT DEFAULT 0,
  `de_type` INT DEFAULT 0 COMMENT '0=文本 1=时间 2=数值 3=地理位置 4=其他',
  `de_extra_type` INT DEFAULT 0,
  `checked` TINYINT DEFAULT 1,
  `column_index` INT DEFAULT 0,
  INDEX idx_table (dataset_table_id),
  INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集字段';

-- ============================================
-- 图表表
-- ============================================

CREATE TABLE IF NOT EXISTS `core_chart_view` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `scene_id` VARCHAR(50) COMMENT '场景ID',
  `table_id` VARCHAR(50) COMMENT '数据集ID',
  `type` VARCHAR(50) COMMENT '图表类型',
  `render` VARCHAR(50) COMMENT '渲染方式',
  `result_count` BIGINT DEFAULT 0,
  `result_mode` VARCHAR(50),
  `title` VARCHAR(255),
  `x_axis` TEXT COMMENT 'X轴配置JSON',
  `x_axis_ext` TEXT,
  `y_axis` TEXT COMMENT 'Y轴配置JSON',
  `y_axis_ext` TEXT,
  `custom_attr` TEXT COMMENT '自定义属性JSON',
  `custom_style` TEXT COMMENT '自定义样式JSON',
  `custom_filter` TEXT COMMENT '过滤条件JSON',
  `drill_fields` TEXT,
  `snapshot` TEXT COMMENT '快照',
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_scene (scene_id),
  INDEX idx_table (table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图表视图';

-- ============================================
-- 仪表板表
-- ============================================

CREATE TABLE IF NOT EXISTS `dashboard` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `pid` VARCHAR(50) DEFAULT '0',
  `node_type` VARCHAR(50) NOT NULL COMMENT 'folder, dashboard',
  `type` VARCHAR(50) COMMENT 'dashboard, dataV',
  `canvas_style_data` LONGTEXT COMMENT '画布样式JSON',
  `component_data` LONGTEXT COMMENT '组件数据JSON',
  `status` INT DEFAULT 0 COMMENT '0=未发布 1=已发布',
  `publish_time` BIGINT DEFAULT 0,
  `sort` INT DEFAULT 0,
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_pid (pid),
  INDEX idx_node_type (node_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仪表板';

-- 仪表板组件
CREATE TABLE IF NOT EXISTS `core_dashboard_component` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dashboard_id` VARCHAR(50) NOT NULL,
  `chart_id` VARCHAR(50),
  `type` VARCHAR(50) COMMENT 'chart, text, image',
  `x` INT DEFAULT 0,
  `y` INT DEFAULT 0,
  `w` INT DEFAULT 0,
  `h` INT DEFAULT 0,
  `config` TEXT COMMENT 'JSON配置',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_dashboard (dashboard_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仪表板组件';

-- ============================================
-- 分享表
-- ============================================

CREATE TABLE IF NOT EXISTS `sys_share` (
  `id` VARCHAR(50) PRIMARY KEY,
  `resource_type` VARCHAR(50) NOT NULL COMMENT 'dashboard, chart',
  `resource_id` VARCHAR(50) NOT NULL,
  `token` VARCHAR(50) UNIQUE,
  `password` VARCHAR(100),
  `expire_time` BIGINT DEFAULT 0 COMMENT '过期时间,0表示永不过期',
  `create_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_token (token),
  INDEX idx_resource (resource_type, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分享表';

-- ============================================
-- 定时任务表
-- ============================================

CREATE TABLE IF NOT EXISTS `sys_schedule_task` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(200) NOT NULL,
  `type` VARCHAR(50) COMMENT 'email_report, snapshot, data_sync',
  `cron_expr` VARCHAR(100),
  `enabled` TINYINT DEFAULT 0,
  `status` VARCHAR(20) COMMENT 'active, inactive, running',
  `config` TEXT COMMENT 'JSON配置',
  `last_run_time` BIGINT DEFAULT 0,
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_enabled (enabled),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';

-- ============================================
-- 初始数据
-- ============================================

-- 插入系统管理员角色
INSERT INTO `sys_role` (`id`, `name`, `description`, `type`, `create_time`, `update_time`)
VALUES ('admin-role', 'Admin', '系统管理员', 'system', UNIX_TIMESTAMP(NOW()) * 1000, UNIX_TIMESTAMP(NOW()) * 1000)
ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入默认管理员用户 (密码: admin123, 需要在代码中hash)
-- INSERT INTO `sys_user` (`id`, `username`, `password`, `email`, `nick_name`, `status`, `create_time`, `update_time`)
-- VALUES ('admin', 'admin', '$2a$10$...hashed...', 'admin@example.com', 'Administrator', 1, UNIX_TIMESTAMP(NOW()) * 1000, UNIX_TIMESTAMP(NOW()) * 1000)
-- ON DUPLICATE KEY UPDATE `username` = `username`;

COMMIT;

-- ============================================
-- 系统管理表 (新增)
-- ============================================

-- 操作日志表
CREATE TABLE IF NOT EXISTS `sys_oper_log` (
  `id` VARCHAR(50) PRIMARY KEY,
  `user_id` VARCHAR(50),
  `username` VARCHAR(100),
  `module` VARCHAR(100) COMMENT 'datasource, dataset, chart, dashboard',
  `action` VARCHAR(50) COMMENT 'create, update, delete, view, export',
  `detail` TEXT COMMENT '操作详情JSON',
  `resource_id` VARCHAR(50),
  `ip` VARCHAR(50),
  `user_agent` VARCHAR(255),
  `status` INT DEFAULT 1 COMMENT '1=成功 0=失败',
  `error_msg` TEXT,
  `create_time` BIGINT,
  INDEX idx_user (user_id),
  INDEX idx_module (module),
  INDEX idx_resource (resource_id),
  INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';

-- 系统设置表
CREATE TABLE IF NOT EXISTS `sys_setting` (
  `id` VARCHAR(50) PRIMARY KEY,
  `type` VARCHAR(50) COMMENT 'email, auth, system, display',
  `setting_key` VARCHAR(100) UNIQUE,
  `value` TEXT,
  `update_time` BIGINT,
  `update_by` VARCHAR(50),
  INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统设置表';

-- 数据集计算字段表
CREATE TABLE IF NOT EXISTS `dataset_table_field_calculated` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dataset_table_id` VARCHAR(50) NOT NULL,
  `field_name` VARCHAR(255),
  `display_name` VARCHAR(255),
  `expression` TEXT COMMENT '计算公式',
  `data_type` VARCHAR(50) COMMENT 'string, long, double, date',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_table (dataset_table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集计算字段';

-- ============================================
-- Dashboard高级功能表 (新增)
-- ============================================

-- Dashboard联动配置表
CREATE TABLE IF NOT EXISTS `dashboard_linkage` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dashboard_id` VARCHAR(50) NOT NULL,
  `source_component_id` VARCHAR(50) NOT NULL,
  `target_component_id` VARCHAR(50) NOT NULL,
  `source_field_id` VARCHAR(50),
  `target_field_id` VARCHAR(50),
  `linkage_type` VARCHAR(50) COMMENT 'click, hover, select',
  `update_type` VARCHAR(50) COMMENT 'replace, append',
  `enable` TINYINT DEFAULT 1,
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_dashboard (dashboard_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仪表板联动配置';

--  Dashboard参数表
CREATE TABLE IF NOT EXISTS `dashboard_parameter` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dashboard_id` VARCHAR(50) NOT NULL,
  `name` VARCHAR(200),
  `param_type` VARCHAR(50) COMMENT 'text, date, select, multiselect',
  `default_value` VARCHAR(500),
  `options` TEXT COMMENT 'JSON数组',
  `required` TINYINT DEFAULT 0,
  `enable` TINYINT DEFAULT 1,
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_dashboard (dashboard_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仪表板参数';

-- Dashboard Tab容器表
CREATE TABLE IF NOT EXISTS `dashboard_tab` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dashboard_id` VARCHAR(50) NOT NULL,
  `name` VARCHAR(200),
  `order` INT DEFAULT 0,
  `component_ids` TEXT COMMENT 'JSON数组',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_dashboard (dashboard_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dashboard Tab容器';

-- 图表联动配置表
CREATE TABLE IF NOT EXISTS `chart_view_linkage` (
  `id` VARCHAR(50) PRIMARY KEY,
  `source_view_id` VARCHAR(50) NOT NULL,
  `target_view_id` VARCHAR(50) NOT NULL,
  `source_field_id` VARCHAR(50),
  `target_field_id` VARCHAR(50),
  `linkage_type` VARCHAR(50) COMMENT 'click, hover',
  `update_type` VARCHAR(50) COMMENT 'replace, append',
  `ext` TEXT COMMENT 'JSON扩展配置',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_source (source_view_id),
  INDEX idx_target (target_view_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图表联动配置';

-- 图表钻取配置表
CREATE TABLE IF NOT EXISTS `chart_view_drill` (
  `id` VARCHAR(50) PRIMARY KEY,
  `chart_view_id` VARCHAR(50) NOT NULL,
  `field_id` VARCHAR(50),
  `drill_fields` TEXT COMMENT 'JSON钻取字段数组',
  `drill_type` VARCHAR(50) COMMENT 'down, up',
  `create_time` BIGINT,
  `update_time` BIGINT,
  INDEX idx_chart (chart_view_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图表钻取配置';

-- 数据集行级权限表
CREATE TABLE IF NOT EXISTS `dataset_row_permissions` (
  `id` VARCHAR(50) PRIMARY KEY,
  `dataset_id` VARCHAR(50) NOT NULL,
  `auth_target_type` VARCHAR(50) COMMENT 'user, role, dept',
  `auth_target_id` VARCHAR(50),
  `where_condition` TEXT COMMENT 'SQL WHERE条件',
  `express_type` VARCHAR(50) COMMENT 'sql, formula',
  `enable` TINYINT DEFAULT 1,
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50),
  INDEX idx_dataset (dataset_id),
  INDEX idx_target (authI_target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集行级权限';

-- 图表模板表
CREATE TABLE IF NOT EXISTS `chart_template` (
  `id` VARCHAR(50) PRIMARY KEY,
  `name` VARCHAR(200),
  `type` VARCHAR(50),
  `category` VARCHAR(50) COMMENT 'business, stat, custom',
  `config` TEXT COMMENT '模板配置JSON',
  `preview` VARCHAR(500),
  `is_system` TINYINT DEFAULT 0,
  `is_public` TINYINT DEFAULT 0,
  `create_time` BIGINT,
  `update_time` BIGINT,
  `create_by` VARCHAR(50)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图表模板';

COMMIT;

