-- 好友关系表（双向好友关系存储）
CREATE TABLE `friends` (
                           `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                           `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
                           `friend_uid` int(11) unsigned NOT NULL COMMENT '好友的用户ID',
                           `remark` varchar(255) DEFAULT NULL COMMENT '好友备注名（用户自定义）',
                           `add_source` tinyint DEFAULT NULL COMMENT '添加来源（0:未知 1:搜索 2:群组 3:二维码...）',
                           `created_at` timestamp NULL DEFAULT NULL COMMENT '好友关系建立时间',
                           PRIMARY KEY (`id`),
                           KEY `idx_user` (`user_id`) COMMENT '用户维度查询索引'
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
                                   PRIMARY KEY (`id`),
                                   KEY `idx_user` (`user_id`) COMMENT '申请人维度索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友请求表';

-- 群组信息表（群基础信息）
CREATE TABLE `groups` (
                          `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                          `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '群名称',
                          `icon` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '群头像URL',
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
                                  PRIMARY KEY (`id`),
                                  KEY `idx_group` (`group_id`) COMMENT '群组维度查询索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加群请求表';
