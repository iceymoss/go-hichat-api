---
name: catchup
description: 恢复上下文。在 /clear 后或新 session 开始时使用。当用户说恢复上下文、继续之前的工作、catchup 时触发。
---

恢复上下文。在 /clear 后或新 session 开始时使用。

## 步骤

1. 检查 PROGRESS.md 是否存在，如存在则读取
2. 运行 `git log --oneline -20` 查看最近提交
3. 运行 `git diff --stat HEAD~5` 查看最近变更的文件
4. 运行 `git status` 查看当前工作状态
5. 运行 `git branch -a` 查看分支状态

## 输出格式

```
## 当前状态
- 分支: xxx
- 最近工作: （从 git log 和 PROGRESS.md 总结）
- 未完成的事: （从 PROGRESS.md 或未提交变更推断）

## 建议下一步
- （基于上下文给出 1-3 个建议）
```

简洁输出，不要重复 git log 的原始内容。
