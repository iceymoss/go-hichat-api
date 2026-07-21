# go-hichat-api

[English](development-guide.md) | [简体中文](development-guide.zh-CN.md)

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

本地开发时，可以使用独立的 Compose 文件一键启动 MySQL、Redis、Etcd、MongoDB 和 Kafka：

```shell
docker compose -f docker-compose.dependencies.yaml up -d --wait
```

该命令会创建依赖网络和持久卷，将各组件端口映射到 `localhost`，并等待所有健康检查通过。它不会启动项目业务服务。

```shell
docker compose -f docker-compose.dependencies.yaml ps      # 查看状态
docker compose -f docker-compose.dependencies.yaml logs -f # 持续查看日志
docker compose -f docker-compose.dependencies.yaml down    # 停止并保留数据
```

默认映射到宿主机的端口为 MySQL `3306`、Redis `6379`、Etcd `2379`、MongoDB `27017` 和 Kafka `9092`。如有需要，可以覆盖对应端口，并在本地服务配置中使用相同端口：

```shell
MYSQL_PORT=3307 docker compose -f docker-compose.dependencies.yaml up -d --wait
```

下面的独立 Docker 命令保留作为自定义安装参考。

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
      # - KAFKA_CFG_LISTENERS=PLAINTEXT://0.0.0.0:9092 不对外使用不配置
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092 # 对外配置填写对应的ip
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
