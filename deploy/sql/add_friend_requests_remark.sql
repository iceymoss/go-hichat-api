-- 好友申请：申请人为对方预设的备注（对方通过后写入申请人侧好友关系的 remark）
-- 兼容 MySQL；SQLite/PostgreSQL 同样支持 ADD COLUMN
ALTER TABLE friend_requests ADD COLUMN remark VARCHAR(64) NOT NULL DEFAULT '' COMMENT '申请人为对方预设的备注';
