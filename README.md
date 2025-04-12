# go-hichat-api

## rpc(user为例)
1. 创建proto
2. 生成代码
> goctl rpc protoc ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
> 
3. 生成数据库crud
> goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models/" -c
