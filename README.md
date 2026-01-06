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


## GO-ZERO框架配置搭建
```sql
# 安装 Go-Zero 核心工具
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 安装 protoc 编译器 (macOS)
brew install protobuf

# 安装 protoc 编译器 (Ubuntu)
sudo apt install -y protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 验证安装
goctl -v
protoc --version
```

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
```
# 创建一个持久化目录
mkdir -p /docker/mysql/data

# 写入配置
mkdir -p /docker/mysql/conf
cat > /docker/mysql/conf/my.cnf <<EOF
[mysqld]
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
default_authentication_plugin=mysql_native_password
max_connections=200
innodb_buffer_pool_size=512M
EOF


# 启动服务
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
```
# 创建持久化目录
mkdir -p /docker/redis/data

# 添加配置
mkdir -p /docker/redis/conf
cat > /docker/redis/conf/redis.conf <<EOF
# 基本配置
bind 0.0.0.0
port 6379
timeout 0
tcp-keepalive 300

# 持久化配置
save 60 1000
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
dir /data

# 内存管理
maxmemory 1gb
maxmemory-policy allkeys-lru

# 安全设置
# requirepass yourpassword  # 取消注释设置密码
EOF

# 启动服务
docker run -d \
  --name redis-hichat \
  -p 6379:6379 \
  -v /docker/redis/data:/data \
  -v /docker/redis/conf:/usr/local/etc/redis \
  --restart=always \
  redis:7.0 redis-server /usr/local/etc/redis/redis.conf
```

#### etcd
```
# 持久化目录
mkdir -p /docker/etcd/data

# 启动服务
docker run -d \
  --name etcd-hichat \
  -p 2379:2379 \
  -p 2380:2380 \
  -v /docker/etcd/data:/etcd-data \
  --restart=always \
  quay.io/coreos/etcd:v3.5.0 \
  /usr/local/bin/etcd \
  --data-dir=/etcd-data \
  --name=etcd-single \
  --initial-advertise-peer-urls=http://127.0.0.1:2380 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --listen-client-urls=http://0.0.0.0:2379 \
  --advertise-client-urls=http://127.0.0.1:2379 \
  --initial-cluster=etcd-single=http://127.0.0.1:2380

```
#### kafka
```
# 创建目录
mkdir -p /docker/kafka

# 创建 docker-compose.yml
cat > /docker/kafka/docker-compose.yml <<EOF
version: '3.8'

services:
  zookeeper:
    image: bitnami/zookeeper:3.8
    container_name: zookeeper
    ports:
      - "2181:2181"
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes
    volumes:
      - zookeeper_data:/bitnami/zookeeper

  kafka:
    image: bitnami/kafka:3.7
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
      - ALLOW_PLAINTEXT_LISTENER=yes
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092
    volumes:
      - kafka_data:/bitnami/kafka
    depends_on:
      - zookeeper

volumes:
  zookeeper_data:
  kafka_data:
EOF

# 启动服务
cd /docker/kafka
docker compose up -d
```

#### mongo
使用以下命令：
```
# 创建持久化
sudo mkdir -p /docker/mongodb/data
sudo chmod 777 /docker/mongodb/data  # 简化权限

# 写配置
sudo mkdir -p /docker/mongodb/conf
sudo tee /docker/mongodb/conf/mongod.conf <<EOF
storage:
  dbPath: /data/db
  journal:
    enabled: true

systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log

net:
  port: 27017
  bindIp: 0.0.0.0

security:
  authorization: enabled
EOF

# 启动服务
docker run -d \
  --name mongodb-hichat \
  -p 27017:27017 \
  -v /docker/mongodb/data:/data/db \
  -v /docker/mongodb/conf:/etc/mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=hichat2 \
  --restart=always \
  mongo:6.0 \
  --config /etc/mongodb/mongod.conf
```

仅做参考：
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

## docker镜像部署
这里以user-rpc服务为例

构建镜像
```
docker build -t hichat2/user-rpc:v1.0 -f deploy/dockerfile/user-rpc.Dockerfile .
```

启动镜像
```
docker run -d   --name user-rpc   --network host   -e ENV_MODE=production   hichat2/user-rpc:v1.0
```

你的理解：
当前好友设置的接口（备注、删除、拉黑、朋友圈权限、通知、置顶、静音、标签、举报等）在 API 层直接用了 db.GetMysqlConn，这是不符合“API 仅调用 RPC，不直接访问 DB”的规范，需要改为调用 social-rpc。当前 social.proto 里没有这些 RPC，必须先扩展 proto，再 goctl 生成，再调整 API 逻辑。
建议调整方案（待你确认后执行）：
在 social.proto 新增 RPC：
FriendUpdateRemark (friend_uid, remark)
FriendDelete (friend_uid) —— 双向删除
FriendBlock (friend_uid, block)
FriendMomentsPermission (friend_uid, permission)
FriendNotification (friend_uid, enabled)
FriendPin (friend_uid, pinned)
FriendMute (friend_uid, muted)
FriendTags (friend_uid, tags[])
FriendReport (friend_uid, reason)
FriendShare (friend_uid, target_uid)（预留）
RPC 返回简单 ack（空 resp），由 RPC 内部访问 MySQL。
social-rpc 内实现上述逻辑，使用 socialmodels 操作 DB。
social-api 的 handler/logic 改为调用对应 RPC，移除所有 db.GetMysqlConn。
需要你确认：
1) 是否按上面这些 RPC 名称/参数扩展 social.proto？有无增减字段要求？
2) 允许我在确认后运行 goctl 生成（请告知命令规范，如 goctl rpc protoc social.proto --go_out=./ --go-grpc_out=./ --zrpc_out=./apps/social/rpc）？
3) 是否保留分享/举报为预留成功返回（无存储/推送），还是需要存表/触发消息？
   确认后我会：
   更新 social.proto
   goctl 生成（在你同意的命令下）
   实现 social-rpc 逻辑
   更新 social-api 调用新 RPC，彻底去掉 DB 直连
   快速自检编译 through go run social-api.

我的回复：OK的，你请继续：需要注意的是，涉及到goctl命令时，你应该停止思考让我来执行命令，我执行成功后，会同步给你，然后让你进行开发，
并且我已经执行了：goctl rpc protoc apps/social/rpc/social.proto --go_out=./apps/social/rpc --go-grpc_out=./apps/social/rpc --zrpc_out=./apps/social/rpc --style=goZero




你是一个高级全栈开发工程师，你需要了解当前项目，并且重点了解user和social模块的前后端功能，然后这个是之前我们为完成的工作，你需要继续完成我们的工作：你的理解：
当前好友设置的接口（备注、删除、拉黑、朋友圈权限、通知、置顶、静音、标签、举报等）在 API 层直接用了 db.GetMysqlConn，这是不符合“API 仅调用 RPC，不直接访问 DB”的规范，需要改为调用 social-rpc。当前 social.proto 里没有这些 RPC，必须先扩展 proto，再 goctl 生成，再调整 API 逻辑。
建议调整方案（待你确认后执行）：
在 social.proto 新增 RPC：
FriendUpdateRemark (friend_uid, remark)
FriendDelete (friend_uid) —— 双向删除
FriendBlock (friend_uid, block)
FriendMomentsPermission (friend_uid, permission)
FriendNotification (friend_uid, enabled)
FriendPin (friend_uid, pinned)
FriendMute (friend_uid, muted)
FriendTags (friend_uid, tags[])
FriendReport (friend_uid, reason)
FriendShare (friend_uid, target_uid)（预留）
RPC 返回简单 ack（空 resp），由 RPC 内部访问 MySQL。
social-rpc 内实现上述逻辑，使用 socialmodels 操作 DB。
social-api 的 handler/logic 改为调用对应 RPC，移除所有 db.GetMysqlConn。
需要你确认：
1) 是否按上面这些 RPC 名称/参数扩展 social.proto？有无增减字段要求？
2) 允许我在确认后运行 goctl 生成（请告知命令规范，如 goctl rpc protoc social.proto --go_out=./ --go-grpc_out=./ --zrpc_out=./apps/social/rpc）？
3) 是否保留分享/举报为预留成功返回（无存储/推送），还是需要存表/触发消息？
   确认后我会：
   更新 social.proto
   goctl 生成（在你同意的命令下）
   实现 social-rpc 逻辑
   更新 social-api 调用新 RPC，彻底去掉 DB 直连
   快速自检编译 through go run social-api.

我的回复：OK的，你请继续：需要注意的是，涉及到goctl命令时，你应该停止思考让我来执行命令，我执行成功后，会同步给你，然后让你进行开发，
并且我已经执行了：goctl rpc protoc apps/social/rpc/social.proto --go_out=./apps/social/rpc --go-grpc_out=./apps/social/rpc --zrpc_out=./apps/social/rpc --style=goZero