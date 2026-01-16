-- 为 friend_requests 表添加 status 字段
-- status: 0-已删除, 1-正常（显示消息气泡）, 2-忽略（不显示消息气泡）
ALTER TABLE `friend_requests` 
ADD COLUMN `status` tinyint DEFAULT 1 COMMENT '消息状态（0:已删除 1:正常显示 2:忽略不显示）' AFTER `req_msg`;

-- 为现有数据设置默认值
UPDATE `friend_requests` SET `status` = 1 WHERE `status` IS NULL;

