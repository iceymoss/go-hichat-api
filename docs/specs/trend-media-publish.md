# 动态媒体发布

## 状态
- 创建日期: 2026-06-05
- 状态: 草稿

## 目标
让登录用户可以在现有动态发布入口中友好地发布图片/视频动态，支持媒体选择、预览、排序、压缩、上传、草稿和安全校验；图片默认最多 9 张，单张默认 50MB，具体限制走系统配置。

## 非目标
- 图片编辑、裁剪、滤镜、美颜先不实现，仅保留后续 TODO。
- 图片/视频内容审核先不实现，仅保留接口和配置扩展点。
- 不允许前端绕过后端校验；前端压缩和限制只用于体验优化，后端仍必须做安全校验。
- 不复用收藏上传作为长期动态上传接口；收藏上传只能作为实现模式参考。

## 用户故事
作为登录用户，我想在动态页点击现有发布入口后选择图片或视频，编辑文字、可见范围并发布，以便好友能在朋友圈看到我的图文或视频动态。

## 核心流程
1. 用户进入动态页，点击现有发布按钮打开发布弹窗。
2. 用户选择动态类型：图文或视频。
3. 图文动态中，用户通过系统文件选择器一次多选图片，或拖拽图片到发布弹窗。
4. 前端按系统配置限制图片数量，默认最多 9 张；超过限制时给出友好提示并阻止继续添加或只保留允许数量。
5. 前端展示九宫格图片预览，用户可删除单张图片、调整图片顺序。
6. 前端对图片做压缩，上传压缩后的文件。
7. 视频动态中，用户选择视频文件；视频限制项全部走系统配置，默认值待定。
8. 前端上传媒体文件，展示每个文件上传进度、成功状态、失败状态和重试入口。
9. 后端上传接口做大小、数量、扩展名、MIME sniffing、图片解码、文件头、存储路径等安全校验。
10. 上传成功后返回资源 URL、文件名、大小、类型等信息。
11. 前端调用现有动态创建接口，传入 `type=2` 图文或 `type=5` 视频，以及上传返回的 `resources`。
12. 发布按钮在上传/发布过程中进入 loading，阻止重复点击。
13. 发布成功后关闭弹窗、清除本地草稿和服务端草稿、刷新动态流。

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 图片数量超过配置限制 | 前端提示最多可选择 N 张，不继续加入超限文件；后端再次校验并拒绝非法请求 |
| 单文件大小超过配置限制 | 前端提示文件过大；后端返回明确错误码和消息 |
| 文件扩展名与 MIME 不匹配 | 后端拒绝上传，提示文件类型不支持 |
| 图片解码失败或疑似恶意图片 | 后端拒绝上传，不进入存储 |
| 上传失败 | 阻止发布，标记失败文件，允许用户重试或删除失败文件 |
| 网络超时 | 提示网络异常，请重试；不自动重复创建动态 |
| 重复点击发布 | 发布按钮 loading，前端阻止；后端保留现有发布频控 |
| 关闭弹窗且存在未发布内容 | 弹出确认：保存草稿、丢弃、取消关闭 |
| 发布成功后草稿未清理 | 前端和服务端都应按动态 ID 或草稿 ID 清理，避免下次误恢复 |
| 服务端草稿保存失败 | 前端保留本地草稿，并提示云草稿同步失败 |
| 视频压缩任务失败 | 标记视频处理失败，允许重试；不创建已发布动态 |

## 技术设计

### 数据模型

#### 现有动态表
`trend` 表已支持媒体动态的基本字段：

| 字段 | 用途 |
|------|------|
| `type` | 动态类型，`2` 图文，`5` 视频 |
| `pic_arr` / `resources` 映射 | 保存图片/视频 URL 数组的 JSON 字符串 |
| `cover` | 视频或文章封面 |
| `content` | 动态正文 |
| `scope` | 可见范围 |
| `circle_state` | 是否进入朋友圈流 |

短期 MVP 可以继续复用 `resources []string`，避免一次性改动动态返回结构。

#### 新增服务端草稿表：`trend_drafts`

数据库 schema 变更需要开发前再次确认。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | bigint | 草稿 ID，GORM 主键 |
| `user_id` | bigint | 草稿所属用户 |
| `type` | tinyint/int | 动态类型 |
| `content` | text | 正文 |
| `title` | varchar/text | 标题 |
| `resources` | text | JSON 字符串，保存媒体 URL 和排序 |
| `cover_url` | varchar/text | 视频封面 URL |
| `share_url` | varchar/text | 分享链接 |
| `position_name` | varchar/text | 位置名称 |
| `longitude` | decimal/double | 经度 |
| `latitude` | decimal/double | 纬度 |
| `scope` | tinyint/int | 可见范围 |
| `open_reply` | tinyint/int | 是否允许评论 |
| `state` | tinyint/int | 草稿状态：1 正常，0 删除，2 已发布 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

#### 系统配置表：`system_settings`

复用现有 `system_settings.key_name/value` 模式。配置 key 建议：

| key | 默认值 | 说明 |
|-----|--------|------|
| `trendPublishEnabled` | `true` | 是否允许发布动态 |
| `trendMediaUploadEnabled` | `true` | 是否允许上传动态媒体 |
| `trendMaxImageCount` | `9` | 图文动态最多图片数量 |
| `trendMaxImageSizeMB` | `50` | 单张图片最大大小 |
| `trendAllowedImageTypes` | `jpg,jpeg,png,webp,gif` | 允许图片格式 |
| `trendMaxVideoCount` | 待定 | 视频数量限制 |
| `trendMaxVideoSizeMB` | 待定 | 单个视频大小限制 |
| `trendMaxVideoDurationSec` | 待定 | 视频时长限制 |
| `trendAllowedVideoTypes` | 待定 | 允许视频格式 |
| `trendImageCompressionEnabled` | `true` | 前端是否启用图片压缩 |
| `trendVideoCompressionEnabled` | `true` | 是否启用视频服务端压缩 |
| `trendMediaReviewEnabled` | `false` | 媒体审核开关，MVP 先 TODO |
| `trendAllowExternalResourceUrl` | `false` | 是否允许非本站资源 URL，默认不允许 |

由于 `SystemConfigModel` 当前在 `apps/user/models` 下，Trend 服务不能直接 import user model。实现时需要将系统配置读取能力抽到 `pkg/` 或新增配置 RPC，避免跨服务直接读对方业务 model。

### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/trend` | 已有创建动态接口，继续用于最终发布 |
| `POST` | `/v1/trend/media/upload` | 新增动态媒体上传接口，multipart 字段名 `file` |
| `POST` | `/v1/trend/media/complete` | 可选：视频服务端压缩/处理完成回调或查询前的完成确认 |
| `GET` | `/v1/trend/config` | 获取动态发布媒体限制配置 |
| `POST` | `/v1/trend/draft` | 创建/保存服务端草稿 |
| `GET` | `/v1/trend/draft` | 获取当前用户最近草稿或指定草稿 |
| `DELETE` | `/v1/trend/draft` | 删除草稿 |

### 上传接口响应

```json
{
  "url": "http://.../trend/image/xxx.webp",
  "name": "xxx.webp",
  "size": 123456,
  "file_type": "image",
  "content_type": "image/webp",
  "width": 1080,
  "height": 720,
  "duration": 0,
  "cover_url": ""
}
```

MVP 创建动态仍只传 `resources: string[]`，但上传响应保留元数据，便于后续升级为结构化资源。

### 前端设计

#### 发布弹窗
- 复用 `MomentsFeed.tsx` 当前发布入口和发布弹窗。
- 将“手动输入图片/视频 URL”替换为媒体选择组件。
- 支持系统文件选择器多选图片。
- 支持拖拽图片/视频添加。
- 图片用九宫格预览。
- 图片支持删除和调整顺序。
- 视频展示封面、文件名、大小、上传/处理状态。
- 上传/发布期间发布按钮 loading，禁止重复点击。

#### 图片压缩
- 前端读取配置后决定是否压缩。
- 压缩后上传压缩文件。
- 前端压缩失败时允许使用原图上传或提示重试，具体行为开发前确认。
- 后端仍按原始上传文件做安全校验，不信任前端压缩结果。

#### 视频压缩
- 视频上传后由服务端进入处理流程。
- MVP 设计为服务端同步或异步压缩，具体实现取决于现有依赖和部署能力。
- 如果异步，前端需要显示“处理中”，处理完成后才允许发布。

#### 草稿
- 本地草稿：使用 `localStorage` 自动保存，建议 key：`hichat_moments_publish_draft:${userId}`。
- 服务端草稿：通过 `/v1/trend/draft` 保存，支持换设备恢复。
- 草稿内容包含：`type`、`content`、`title`、`resources`、`coverUrl`、`shareUrl`、`positionName`、`scope`、`openReply`、`updatedAt`。
- 发布成功后清理本地草稿和服务端草稿。
- 关闭弹窗时如果存在未发布内容，弹确认：保存草稿、丢弃、取消。

### 后端安全设计
- 上传接口使用 `http.MaxBytesReader` 限制请求体大小。
- 不只依赖扩展名，必须做 MIME sniffing 和文件头校验。
- 图片必须能被安全解码，拒绝损坏或异常图片。
- 校验图片像素尺寸上限，避免解压炸弹类攻击；具体上限走配置或待定。
- 对 GIF 等多帧图片限制大小和帧数，具体策略待定。
- 视频校验大小、格式、时长，时长默认待定。
- 资源 URL 默认只允许本站上传返回的 URL，不允许客户端直接提交外链资源。
- 文件保存路径按用户和媒体类型隔离，避免路径穿越。
- 文件名由服务端生成，不信任用户文件名。
- 上传成功但未发布的资源需要清理策略，MVP 可通过草稿关联和定时任务后续补充。
- 图片/视频内容审核保留 TODO 和配置开关。

### 实现步骤（每步可独立 commit）
1. [ ] 配置设计：定义动态媒体相关系统配置 key 和默认值。
2. [ ] 数据模型：新增 `trend_drafts` 草稿表和 model，schema 变更前再次确认。
3. [ ] API 契约：在 `apps/trend/api/trend.api` 新增媒体上传、配置、草稿接口。
4. [ ] RPC 契约：如上传/草稿核心逻辑需放 RPC，则在 `apps/trend/rpc/trend.proto` 追加方法和 message。
5. [ ] 后端上传：新增动态媒体上传 handler/logic，复用 `pkg/storage`。
6. [ ] 后端安全校验：实现图片/视频类型、大小、MIME、解码、URL 白名单校验。
7. [ ] 后端草稿：实现保存、读取、删除、发布成功清理。
8. [ ] 视频压缩：实现服务端压缩/处理流程，异步方案需增加状态查询。
9. [ ] 前端 API：新增 trend 媒体上传、配置、草稿 client 和 Next.js multipart proxy。
10. [ ] 前端发布弹窗：实现选图、拖拽、九宫格、删除、排序、上传进度。
11. [ ] 前端压缩：实现图片压缩并接入上传链路。
12. [ ] 前端草稿：实现本地自动保存、服务端同步、关闭确认、恢复草稿。
13. [ ] 发布串联：上传成功资源写入 `createTrend.resources`，发布成功刷新动态流。
14. [ ] 文档和测试：补充 API 文档、单元测试、前端交互测试。

### 参考的现有模式
- `apps/trend/api/trend.api` — 动态 HTTP 契约，已有 `POST /v1/trend`。
- `apps/trend/rpc/trend.proto` — 动态类型、可见范围和资源相关 proto 定义。
- `apps/trend/api/internal/logic/trend/createtrendlogic.go` — HTTP 创建动态校验和 RPC 调用模式。
- `apps/trend/rpc/internal/logic/createtrendlogic.go` — 发布频控、敏感词、落库模式。
- `apps/trend/models/trendmodel_gen.go` — `resources` 当前落到 `PicArr` JSON 字段的模式。
- `pkg/storage/storage.go` — 文件存储抽象。
- `pkg/storage/local.go` — 本地存储实现。
- `pkg/storage/filetype.go` — 媒体类型分类。
- `apps/user/api/internal/logic/favorite/uploadfilelogic.go` — 50MB 上传、分类、存储模式参考。
- `apps/user/api/internal/handler/favorite/uploadhandler.go` — multipart 上传 handler 和 `MaxBytesReader` 模式参考。
- `pkg/db/objects/system.go` — `system_settings` 配置表结构。
- `apps/user/models/systemconfigmodel.go` — 系统配置读取/写入模式，需抽象到共享层或通过 RPC 使用。
- `web/src/app/api/user/favorites/upload/route.ts` — Next.js multipart 上传代理参考。
- `web/src/lib/trend-api.ts` — 前端动态 API client。
- `web/src/components/im/MomentsFeed.tsx` — 当前发布弹窗和动态发布入口。

## 测试计划
- [ ] API 单元测试：创建图文动态时 `resources` 为空/非空校验。
- [ ] API 单元测试：超过图片数量限制时拒绝。
- [ ] API 单元测试：超过单图大小限制时拒绝。
- [ ] API 单元测试：伪造扩展名、伪造 MIME、损坏图片拒绝上传。
- [ ] API 单元测试：视频大小/格式/时长按配置校验。
- [ ] API 单元测试：外链资源 URL 默认拒绝。
- [ ] API 单元测试：草稿保存、读取、删除、发布成功清理。
- [ ] 前端测试：系统多选最多 9 张。
- [ ] 前端测试：拖拽添加、删除、排序。
- [ ] 前端测试：上传失败阻止发布并可重试。
- [ ] 前端测试：关闭弹窗时保存/丢弃/取消逻辑。
- [ ] 集成测试：图片上传成功后发布动态，朋友圈流展示九宫格。
- [ ] 集成测试：视频上传处理成功后发布动态，朋友圈流展示视频封面。

## 待定事项
- 视频默认数量、大小、时长限制。
- 允许的视频格式列表。
- 图片像素尺寸上限、GIF 帧数限制。
- 前端图片压缩质量、最大宽高策略。
- 视频服务端压缩采用同步还是异步，是否引入 FFmpeg 或队列。
- 上传成功但未发布资源的清理策略。
- 系统配置读取能力放到 `pkg/` 还是通过 RPC 暴露。
- `trend` 表位置字段 `position_point` 与生成 model 的 `position` 命名差异需要核对。
- 是否要将 `resources []string` 升级为结构化资源列表。

## MVP 范围
- 复用现有动态发布入口。
- 支持图片动态：系统多选、拖拽添加、九宫格预览、删除、排序、最多 9 张默认配置、单张 50MB 默认配置。
- 支持视频动态：视频限制全部配置化，默认值待定。
- 完整媒体上传链路：前端上传、Next.js multipart proxy、Trend 后端上传接口、存储返回 URL。
- 前端图片压缩。
- 服务端视频压缩/处理能力。
- 本地草稿和服务端草稿。
- 后端安全校验，防止图片攻击、伪 MIME、大文件和外链资源绕过。
- 发布成功后进入朋友圈流，使用现有 `POST /v1/trend` 和 `resources` 字段完成动态创建。
