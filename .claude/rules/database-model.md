---
description: 数据模型和数据库迁移规则
globs: ["model/**/*.go"]
---

- 数据库 schema 变更前必须先和用户确认
- 迁移必须同时兼容 SQLite/MySQL/PostgreSQL
- SQLite 不支持 ALTER COLUMN，只能用 ADD COLUMN
- 不要使用 AUTO_INCREMENT 或 SERIAL，让 GORM 处理主键
- JSON 字段存储用 TEXT 类型，不用 JSONB
- 使用 common.UsingPostgreSQL/UsingSQLite/UsingMySQL 做数据库特定分支
- 不要修改 .env 文件
