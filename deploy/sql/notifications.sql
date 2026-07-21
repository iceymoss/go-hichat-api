-- 公共通知表（业务方投递通知事件 -> task/mq 消费 -> im-rpc 落本表 -> 在线推 ws）
-- 幂等键 (receiver_id, notify_type, biz_id) 唯一，对付 Kafka 重投与消费重试
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `receiver_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收者UID',
  `notify_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知类型 friend.apply/group.accept...',
  `biz_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '业务关联ID(幂等键)',
  `actor_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '触发者UID',
  `group_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '群ID(群相关通知)',
  `title` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标题(可空,前端也可用type+i18n渲染)',
  `content` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '正文摘要',
  `payload` text COLLATE utf8mb4_unicode_ci COMMENT '扩展字段JSON',
  `is_read` tinyint NOT NULL DEFAULT '0' COMMENT '0未读 1已读',
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `read_at` timestamp NULL DEFAULT NULL COMMENT '已读时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recv_type_biz` (`receiver_id`,`notify_type`,`biz_id`),
  KEY `idx_recv_read` (`receiver_id`,`is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公共通知表';
