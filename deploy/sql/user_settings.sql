-- 用户配置表（key/value 以 JSON 整体存储，每个用户一行）
-- 对应 apps/user/models/usersettingsmodel.go
CREATE TABLE IF NOT EXISTS `user_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `settings` json NOT NULL COMMENT '用户配置JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户配置表';
