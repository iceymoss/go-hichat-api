# HiChat 2.0 — One-Click Deploy (Docker Compose)

**English** | [简体中文](README.zh-CN.md)

Bring up the entire HiChat stack (6 microservices + middleware + web client) with a single command and try every feature out of the box.

## Prerequisites

- **Docker** and **Docker Compose v2** installed (`docker compose version` works)
- Host memory ≥ 4 GB recommended (Kafka / MySQL / the frontend build are resource-hungry)

## Clone & run (from scratch)

```bash
git clone https://github.com/iceymoss/go-hichat-api.git
cd go-hichat-api
docker compose up -d --build
```

The first run compiles all Go services and the frontend (takes a while, depending on machine and network). Once everything is up:

- Web: **http://localhost:2470**
- On first use click **Register** — in demo mode the verification code is **auto-filled** into the input box (no real SMS needed); just enter a nickname and password to sign up and log in.

> Demo mode: when no real SMS provider (aliyun/tencent) is configured, the backend returns the registration code in the API response and the frontend fills it into the code box automatically. Setting `verification.sms.provider` to a real provider in the global config turns this off.

Status / logs / stop:

```bash
docker compose ps              # service status
docker compose logs -f web     # follow a service's logs
docker compose restart web     # restart one service
docker compose down            # stop (keep volumes and images)
docker compose down -v         # stop and wipe all data volumes (MySQL/Mongo/uploads...)
```

## One-shot cleanup / uninstall

```bash
# 1) Stop and remove containers, networks and data volumes (wipes all data)
docker compose down -v --remove-orphans

# 2) Remove the images this project built (the hichat-* set: services + migrate + web)
docker images 'hichat-*' -q | xargs -r docker rmi

# 3) Optional: reclaim BuildKit build cache / disk
docker builder prune -f
```

> Step 2 only removes this project's `hichat-*` app images; shared base images (mysql/redis/mongo/etcd/kafka) are left intact (other projects may reuse them).
> To also drop those, use `docker compose down -v --rmi all --remove-orphans` (note: this removes the middleware images too).


## Architecture & ports

| Tier | Service | Container port | Published to host |
|------|---------|----------------|-------------------|
| Web | web (Next.js) | 3001 | ✅ host 2470 |
| User | user-api / user-rpc | 8887 / 10000 | api ✅ 8887 |
| Social | social-api / social-rpc | 8889 / 10001 | api ✅ 8889 |
| IM | im-api / im-rpc / im-ws | 8890 / 10002 / 10090 | api ✅ 8890, ws ✅ 10090 |
| Trend | trend-api / trend-rpc | 8891 / 10003 | api ✅ 8891 |
| Task | task-mq | 10091 | no (internal Kafka consumer) |
| Streaming | streaming | 10093 | ✅ 10093 |
| Middleware | mysql / redis / mongo / etcd / kafka | — | no (container network only) |

> The browser connects directly to `ws://localhost:10090` (IM long connection) and `ws://localhost:10093` (audio/video signaling), and fetches uploads via `http://localhost:8887/static`, so those ports must be published — already configured in compose.
>
> Custom host ports: copy `.env.example` to `.env` and edit `WEB_PORT` / `USER_API_PORT` / etc. (defaults match the table above).

## How the deployment is organized

- **Unified image**: every Go service is built from a single parameterized Dockerfile, `deploy/dockerfile/service.Dockerfile`, selected via `--build-arg SVC_DIR=apps/<svc>/<layer>`. Dependency and source layers are shared across services, and BuildKit cache mounts compile dependencies only once.
- **Config isolation**: Docker-specific config lives under `deploy/docker/`:
  - `config.yaml` — global config (MySQL/Redis/Mongo connections), mounted over the container's `config/config-local.yaml`
  - `etc/*.yaml` — each service's go-zero config, with hostnames pointing at container service names (etcd/kafka/...)
  - These are mounted **read-only**, leaving the repo's local-dev `*-sample.yaml` untouched.
- **Startup order**: strictly mirrors `hichat2.sh` —
  `user-rpc → user-api → social-rpc → social-api → im-rpc → im-api → im-ws → task-mq → trend-rpc → trend-api → streaming → web`.
  Each backend service has a port health check and downstream services `depends_on` upstream `service_healthy`, so a service starts only after the ones it calls (RPC/WS) are actually ready.
- **Database init**: a one-shot `migrate` container runs `deploy/sql_init.go` (GORM AutoMigrate) to create all MySQL tables, then exits; other services `depends_on` it completing. MongoDB collections are created lazily on first write.
- **Service discovery**: every RPC service registers in etcd; callers discover via etcd — no hardcoded addresses.
- **Persistence**: MySQL/Mongo/Redis/Kafka/etcd data and uploaded files live in named volumes — kept across `down`, wiped by `down -v`.

## Server / domain deployment (optional reverse proxy)

For a server with a public domain, enable the built-in Caddy single entrypoint with automatic HTTPS:

```bash
cp .env.example .env          # set DOMAIN=your-domain.com
# change Upload.BaseURL in deploy/docker/etc/im-api.yaml to https://your-domain.com/static
docker compose --profile proxy up -d --build
```

Caddy (`deploy/docker/Caddyfile`) routes a single domain: `/ws → im-ws`, `/streaming/* → streaming`, `/static/* → user-api`, everything else → the frontend. When accessed via a non-localhost host, the frontend automatically switches to production mode (`wss://domain/ws`, `wss://domain/streaming/ws`) — no frontend changes needed.

## Audio / video calls

- Same machine / same LAN: the built-in public STUN servers are enough for direct calls.
- **Across the public internet**: you need your own TURN server — add it to `WebRTC.IceServers` in `deploy/docker/etc/streaming.yaml`. Browsers also require **HTTPS** for camera/microphone access (localhost is exempt), so use the reverse proxy + domain HTTPS above for public deployments.

## FAQ

- **First build is slow**: it compiles 12 Go services + the frontend; BuildKit cache mounts compile dependencies once, so subsequent builds are fast.
- **Rebuild one service**: `docker compose up -d --build <service>` (e.g. `web`, `user-api`).
- **Port already in use**: copy `.env.example` to `.env` and change the relevant port, or edit the host-side port in `docker-compose.yaml` `ports`.
- **Start over**: `docker compose down -v` to wipe volumes, then `up` again.
- **Check startup order / health**: `docker compose ps` shows each service's `healthy` state; the compose file guarantees readiness in `hichat2.sh` order.
