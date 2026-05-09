---
name: goreview
description: 审查当前 git diff 中的 Go 代码。当用户说审查 Go 代码、review Go、检查代码时触发。
---

审查当前 git diff 中的 Go 代码。

## 步骤

1. 运行 `git diff HEAD` 获取变更
2. 检查以下方面：
   - 错误处理是否完整（有没有吞掉 error）
   - 并发安全性（共享状态是否有锁保护）
   - 资源泄漏（goroutine、连接、文件句柄）
   - 接口设计是否符合 Go 惯例
   - 是否违反 CLAUDE.md 规则（encoding/json、三数据库兼容等）
   - 测试覆盖是否充分
3. 只报告真正的问题，不报告纯风格建议

## 输出格式

按严重程度排序，每个问题一行：
```
SEVERITY | file:line | 描述
```
