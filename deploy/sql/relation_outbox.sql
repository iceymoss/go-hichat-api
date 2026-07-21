-- 关系变更事务性发件箱（踢人/退群/解散/删好友 与本表插入同事务提交，relay 异步投递 Kafka）
-- 自增 id 兼作单调版本号（消费端版本门用）
CREATE TABLE IF NOT EXISTS `relation_outbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键，兼作单调版本号',
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件类型 group.member.removed/group.disbanded/friend.deleted',
  `group_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '群ID（好友事件留空，作分区/索引）',
  `payload` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件JSON载荷',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '0待投递 1已投递',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `sent_at` timestamp NULL DEFAULT NULL COMMENT '投递完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_group` (`group_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关系变更事务性发件箱';
