# 添加好友 & 加群（搜索 + 资料卡片 + 申请）

## 状态
- 创建日期: 2026-06-02
- 状态: 草稿
- 关联分支约定: `feat-im-add-friend-...` / `feat-im-join-group-...`（英文提交、无署名、专属分支）

## 目标
把现有过于简单的「添加好友」面板（`web/src/components/im/AddFriendPanel.tsx`）升级为完整的「找人 / 找群」能力：
搜索结果支持**滚动分页**；点击结果弹出**资料卡片**（用户/群）展示完整资料并据状态发起**加好友 / 申请加群**。

## 非目标
- 不做好友/群申请的**审核处理**流程（已存在：好友 `/friend/putIn` PUT、群 `GroupPutInHandle`），本 spec 只负责「发起申请」。
- 不做通讯录/群聊主列表的重构；不改已读未读、置顶免打扰等既有逻辑。
- 不做二维码/邀请链接入群（已存在 `joinByToken`，保持现状）。
- 不做按职业/标签等高级筛选；搜索仅限：用户(昵称模糊/手机号·邮箱精确)、群(群号精确/群名模糊)。

## 用户故事
- 作为用户，我在会话列表右上角点「+」，打开统一搜索面板，可在「用户 / 群聊」两个 tab 间切换。
- 作为用户，我搜索用户/群，结果**分页加载**（滚动到底自动加载下一页），不一次性拉全量。
- 作为用户，我点某个搜索结果，弹出**资料卡片**看头像/昵称/性别/地区/签名（用户）或群头像/群名/人数/简介/群主（群），并据当前关系状态操作。
- 作为用户，我对陌生人点「添加好友」（可填附言）发起申请；对未加入的群点「申请加入」（可填附言）发起入群申请。

## 核心流程

### A. 加好友
1. 会话列表 `ChatListToolbar` 的 `UserPlus` → 打开 `AddFriendPanel`，默认「用户」tab。
2. 输入关键词（昵称模糊 / 手机号·邮箱精确）→ 防抖 → 调分页搜索接口，第 1 页。
3. 结果列表项展示头像+昵称+地区；滚动到底 → 加载下一页（有更多才加载）。
4. 点结果项 → 弹 `UserProfileCard`（陌生人模式）展示资料 → 底部按钮按状态：
   - 已是好友 → 「发送消息」（跳到/打开会话）
   - 已发过申请且待审核 → 「已发送」（禁用）
   - 自己 → 不显示加好友按钮
   - 其它 → 「添加好友」→ 填附言 → `POST /api/social/friend/putIn` → toast「已发送」。

### B. 加群
1. `AddFriendPanel` 切到「群聊」tab（或通讯录群聊页入口，见 UX：统一面板两 tab）。
2. 输入关键词（群号精确 / 群名模糊）→ 防抖 → 调群分页搜索接口（**新增后端**），第 1 页。
3. 结果项展示群头像+群名+人数；滚动分页同上。
4. 点结果项 → 弹**群资料卡**（复用 `/group/detail`）展示群名/头像/人数/简介/群主 → 底部按钮按状态：
   - 已在群内 → 「进入群聊」
   - 已申请待审核 → 「已申请」（禁用）
   - 其它 → 「申请加入」→ 填附言 → `POST /api/social/group/putIn`（`join_source=1`）→ toast「已发送」。

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 搜索无结果 | 空态：「未找到用户 / 未找到群聊」 |
| 空输入 | 提示「输入手机号/邮箱/昵称」或「输入群号/群名」，不发请求 |
| 加载失败/超时 | 列表底部错误态 + 「重试」；首页失败显示重试块 |
| 搜到自己 | 用户结果中排除自己或卡片隐藏加好友按钮 |
| 重复申请 | 后端幂等；前端按状态显示「已发送/已申请」并禁用 |
| 已是好友/已在群 | 按钮变「发送消息/进入群聊」 |
| 手机号隐私 | 仅当**手机号精确搜索命中**时结果/卡片才返回并展示 `mobile`；昵称模糊搜索不返回手机号 |
| 翻页并发 | 请求带页码，丢弃过期响应（关键词变化时重置分页、取消在途） |

## 技术设计

### 数据模型
- **无 schema 变更**。用户、群、好友申请、群申请表均已存在；搜索只读查询，分页用 `LIMIT/OFFSET`（注意三库兼容，走 GORM）。
- 好友状态 / 群成员状态：前端用**已加载数据派生**（`im-store` 的 `friends`、群列表），不新增接口。
  申请「待审核」态：复用现有 `/friend/putIns`、`/group/putInsByUid` 列表派生。

### API 接口

复用（不改）：
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/social/friend/putIn | 发起好友申请（user_uid, req_msg?） |
| POST | /api/social/group/putIn | 发起入群申请（group_id, req_msg?, join_source=1） |
| GET  | /api/social/group/detail | 群资料卡数据 |
| GET  | /api/social/friend/putIns | 派生「好友申请待审核」状态 |
| GET  | /api/social/group/putInsByUid | 派生「入群申请待审核」状态 |

改造 / 新增（后端）：
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/user/search | **加分页**：新增 `page,size`（可选，默认 1/20），响应加 `total`；保持 `name/phone/email/ids` 语义；`mobile` 仅在 phone 精确命中时回填 |
| GET | /api/v1/social/group/search | **新增**：`keyword`(群号精确+群名模糊)、`page,size`；响应 `list[]{id,name,icon,memberCount,notification/intro,createUid}` + `total` |

> 对应 rpc：`user.SearchUser`（若现有搜索在 rpc 层则加分页字段；否则在 user-api logic 内分页）、`social` 新增 `GroupSearch` rpc（proto 追加 message + 方法，字段号不复用，goctl 生成）。

### 实现步骤（每步可独立 commit）

**后端**
1. [ ] user `search` 接口加分页：`.api` 加 `page/size` + resp `total` → goctl 生成 → logic 分页查询（GORM LIMIT/OFFSET，三库兼容）；mobile 隐私回填规则
2. [ ] social 新增群搜索：`.proto` 加 `GroupSearch`（keyword+分页）→ goctl → rpc logic（群号精确 OR 群名模糊 + 分页 + total）
3. [ ] social `.api` 加 `GET /group/search` → goctl → api logic 调 rpc，回填 memberCount（复用群成员统计）

**前端**
4. [ ] `api-client`/social 封装：`searchUsers(page,size)`、`searchGroups(keyword,page,size)`，类型补 `total`
5. [ ] `AddFriendPanel` 升级为两 tab（用户/群）+ 通用「无限滚动列表」（IntersectionObserver 触底加载、关键词变化重置、在途请求作废）
6. [ ] 用户结果点项 → `UserProfileCard` 陌生人模式：**微调使其展示** 头像/昵称/性别/地区/签名/职业/标签（mobile 仅精确搜索时）；按好友状态渲染底部按钮
7. [ ] 群结果点项 → 群资料卡（复用 `/group/detail` 数据）+ 按群成员状态渲染「申请加入/进入群聊/已申请」
8. [ ] 状态派生与边界：排除自己、已是好友/已在群、待审核禁用、空态/错误态/重试

### 参考的现有模式
- `web/src/components/im/AddFriendPanel.tsx` — 现有搜索/发申请面板（本 spec 的升级基线）
- `web/src/components/im/UserProfileCard.tsx` — `isStranger` + `onAddFriend`（注意：陌生人模式当前隐藏 region/phone，需放开展示规则，见步骤 6）
- `web/src/components/im/FriendRequestList.tsx` — `apiSendFriendRequest`/`apiSearchUsers` 调用与刷新模式
- `web/src/components/im/GroupList.tsx` — 既有加群弹窗、群申请发送、群详情展示（复用其交互/接口）
- `apps/user/api/...searchUser` logic、`apps/social/rpc/social.proto`（`GroupPutin`/`GroupDetail`/`FindGroupList`）— 改造/新增参照
- 规则：`.claude/rules/go-zero.md`（goctl 生成）、`database-model.md`（三库兼容、无 schema 变更确认）、`rpc-client.md`、`frontend.md`

## 测试计划
- [ ] user search 分页：page1/page2 边界、total 正确、`phone` 精确才返回 mobile、`name` 模糊不返回 mobile
- [ ] group search：群号精确命中、群名模糊命中、分页、total、无结果
- [ ] 前端无限滚动：触底加载、最后一页不再请求、切换关键词重置、快速输入不串页（在途作废）
- [ ] 用户卡片：陌生人字段齐全；已是好友→发消息；待审核→已发送禁用；自己→无按钮
- [ ] 群卡片：未入群→申请加入；已入群→进入群聊；待审核→已申请
- [ ] 申请发送：好友/群 toast 成功；后端幂等（重复申请不重复产生）
- [ ] `go test ./apps/user/... ./apps/social/... -count=1`；前端 `bunx tsc --noEmit` 无新报错
- [ ] 账号 `17585710998 / dsafasf` 端到端走查两条线

## 待定事项
- ~~user search 现有实现是否在 rpc 层~~ → **已确认：搜索在 rpc 层**，分页字段加在 `user.proto` 的搜索 message + rpc logic，api 透传。
- 群「人数」从何处取（群成员表 count vs 群表冗余字段）— 步骤 2/3 确认，避免跨服务直查
- 群名模糊是否需要可见性限制（私密群是否可被搜到）— 默认仅 `status` 正常且非私密群可搜，待确认
- 资料卡是否展示「共同好友 / 共同群」— 暂不做，列入后续

## MVP 范围
- ✅ 用户搜索分页 + 用户资料卡（陌生人，完整字段）+ 加好友（带附言、状态区分、排除自己）
- ✅ 群搜索（群号精确+群名模糊）分页 + 群资料卡 + 申请加群（带附言、状态区分）
- ✅ 空态/错误态/重试、防抖、在途请求作废
- ⏭ 后续：共同好友/共同群、二维码入群整合、按标签筛选
