# Trend 模块联调 Bug 批次修复

## 状态
- 创建日期: 2026-06-04
- 状态: 草稿
- 分支: `feat-trend-api-frontend-integration`（续）

## 目标
修复 trend 前后端联调中暴露的 5 个问题，使发布、点赞、用户信息展示、朋友圈背景、头像跳转全部按预期工作。

## 非目标
- 不重做点赞/动态的整体架构。
- 不引入新 UI 库。
- 背景图本期只做「单张封面」，不做相册/多图轮播。

## 问题清单与根因

| # | 现象 | 根因（已查） | 层 |
|---|------|--------------|----|
| B4 | 发布动态前端报「请求失败」，但其实已创建(`trend_id:27`) | 代理 `web/src/app/api/trend/[...path]/route.ts` 把响应体自身的 `code:1000`(CreateTrendResponse.code) 当成 envelope 错误码(`code!==200`) | 前端代理 |
| B1 | 点赞数每点一次只减不增 10→9→…→0 | 后端 toggle 正确(`GetTrendAgree` 过滤 `state=1`，`AgreeInc` 翻转 state)。疑在前端：`likedTrends` 派生 / 计数来源 / 乐观更新与刷新交互。需实跑定位 | 前端为主 |
| B2 | 顶部封面缺少本人信息 | `login` 已回填 currentUser(name/avatar/introduction)。signature 为空时我移除了 mock 兜底→显示空。需核实封面是否正确渲染 meName/meAvatar，并审计其他位置 | 前端 |
| B3 | 朋友圈背景图应为「查看者本人」设置的图 | 当前 `renderCoverSection` 硬编码 `picsum.photos/...`；user 模型**无封面字段** | 全栈(含 schema) |
| B5 | 点动态里头像/昵称应进好友/用户详情页 | `ContactDetailPanel` 存在，经 `setActiveTab('contacts')+setSelectedContactId(id)` 打开；feed 未接 onClick | 前端 |

## 技术设计

### B4 代理 envelope 判定（先做，阻断）
`route.ts` 仅当响应体**同时含** `code:number` 且 `msg:string` 时才按 go-zero envelope 处理；否则视为原始 payload 直接放进 `data`。
```ts
const isEnvelope = typeof obj.code === 'number' && typeof obj.msg === 'string';
if (isEnvelope && obj.code !== 200) return fail(obj.msg);
return { success: true, data: isEnvelope ? (obj.data ?? {}) : obj };
```

### B1 点赞计数（实跑定位 + 修复）
- 后端已确认正确，重点查前端：
  1. 点赞数显示来源（`trend.agreeCount` vs `likeUsersMap[id].length`）；
  2. `likedTrends` 初始/刷新派生（`loadFeed` 用 batch-summary「me 是否在 likers」）；
  3. `handleToggleLike` 乐观更新与 `bumpTrendVersion`/`seenVersionsRef` 跳过刷新的交互；
  4. trend 模型 `agree_count` 走缓存导致回读旧值的可能。
- 修复目标：连续点击应在 已赞/未赞 间正确切换，计数 +1/−1 振荡。

### B2 封面本人信息
- 进入已登录态时（`page.tsx`）拉 `/api/v1/user/detail` 并 `updateUser` 回填，避免 persist 旧数据缺字段。
- 核实 `renderCoverSection` 是否渲染 `meName/meAvatar`；signature 为空显示占位（不复活 mock）。
- 审计 feed 卡片、评论、点赞列表是否有其他「ID 当名字」的位置并修正。

### B3 朋友圈背景图（schema 变更，已确认）
- **数据库**：user 表新增 `moments_cover TEXT`（默认空串）。迁移走 ADD COLUMN，兼容 SQLite/MySQL/PostgreSQL，让 GORM 处理。
- **RPC/API**：`user/detail` 返回 `moments_cover`；`user/update` 接受可选 `moments_cover` 写入。
- **前端**：`renderCoverSection` 用 `currentUser.momentsCover`，空则回退到现有渐变/默认图；自己可点封面→复用 `/api/v1/user/avatar/upload`（或 im 上传）上传后 `user/update` 落库并 `updateUser`。
- AuthUser 增加 `momentsCover` 字段，登录/detail 回填。

### B5 头像/昵称跳详情页（统一）
- feed 与详情页的头像/昵称加 onClick：`setActiveTab('contacts'); setSelectedContactId(backendUid)`。
- 「自己」也统一跳详情页（决策）。注意 `ContactDetailPanel` 从 `friends` 按 id 找联系人；自己/非好友不在列表 → 需让详情页支持按 uid 拉取用户资料（`user/search?ids=` 或 `user/detail`）。本期最小实现：好友直接跳；自己/非好友走「按 uid 取资料」的详情。

### 参考的现有模式
- `apps/user` user/detail、user/update、avatar/upload 既有链路。
- `.claude/rules/database-model.md` — ADD COLUMN、TEXT、三库兼容、不改 .env。
- 上一轮 `docs/specs/trend-frontend-integration.md` 的联调方法（临时实例 :8896 + 真实 token curl）。

## 异常处理
| 场景 | 处理 |
|------|------|
| 发布成功但 data 含 code 字段 | 代理不再误判为失败 |
| moments_cover 为空 | 前端回退默认封面 |
| 点击非好友头像 | 详情页按 uid 拉资料，无好友关系也能看基础资料 |
| 上传封面失败 | toast 报错，不改本地状态 |

## 测试计划（实跑，账号 17585710998）
- [ ] B4：前端发纯文字动态 → 成功提示，不再「请求失败」
- [ ] B1：连续点赞同一动态 → 计数在 n/n+1 间振荡，刷新后一致
- [ ] B2：封面显示本人头像/昵称；其他位置无「ID 当名字」
- [ ] B3：设置封面图后，自己与其他设备看到的「我的朋友圈」封面为该图
- [ ] B5：点头像/昵称 → 进入该用户详情页

## 待定事项
- B5「自己/非好友」详情页的资料拉取方式（user/detail 仅本人；他人用 user/search?ids=）——实现时定。
- B1 具体根因以实跑为准。

## 实现步骤（每步可独立 commit）
1. [ ] B4 代理 envelope 判定
2. [ ] B2 封面本人信息 + 审计其他位置
3. [ ] B1 点赞计数实跑定位并修复
4. [ ] B5 头像/昵称跳详情页
5. [ ] B3 user.moments_cover：model + rpc/api + 前端展示与上传
6. [ ] 起服务端到端验证全部 5 项

## MVP 范围
全部 5 项（阻断 B4/B1 优先）。B3 含 schema 变更（已确认）。B5 自己/非好友详情若超时则先保证好友可跳。
