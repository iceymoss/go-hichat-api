-- 好友关系表（双向好友关系存储）
CREATE TABLE `friends` (
                           `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                           `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
                           `friend_uid` int(11) unsigned NOT NULL COMMENT '好友的用户ID',
                           `remark` varchar(255) DEFAULT NULL COMMENT '好友备注名（用户自定义）',
                           `add_source` tinyint DEFAULT NULL COMMENT '添加来源（0:未知 1:搜索 2:群组 3:二维码...）',
                           `blacklisted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否拉黑 0否 1是',
                           `moments_permission` tinyint NOT NULL DEFAULT '0' COMMENT '朋友圈权限 0允许 1仅聊天 2屏蔽朋友圈',
                           `notify_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '消息通知开关 1开 0关',
                           `pinned` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否置顶 0否 1是',
                           `muted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否静音 0否 1是',
                           `friend_tags` text COMMENT '好友标签(JSON数组)',
                           `created_at` timestamp NULL DEFAULT NULL COMMENT '好友关系建立时间',
                           PRIMARY KEY (`id`),
                            KEY `idx_user` (`user_id`) COMMENT '用户维度查询索引',
                            UNIQUE KEY `uk_friends_user_friend` (`user_id`,`friend_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

-- 好友请求表（好友申请记录）
CREATE TABLE `friend_requests` (
                                   `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                   `user_id` int(11) unsigned NOT NULL COMMENT '申请人用户ID',
                                   `req_uid` int(11) unsigned NOT NULL COMMENT '被申请人用户ID',
                                   `req_msg` varchar(255) DEFAULT NULL COMMENT '好友申请留言',
                                   `req_time` timestamp NOT NULL COMMENT '申请发起时间',
                                   `handle_result` tinyint DEFAULT NULL COMMENT '处理结果（0:待处理 1:同意 2:拒绝）',
                                   `handle_msg` varchar(255) DEFAULT NULL COMMENT '处理结果备注',
                                   `handled_at` timestamp NULL DEFAULT NULL COMMENT '处理操作时间',
                                   `read_state` tinyint NOT NULL DEFAULT '0' COMMENT '读取状态（0:未读 1:已读）',
                                   `receiver_read` tinyint NOT NULL DEFAULT '0' COMMENT '接收方已读（0:未读 1:已读）',
                                    `sender_read` tinyint NOT NULL DEFAULT '0' COMMENT '发起方已读处理结果（0:未读 1:已读）',
                                    `status` int DEFAULT NULL COMMENT '消息状态（0:已删除 1:正常显示 2:忽略不显示）',
                                    `remark` varchar(64) NOT NULL DEFAULT '' COMMENT '申请人为对方预设的备注',
                                    `active_key` varchar(160) DEFAULT NULL COMMENT '待处理申请唯一键',
                                    PRIMARY KEY (`id`),
                                    KEY `idx_user` (`user_id`) COMMENT '申请人维度索引',
                                    UNIQUE KEY `uk_friend_requests_active_key` (`active_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友请求表';

-- 好友举报表
CREATE TABLE IF NOT EXISTS `friend_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `reporter_uid` bigint unsigned NOT NULL COMMENT '举报人',
  `target_uid` bigint unsigned NOT NULL COMMENT '被举报人',
  `reason` varchar(512) DEFAULT NULL COMMENT '举报原因',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_reporter` (`reporter_uid`),
  KEY `idx_target` (`target_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友举报记录';

-- 群组信息表（群基础信息）
CREATE TABLE `groups` (
                          `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                          `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '群名称',
                          `icon` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '群头像URL',
                          `description` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '群描述',
                          `status` tinyint DEFAULT NULL COMMENT '群状态（0:正常 1:已解散 2:封禁）',
                          `creator_uid` int(11) unsigned NOT NULL COMMENT '群主用户ID',
                          `group_type` int(11) NOT NULL COMMENT '群类型（1:普通群 2:企业群 3:粉丝群...）',
                          `is_verify` tinyint NOT NULL COMMENT '入群验证（0:不需要 1:需要）',
                          `notification` varchar(255) DEFAULT NULL COMMENT '群公告内容',
                          `notification_uid` int(11) unsigned DEFAULT NULL COMMENT '最后更新公告的用户ID',
                          `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                          `updated_at` timestamp NULL DEFAULT NULL COMMENT '最后更新时间',
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组信息表';

-- 群成员表（群成员关系）
CREATE TABLE `group_members` (
                                 `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                 `group_id` int(11) unsigned NOT NULL COMMENT '关联群ID',
                                 `user_id` int(11) unsigned NOT NULL COMMENT '成员用户ID',
                                 `role_level` tinyint NOT NULL COMMENT '成员角色（0:普通成员 1:管理员 2:群主）',
                                 `join_time` timestamp NULL DEFAULT NULL COMMENT '加入群聊时间',
                                 `join_source` tinyint DEFAULT NULL COMMENT '加入来源（1:扫码 2:邀请 3:搜索...）',
                                 `inviter_uid` int(11) unsigned DEFAULT NULL COMMENT '邀请人用户ID',
                                 `operator_uid` int(11) unsigned DEFAULT NULL COMMENT '操作人用户ID',
                                 `group_nickname` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '群内昵称',
                                 `group_remark` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '群备注（仅自己可见）',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_member` (`group_id`,`user_id`) COMMENT '群内成员唯一性约束'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员表';

-- 加群请求表（入群申请记录）
CREATE TABLE `group_requests` (
                                  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                  `req_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '业务请求ID（唯一标识）',
                                  `group_id` int(11) unsigned NOT NULL COMMENT '目标群ID',
                                  `req_msg` varchar(255) DEFAULT NULL COMMENT '入群申请留言',
                                  `req_time` timestamp NULL DEFAULT NULL COMMENT '申请时间',
                                  `join_source` tinyint DEFAULT NULL COMMENT '申请来源（1:扫码 2:邀请 3:搜索...）',
                                  `inviter_user_id` int(11) unsigned DEFAULT NULL COMMENT '邀请人ID',
                                  `handle_user_id` int(11) unsigned DEFAULT NULL COMMENT '请求处理人ID',
                                  `handle_time` timestamp NULL DEFAULT NULL COMMENT '处理时间',
                                  `handle_result` tinyint DEFAULT NULL COMMENT '处理结果（0:待处理 1:同意 2:拒绝）',
                                   `receiver_read` tinyint NOT NULL DEFAULT '0' COMMENT '接收方(群主/管理员)已读 0未读 1已读',
                                   `active_key` varchar(160) DEFAULT NULL COMMENT '主动待处理申请唯一键',
                                   `source_type` tinyint NOT NULL DEFAULT '1' COMMENT '来源类型 1主动申请 2成员邀请',
                                   `source_invitation_id` bigint unsigned DEFAULT NULL COMMENT '来源邀请ID',
                                   `actual_join_source` tinyint DEFAULT NULL COMMENT '最终实际入群来源',
                                   `invalid_reason` varchar(128) NOT NULL DEFAULT '' COMMENT '系统失效原因',
                                   `handle_msg` varchar(255) NOT NULL DEFAULT '' COMMENT '审批附言或拒绝原因',
                                   PRIMARY KEY (`id`),
                                   KEY `idx_group` (`group_id`) COMMENT '群组维度查询索引',
                                   KEY `idx_group_request_lookup` (`group_id`,`req_id`,`handle_result`),
                                   UNIQUE KEY `uk_group_requests_active_key` (`active_key`),
                                   UNIQUE KEY `uk_group_requests_source_invitation` (`source_invitation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加群请求表';

CREATE TABLE `group_invitations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` bigint unsigned NOT NULL,
  `inviter_uid` bigint unsigned NOT NULL,
  `invitee_uid` bigint unsigned NOT NULL,
  `inviter_role_snapshot` tinyint NOT NULL DEFAULT '0',
  `message` varchar(255) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT '0',
  `reject_reason` varchar(255) NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL,
  `handled_at` timestamp NULL DEFAULT NULL,
  `expires_at` timestamp NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_invitation_invitee` (`invitee_uid`,`status`,`created_at`),
  KEY `idx_group_invitation_group_invitee` (`group_id`,`invitee_uid`,`status`),
  KEY `idx_group_invitation_expiry` (`status`,`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员邀请';

CREATE TABLE `social_request_receipts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `request_type` varchar(16) NOT NULL,
  `request_id` bigint unsigned NOT NULL,
  `receiver_id` varchar(64) NOT NULL,
  `receipt_kind` varchar(16) NOT NULL,
  `is_read` tinyint NOT NULL DEFAULT '0',
  `is_actionable` tinyint NOT NULL DEFAULT '0',
  `result` tinyint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL,
  `read_at` timestamp NULL DEFAULT NULL,
  `resolved_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_social_request_receipt` (`request_type`,`request_id`,`receiver_id`,`receipt_kind`),
  KEY `idx_social_receipt_unread` (`receiver_id`,`is_read`,`request_type`),
  KEY `idx_social_receipt_actionable` (`receiver_id`,`is_actionable`,`request_type`),
  KEY `idx_social_receipt_request` (`request_type`,`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='社交申请个人回执';

CREATE TABLE `social_notification_outbox` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `notify_type` varchar(64) NOT NULL,
  `receiver_id` varchar(64) NOT NULL,
  `actor_id` varchar(64) NOT NULL,
  `biz_id` varchar(128) NOT NULL,
  `group_id` varchar(64) NOT NULL DEFAULT '',
  `payload` text NOT NULL,
  `status` tinyint NOT NULL DEFAULT '0',
  `attempts` int NOT NULL DEFAULT '0',
  `next_retry_at` timestamp NULL DEFAULT NULL,
  `last_error` varchar(512) NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL,
  `sent_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_social_notification` (`notify_type`,`receiver_id`,`biz_id`),
  KEY `idx_social_notification_retry` (`status`,`next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='社交通知事务发件箱';

-- 群邀请链接表（链接/二维码统一用 token 表示）
CREATE TABLE IF NOT EXISTS `group_invite_links` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `group_id` int(11) unsigned NOT NULL COMMENT '群ID',
  `token` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邀请token（唯一）',
  `created_by` int(11) unsigned NOT NULL COMMENT '创建人（群成员ID）',
  `expire_at` timestamp NULL DEFAULT NULL COMMENT '过期时间（NULL=永不过期）',
  `max_uses` int unsigned NOT NULL DEFAULT '0' COMMENT '最大可用次数（0=无限）',
  `used_count` int unsigned NOT NULL DEFAULT '0' COMMENT '已使用次数',
  `revoked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否撤销 0否 1是',
  `revoked_at` timestamp NULL DEFAULT NULL COMMENT '撤销时间',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_group` (`group_id`),
  KEY `idx_creator` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群邀请链接表';

-- 群成员资料/设置（群内昵称、群备注等）
CREATE TABLE IF NOT EXISTS `group_member_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `group_id` int(11) unsigned NOT NULL COMMENT '群ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
  `group_nickname` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '我在本群的昵称',
  `group_remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '群备注（仅自己可见）',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`,`user_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员资料/设置';

-- 群公告历史表（支持置顶）
CREATE TABLE IF NOT EXISTS `group_announcements` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `group_id` int(11) unsigned NOT NULL COMMENT '群ID',
  `content` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '公告内容',
  `created_by` int(11) unsigned NOT NULL COMMENT '发布人',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
  `pinned` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否置顶 0否 1是',
  `pinned_at` timestamp NULL DEFAULT NULL COMMENT '置顶时间',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除 0否 1是',
  PRIMARY KEY (`id`),
  KEY `idx_group` (`group_id`),
  KEY `idx_group_pinned` (`group_id`,`pinned`),
  KEY `idx_creator` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群公告历史';
