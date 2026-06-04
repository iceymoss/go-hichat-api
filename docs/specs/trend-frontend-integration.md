# Trend 模块 API 与文档一致性核对 + 前后端联调

## 状态
- 创建日期: 2026-06-04
- 状态: 草稿（待审阅）
- 分支: `feat-trend-api-frontend-integration`

## 目标
让 trend（动态）模块的三方契约——`docs/specs/api.md` 文档、后端 `apps/trend/api` 真实实现、前端 `web/` 调用层——完全对齐，并通过实跑服务端到端验证「发动态 / 评论 / 点赞 / 未读通知」核心路径全链路可用。

## 非目标
- 不改动 trend **rpc** 层已有 proto 字段编号 / 方法签名（只在必须时追加，不复用编号）。
- 不重构动态推荐算法 / 扩散逻辑（Kafka 扩散不在本次范围）。
- 不引入新的 UI 组件库（沿用 Semi + Tailwind）。
- 不做数据库 schema 变更。

## 用户故事
作为已登录用户，我希望在「朋友圈 / 动态」页发布纯文字、图文、视频、文章、分享动态，浏览好友与自己的最新动态流，对动态点赞 / 取消、发表与删除评论、查看点赞用户与未读通知，且所有操作都真实落库并即时反映在 UI 上。

---

## 一、审计结论（文档 ↔ 后端 ↔ 前端）

### 路由覆盖
`docs/specs/api.md` 第 3 章列出的 **20 个 trend 接口全部存在对应 handler**，路径 / 方法一致，无缺失、无多余文档。

### 关键不一致点

| # | 严重度 | 位置 | 问题 | 前端是否受影响 |
|---|--------|------|------|----------------|
| E | 🔴 阻断 | `createtrendlogic.go:42` | 校验 `len(Title)==0 && CoverUrl==""` 即报错；前端仅对 `type===3`(文章) 传 title/cover，**发纯文字(1)/图文(2)/视频(5)/分享(4) 动态全部被拒** | 是，核心发布路径断 |
| D | 🟠 高 | `batchgettrendlikesummarylogic.go:28` | TODO 桩，恒返回空 map `{trend_likes:{}}` | 是，Feed 点赞头像永远为空 |
| F | 🟡 中 | `createtrendlogic.go:34,79` | `type 0` 与 `type 6`(广告) 被拒，但文档把 6=广告 列为合法类型 | 否（前端未发 6），文档误导 |
| A | 🟡 中 | `listtrendslogic.go:28` | `GET /v1/trends` 是空桩（返回 nil） | 否，前端用 `/trends/latest` |
| B | 🟡 中 | `gettrendlikesummarylogic.go:28` | `GET /v1/trend/like/summary` 空桩 | 否，前端用 `/like/batch-summary` |
| C | 🟡 中 | `getunreadlikeslogic.go:28` | `GET /v1/trend/like/unread` 空桩 | 否，前端用 `/trend/unread` |
| G | 🟡 中 | `getusertrendslogic.go:55,108` | 残留 `fmt.Println("len:",tops)` 调试；`resp.LastTime` 从不赋值（文档声明返回 last_time）；topErr 时 `return nil,nil` 吞错 | 文档 last_time 不准 |
| H | 🟢 低 | `getunreadreplieslogic.go:57` | `userIds := make([]string, len(...))` 后再 append → 给 `FindUser` 传入前导空串；`GetUnreadReplies` 报错后未 return，`unRead` 可能 nil 解引用 | 健壮性 |
| I | 🟢 低 | `getlikeduserslogic.go:54` | `FindUser` 失败时 `return nil,nil` 静默吞错 | 健壮性 |
| K | 🟢 低 | `trend.api:213,381-384` | `.api` 源用小写字段名 `isTop` / `Like{id,user,...}`（非导出）；当前 `types.go` 已是导出正确版，但**重跑 goctl 会再生成不可序列化字段** | 潜在回归 |
| L | 🟢 低 | `trend.api:402-410` | `BatchTrendLikeSummaryRequest/Response` 为无人使用的死类型 | 清理项 |
| N | 🟡 中 | 文档 3.3.1 | `like_type` 文档仅写「点赞类型」，实际语义 `>0=点赞, 0=取消`（前端 `MomentsFeed.tsx:1284` 注释印证） | 文档不清 |
| O | 🟢 低 | 文档 3.1.5 | `/trends/latest` 实际是「好友 + 自己」动态，并非全站推荐流 | 文档语义 |

### 前端现状（已基本接好）
- `web/src/lib/trend-api.ts`：覆盖全部接口的 client + backend→domain 映射器。
- `web/src/app/api/trend/[...path]/route.ts`：Next.js 透传代理 → 后端 `127.0.0.1:8891`，统一包成 `{success,data,message}`。
- 消费方：`MomentsFeed.tsx`（发布/列表/点赞/评论/未读/删除）、`TrendDetailPanel.tsx`（详情/评论/点赞）。
- mock 残留：仅 `currentUser`(兜底名/头像/签名) 与 `contacts`、以及若干 `type` 定义仍来自 `lib/mock-data`。

---

## 二、技术设计 / 改动清单

> 约定：改 `.proto` 后 `goctl rpc protoc`；改 `.api` 后 `goctl api go ... -style gozero`；只填 logic，不手写骨架。

### 后端（`apps/trend`）
1. **[E 阻断] 修复 `createTrend` 校验**：按动态类型校验「内容是否为空」——
   - type 1 纯文本：`content` 必填；
   - type 2 图文：`content` 或 `resources` 至少其一；
   - type 3 文章：`title` + (`content`/`resources`) ；
   - type 4 分享：`share_url` 必填；
   - type 5 视频：`resources` 必填。
   - 移除「必须有 title 或 cover_url」的错误前置条件。
2. **[D] 实现 `batchGetTrendLikeSummary`**：对入参 `trend_id[]` 调 rpc 取每条动态点赞用户（复用 `GetLikedUsers` 或确认 rpc 是否已有批量方法），聚合用户信息返回 `map[trend_id][]User`。⚠️ 待定见下。
3. **[F] type 校验与文档对齐**：明确 `type 6/0` 是否允许；若产品上「广告(6)」由运营侧发布而非普通用户，则后端保留拒绝、文档标注「6 不对普通用户开放」。
4. **[G] 清理 `getUserTrends`**：删 `fmt.Println`；正确赋值 `resp.LastTime`；`topErr` 时返回真实 error。
5. **[H/I] 健壮性**：`getUnreadReplies` 用 `make([]string,0,n)` + nil 判定；`getLikedUsers` 不吞错。
6. **[A/B/C] 三个空桩接口处置**（待定见下）：实现 or 文档正式标注「未实现/已被替代」。
7. **[K] 修正 `.api` 字段名**为 PascalCase（`IsTop`、`Like{Id,User,TrendId,LikeTime}`），保 json tag 不变，重跑 goctl 验证 `types.go` 不回退。
8. **[L] 删除死类型** `BatchTrendLikeSummary*`。

### 文档（`docs/specs/api.md` 第 3 章）
9. 补 `like_type` 枚举说明（`1=点赞, 0=取消`）。
10. 3.1.1 标注 `type` 合法范围与 6/0 限制；明确各类型必填字段。
11. 3.1.5 标注 latest 流为「好友+自己」。
12. 对 A/B/C 三接口：按最终处置补「未实现/替代接口」说明；batch-summary 行为与最终实现一致。
13. 3.1.6 `last_time` 行为与代码修复后保持一致。

### 前端（`web/`）
14. 复核 `trend-api.ts` 映射在后端修复后无偏差（重点 create / batch-summary / user-trends 分页）。
15. **清理 mock 残留**：`currentUser`/`contacts` 兜底改为取自真实登录态（`im-store` / user RPC），保留纯 `type` 定义或迁出到独立 types 文件。

### 参考的现有模式
- `getlatesttrendslogic.go` — 「拉 RPC → 批量 FindUser → 组装 User/at_user」标准三段式，新逻辑沿用。
- `social` 模块 api.md 文档密度与字段表风格 — 文档补全对齐该风格。

---

## 三、异常处理

| 场景 | 处理方式 |
|------|---------|
| 发布内容为空 | 按类型返回明确 4xx 业务错误（非 500） |
| 动态不可见(scope) | `getTrendDetail` 已有好友/作者判定，保持 |
| RPC 失败 | 用 `pkg/xerr` 包装，proxy 解析 `rpc error` msg 回传前端 toast |
| 未读列表上游为空/报错 | 返回空数组而非 nil，避免前端崩 |

## 四、端到端验证计划（账号 17585710998 / dsafasf）

前置：MySQL/Redis/Etcd/Mongo/Kafka 已起。

- [ ] `go build ./apps/trend/...` 通过；`bun run build`（或 `tsc --noEmit`）通过
- [ ] 起 trend rpc(`:10003`) + trend api(`:8891`) + `cd web && bun dev`
- [ ] 登录 → 进动态页，`/trends/latest` 正常加载好友+自己动态
- [ ] 发布纯文字动态（type1）→ 成功落库、Feed 出现（验证 E 修复）
- [ ] 发布图文动态（type2，仅图片无标题）→ 成功（验证 E）
- [ ] 点赞 / 取消点赞 → `agree_count` 变化、头像出现（验证 D）
- [ ] 发评论 / 回复 / 删除评论 → 评论树更新
- [ ] 未读通知（评论+点赞）→ `/trend/unread` 返回并可标记已读
- [ ] 个人主页动态列表（含置顶）分页 `last_id/last_time` 正常翻页（验证 G）

## 五、已决策事项（2026-06-04 确认）
1. **空桩接口 A/B/C（`/trends`、`/like/summary`、`/like/unread`）**：✅ 文档标注「未实现，已被替代」，本次不实现（前端无依赖）。
2. **batch-summary（D）**：✅ 在 `trend.proto` **新增批量点赞用户 rpc 方法**，重生成 rpc，rpc 侧实现批量查询，api 侧聚合用户信息。
3. **type 6 广告（F）**：✅ 后端保留拒绝普通用户发布 type 6/0，文档标注「6 不对普通用户开放」。
4. **mock 清理（15）**：✅ 彻底移除 `currentUser`/`contacts` 兜底，缺数据则显示空/降级到登录态真实值。

## 六、MVP 范围
本次 MVP = 修复 E（阻断）+ D（点赞头像）+ G（个人动态分页/调试残留）+ 文档补全（9-13）+ 端到端验证核心 4 路径（发布/点赞/评论/未读）。
A/B/C 与 K/L/H/I 视待定事项结论纳入或转后续迭代。

## 七、实现步骤（每步可独立 commit）
1. [x] 后端修复 E（create 校验）+ G（user-trends）+ H/I 健壮性 — commit `fix(trend): align api contract`
2. [x] 后端实现 D（batch-summary，复用已存在的 rpc `BatchGetTrendLikeSummary`）— 同上
3. [~] `.api` 字段名修正 K + 删死类型 L — **已撤销**，见下「联调发现」
4. [x] 文档补全 `docs/specs/api.md` 第 3 章（9-13）— commit `docs(trend): sync api.md`
5. [x] 前端复核 + mock 清理（14-15）— commit `refactor(web): drop mock-data fallbacks`
6. [x] 起服务端到端联调验证 — 见下「联调结果」

## 八、联调发现：goctl 重生成会破坏 GET 查询绑定（重要）
- 本仓库的 `apps/trend/api/internal/types/types.go` 中，所有 **GET 请求体**字段用的是 `form:"xxx"` 标签（由较老版本 goctl 生成）。
- 用 **goctl 1.8.2** 重新生成会把这些 `form:` 标签改成 `json:`，导致 go-zero **无法再从 query 绑定参数**——实测每个 GET 接口都返回 `field "xxx" is not set` / `type mismatch`。
- 因此 K（修 `.api` 小写字段名）+ L（删死类型）所触发的 `goctl api go` 重生成被**整体撤销**（commit `revert(trend): restore form tags`）。原始 `types.go` 本就已是导出的 `IsTop`/`Like{Id,User,Trend_id,Like_time}`，K 实际上没有必要。
- ⚠️ 后续若需再次 `goctl api go`，必须先解决 goctl 版本/模板把 GET 字段 `form→json` 降级的问题，否则会回归此 bug。

## 九、联调结果（2026-06-04，账号 17585710998 / user_id=11，临时实例 :8896 跑新代码）
| 验证项 | 结果 |
|--------|------|
| 发纯文字动态 type1 | ✅ `trend_id:25` 创建成功（E 修复前会被拒） |
| 发图文 type2（仅图无标题）| ✅ 通过 api 校验（命中 rpc「发布太频繁」限流，校验层 OK）|
| type1 空内容 / type6 广告 | ✅ 分别被拒「请填写动态内容」/「不支持该类型的动态」(F) |
| 最新动态流 `/trends/latest` | ✅ 返回好友+自己动态 |
| 动态详情 `/trend/detail` | ✅ 返回 trend 25 |
| 点赞 `/trend/like` | ✅ 成功 |
| 批量点赞摘要 `/like/batch-summary` | ✅ `{"trend_likes":{"25":[{user 11}]}}`（D 实现验证）|
| 点赞用户列表 `/like/users` | ✅ total=1 |
| 发评论 / 评论树 / 根评论 | ✅ 评论 id=38 正常 |
| 未读 `/trend/unread` | ✅ 返回 replies |
| 用户动态 `/user/trends` | ✅ `last_time=1751090955` 已赋值（G 修复验证）|
