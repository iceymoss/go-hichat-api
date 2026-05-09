---
description: Go 后端代码规则
globs: ["*.go"]
---

- JSON 操作用 common.Marshal/Unmarshal，不要导入 encoding/json 做序列化
- 错误必须处理，不要用 _ 忽略 error 返回值
- 数据库操作优先用 GORM 方法，避免原始 SQL
- 原始 SQL 必须兼容 SQLite/MySQL/PostgreSQL 三种数据库
- 保留字列名用 model/main.go 的 commonGroupCol/commonKeyCol
- 布尔值用 commonTrueVal/commonFalseVal
- 新增数据库字段用 TEXT 类型存 JSON，不用 JSONB
- goroutine 注意并发安全，共享状态要加锁
- 资源（连接、文件句柄）必须正确关闭，用 defer
