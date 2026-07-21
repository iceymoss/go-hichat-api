---
name: new-service
description: 在 apps/ 下新建一个 go-zero 微服务（rpc + api + model）。当用户说新服务、新增微服务、scaffold 服务时触发。
---

按 go-hichat-api 的约定脚手架一个新微服务。**不假设需求**，先访谈再生成。

## 步骤

### 1. 访谈（用 AskUserQuestion，逐个问）
- 服务名（小写单词，如 `payment`、`notify`）
- 需要哪几层：`rpc` / `api` / `ws` / `mq` / `cron`（多选）
- 主要数据存在哪里：MySQL / MongoDB / Redis / 多种
- 是否需要被其它服务通过 RPC 调用（决定要不要 .proto + etcd 注册）
- 对外暴露 HTTP 路径前缀（如 `v1/<svc>`）

### 2. 参考已有服务
读 `apps/user/` 或 `apps/trend/` 作为模板：
- `<svc>.proto` / `<svc>.api` 的写法
- `etc/*-sample.yaml` 的字段
- `internal/{config,handler,logic,svc,types}` 目录骨架

### 3. 创建文件骨架
按 go-zero 约定，**用 `goctl` 生成，不要手写骨架**：

```bash
# 1) 创建 .proto / .api
mkdir -p apps/<svc>/{rpc,api,models}
# 写 apps/<svc>/rpc/<svc>.proto 和 apps/<svc>/api/<svc>.api（参考同类服务）

# 2) 生成 RPC
cd apps/<svc>/rpc && goctl rpc protoc ./<svc>.proto --go_out=. --go-grpc_out=. --zrpc_out=.

# 3) 生成 API
cd apps/<svc>/api && goctl api go -api <svc>.api -dir . -style gozero

# 4) 生成 model（如有 SQL）
goctl model mysql ddl -src=./deploy/sql/<svc>.sql -dir=./apps/<svc>/models/ -c
```

### 4. 注册到启动脚本
在 `hichat2.sh` 的 `SERVICES=(...)` 里加上：
```
"rpc <svc>"
"api <svc>"
```

### 5. 写最小 sample yaml
- `apps/<svc>/rpc/etc/<svc>-sample.yaml`（含 etcd 注册）
- `apps/<svc>/api/etc/<svc>-sample.yaml`（含 JWT 配置 + 上游 RPC 地址）
- 参考其它服务，**端口必须不冲突**（grep 现有 yaml 找空闲端口）

### 6. 跨服务调用
如果该服务要调其它服务，按 [`.claude/rules/rpc-client.md`](../../rules/rpc-client.md) 的方式注入对方 `xxxclient`，不要直连数据库。

### 7. 收尾
- 跑 `go build ./...` 确认无编译错误
- 跑 `./hichat2.sh` 确认服务起得来
- 提示用户去 `/spec` 写第一个具体功能

## 严格约束

- **不要**绕过 goctl 手写 handler / logic 骨架
- **不要**直接读其它服务的数据库表
- **不要**为了 demo 把端口写死成已有服务的端口
- 新增 schema 必须先与用户确认（[`database-model.md`](../../rules/database-model.md)）
