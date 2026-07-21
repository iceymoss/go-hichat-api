-- 用户自定义表情表
-- 对应 apps/user/models/useremojimodel.go
CREATE TABLE IF NOT EXISTS `user_emojis` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表情URL',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '名称',
  `thumbnail` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '缩略图URL',
  `width` int unsigned NOT NULL DEFAULT '0' COMMENT '宽',
  `height` int unsigned NOT NULL DEFAULT '0' COMMENT '高',
  `size` int unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  `file_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'image' COMMENT '文件类型',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_user_sort` (`user_id`,`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户自定义表情表';
