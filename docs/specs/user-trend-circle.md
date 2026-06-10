# 我的朋友圈 / 好友朋友圈（按用户查看动态）

## 状态
- 创建日期: 2026-06-10
- 状态: 草稿

## 目标
让用户能查看「指定用户（自己或好友）」的动态集合（列表 + 详情 + 点赞列表 + 评论列表），样式完全复用现有公共动态流与详情组件；并在打通入口的同时补齐可见性边界，保证：
1. 看别人朋友圈时遵循微信式可见性规则，不泄露对方「仅自己可见」的动态、非好友拒绝访问；
2. 不能在好友/他人的朋友圈里发动态（发布入口仅在自己朋友圈出现，作者由 JWT 强制）。

## 非目标
- **逐好友动态可见性设置**（类微信「不看TA / 不让TA看我」黑白名单）——后续迭代，本次不做。
- 不新建独立 URL 路由（前端沿用 `MomentsFeed` 内 `userTrends` 视图状态机）。
- 不改动现有公共动态流（`getLatestTrends`）的查询逻辑。
- 不改任何数据库 schema（本需求不需要建表/加字段）。

## 用户故事
- 作为登录用户，我想在「我的-朋友圈」或自己的资料卡里查看**我自己的全部动态**（含仅自己可见），以便回顾和管理。
- 作为登录用户，我想在好友的资料卡/详情页、或公共动态流里点头像，进入**这位好友的朋友圈**，只看到他对我可见的动态，以便了解他的近况。
- 作为登录用户，进入别人的朋友圈时**不应看到发布按钮**，也无法把动态发到别人名下。
- 作为登录用户，从他人朋友圈点开任意一条动态详情/点赞列表/评论列表时，体验与公共动态流一致。

## 核心流程

### A. 我的朋友圈（happy path）
1. 用户在「我的-朋友圈」入口 或 自己资料卡「查看更多」点击进入。
2. 前端 `setUserTrendsUserId(meUserId)` + `setView('userTrends')`，调用 `getUserTrends(token, meUserId)`。
3. 后端识别 `targetUserId == currentUid` → 返回**全部** scope（含 scope=1 仅自己可见）+ 置顶动态。
4. 视图顶部显示发布入口（仅自己朋友圈）；列表/详情/点赞/评论复用现有组件。

### B. 好友朋友圈（happy path）
1. 用户从好友资料卡/详情页「查看TA的朋友圈」，或公共流点击好友头像后的卡片「进入TA的朋友圈」进入。
2. 前端 `setUserTrendsUserId(friendId)` + `setView('userTrends')`，调用 `getUserTrends(token, friendId)`。
3. 后端识别 `targetUserId != currentUid` → 先校验好友关系：
   - 非好友 → 返回业务错误（前端提示「你们还不是好友」/无权限）。
   - 是好友 → 只返回对方**对外可见**的动态（`circle_state=1`，即 scope=好友/所有人），过滤掉「仅自己可见」。
4. 视图**不显示发布入口**；详情点开走带可见性校验的 `getTrendDetail`。

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 查看非好友的朋友圈 | 后端 `GetUserTrends` 返回业务错误（`pkg/xerr` 包装），前端展示空态+提示，不暴露任何动态 |
| 好友把动态设为「仅自己可见」 | 后端过滤掉（不进结果集） |
| 看自己的朋友圈 | 放行全部 scope，含仅自己可见 |
| 列表为空 | 前端显示空态（复用现有空态） |
| 进入他人朋友圈尝试发动态 | 前端无发布按钮；即便构造请求，后端 `createTrend` 作者取 JWT，仍落到自己名下，不会污染他人朋友圈 |
| 从他人朋友圈点开「仅自己可见」的详情（理论上不会出现，因列表已过滤） | `getTrendDetail` 已有 scope 校验兜底，返回「仅作者可见」错误 |
| social FriendList RPC 失败 | 错误向上传播（不吞），前端提示加载失败 |

## 技术设计

### 数据模型
**无 schema 变更。** 复用现有表：
- `trend`：`userid`(作者)、`scope`(1仅自己/2仅好友/3所有人)、`circle_state`(1可见/2不可见/0删除)、`state`、`is_top`。
- `trend_agree`(点赞)、`trend_discuss`(评论) 复用现有。

可见性已由数据层 `circle_state` 表达：scope=1 → circle_state=2；scope=2/3 → circle_state=1。

### API 接口
**复用现有接口，不新增路由。** 仅修正后端 logic 行为：

| 方法 | 路径 | 说明 | 改动 |
|------|------|------|------|
| GET | /v1/user/trends | 指定用户动态列表(+置顶) | **加可见性逻辑**（self vs friend vs 非好友） |
| GET | /v1/trend/detail | 动态详情 | 已有 scope 校验，无需改 |
| GET | /v1/trend/like/users | 点赞用户列表 | 复用 |
| GET | /v1/trend/comment/root / children / tree | 评论 | 复用 |
| POST | /v1/trend | 发动态 | 作者已取 JWT，无需改（前端按入口隐藏按钮） |

RPC 侧 `GetUserTrends` 已支持 `Scope` 参数（VisibilityScope），底层 `ListByUserIds` 按 scope 映射 `circle_state` 过滤，无需改 proto。

### 实现步骤（每步可独立 commit）

**后端（apps/trend）**
1. [ ] `apps/trend/api/internal/logic/trend/getusertrendslogic.go`：
   - 取当前用户 `currentUid := utils.GetUser(l.ctx)`。
   - `targetUserId == currentUid` → 传 `Scope: 3`（全部，保留现状），并查置顶。
   - `targetUserId != currentUid` → 调 `social.FriendList` 校验好友关系：
     - 非好友 → 返回 `pkg/xerr` 业务错误。
     - 是好友 → 传 `Scope: 2`（仅 circle_state=1 对外可见）。置顶同样按对外可见过滤。
   - 校验逻辑参考 `gettrenddetaillogic.go:56` 现有好友判定写法。
2. [ ] （如需）核对 `GetUserTopTrend` 对他人是否也应按可见性过滤——他人置顶若 scope=1 需排除。

**前端（web）**
3. [ ] **入口 1 - 我的朋友圈（个人中心）**：`ProfilePage.tsx` 菜单加「我的朋友圈」项 → `IMLayout` 新增 subPage 分支，进入 `userTrends` 视图并传 `meUserId`。
4. [ ] **入口 2 - 资料卡/详情页**：`UserProfileCard.tsx` 的「朋友圈」区块「查看更多」绑定 `onViewMoments` 回调（`FloatingProfileCard` / `ContactDetailPanel` 透传），进入 `userTrends` 视图并传该 userId。
5. [ ] **入口 3 - 公共流点头像**：`FloatingProfileCard` 加「进入TA的朋友圈」按钮，回调 `setView('userTrends')` + `setUserTrendsUserId(userId)`。
6. [ ] **发布边界**：`MomentsFeed` 的 `userTrends` 视图中，发布按钮仅当 `userTrendsUserId === meUserId`（或 `'me'` 哨兵）时渲染。
7. [ ] **「我的/好友」标题与空态**：视图标题按是否本人显示「我的朋友圈 / TA的朋友圈」；非好友错误展示空态提示。

### 参考的现有模式
- `apps/trend/api/internal/logic/trend/gettrenddetaillogic.go:56` — 好友关系校验 + scope 鉴权写法，直接照搬到 list。
- `apps/trend/api/internal/logic/trend/getlatesttrendslogic.go` — `utils.GetUser(ctx)` 取当前用户 + `social.FriendList` 调用方式。
- `apps/trend/models/trendmodel_gen.go` `ListByUserIds` — scope→circle_state 映射（scope=2 → circle_state=1，scope=3 → in(1,2)）。
- 前端 `web/src/lib/trend-api.ts` `getUserTrends` — 已封装，直接复用。
- 前端 `web/src/components/im/MomentsFeed.tsx` `view==='userTrends'` — 已完整实现的用户动态视图，复用为两种朋友圈的统一目标。
- 前端 `web/src/components/im/UserProfileCard.tsx` L682-746 — 已有朋友圈区块，「查看更多」当前空实现。

## 测试计划
- [ ] 后端 table-driven：`GetUserTrends` 三种关系——本人（含 scope=1）、好友（仅对外可见、不含 scope=1）、非好友（报错）。
- [ ] 后端：置顶动态在「看他人」时不返回 scope=1 的置顶。
- [ ] 后端测试需在 SQLite/MySQL/PostgreSQL 均通过，测试数据用后清理，不 mock 数据库。
- [ ] 前端手动验证（账号 17585710998）：三个入口均能进入对应朋友圈；自己朋友圈有发布按钮、他人朋友圈无；点详情/点赞/评论正常；非好友进入有空态提示。

## 待定事项
- 「看他人朋友圈」时，置顶动态是否需要同样过滤 scope=1（步骤 2）——倾向需要，开发时确认 `GetUserTopTrend` 的 scope 行为。
- 非好友访问的错误文案 / 是否区分「未关注」与「无权限」——前端确认。
- 公共流点头像：是保留现有浮动卡片 + 卡内按钮，还是直接跳朋友圈——本 spec 采用「卡内加按钮」，避免改变现有点头像行为。

## MVP 范围
1. 后端 `GetUserTrends` 可见性过滤（self / friend / 非好友）。
2. 前端三个入口打通到统一 `userTrends` 视图。
3. 发布按钮仅在自己朋友圈出现（数据边界）。
4. 详情 / 点赞列表 / 评论列表直接复用现有组件。

**不在 MVP**：逐好友动态可见性黑白名单设置（后续迭代）。
