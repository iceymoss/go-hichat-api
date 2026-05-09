---
name: new-rpc
description: 在已有 go-zero 服务里新增一个 RPC 方法。当用户说加 rpc 方法、新增 grpc、新加 proto 接口时触发。
---

在 `apps/<svc>/rpc` 里新增 gRPC 方法。

## 步骤

### 1. 确认输入
- 哪个服务
- 方法名（驼峰，如 `GetUserByPhone`）
- 请求 / 响应消息字段
- 是否会被多个服务调用（影响以后的兼容性策略）

### 2. 修改 `.proto`
编辑 `apps/<svc>/rpc/<svc>.proto`：
- **新方法只能加在 service 末尾**，不要插队（避免序号变动）
- **新字段在 message 末尾追加**，并保留旧字段编号（向后兼容）
- 不要随意改字段编号或类型；废弃字段加 `// deprecated`，不要删

### 3. 重新生成
```bash
cd apps/<svc>/rpc && goctl rpc protoc ./<svc>.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

会生成 `internal/logic/<xxx>logic.go` 骨架。已存在的方法不会被覆盖。

### 4. 写 logic
`internal/logic/<xxx>logic.go`：
- DB 操作走 `models/`，优先 GORM（[`go-backend.md`](../../rules/go-backend.md)）
- 原始 SQL 必须三库兼容（[`database-model.md`](../../rules/database-model.md)）
- 错误用 `pkg/xerr`

### 5. 在调用方注入 client
被调服务的 `<svc>client` 已生成在 `apps/<svc>/rpc/<svc>client/`。在调用方：
- 配置里加 `XxxRpc zrpc.RpcClientConf`
- `internal/svc/servicecontext.go` 里 `XxxRpc: <svc>.New<Svc>(zrpc.MustNewClient(c.XxxRpc))`
- `internal/config/config.go` 加字段

### 6. 测试
- RPC 单元测试：在 `internal/logic/*_test.go` 用 table-driven
- 集成：起 rpc + 调用方，跑端到端

## 严格约束

- proto 字段**只追加**，不重排不删除编号
- 废弃方法保留 stub，不要从 proto 直接删（外部可能仍在调）
- 重新生成 proto 后**必须** `git diff` 检查 `*_grpc.pb.go` 是否引入意外破坏性变更
