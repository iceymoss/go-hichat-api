
CREATE TABLE `trend_discuss` (
                                 `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                 `trendid` INT NOT NULL DEFAULT 0 COMMENT '所属动态ID（一级评论关联的源动态）',
                                 `father` INT NOT NULL DEFAULT 0 COMMENT '父评论ID（0表示一级评论）',
                                 `rootid` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '根评论ID（多级评论的祖先ID）',
                                 `replyer` INT UNSIGNED NOT NULL COMMENT '评论者用户ID',
                                 `userid` INT UNSIGNED NOT NULL COMMENT '被评论者用户ID',
                                 `level` TINYINT NOT NULL DEFAULT 1 COMMENT '评论层级（1=一级评论，2=二级评论...）',
                                 `content` VARCHAR(2000) NOT NULL COMMENT '评论内容',
                                 `idlist` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '@用户的ID列表',
                                 `agree_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
                                 `discuss_count` INT NOT NULL DEFAULT 0 COMMENT '子评论数量',
                                 `state` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态（0=已删除, 1=正常）',
                                 `read` TINYINT NOT NULL DEFAULT 1 COMMENT '已读状态（0=已读, 1=未读）',
                                  path          VARCHAR(1000)    DEFAULT '/'               NOT NULL COMMENT '评论路径，例如:评论1 (id=1, path=/) 评论2 (id=2, path=/1/) 评论3 (id=3, path=/1/2/)',
                                 `createtime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 `updatetime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',

                                 PRIMARY KEY (`id`),
                                 INDEX `idx_trendid` (`trendid`),
                                 INDEX `idx_father` (`father`),
                                 INDEX `idx_rootid` (`rootid`),
                                 INDEX `idx_replyer` (`replyer`),
                                 INDEX `idx_userid` (`userid`),
                                 INDEX `idx_level` (`level`),
                                 INDEX `idx_state` (`state`),
                                 INDEX `idx_createtime` (`createtime`),
                                 INDEX `idx_trendid_createtime` (`trendid`, `createtime`),
                                 INDEX `idx_father_level` (`father`, `level`),
                                 INDEX `idx_trendid_updatetime` (`trendid`, `updatetime`),
                                 INDEX `idx_updatetime_state` (`updatetime`, `state`),
                                 INDEX `idx_path_level` (`path`(100), `level`),
                                 INDEX  `idx_path` (`path`(255));
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='动态评论表（支持多级嵌套结构）';

-- ID  | trendid | father | rootid | content
-- ----|--------|--------|--------|---------
-- 100 | 888    | 0      | 100    | 第一级评论
-- 101 | 888    | 100    | 100    | 回复第一级评论
-- 102 | 888    | 101    | 100    | 回复回复评论
-- 103 | 888    | 101    | 100    | 另一个回复回复
-- 104 | 888    | 100    | 100    | 另一个回复第一级
-- 200 | 888    | 0      | 200    | 另一个主题评论
-- 201 | 888    | 200    | 200    | 回复另一个主题