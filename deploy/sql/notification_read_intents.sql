-- Durable business notification read intents. The unique key is also the
-- serialization point between mark-before-insert and insert-before-mark.
CREATE TABLE IF NOT EXISTS `notification_read_intents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `receiver_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `notify_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `biz_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_notification_read_intent` (`receiver_id`,`notify_type`,`biz_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Durable notification read intents';
