# HiChat 2.0 一键部署（Docker Compose）

用一条命令把整套 HiChat（6 个微服务 + 中间件 + 前端）跑起来，开箱体验全部功能。

## 先决条件

- 已安装 **Docker** 与 **Docker Compose v2**（`docker compose version` 可用）
- 建议宿主机内存 ≥ 4G（Kafka / MySQL / 前端构建较吃资源）

## 本地一键启动

在项目根目录执行：

```bash
docker compose up -d --build
```

首次会编译所有 Go 服务与前端，耗时较久（取决于机器与网络）。全部就绪后访问：

- 前端： **http://localhost:2470**
- 首次使用点「注册」即可：演示模式下**验证码会自动填入**输入框（无需真实短信），填昵称密码直接注册登录。

> 演示模式说明：未配置真实短信服务商（aliyun/tencent）时，后端把注册验证码随接口返回，前端自动填进验证码框，
> 方便部署体验。生产环境在全局配置里设置 `verification.sms.provider` 为真实服务商后，此行为自动关闭。

查看状态 / 日志、停止：

```bash
docker compose ps              # 查看各服务状态
docker compose logs -f web     # 跟踪某个服务日志
docker compose down            # 停止（保留数据）
docker compose down -v         # 停止并清空所有数据卷（MySQL/Mongo/上传文件等）
```

## 架构与端口

| 类别 | 服务 | 容器内端口 | 是否发布到宿主机 |
|------|------|-----------|----------------|
| 前端 | web (Next.js) | 3001 | ✅ 宿主机 2470 |
| 用户 | user-api / user-rpc | 8887 / 10000 | api ✅ 8887 |
| 社交 | social-api / social-rpc | 8889 / 10001 | api ✅ 8889 |
| IM | im-api / im-rpc / im-ws | 8890 / 10002 / 10090 | api ✅ 8890，ws ✅ 10090 |
| 动态 | trend-api / trend-rpc | 8891 / 10003 | api ✅ 8891 |
| 任务 | task-mq | 10091 | 否（内部消费 Kafka） |
| 音视频 | streaming | 10093 | ✅ 10093 |
| 中间件 | mysql / redis / mongo / etcd / kafka | — | 否（仅容器网络内） |

> 浏览器会直连 `ws://localhost:10090`（IM 长连接）和 `ws://localhost:10093`（音视频信令），
> 上传文件经 `http://localhost:8887/static` 回源，所以这几个端口必须发布——已在 compose 中配置好。

## 这套部署是怎么组织的

- **统一镜像**：所有 Go 服务由 `deploy/dockerfile/service.Dockerfile` 一份参数化 Dockerfile 构建，
  通过 `--build-arg SVC_DIR=apps/<svc>/<layer>` 区分；依赖与源码层在各服务间复用，构建高效。
- **配置隔离**：docker 专用配置集中在 `deploy/docker/`：
  - `config.yaml` —— 全局配置（MySQL/Redis/Mongo 连接），挂载覆盖容器内 `config/config-local.yaml`
  - `etc/*.yaml` —— 各服务 go-zero 配置，主机名指向容器服务名（etcd/kafka/...）
  - 这些文件以**只读卷**挂载进容器，**不改动**仓库里本地开发用的 `*-sample.yaml`。
- **数据库初始化**：`migrate` 一次性容器运行 `deploy/sql_init.go`（GORM AutoMigrate）建好所有 MySQL 表，
  完成后退出；其余服务 `depends_on` 它成功后才启动。MongoDB 集合首次写入自动创建。
- **服务发现**：各 RPC 服务注册到 etcd，调用方经 etcd 发现，无需写死地址。
- **数据持久化**：MySQL/Mongo/Redis/Kafka/etcd 数据与上传文件分别落在命名卷，`down` 不丢、`down -v` 清空。

## 服务器域名部署（可选反向代理）

面向有公网域名的服务器，启用内置 Caddy 统一入口并自动签发 HTTPS：

```bash
cp .env.example .env          # 填写 DOMAIN=your-domain.com
# 把 deploy/docker/etc/im-api.yaml 里 Upload.BaseURL 改为 https://your-domain.com/static
docker compose --profile proxy up -d --build
```

Caddy（`deploy/docker/Caddyfile`）会把单域名路由到：`/ws → im-ws`、`/streaming/* → streaming`、
`/static/* → user-api`、其余 → 前端。前端在非 localhost 访问时自动切换到 `wss://域名/ws`、
`wss://域名/streaming/ws` 的生产模式，无需改前端代码。

## 音视频通话说明

- 同机 / 同局域网：内置公共 STUN 即可直接通话。
- **跨公网**：需自建 TURN 服务器，把地址填入 `deploy/docker/etc/streaming.yaml` 的 `WebRTC.IceServers`。
  另外浏览器使用摄像头/麦克风要求 **HTTPS**（localhost 例外），公网部署请走上面的反向代理 + 域名 HTTPS。

## 常见问题

- **首次启动后端报连不上中间件**：compose 已用 healthcheck + `depends_on` 控制顺序；
  若机器较慢，个别服务会自动重启重连，稍等即可。`docker compose ps` 看健康状态。
- **想重建某个服务**：`docker compose up -d --build <service>`（如 `web`、`user-api`）。
- **端口被占**：修改 `docker-compose.yaml` 里对应服务的 `ports` 左侧宿主机端口。
- **彻底重来**：`docker compose down -v` 清空数据卷后再 `up`。
