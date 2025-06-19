CREATE TABLE `trend_agree` (
                               `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `userid` INT UNSIGNED NOT NULL COMMENT '点赞用户ID',
                               `author_id` INT UNSIGNED NOT NULL COMMENT '被赞动态的作者ID',
                               `trend_id` INT UNSIGNED NOT NULL COMMENT '动态ID',
                               `op_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间（点赞/取消时间）',
                               `state` TINYINT UNSIGNED NOT NULL DEFAULT '1' COMMENT '点赞状态 0:取消，1:有效',
                               `is_read` TINYINT UNSIGNED NOT NULL DEFAULT '1' COMMENT '已读状态 0:已读，1:未读',
                               `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
                               `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `uniq_user_trend` (`userid`, `trend_id`),
                               KEY `idx_author_read_state` (`author_id`, `is_read`, `state`),
                               KEY `idx_trend_state` (`trend_id`, `state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态点赞表';