CREATE TABLE trend (
                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '动态ID',
                       userid INT UNSIGNED NOT NULL COMMENT '发表用户ID',
                       type TINYINT UNSIGNED NOT NULL COMMENT '动态类型：1文本，2混合(图片)，3长文，4第三方分享(如B站视频)，5视频，6置顶广告',
                       content TEXT NOT NULL COMMENT '动态内容',
                       position VARCHAR(255) DEFAULT '' COMMENT '位置信息',
                       reply_count INT UNSIGNED DEFAULT 0 COMMENT '评论数量',
                       agree_count INT UNSIGNED DEFAULT 0 COMMENT '点赞数量',
                       createtime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '原始创建时间',
                       updatetime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',

    -- 双平台状态控制 --
                       circle_state TINYINT(1) NOT NULL DEFAULT 1 COMMENT '朋友圈状态：0-不可见，1-可见，2-朋友圈删除',
                       public_state TINYINT(1) NOT NULL DEFAULT 0 COMMENT '公共论坛状态：0-未发布，1-已发布，2-审核中，3-论坛删除',

    -- 双平台时间管理 --
                       circle_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '朋友圈发布时间',
                       public_time DATETIME DEFAULT NULL COMMENT '公共论坛发布时间',

    -- 基础状态 --
                       is_ad TINYINT DEFAULT 0 COMMENT '是否广告：0-普通，1-广告',
                       url VARCHAR(255) DEFAULT '' COMMENT '广告/视频链接（类型5使用）',
                       ad_end_time DATETIME DEFAULT NULL COMMENT '广告展示截止时间',
                       open_reply TINYINT(1) DEFAULT 1 COMMENT '是否开启评论：0-关闭，1-开启',
                       is_top TINYINT(1) DEFAULT 0 COMMENT '是否置顶：0-否，1-是',
                       title VARCHAR(255) DEFAULT '' COMMENT '长文标题（类型3使用）',
                       idlist TEXT DEFAULT NULL COMMENT '@用户ID列表（使用JSON数组）',
                       pic_sort TEXT DEFAULT NULL COMMENT '图片ID排序',
                       share_id INT DEFAULT 0 COMMENT '第三方内容ID（类型4使用）',
                       cover VARCHAR(255) DEFAULT '' COMMENT '封面图URL（类型3/5使用）',
                       ip VARCHAR(45) DEFAULT '' COMMENT '发布者IP地址',
                       device VARCHAR(100) DEFAULT '' COMMENT '发布设备标识',

    -- 索引优化 --
                       PRIMARY KEY (id),
                       INDEX idx_userid_circle (userid, circle_time) COMMENT '用户朋友圈动态索引',
                       INDEX idx_userid_public (userid, public_time) COMMENT '用户公开动态索引',
                       INDEX idx_circle_time (circle_time) COMMENT '朋友圈时间排序索引',
                       INDEX idx_public_time (public_time) COMMENT '论坛时间排序索引',
                       INDEX idx_public_state (public_state) COMMENT '论坛状态索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='双平台社交动态表（支持朋友圈+公共论坛同步发布）';