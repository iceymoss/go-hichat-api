---
name: onboard
description: 帮助新成员上手 go-hichat-api。当用户说新人上手、了解项目、项目介绍时触发。
---

帮助新成员快速理解 go-hichat-api 的架构与开发流程。

## 步骤

1. 读取根 `CLAUDE.md`（若不存在，先建议运行 `/init`）和 `README.md`，提炼项目简介
2. 列出 `apps/*` 各服务的层（`api / rpc / ws / mq / cron / models`）
3. 说明数据流：
   - HTTP：客户端 → `apps/<svc>/api`（go-zero handler/logic）→ 调 `apps/<svc>/rpc/<svc>client` → `apps/<svc>/rpc/internal/logic` → `models`（DB）
   - WS：客户端 → `apps/im/ws`（连接管理 + 心跳 + ack）→ 业务逻辑 → 投递到 Kafka 或直接转发
   - 异步：业务投递 Kafka → `apps/task/mq` 消费 → 写库或回推 ws
4. 说明契约来源：`.api`（HTTP）和 `.proto`（gRPC）是唯一真相，代码用 `goctl` 生成
5. 列出最常见的 5 类任务及对应入口（见输出格式）
6. 列出最重要的 5 条规则（来自 `.claude/rules/`），不复制内容只指路

## 输出格式

```
## 项目简介
（2-3 句）

## 架构数据流
（一段话或一个 ASCII 图）

## 服务速览
| 服务 | 层 | 一句话用途 |

## 开发环境
- 依赖: MySQL / Redis / Etcd / Mongo / Kafka（详见 README）
- 启动: ./hichat2.sh
- 测试: go test ./... -count=1
- 前端: cd web && bun dev

## 日常开发入口
- 新接口: /new-api
- 新 RPC: /new-rpc
- 新表:   /new-model
- 起服务: /run-services
- 跑测试: /gotest
- 审 diff: /goreview

## 必读规则
（5 条最重要的，每条一行 + 链接到 .claude/rules/<file>.md）
```

简洁输出，不展开任何细节，需要时让用户读对应规则文件。
