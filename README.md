# go-hichat-api
go-hichat-api是HiChat的2.0版本，其模块拆分，使用微服务架构，功能点：优化社交模块、记录重构聊天存储项目、添加用户在线/离线，消息已读/未读状态、添加动态空间模块。
### 调整点
* 调整为微服务架构
* 项目前后端分离

### 优化点
* 优化社交模块，添加或者好友申请，管理员，以及相应消息实时通知
* 优化文件消息存储方式
* 重构聊天模块，修复内存泄漏问题，优化消息流，解耦和异步话聊天模块
* 优化心跳检查，添加消息可靠性ack确认机制
* 完善聊天记录持久化

### 新增功能点
* 添加消息已读/未读功能
* 添加好友在线状态
* 添加动态空间模块，点赞，评论，屏蔽动态等

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

## 需要的配置
#### mysql
1. 创建一个持久化目录
   ```
   mkdir -p /docker/mysql/data
   ```
2. 写入配置
   ```
   mkdir -p /docker/mysql/conf
   cat > /docker/mysql/conf/my.cnf <<EOF
   [mysqld]
   character-set-server=utf8mb4
   collation-server=utf8mb4_unicode_ci
   default_authentication_plugin=mysql_native_password
   max_connections=200
   innodb_buffer_pool_size=512M
   EOF
   ```

3. 启动服务
   ```
   docker run -d \
   --name mysql-hichat2 \
   -p 3306:3306 \
   -e MYSQL_ROOT_PASSWORD=123456789 \
   -e MYSQL_DATABASE=hichat2 \
   -v /docker/mysql/data:/var/lib/mysql \
   -v /docker/mysql/conf:/etc/mysql/conf.d \
   --restart=always \
   mysql:8.0
   ```

#### redis

#### etcd

#### kafka

#### mongo
MongoDB：
```shell
docker run -d \
  --name mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=hichat2 \
  -v "/Users/iceymoss/docker-volume/mongo-data:/data/db" \
  --restart always \
  mongo:4.0
```

## 如何启动
直接运行hichat2.sh启动
```shell
./hichat2.sh
```


