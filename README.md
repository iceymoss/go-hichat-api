# go-hichat-api
go-hichat-api是HiChat的2.0版本，其模块拆分，使用微服务架构，功能点：优化社交模块、记录重构聊天存储项目、添加用户在线/离线，消息已读/未读状态、添加动态空间模块。

## 如何快速进行模块开发
生成代码模块rpc/api/model(user为例)

1. 创建proto
2. 生成代码
> goctl rpc protoc ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
> 
3. 生成数据库crud(mysql)
> goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models/" -c

4. 生成数据库模型(mongo)
> goctl model mongo --type chatLog --dir ./apps/im/models/

5. 生成api
> goctl api go -api apps/user/api/user.api -dir apps/user/api -style gozero
6. token验证方式
> 通过http header传递
> 例如：
> GET /v1/user/detail HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...


## 如何启动
直接运行hichat2.sh启动
```shell
./hichat2.sh
```
