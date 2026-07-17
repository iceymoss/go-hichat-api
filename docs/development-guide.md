# go-hichat-api Development Guide

[English](development-guide.md) | [简体中文](development-guide.zh-CN.md)

go-hichat-api is HiChat 2.0. It splits the application into a microservice architecture, improves the social module, restructures chat storage, tracks user presence and message read state, and adds an activity feed.

### Architectural Changes

* Migrated to a microservice architecture
* Separated the frontend and backend

### Improvements

* Improved the social module with friend requests, administrators, and real-time notifications
* Improved file message storage
* Refactored the chat module to fix memory leaks, improve message flows, and decouple chat processing through asynchronous operations
* Improved heartbeat checks and added reliable message delivery through ACK confirmation
* Improved chat history persistence

### New Features

* Message read and unread state
* Friend online status
* Activity feeds with likes, comments, and content blocking

## Set Up the go-zero Toolchain

```shell
# Install the core go-zero tool
go install github.com/zeromicro/go-zero/tools/goctl@latest

# Install the protoc compiler on macOS
brew install protobuf

# Install the protoc compiler on Ubuntu
sudo apt install -y protobuf-compiler

# Install the Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify the installation
goctl -v
protoc --version
```

## Develop a Module

The following examples generate the RPC, API, and model code for the user service.

1. Create the `.proto` file.
2. Generate the RPC code:

   ```shell
   goctl rpc protoc ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
   ```

3. Generate MySQL CRUD code:

   ```shell
   goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models/" -c
   ```

4. Generate a MongoDB model:

   ```shell
   goctl model mongo --type chatLog --dir ./apps/im/models/
   ```

5. Generate API code:

   ```shell
   goctl api go -api apps/user/api/user.api -dir apps/user/api -style gozero
   ```

6. Pass the access token through the HTTP `Authorization` header:

   ```http
   GET /v1/user/detail HTTP/1.1
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

## Required Services

For local development, start MySQL, Redis, Etcd, MongoDB, and Kafka with the standalone Compose file:

```shell
docker compose -f docker-compose.dependencies.yaml up -d --wait
```

This command creates the dependency network and persistent volumes, exposes the services on `localhost`, and waits for every health check to pass. It does not start the application services.

```shell
docker compose -f docker-compose.dependencies.yaml ps      # Show status
docker compose -f docker-compose.dependencies.yaml logs -f # Follow logs
docker compose -f docker-compose.dependencies.yaml down    # Stop and preserve data
```

The default host ports are MySQL `3306`, Redis `6379`, Etcd `2379`, MongoDB `27017`, and Kafka `9092`. Override a port if needed, and use the same port in your local service configuration:

```shell
MYSQL_PORT=3307 docker compose -f docker-compose.dependencies.yaml up -d --wait
```

The individual Docker commands below are retained as reference for custom installations.

#### MySQL

```shell
# Create persistent storage
mkdir -p /docker/mysql/data

# Write the configuration
mkdir -p /docker/mysql/conf
cat > /docker/mysql/conf/my.cnf <<EOF
[mysqld]
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
default_authentication_plugin=mysql_native_password
max_connections=200
innodb_buffer_pool_size=512M
EOF

# Start the service
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

#### Redis

```shell
# Create persistent storage
mkdir -p /docker/redis/data

# Add the configuration
mkdir -p /docker/redis/conf
cat > /docker/redis/conf/redis.conf <<EOF
# Basic settings
bind 0.0.0.0
port 6379
timeout 0
tcp-keepalive 300

# Persistence
save 60 1000
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
dir /data

# Memory management
maxmemory 1gb
maxmemory-policy allkeys-lru

# Security
# requirepass yourpassword  # Uncomment to set a password
EOF

# Start the service
docker run -d \
  --name redis-hichat \
  -p 6379:6379 \
  -v /docker/redis/data:/data \
  -v /docker/redis/conf:/usr/local/etc/redis \
  --restart=always \
  redis:7.0 redis-server /usr/local/etc/redis/redis.conf
```

#### Etcd

```shell
# Create persistent storage
mkdir -p /docker/etcd/data

# Start the service
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

#### Kafka

```shell
# Create the directory
mkdir -p /docker/kafka

# Create docker-compose.yml
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
      # Do not configure KAFKA_CFG_LISTENERS when the broker is not externally accessible
      # - KAFKA_CFG_LISTENERS=PLAINTEXT://0.0.0.0:9092
      # Replace the IP address when exposing the broker externally
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://127.0.0.1:9092
    volumes:
      - kafka_data:/bitnami/kafka
    depends_on:
      - zookeeper

volumes:
  zookeeper_data:
  kafka_data:
EOF

# Start the services
cd /docker/kafka
docker compose up -d
```

#### MongoDB

Use the following commands:

```shell
# Create persistent storage
sudo mkdir -p /docker/mongodb/data
sudo chmod 777 /docker/mongodb/data  # Simplified permissions for this example

# Write the configuration
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

# Start the service
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

For reference, the following shorter command also starts MongoDB:

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

## Start the Services

Run `hichat2.sh` directly:

```shell
./hichat2.sh
```

## Deploy a Docker Image

The following example uses the `user-rpc` service.

Build the image:

```shell
docker build -t hichat2/user-rpc:v1.0 -f deploy/dockerfile/user-rpc.Dockerfile .
```

Start the container:

```shell
docker run -d --name user-rpc --network host -e ENV_MODE=production hichat2/user-rpc:v1.0
```
