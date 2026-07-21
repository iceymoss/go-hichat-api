---
name: new-api
description: 在已有 go-zero 服务里新增一个 HTTP 接口。当用户说加接口、新增 api、加路由时触发。
---

在 `apps/<svc>/api` 里新增一个 HTTP 接口，使用 goctl 重新生成。

## 步骤

### 1. 确认输入
- 哪个服务（`apps/` 下选一个）
- 路径与方法（如 `POST /v1/im/conversation/mute`）
- 是否需要 JWT
- 请求 / 响应字段
- 业务逻辑要不要调用某个 RPC（如果要，确认对应 `xxxclient` 已经注入到 `svc.ServiceContext`）

### 2. 修改 `.api`

编辑 `apps/<svc>/api/<svc>.api`：
- 在 `type (...)` 块里加 `XxxReq` / `XxxResp` —— **可选字段加 `optional`**，可选标量在 Go 端用指针 + omitempty
- 在对应 `@server` 块下加路由：

```
@doc "中文描述"
@handler xxxHandler
post /v1/<svc>/path (XxxReq) returns (XxxResp)
```

### 3. 重新生成代码

```bash
goctl api go -api apps/<svc>/api/<svc>.api -dir apps/<svc>/api -style gozero
```

goctl 只会生成**缺失的** handler/logic 骨架，不会覆盖你已有的实现。生成完检查 `internal/handler/routes.go` 已加路由。

### 4. 写 logic
在 `apps/<svc>/api/internal/logic/<xxx>logic.go` 写业务：
- 调用 RPC：`l.svcCtx.UserRpc.GetUser(l.ctx, &user.GetUserReq{...})`
- 错误用 `pkg/xerr` 封装，不要返回原始 error 给客户端
- JSON 序列化用 `common.Marshal/Unmarshal`（[`go-backend.md`](../../rules/go-backend.md)）

### 5. 测试
- 单元测试：在 `internal/logic/<xxx>logic_test.go` 写 table-driven 测试
- 手工：`go run apps/<svc>/api/<svc>.go -f apps/<svc>/api/etc/<svc>-sample.yaml`，再 curl

### 6. 同步文档
跑 `/sync-api-docs` 更新 `docs/api.md`。

## 严格约束

- 改 `.api` 后**必须**重跑 goctl，不要手写 handler
- 业务**只**在 logic 层；handler 只做参数绑定 + 调 logic
- 跨服务调用走 RPC，不要直接读对方数据库（[`microservice.md`](../../rules/microservice.md)）
- 可选请求字段：DTO 用指针 + omitempty
