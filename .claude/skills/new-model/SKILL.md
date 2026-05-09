---
name: new-model
description: 新增数据库模型/表。当用户说加表、新增 model、加数据库字段时触发。
---

新增数据库表 / 字段。**结构变更前必须先与用户对齐 schema**。

## 步骤

### 1. 与用户对齐 schema（必须）
按 [`database-model.md`](../../rules/database-model.md)，问清楚：
- 表名、字段、类型、索引、是否软删除
- JSON 字段统一用 `TEXT`（不要用 JSONB）
- 主键交给 GORM 处理（不写 AUTO_INCREMENT / SERIAL）
- 三库兼容：MySQL / PostgreSQL / SQLite 都跑得过

### 2. 写 DDL
- MySQL：`deploy/sql/<svc>.sql`（追加新表/`ALTER TABLE`）
- 注意：SQLite 不支持 `ALTER COLUMN`，只能 `ADD COLUMN`
- 保留字列名用 `commonGroupCol` / `commonKeyCol`（见 `model/main.go` 约定）
- 布尔值用 `commonTrueVal` / `commonFalseVal`

### 3. 生成 GORM model
```bash
goctl model mysql ddl -src=./deploy/sql/<svc>.sql -dir=./apps/<svc>/models/ -c
```

如果是 MongoDB：
```bash
goctl model mongo --type <Type> --dir ./apps/<svc>/models/
```

### 4. 迁移 / 初始化
如果项目用代码自动迁移，确认 `AutoMigrate` 列表里加上新模型。
否则在 `deploy/sql/` 提供 idempotent 的 SQL 脚本。

### 5. 数据库分支
需要数据库特定 SQL 时用：
```go
common.UsingPostgreSQL { ... }
common.UsingSQLite     { ... }
common.UsingMySQL      { ... }
```

不要把 MySQL 方言的语法直接写死。

### 6. 测试
- 单元测试不 mock 数据库（[`test-files.md`](../../rules/test-files.md)）
- 测试结束清理数据
- 三种数据库环境都跑通

## 严格约束

- **绝对不**修改 `.env`
- **绝对不**直接 drop / truncate 现有表
- 新增字段如果是 NOT NULL，必须给默认值（兼容旧行）
- 涉及生产数据的迁移必须在 PR 描述里说明回滚方案
