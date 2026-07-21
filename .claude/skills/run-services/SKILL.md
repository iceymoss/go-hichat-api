---
name: run-services
description: 启动 / 停止 / 调试 go-hichat-api 的多个微服务。当用户说起服务、跑后端、看日志、停服务时触发。
---

go-hichat-api 是 go-zero 多服务架构，本 skill 提供本地启停 / 看日志的标准操作。

## 前置依赖（确认已起）
MySQL / Redis / Etcd / MongoDB / Kafka — 详见 README.md。
快速检查：`docker ps | grep -E "mysql|redis|etcd|mongo|kafka"`

## 一键启动所有服务
```bash
./hichat2.sh
```

`hichat2.sh` 会按顺序起：
- user (rpc + api)
- social (rpc + api)
- im (rpc + api + ws)
- task (mq 消费 + 定时)
- trend (rpc + api)

日志写到 `logs/<svc>-<layer>/<svc>-<layer>.log`。`Ctrl+C` 一次性停掉全部。

## 单独起一个服务
```bash
go run apps/<svc>/<layer>/<svc>.go -f apps/<svc>/<layer>/etc/<svc>-sample.yaml
```

`<layer>` 是 `api` / `rpc` / `ws`（im 专属）/ `mq`（task 专属）。

## 看实时日志
```bash
tail -f logs/<svc>-<layer>/<svc>-<layer>.log
```

## 端口冲突排查
```bash
grep -RnE "Port|Listen" apps/*/etc/*.yaml
```

## 停某个服务
```bash
ps aux | grep "apps/<svc>/<layer>/<svc>.go" | grep -v grep
kill <pid>
```

## 严格约束

- 不要 `kill -9 -1` / `pkill -f go`，会把别人的进程一起带走
- 起 / 停服务都不需要 sudo；如果遇到权限问题先排查 docker 用户组
- 不要修改 `etc/*.yaml` 然后忘记 git restore——这些是 sample，提交前确认
