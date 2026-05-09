---
description: 测试文件规则
globs: ["*_test.go"]
---

- 测试函数命名用 TestXxx 格式
- 使用 table-driven tests 组织多个测试用例
- 测试必须能在三种数据库环境下通过
- 不要 mock 数据库，除非有明确理由
- 测试数据在测试结束后清理
