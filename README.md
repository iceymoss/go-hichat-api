# go-hichat-api

## rpc(user为例)
1. 创建proto
2. 生成代码
> goctl rpc protoc ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
> 
3. 生成数据库crud
> goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models/" -c
4. 生成api
> goctl api go -api apps/user/api/user.api -dir apps/user/api -style gozero
5. token验证方式
> 通过http header传递
> 例如：
> GET /v1/user/detail HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...