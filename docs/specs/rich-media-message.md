# 会话富媒体消息（图片 / 视频 / 文件 / 语音 / 表情包 / emoji）

## 状态
- 创建日期: 2026-06-02
- 状态: 草稿
- 分支: `feat-im-message-media`

## 目标
让会话模块在现有「纯文字」基础上，支持发送 **emoji、图片、视频、普通文件、语音、表情包（含个人收藏表情）**，覆盖主流 IM 的富媒体消息能力。

## 非目标
- 不做服务端转码 / 压缩 / 缩略图生成（缩略图、视频封面由前端生成后随文件一并上传）。
- 不做表情包「商店 / 运营后台」（仅做内置一套 + 个人收藏）。
- 不做对象存储（OSS/S3）接入，MVP 用本地存储，保留后续可切换的抽象接口。
- 不改动消息可靠性机制（ack / 重发 / 已读位图）本身，富媒体复用现有链路。
- 不做大文件分片 / 断点续传（100MB 以内一次性上传）。

## 用户故事
- 作为聊天用户，我想在输入框旁点击 emoji 面板插入表情，以便快速表达情绪。
- 作为聊天用户，我想发送本地照片 / 视频 / 文件，以便分享内容；对方能看到缩略图/封面，点击查看大图或下载。
- 作为聊天用户，我想按住录音发送语音消息，以便不打字也能沟通。
- 作为聊天用户，我想把喜欢的图片/GIF 收藏为「我的表情」，并从表情面板里点选发送；也能把聊天里别人发的图片一键添加到我的表情。

## 核心流程

### A. emoji（零后端）
1. 前端输入框旁 emoji 面板，引入**全量开源 emoji 数据集**（按分组展示 + 搜索）。
2. 点击 emoji 插入到文本输入框，作为普通文本消息（`mType=1`）发送。后端无任何改动。

### B. 图片 / 视频 / 文件 / 语音（上传 + 发送）
1. 用户在会话内选择文件 / 录音。
2. 前端校验类型与大小（统一上限 **100MB**），对图片生成缩略图、对视频抽首帧生成封面。
3. 前端把「原文件」「缩略图/封面」上传到 **im 服务新增的上传接口** `POST /v1/im/upload`，拿到 URL。
4. 前端把元数据拼成 **JSON 字符串**塞进消息 `content`，并设置对应 `mType`，走现有 WebSocket `chat.user` 发送。
5. 现有链路（ws → MQ → MongoDB → 推送）原样透传 `content`，**发送侧后端零改动**。
6. 接收方/发送方前端按 `mType` 解析 `content` JSON 并渲染（图片缩略图、视频播放器、文件卡片、语音条）。

### C. 表情包收藏（user 服务）
1. **内置表情包**：前端内置一套，点选后按图片消息（`mType=5` MemesMType）发送。
2. **自定义表情上传**：用户上传图片/GIF → user 上传接口 → 存入个人表情库（user 服务）。
3. **从聊天添加到表情**：长按/右键聊天里的图片消息 → 调 user 接口把该 URL 存为个人表情。
4. **表情面板**：拉取「内置 + 我的收藏」展示，点选发送（`mType=5`）。

## `content` JSON 约定（核心技术约定）

`msgContent` 当前是 `string`，纯文本直接存文本（**向后兼容，旧数据不变**）。富媒体把元数据 JSON 序列化为字符串存入 `content`，按 `mType` 区分：

| mType | 类型 | content JSON 结构 |
|-------|------|-------------------|
| 1 TextMType  | 文本 | 纯文本字符串（不变） |
| 2 FileMType  | 文件 | `{"url","name","size","ext"}` |
| 3 VoiceMType | 语音 | `{"url","duration","size"}`（duration 秒） |
| 4 ImageMType | 图片 | `{"url","thumbUrl","width","height","size"}` |
| 5 MemesMType | 表情包 | `{"url","width","height"}` |

- 前后端共用一份字段命名；序列化用 `common.Marshal`（后端如需解析）/ `JSON.stringify`（前端）。
- **兼容渲染**：前端遇到未知 `mType` 或 `content` 解析失败时，降级显示「[当前版本不支持的消息类型]」，不崩溃。
- 会话列表「最后一条消息」预览：图片→`[图片]`、视频→`[视频]`、文件→`[文件] name`、语音→`[语音]`、表情→`[表情]`。

## 技术设计

### 1. 上传接口（im 服务新增）
- 契约（`apps/im/api/im.api`）：
  - `POST /v1/im/upload`，`multipart/form-data`，字段 `file`。
  - 响应：`{url, name, size, fileType}`（`fileType` ∈ image/video/file/voice）。
- 复用 `pkg/storage.FileStorage`（`LocalStorage`），落盘目录按类型分：`im/image` / `im/video` / `im/file` / `im/voice`。
- 大小上限 100MB（在 logic 校验）；类型识别复用 user 服务 `uploadfilelogic.go` 的扩展名映射思路，抽到 `pkg/` 或在 im logic 内实现。
- 鉴权走现有 JWT 中间件；`internal/config` + `etc/*-sample.yaml` 增加 `Upload`（basePath/baseURL）配置。
- 参考：`apps/user/api/internal/logic/favorite/uploadfilelogic.go`、`apps/user/api/internal/handler/user/uploadavatarhandler.go`。

### 2. 消息发送链路（后端基本零改动）
- `ws.Msg.Content`(string) → `mq.MsgChatTransfer.MsgContent`(string) → MongoDB `ChatLog.MsgContent`(string) 全程透传，富媒体只是 `content` 里换成 JSON。
- 验证 ws / MQ / push 不对 `content` 做任何针对文本的假设（调研确认为纯透传）。
- 枚举已就绪：`pkg/constants/im.go` 已有 `FileMType=2 / VoiceMType=3 / ImageMType=4 / MemesMType=5`；前端 `web/src/lib/ws-client.ts` `MsgType` 已对齐。

### 3. 表情包收藏（user 服务，已存在，直接复用）
> ✅ **后端已实现**：表情包收藏在 main 分支即有，模型为 `apps/user/models/useremojimodel.go` 的 `UserEmoji`（表 `user_emojis`），是独立于 `favorites` 收藏的功能，恰好满足「表情包收藏 ≠ 用户收藏」。**不再新建 `user_stickers`，直接复用。**

现有接口（前缀 `/api/v1/user`，JWT）：
- `POST   /emoji/upload`  上传表情文件拿 URL（`apps/user/api/internal/logic/emoji/uploademojilogic.go`）
- `POST   /emoji`         添加表情：`{url, name, thumbnail, width, height, size, fileType}`，按 url 去重；**从聊天图片收藏也走这个**（传图片消息里的 url）
- `GET    /emojis`        我的表情列表（按 `sort_order DESC, created_at DESC`）
- `DELETE /emoji`         删除（仅限本人）

`user_emojis` 字段：id / user_id / url / name / thumbnail / width / height / size / file_type / sort_order / created_at / updated_at —— 比原 spec 设计的 `user_stickers` 更全。

> ⚠️ 已知缺口（本期按用户决定暂不处理）：`user_emojis`（及 `favorites`）未登记进 `pkg/db/objects` + `deploy/sql_init.go` 迁移注册表，新环境部署需手工建表。

### 4. 前端（web/）
- 输入区组件扩展：emoji 面板、表情面板、附件按钮（图片/视频/文件）、语音录制按钮。
- 上传封装：`web/src/lib/api-client.ts` 增加 im 上传调用；缩略图/封面用 canvas / `<video>` 抽帧。
- 发送：`web/src/lib/chat-store.ts` `sendMessage` 支持把 JSON content + 对应 mType 发出（乐观更新沿用现有 `local_` 占位 → MsgAck 升级 ObjectID 机制）。
- 渲染：消息气泡按 `mType` 分支渲染图片/视频/文件卡片/语音条/表情。
- 所有用户可见文本走 `useTranslation()` + `web/src/i18n/locales/*.json`（中英）。
- UI 用 Semi Design 组件 + Tailwind。

### 实现步骤（每步可独立 commit）
1. [x] 后端：im 新增 `POST /v1/im/upload`（改 `.api` → goctl 生成 → 填 logic → 配 yaml/config）
2. [x] 后端：抽取/实现文件类型识别 + 100MB 校验，落盘分目录（`pkg/storage.ClassifyMedia` + `uploadMedia`）
3. [x] 后端：表情包收藏 —— 复用已存在的 `user_emojis` 功能（`UserEmoji` model + `/emoji*` 接口），不新建表
4. [x] 前端：富媒体消息渲染（图片/视频/文件/语音/表情，含未知类型降级）—— `MessageContent`
5. [x] 前端：上传 + 发送图片/视频/文件（`media-message.ts` + `imUpload` + ChatDetail 附件入口）
6. [x] 前端：语音录制（Safari→m4a / 其余→webm）+ 发送
7. [x] 前端：emoji 面板（emoji-mart 全量开源数据集）—— `ChatEmojiPanel`
8. [x] 前端：表情收藏面板（发送收藏表情 + 从聊天右键「添加到表情」，复用 `/api/user/emojis`）
9. [x] 前端：会话列表富媒体预览（`mediaPreview`）；ChatDetail 文案沿用组件内中文硬编码约定
10. [ ] 后端：孤儿文件清理定时任务（apps/task/cron，新 task 注册进 registry，Redis 单实例锁）
11. [ ] 文档：`/sync-api-docs` 更新 `docs/specs/api.md`

### 参考的现有模式
- `apps/user/api/internal/logic/favorite/uploadfilelogic.go` — 上传 + 类型识别 + 大小校验
- `pkg/storage/storage.go` / `pkg/storage/local.go` — 存储抽象 + 本地实现
- `pkg/constants/im.go` — MType 枚举（已预留富媒体类型）
- `apps/im/ws/internal/handler/conversation/conversation.go` + `apps/task/mq/.../msg_chat_transfer.go` — 发送链路（透传 content）
- `web/src/lib/chat-store.ts` / `web/src/lib/ws-client.ts` — 前端发送/接收/类型映射
- `apps/user/models/favoritesmodel.go` — 收藏表 GORM 模式

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 文件超过 100MB | 上传接口返回业务错误（pkg/xerr），前端提示「文件不能超过 100MB」 |
| 不支持的文件类型 | 后端归类为 `file` 兜底；前端按文件卡片渲染 |
| 上传失败 / 超时 | 前端消息标记 `failed`，可重试；不写 MongoDB |
| 上传成功但发送失败 | 文件已在存储，消息走现有失败重发；产生孤儿文件（MVP 接受，后续加清理任务） |
| 接收方旧版本收到富媒体 | 未知 mType 降级显示「[不支持的消息类型]」，不崩溃 |
| content JSON 解析失败 | 同上降级，并打日志 |
| 重复上传同一文件 | 文件名加时间戳，天然不冲突（沿用 LocalStorage 命名） |
| 语音录制无权限 | 前端提示授权麦克风 |

## 测试计划
- [ ] im 上传接口：正常上传 / 超 100MB 拒绝 / 各类型归类正确（table-driven）
- [ ] 文件类型识别函数单测（image/video/voice/file 边界扩展名）
- [ ] user sticker CRUD：增删查 + 仅本人可删（三库 SQLite/MySQL/PostgreSQL 通过）
- [ ] 发送链路：富媒体 content 透传后 MongoDB 存取一致、推送 mType/content 不丢
- [ ] 前端：各类型消息渲染快照 + 未知类型降级
- [ ] 测试数据测试后清理

## 待定事项（全部已确认）
1. ✅ 表情包收藏表：新建 `user_stickers` 专表，独立于 `favorites` 收藏功能，不复用。
2. ✅ 内置 emoji：使用全量开源 emoji 数据（前端引入开源 emoji 数据集，按分组展示 + 搜索）。
3. ✅ 语音格式：需兼容 iOS Safari。前端按浏览器能力择优录制（Safari 用 `audio/mp4`/`m4a`，其余 `audio/webm;codecs=opus`），上传时带真实扩展名，渲染用 `<audio>` 直接播放。
4. ✅ 孤儿文件清理：本期加定时清理任务到 `apps/task/cron`，扫描存储目录中未被任何 chatLog/sticker 引用且超过保留期的文件并删除（参考 `apps/task/cron/tasks/data_cleanup_task.go` + `registry.go`，单实例 Redis 锁、ctx.Done 可中断）。

## MVP 范围
全部纳入 MVP（用户已确认）：emoji 面板 + 内置 emoji 表情包 + 图片 + 视频 + 文件 + 语音 + 表情收藏（自定义上传 / 从聊天添加 / 内置一套）。存储用本地，缩略图/封面前端生成，单文件上限 100MB，上传接口在 im 服务。
