# 好友与群申请交互可靠性修复

## 状态
- 创建日期: 2026-07-17
- 状态: 草稿
- 关联文档: `docs/specs/add-friend-and-group.md`、`docs/specs/common-notify-channel.md`
- Schema 变更: 已确认，迁移必须兼容 SQLite/MySQL/PostgreSQL

## 目标
修复好友申请和入群申请在授权、并发、通知可靠性、个人未读、实时刷新及关系同步上的断点，使双方在线时及时看到一致状态，离线后也能从持久化数据完整恢复。

## 设计原则
- Social 数据库中的申请、关系和个人回执是业务真相；WebSocket 只发送失效通知，不承载完整申请状态。
- 前端收到 `friend.*` / `group.*` 通知后，使相关查询失效并通过 REST 重拉，不直接猜测服务端终态。
- 所有“只能处理一次”的状态转换必须由数据库 CAS 或唯一约束保证，不能依赖先查后写。
- 申请状态、关系变更、个人回执和待投递通知必须在同一个 Social 数据库事务中提交。
- 公共通知中心与业务申请回执保留各自数据，但用户进入业务页面时联动标记对应通知已读，避免两个红点长期冲突。
- 公开 HTTP 接口中的操作者身份只能来自 JWT；RPC 仍须做纵深授权校验。

## 已确认的群申请与邀请规则
- 用户主动申请 `is_verify=0` 的群时直接入群，不进入管理员审批列表，并立即可靠创建群会话。
- 被拒绝后允许重新申请并保留每次历史；同一用户对同一群 60 秒内最多提交一次主动申请，同一时间最多一条主动 pending。
- 群主和管理员都可审批，审批时按当前角色重新鉴权。
- 群主和每个管理员维护独立已读；一人查看不替其他人标读。
- 管理员 A 处理后，管理员 B 的待处理动作立即结束，但仍能看到终态、处理时间和“未查看结果”红点；申请者不看到具体处理管理员。
- 拒绝申请的原因可选并对申请者可见；不展示具体处理管理员身份。
- 本期不支持申请者撤回 pending 入群申请，也不预留对外取消接口。
- 审批时群已解散、封禁或达到人数上限，申请转为失效终态并通知申请者。
- pending 期间通过其他入口入群时，原申请自动转 accepted/resolved，并记录实际入群来源。
- 新管理员上任后可看到全部现存 pending，并为其补个人未读回执；被撤销管理员立即失去审批权。
- 所有群成员都可邀请其他用户；被邀请人必须确认，不允许无感直接拉人入群。
- 普通成员邀请被确认后，生成/进入正常管理员审批流程，并向管理员展示申请人和邀请人。
- 用户已有主动 pending 时接受普通成员邀请，仍为该邀请创建第二条独立审批；不合并来源。任意一条审批通过后，其余同群同用户 pending 自动 accepted/resolved；拒绝一条不影响其他审批。
- 群主或管理员邀请被确认后直接入群；邀请人的权限以被邀请人确认时为准，若已降为普通成员则转管理员审批。
- 每一条邀请独立存在；同一用户可同时收到同群不同成员的多条邀请。接受任意一条后，其他 active 邀请自动失效。
- 邀请 7 天过期；拒绝邀请的原因可选；接受或拒绝均不通知邀请人。
- 被邀请人在群申请面板的独立“收到的邀请”Tab 查看和处理邀请。
- 所有申请、邀请及其终态历史全部保留并分页，本期不自动清理。
- 进入群申请/邀请页面后，按本次成功展示的记录联动清除对应业务回执和通知中心未读。

## 非目标
- 不把申请完整数据放进 WebSocket 帧，不实现纯前端增量合并协议。
- 不统一动态、系统设置等其他业务通知。
- 不增加通知偏好、免打扰、通知聚合或通知清理策略。
- 不重做好友、群资料卡和联系人页面的视觉设计。
- 不改变群主、管理员的既有审批角色定义。
- 不在本期解决多 WebSocket 节点间的用户路由；本期只保证当前节点内同一用户的全部连接都收到通知。
- 不通知邀请人“对方接受/拒绝邀请”的结果。

## 用户故事
- 作为被申请人，我收到好友申请时，希望“新的朋友”红点和已打开的申请列表及时更新，并且刷新或重新登录后仍能看到申请。
- 作为好友申请人，我希望对方同意或拒绝后立即看到结果、附言和正确红点；同意后联系人列表应及时出现新好友。
- 作为群主或管理员，我希望只看到和处理自己有权限管理的群申请，并拥有独立于其他管理员的未读状态。
- 作为入群申请人，我希望审批结果及时更新；通过后群列表和群会话可可靠出现，不依赖一次性后台 goroutine。
- 作为多端用户，我希望当前 WebSocket 节点上的所有在线设备都收到相同通知，离线设备上线后可以通过持久化列表恢复。
- 作为系统维护者，我希望重复请求、并发审批、Kafka 故障和进程重启都不会造成重复关系、状态冲突或永久丢失通知。

## 当前问题与具体修复方案

| # | 当前问题 | 根因 | 具体修复方案 | 验收结果 |
|---|---|---|---|---|
| P01 | 好友申请人可同意或拒绝自己发出的申请 | `FindOne` 允许申请双方查询，处理逻辑未强制当前用户等于 `req_uid` | HTTP 用户 ID 仅取 JWT；RPC 加 `actor_uid == request.req_uid` 校验；申请人取消申请走独立 cancel 语义，不复用审批接口 | 申请人调用审批接口返回 403/业务无权限，且不改变申请和好友关系 |
| P02 | 群申请可伪造 `req_id`、`join_source`、`inviter_uid`，甚至替其他用户直接入群 | 公开申请接口信任客户端身份和来源字段 | 普通申请请求仅接受 `group_id`、`req_msg`；`applicant_uid` 和主动申请来源由服务端生成；邀请、token 入群继续走独立入口；RPC 增加 actor 并复核来源权限 | 普通用户无法替他人申请或入群，伪造字段无效或直接拒绝 |
| P03 | `/group/putInsByUid` 可通过 `ids` 查询他人申请 | API 仅在 `ids` 为空时使用 JWT | 公开 API 忽略/移除 `ids`，始终以 JWT UID 查询；内部批量能力只保留 RPC，不暴露给终端用户 | 任意 `ids` 参数都不能扩大当前用户可见范围 |
| P04 | `handle_result` 可写任意整数 | API/RPC 无枚举校验 | 好友和群审批均只接受 `1=accept`、`2=reject`；其他值在 API 和 RPC 两层拒绝；`ignored` 只允许服务内部状态转换 | `0/3/负数/99` 均不写数据库、不发通知 |
| P05 | 可以向自己发好友申请，目标用户为空时仍可能继续 | self-check 和目标用户存在性校验不完整 | API 和 RPC 均拒绝 `actor == target`；User RPC 返回空用户按不存在处理；已是好友返回明确的幂等结果码 | 自申请不落库；不存在用户不落库；已是好友不显示“已发送申请” |
| P06 | 并发发起会产生重复 pending 申请 | 先查后插且没有唯一约束 | 好友申请和群主动申请使用 nullable `active_key` 防重；邀请转审批使用稳定 `source_invitation_id` 唯一约束，每个邀请允许独立 pending；事务内 create-or-return-existing | 同一好友申请或群主动申请并发只保留一个 pending，每条邀请最多生成一条独立审批 |
| P07 | 群重复申请先物理删除旧记录，丢失历史和审计 | 查询旧行后事务外删除再插入 | 删除物理删除逻辑；已处理后重申创建新行；已有 pending 返回现有记录或更新留言，行为由服务端统一 | 历史申请保留，通知可追踪到稳定 request ID |
| P08 | 并发同意/拒绝可导致重复关系或“已拒绝但已是好友/群成员” | 事务外读取、无行锁/CAS、关系唯一约束不完整 | 在事务内执行 `UPDATE ... WHERE id=? AND handle_result=0` 并检查 `RowsAffected`；赢得 CAS 后才能写关系、回执和 outbox；相同终态重复请求幂等成功，不同终态返回 409 | 同一申请只有一个终态；不会同时产生 accept/reject 或冲突关系 |
| P09 | 好友关系可能重复 | `friends` 缺少 `(user_id, friend_uid)` 唯一约束 | 迁移前清理有向重复记录，新增联合唯一索引；同意时双向插入使用幂等 create，任何异常回滚整个事务 | 每个有向好友关系最多一行，双向关系与申请终态原子一致 |
| P10 | 群事务使用 `defer tx.Commit()` 或忽略 Commit 错误 | 裸 Begin/Commit 控制不完整 | 全部改为 `db.Transaction(func(tx *gorm.DB) error)`；所有写入检查错误和 RowsAffected；仅事务成功返回后响应成功 | Commit 失败不会发送成功响应或成功通知 |
| P11 | Social 状态写成功后 Kafka 失败会永久丢通知；群 apply 甚至可能在 Commit 前发通知 | 直接 best-effort Push | 新增 `social_notification_outbox`；申请、审批与每个接收者的通知事件同事务写 outbox；relay 重试投递 Kafka，成功后标 sent | Kafka 中断期间业务可提交；恢复后通知最终进入 `notifications` |
| P12 | 群管理员共用 `receiver_read`，一人查看会替所有管理员已读 | 已读状态存在申请行上，没有 receiver 维度 | 新增统一 `social_request_receipts`，键为 `(request_type, request_id, receiver_id, receipt_kind)`；群申请为每个群主/管理员建立独立 apply receipt；结果为申请人建立 result receipt | 管理员 A 已读不影响管理员 B；申请人结果也有业务红点 |
| P13 | 好友申请 `receiver_read/sender_read` 与群回执模型不一致 | 两套历史模型分别演进 | 新逻辑和计数统一读取 receipt 表；旧字段在兼容迁移期回填并停止写入，本期不删除旧列；后续单独清理 | 好友和群申请使用同一套个人未读语义 |
| P14 | 公共通知已读与业务申请已读长期冲突 | `notifications.is_read` 和申请 read 独立操作 | 进入好友/群申请页面后，先拉取并标记对应 receipt，再调用 IM 批量按 `notify_type + biz_id` 标读；点击通知中心仍先标该 notification，再导航；两个操作均幂等 | 进入业务页面后对应业务红点与铃铛通知同步消除，不清除无关通知 |
| P15 | `friend.accept/reject`、`group.accept/reject` 不实时增加业务红点 | WS switch 只给 apply 类型加计数 | 不再手工 `+1` 作为最终真相；所有好友、群申请和邀请通知都 bump 对应 request version 并重拉 unread count；短暂 UI 可保留乐观 +1，但必须被服务端 count 覆盖 | 在线结果通知后“新的朋友/群申请”红点与后端一致 |
| P16 | 已打开的好友或群申请页不会实时刷新 | 页面只在 mount 拉取，未订阅业务版本 | store 增加 `friendRequestsVersion`、`groupRequestsVersion`；收到相应通知后递增；页面订阅版本并重拉，使用请求序号或 AbortController 丢弃过期响应 | 页面保持打开时，新申请和审批结果自动出现 |
| P17 | 同意好友后申请者联系人列表不刷新 | 只有审批方调用 `invalidateFriends()` | 收到 `friend.accept` 后同时 bump friend request version 和 `invalidateFriends()`；关系 `friend.added` 事件也作为可靠兜底触发刷新 | 双方联系人列表及时出现新好友 |
| P18 | 入群通过后申请者群列表、群会话和成员信息不刷新 | 没有 groups version；建会话依赖 API goroutine | 增加 `groupsVersion` / `invalidateGroups()`；`group.member.added` 消费端幂等确保群会话存在并推 `relation.changed`；前端刷新群列表、会话列表及当前群详情 | 申请通过后群和会话可靠出现，IM 暂时失败可重试 |
| P19 | 免验证群已直接入群，前端仍提示“申请已发送” | helper 只返回 boolean，丢弃 `is_pass` | `sendGroupRequest` 返回 `{groupId,isPass,requestId}`；`isPass=true` 时提示已加入、刷新群和会话并切换按钮状态 | 直接入群不展示待审批状态 |
| P20 | 好友处理附言和处理时间在列表中丢失 | proto/API 未透传，API 将 `handle_msg` 写死为空 | RPC/API response 增加 `handle_msg`、`handled_at`；由 model 原样映射；新增字段保持 optional/omitempty | 刷新后仍能看到同意附言或拒绝原因及处理时间 |
| P21 | “我发起”的好友申请可能显示自己的 ID | 前端始终用 `user_id` 作为对方 ID | sent 使用 `req_uid`，received 使用 `user_id`；后端新增明确 `peer_uid`，前端优先使用，避免继续依赖方向判断 | 两个 Tab 都显示正确对方资料和 HiChat ID |
| P22 | 好友页面挂载时标读和未读 count 并发，旧响应可覆盖 0 | 两个 effect 独立执行 | 合并为一个顺序流程：拉列表和 count -> 渲染 -> 标 receipt/notification 已读 -> 以标读响应中的最新 count 更新 store；版本变化时复用同一 refresh 函数 | 不再出现已读后红点回弹 |
| P23 | 批量已读后本地卡片仍保留 unread | 只清 store count | 标读成功后同步更新当前视角下匹配记录的 `readState=true`；失败保留未读并提供重试 | 卡片样式、Tab 红点和总红点一致 |
| P24 | 好友列表一次发 8 个请求，好友和群申请均无分页 | 按 class/status 笛卡尔积拉全量历史 | 新增统一分页查询参数 `class/status/page/size`；`status` 可选表示全部；好友页面每个 Tab 一次请求；badge 走独立 count API，不从全量列表计算 | 首屏每个 Tab 最多一次列表请求，历史增长后仍可分页 |
| P25 | 群申请 badge 在 IMLayout、ContactList、GroupList 多处重复拉取 | 缺少统一 query/version 所有者 | badge 拉取收敛到 store action；登录/重连/版本变化触发一次；组件只订阅状态，不自行重复计算全量列表 | 相同状态变化不产生三份重复请求和竞态写入 |
| P26 | WS 只发送 `rconn[0]` | handler 只选择第一条连接 | 当前节点内遍历该用户全部连接逐一发送；单连接失败不阻断其他连接；离线仍依赖通知表恢复 | 同一用户多个标签页/设备均收到通知 |
| P27 | Kafka 重投时通知已落库会跳过 WS 重试，语义不明确 | 落库幂等和实时 delivery 混在一起 | 明确 WS 为 best-effort：通知落库成功即确认 Kafka；WS 失败记录指标但不返回重试错误；客户端重连后立即重拉 unread/list | 不产生无效重投，离线/推送失败由持久化 REST 恢复 |
| P28 | 通知点击只能进入泛化 Tab，不能定位申请 | 导航只传 notifyType | outbox payload 固定带 `requestType/requestId/groupId/result`；store 导航目标增加 requestId/groupId；页面重拉后定位并高亮目标 | 点击 Toast/通知可定位对应申请，记录不存在时显示明确提示并刷新关系列表 |
| P29 | 前端实时通知和部分群资料卡文案硬编码中文 | 未统一走 i18n | 为好友/群申请/邀请通知、错误、重复、并发冲突和直接入群增加中英文 key；业务 store 不直接拼中文 | 中英文环境显示对应文案，无新增硬编码业务文案 |
| P30 | 列表错误被吞并显示为空，用户无法区分故障与无数据 | `.catch(() => {})` 和 RPC 吞 DB error | 后端返回真实错误；前端保留最后成功数据并显示 stale/retry 状态；失败不把 badge 强制清零 | API 故障时不误显示“没有申请”，用户可重试 |
| P31 | 多个 Social API logic 用 JWT claim 强制类型断言，异常上下文会 panic | 直接调用 `ctx.Value(...).(string)` | 好友和群申请入口统一使用 `ctxdata.GetUId` 或返回 error 的安全 helper；缺失、空值、类型错误统一返回 401；禁止 handler/logic panic | 缺失或畸形 claim 返回 401，服务进程稳定且不调用 RPC |
| P32 | 好友审批的 `tags` 可能写到申请人一侧，而不是审批人自己的好友记录 | 两条有向 friend 记录的字段归属混淆 | 明确 `remark/tags` 属于提交审批的当前用户视角，写入 `friends(user_id=actor,friend_uid=applicant)`；申请时预设备注只写申请人视角；增加双向字段断言测试 | 双方看到各自独立备注/标签，不互换、不覆盖 |
| P33 | `.api`、proto、生成类型、GORM object 和基础 SQL 存在漂移 | 手工定义与生成/迁移来源未同步，如 `handle_msg`、`status`、`remark` | 先修改 `.api/.proto` 和 `pkg/db/objects` 作为契约源，再运行 goctl 重新生成非手写文件；版本化迁移补齐 schema；CI 增加生成后 diff/schema 校验；不直接修改 `_gen.go` | 重新生成不会丢字段，三库新建环境与升级环境 schema 一致 |
| P34 | task/mq 直接使用 IM notification model 写表，弱化服务边界 | 公共通知消费者绕过 IM RPC | 补齐并使用幂等 `im-rpc CreateNotification`；task 只消费事件并调用 IM RPC，notifications 表继续由 IM 服务拥有；RPC 可重试错误交给 Kafka，参数错误进入 dead-letter/告警 | task 不再直接持有或写 IM model，重复消费仍只创建一条通知 |
| P35 | 当前申请和邀请复用 `GroupPutin` 字段，无法表达“用户确认”和两段式审批 | `req_id/join_source/inviter_uid` 同时承担操作者、申请人和来源 | 将主动申请与邀请拆成独立命令和状态机；邀请记录保存 inviter/invitee、确认终态、邀请时角色快照和 7 天有效期；确认时重新查邀请人当前角色 | 普通成员邀请确认后进入审批，管理角色邀请确认后直接入群，权限变化不会越权 |
| P36 | 同一用户收到同群多个邀请时可能互相覆盖 | 旧群申请按 `(group_id,req_id)` 删除重建，没有独立邀请 ID | 每次邀请创建独立记录和稳定 ID，不对邀请建立“同群同用户唯一 active”约束；接受一条时事务内将同群同 invitee 的其他 active 邀请置 invalidated | 多个邀请均可展示；接受一个后不能再接受其余邀请 |
| P37 | 邀请不会自动过期 | 模型没有 `expires_at` 和过期状态 | 邀请创建时设置 `expires_at=created_at+7d`；列表查询和确认接口均执行过期判定；cron 批量转 expired 负责及时红点收敛，确认接口的条件更新负责最终正确性 | 7 天后邀请不可接受，状态稳定显示为已过期 |
| P38 | 新管理员看不到任职前 pending，或旧管理员仍可处理 | 管理员 receipt 只在申请创建时生成，角色变更未同步 | 管理员授予事务/事件为所有现存 pending 幂等补 apply receipt；撤销时将其 action 状态置 resolved/revoked；审批仍按当前角色鉴权 | 新管理员立即看到现存待审批，撤销管理员无法处理 |
| P39 | “已处理”和“已阅读”混成一个 receipt state，无法表达未查看终态 | 单一枚举 `unread/read/resolved` 丢失两个维度 | receipt 改为独立 `is_read` 与 `is_actionable`，另存 `result/resolved_at`；审批后所有管理员 `is_actionable=0`，仅审批人 `is_read=1`，其他人保留 `is_read=0` | 管理员 B 无待处理按钮，但仍有未查看终态红点 |
| P40 | 群异常、其他路径已入群或并行审批通过时 pending 无稳定收口 | 只更新当前人工审批记录 | 增加 `invalidated` 终态及 reason；群异常时转 invalidated 并通知申请者；任一合法入口成功入群时，事务内将同群同用户的其他 pending 转 accepted/resolved，记录 actual_join_source 并关闭管理员 action receipt | 不遗留永远 pending 的申请，成员关系和所有并行申请终态一致 |

## 核心流程

### A. 发起好友申请
1. 前端仅提交目标 UID、验证消息和预设备注。
2. Social API 从 JWT 取得申请人，校验目标存在、状态正常、不是自己。
3. Social RPC 再次校验 actor、目标和既有好友关系。
4. RPC 在事务内尝试创建带 `active_key=friend:<sender>:<receiver>` 的 pending 申请。
5. 若 active key 冲突，读取并返回已有 pending；不插入第二行，不重复创建 apply receipt。产品语义为“申请仍在等待”，可更新留言但不自动重复通知；显式“再次提醒”后续单独设计。
6. 同事务为被申请人创建 apply receipt，并写一条 `friend.apply` notification outbox。
7. 事务成功后返回稳定 `request_id/status`；relay 异步投 Kafka。
8. 被申请人在线收到 `notify`，前端 bump version 并重拉申请列表和 unread count；离线时登录后通过 REST 恢复。

### B. 处理好友申请
1. 被申请人提交 `request_id`、`result=accept|reject`、可选 `handle_msg/remark/tags`。
2. API 从 JWT 取得 actor；RPC 验证 actor 必须等于申请的 `req_uid`。
3. 在事务内 CAS 将 pending 更新为终态，同时清空 `active_key`。
4. accept 时幂等写入双向好友关系和 relation outbox；reject 不写好友关系。
5. 同事务将接收方 apply receipt 标记 handled/read，为申请人创建 result receipt，并写 `friend.accept/reject` notification outbox。
6. 相同结果重复提交返回当前成功结果；不同结果重复提交返回 409，不覆盖先到终态。
7. 申请人收到通知后重拉“我发起”、unread count；accept 同时刷新好友列表。

### C. 发起入群申请
1. 普通申请接口只接受 `group_id/req_msg`，申请人和来源来自 JWT/服务端。
2. RPC 校验群存在、未解散、申请人不是成员。
3. 检查同一用户和群最近一次申请时间；60 秒内重复提交返回 429 和可重试时间，不创建新历史。已有 pending 优先返回原 request ID。
4. 需审批时，在事务内创建带 `active_key=group:direct:<groupId>:<applicant>` 的主动 pending 申请。
5. 查询事务时点上的群主和管理员，为每人创建独立 apply receipt，并为每人写 notification outbox。
6. 免验证时直接在事务内创建 accepted 申请、群成员和 `group.member.added` relation outbox，不创建管理员 apply receipt。
7. 前端依据 `is_pass` 区分“已加入”与“等待审批”。

### D. 处理入群申请
1. 群主/管理员提交 request ID 和 accept/reject；请求不再依赖客户端 group ID。
2. RPC 根据申请记录的 group ID 验证 actor 当前仍具有审批权限。
3. 事务内 CAS pending 到终态并清空 active key。
4. accept 时幂等创建群成员和 relation outbox；reject 不创建成员。
5. 同事务将审批人的 receipt 标记 read，并将全部管理员 apply receipt 标记 resolved；为申请人创建 result receipt，写结果 notification outbox。
6. 所有管理员 receipt 的 `is_actionable` 置 0；审批人 `is_read=1`，其他未查看管理员保持 `is_read=0`，以显示“未查看终态”。
7. relation consumer 幂等确保群会话存在并推关系变化；申请人前端刷新群列表、会话列表和申请状态。

### E. 邀请入群
1. 任意当前群成员可以提交 invitee UID；RPC 校验邀请人确实在群内、目标不是成员、群状态允许邀请。
2. 每次邀请创建独立 `group_invitation`，即使同一 invitee 已有同群其他 active 邀请也不合并；设置 7 天有效期。
3. 同事务为 invitee 创建 invite receipt 和 `group.invite` notification outbox；不向邀请人发送结果通知。
4. 被邀请人在“收到的邀请”Tab 查看，接受或拒绝时 actor 必须等于 invitee；拒绝原因可选。
5. 接受时重新读取邀请人的当前群角色：群主/管理员邀请直接创建成员；普通成员邀请为该 invitation 创建独立 pending 审批，不复用已有主动申请或其他邀请审批，管理员侧展示申请人和邀请人。
6. 接受任意一条邀请时，同事务将同群同 invitee 的其他 active 邀请置 invalidated；普通成员邀请对应的审批通过后，将同群同申请人的其他 pending 审批联动为 accepted/resolved；拒绝该审批不影响主动申请或其他已形成的审批。
7. 邀请过期由确认接口 CAS 兜底，并由 cron 定时批量转 expired、关闭回执红点。
8. 邀请人的角色变化按确认时角色生效；邀请人已退群时邀请转 invalidated，不能借历史权限入群。

### F. 已读联动
1. 好友或群申请页按当前 Tab 拉取申请列表和个人 unread count。
2. 页面可见且列表成功加载后，调用 Social receipt 标读接口，只标记当前用户、当前方向、当前已展示申请。
3. Social 标读响应返回最新业务 unread count，前端以响应覆盖本地值。
4. 前端随后调用 IM 通知标读接口，按当前用户和对应 `notify_type + biz_id[]` 标记通知中心记录；不清除其他类型通知。
5. 任一步失败均可幂等重试；业务列表不因标读失败而隐藏。

## 异常处理

| 场景 | 处理方式 |
|---|---|
| 重复发起同一 pending 申请 | 唯一 active key 防重，返回已有 request ID 和 pending 状态 |
| 同一审批请求重复提交相同结果 | 返回幂等成功和已存在终态，不重复写关系、receipt 或 outbox |
| 两个处理人提交不同结果 | CAS 首个成功；后续返回 409 `already_handled` 和当前终态 |
| 关系唯一键冲突但申请仍 pending | 事务内复核既有关系；关系已存在则规范申请为 accepted，否则返回错误并回滚 |
| Kafka 不可用 | outbox 保持 pending，relay 指数退避重试；业务请求不因 Kafka 短暂不可用失败 |
| outbox 重复投递 | `event_id` 和 notifications 唯一键双重幂等，最多一条持久化通知 |
| 用户离线 | 通知已持久化；登录或 WS 重连后重拉 unread/list |
| 单个 WebSocket 连接发送失败 | 继续发送其他连接，记录指标；不回滚通知落库 |
| 列表请求乱序 | AbortController 或 request sequence 丢弃旧响应 |
| 标读失败 | 保留未读状态并后台重试；不把 count 乐观清零 |
| 群管理员角色在申请后被撤销 | 不再允许审批；其历史 notification 可见，但业务列表只返回当前有权范围或本人历史 receipt |
| 新管理员上任 | 角色授予流程为现存 pending 幂等补 receipt 和未读；无需等待新申请 |
| 群已解散、封禁或已满 | pending 转 invalidated，记录 reason，关闭管理员操作并通知申请者 |
| 用户已通过其他方式入群 | pending 转 accepted，记录实际来源，不重复加成员；同群 active 邀请全部失效 |
| 邀请超过 7 天 | 确认接口拒绝并 CAS 为 expired；cron 负责批量及时收敛 |
| 邀请人已退群 | 邀请转 invalidated；邀请人仅被降级为普通成员时，确认后转管理员审批 |
| 同群存在多条邀请 | 每条独立展示；接受一条后其余 active 邀请同事务失效 |
| 迁移发现重复数据 | 先生成审计报告并按迁移规则归并，未完成清理前不创建唯一索引 |

## 技术设计

### 1. 数据模型

#### 1.1 `friend_requests` / `group_requests`

两表新增：

| 字段 | 类型 | 说明 |
|---|---|---|
| `active_key` | varchar(160), nullable | 好友和群主动申请 pending 时非空；邀请审批不用该字段 |

`group_requests` 额外新增：

| 字段 | 类型 | 说明 |
|---|---|---|
| `source_type` | tinyint | `1=direct_apply,2=member_invite` |
| `source_invitation_id` | uint64 nullable | 普通成员邀请确认后生成审批时关联邀请 |
| `actual_join_source` | tinyint nullable | 最终实际入群来源 |
| `invalid_reason` | varchar(128) | 群异常等系统失效原因 |

索引：

```text
UNIQUE(active_key)
```

规则：
- 好友：`friend:<sender_uid>:<receiver_uid>`，保留方向语义。
- 群主动申请：`group:direct:<group_id>:<applicant_uid>`。
- 群邀请审批：`active_key=NULL`，以 `UNIQUE(source_invitation_id)` 保证一条 invitation 只生成一条审批。
- 对应申请进入任意终态时，`active_key` 与状态更新同事务置为 NULL。
- 三库普通唯一索引均允许多个 NULL；实现仍需在 SQLite/MySQL/PostgreSQL 集成测试验证。
- 不删除旧的 `receiver_read/sender_read/read_state` 列；迁移后新代码停止依赖和写入，待后续兼容窗口结束再清理。

`friends` 新增：

```text
UNIQUE(user_id, friend_uid)
```

`group_members` 已有 `(group_id,user_id)` 唯一约束，迁移脚本须验证真实环境索引存在。

`group_requests` 新增索引：

```text
UNIQUE(source_invitation_id)
INDEX(group_id, req_id, handle_result)
```

`source_invitation_id` 为 nullable，历史和主动申请均为 NULL；三库唯一索引允许多个 NULL。

#### 1.2 `group_invitations`

邀请与主动申请拆表，避免继续复用语义混乱的 `group_requests.req_id/join_source/inviter_user_id`：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uint64 | 稳定邀请 ID |
| `group_id` | uint64 | 目标群 |
| `inviter_uid` | varchar(64) | 邀请人 |
| `invitee_uid` | varchar(64) | 被邀请人 |
| `inviter_role_snapshot` | tinyint | 发出邀请时角色，仅用于展示/审计，不作为确认时授权依据 |
| `message` | varchar(255) | 邀请附言，可空 |
| `status` | tinyint | `0=pending,1=accepted,2=rejected,3=invalidated,4=expired` |
| `reject_reason` | varchar(255) | 被邀请人拒绝原因，可空 |
| `created_at` | timestamp | 创建时间 |
| `handled_at` | timestamp nullable | 接受、拒绝、失效或过期时间 |
| `expires_at` | timestamp | 创建后 7 天 |

索引：

```text
INDEX(invitee_uid, status, created_at)
INDEX(group_id, invitee_uid, status)
INDEX(status, expires_at)
```

不增加 `(group_id,invitee_uid)` active 唯一约束，因为产品规则要求每个邀请独立。接受一条后失效其他 active 邀请由同一数据库事务完成；确认操作使用 `WHERE id=? AND invitee_uid=? AND status=0 AND expires_at>now` 的 CAS。

#### 1.3 `social_request_receipts`

新增 Social 所有的个人业务回执表：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uint64 | GORM 主键 |
| `request_type` | varchar(16) | `friend` / `group` / `group_invite` |
| `request_id` | uint64 | 对应申请 ID |
| `receiver_id` | varchar(64) | 该回执所属用户 |
| `receipt_kind` | varchar(16) | `apply` / `result` / `invite` |
| `is_read` | tinyint | 个人是否已查看：`0=unread, 1=read` |
| `is_actionable` | tinyint | 当前用户是否仍可操作：`0=no, 1=yes` |
| `result` | tinyint | `0=pending, 1=accept, 2=reject, 3=invalidated, 4=expired` |
| `created_at` | timestamp | 创建时间 |
| `read_at` | timestamp nullable | 首次阅读时间 |
| `resolved_at` | timestamp nullable | 申请终态时间 |

约束与索引：

```text
UNIQUE(request_type, request_id, receiver_id, receipt_kind)
INDEX(receiver_id, is_read, request_type)
INDEX(receiver_id, is_actionable, request_type)
INDEX(request_type, request_id)
```

语义：
- 好友申请创建：给被申请人写 apply receipt。
- 好友处理完成：apply receipt resolved；给申请人写 result receipt。
- 群申请创建：给当时群主和管理员分别写 apply receipt。
- 群处理完成：所有 apply receipt resolved；给申请人写 result receipt。
- 群邀请创建：给被邀请人写 invite receipt。
- `is_read/read_at` 只记录个人是否阅读；`is_actionable/resolved_at` 表示是否仍可操作，两组字段不可互相替代。
- 管理员 A 审批后，所有管理员 receipt 都设 `is_actionable=0`；只有 A 自动 `is_read=1`，其他管理员保持原个人阅读状态。

#### 1.4 `social_notification_outbox`

新增通知事务性发件箱，不复用 `relation_outbox`，避免两种 topic、payload 和重试生命周期耦合：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | uint64 | 主键及稳定 event ID |
| `notify_type` | varchar(64) | friend/group apply/accept/reject、`group.invite`、`group.invalidated` |
| `receiver_id` | varchar(64) | 单接收者，群管理员在事务内展开多行 |
| `actor_id` | varchar(64) | 操作者 |
| `biz_id` | varchar(128) | `<request_type>:<request_id>:<phase>` |
| `group_id` | varchar(64) | 群通知可选 |
| `payload` | TEXT | request ID、result 等向前兼容 JSON |
| `status` | tinyint | `0=pending,1=sent,2=dead` |
| `attempts` | int | 投递次数 |
| `next_retry_at` | timestamp nullable | 下次重试时间 |
| `last_error` | varchar(512) | 最近错误摘要 |
| `created_at` | timestamp | 创建时间 |
| `sent_at` | timestamp nullable | 成功投递时间 |

约束：

```text
UNIQUE(notify_type, receiver_id, biz_id)
INDEX(status, next_retry_at)
```

relay：
- 复用 `apps/social/rpc/internal/relay` 的单实例锁和批量轮询模式，但实现独立 notification relay。
- 使用项目 MQ 命名规范的新 topic `social.request.notification.v1`；兼容期消费者可同时监听旧 `commonNotifyTransfer` 和新 topic。
- Kafka 消息增加 `eventId`、`requestType`、`requestId`、`result`，旧字段继续保留。
- 成功后标 sent；可重试错误指数退避；超过阈值标 dead 并记录告警，提供人工重放命令或管理脚本。

### 2. 授权与状态机

好友申请状态机：

```text
pending -> accepted
pending -> rejected
pending -> cancelled    仅申请人
pending -> ignored      仅服务内部归并
```

群申请状态机：

```text
pending -> accepted     当前群主/管理员
pending -> rejected     当前群主/管理员
pending -> invalidated  群异常，或系统判定不可继续处理
pending -> accepted     通过其他合法入口已经入群
```

群邀请状态机：

```text
pending -> accepted     仅被邀请人；随后直入群或转管理员审批
pending -> rejected     仅被邀请人，原因可选
pending -> expired      超过 7 天
pending -> invalidated  接受同群其他邀请、邀请人退群、群异常或用户已入群
```

要求：
- API 从 JWT 取 actor，不接受 body 覆盖。
- RPC request 显式携带 `actor_uid` 并执行同等校验。
- 审批 CAS 和关系写入位于同一事务。
- `FindOne` 可以服务详情读取，但不能作为审批授权本身。
- 数据层返回 `not_found`、`forbidden`、`already_handled`、`conflict` 等稳定错误，API 映射为 404/403/409。

### 3. API 与 RPC 契约

#### 3.1 好友申请

| 方法 | 路径 | 变更 |
|---|---|---|
| POST | `/v1/social/friend/putIn` | 响应增加 `request_id/status/already_pending/already_friend`，不再只有空成功 |
| PUT | `/v1/social/friend/putIn` | 仅接收 accept/reject；透传明确错误；`handle_msg` 补回 `.api` |
| GET | `/v1/social/friend/putIns` | `status` 可选、增加 `page/size`；返回 `peer_uid/handle_msg/handled_at/read_state` |
| PUT | `/v1/social/friend/putIn/read` | 支持 `request_ids[]` 或当前筛选批量标 receipt；响应最新 unread count |
| GET | `/v1/social/friend/putIn/messageCount` | 改为 receipt count，返回总数及 apply/result 分类 |

#### 3.2 群申请

| 方法 | 路径 | 变更 |
|---|---|---|
| POST | `/v1/social/group/putIn` | 移除公开 `req_id/inviter_uid/join_source` 控制；响应 `request_id/group_id/is_pass/status` |
| PUT | `/v1/social/group/putIn` | 请求只需 `group_req_id/handle_result/handle_msg?`；group ID 取申请记录 |
| GET | `/v1/social/group/putInsByUid` | 移除公开 `ids`；增加 `class/status/page/size` 和明确 `peer/applicant` 字段 |
| PUT | `/v1/social/group/putIns/read` | 按当前用户 receipt 标读，支持 request IDs，响应最新 unread count |
| GET | `/v1/social/group/putIn/messageCount` | 新增个人业务未读数，包含管理员 apply 和申请人 result |

#### 3.3 群邀请

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/v1/social/group/invitations` | 任意群成员发起邀请；body 仅含 `group_id/invitee_uid/message?` |
| GET | `/v1/social/group/invitations` | 被邀请人分页查询“收到的邀请”，支持 status/page/size |
| PUT | `/v1/social/group/invitations/:id` | 被邀请人接受或拒绝；拒绝原因可选 |
| PUT | `/v1/social/group/invitations/read` | 按本次展示的 invitation IDs 标记个人回执并返回最新 count |

接受响应明确区分：

```json
{
  "invitation_id": 123,
  "status": "accepted",
  "join_state": "joined|pending_approval",
  "group_request_id": 456
}
```

邀请人不提供“我发出的邀请”结果通知入口；记录仍保留用于审计和被邀请人历史分页。

#### 3.4 通知标读

扩展 IM 标读 API，保留已有按 notification ID 的能力，并增加受 JWT 限制的业务筛选：

```json
{
  "notify_types": ["friend.apply"],
  "biz_ids": ["friend:123:apply"]
}
```

不得只按 notify type 无界清除全部历史；页面应传本次已成功展示的 biz IDs。

#### 3.5 兼容策略
- Proto 只追加字段和方法，不复用字段号；可选字段使用 `optional`。
- HTTP 新字段向后兼容；前端迁移完成后再移除危险输入字段。
- 上线顺序为“新 schema -> 可读新旧字段的后端 -> 新前端 -> 停止旧写法”；不添加长期无依据兼容分支。

### 4. 前端状态设计

在 `web/src/lib/im-store.ts` 增加：

```ts
friendRequestsVersion: number;
invalidateFriendRequests(): void;
groupRequestsVersion: number;
invalidateGroupRequests(): void;
groupsVersion: number;
invalidateGroups(): void;
friendRequestUnread: { total: number; apply: number; result: number };
groupRequestUnread: { total: number; apply: number; result: number; invite: number };
```

通知分发规则：

| notifyType | 失效动作 |
|---|---|
| `friend.apply` | invalidate friend requests + refresh friend unread |
| `friend.accept` | invalidate friend requests + refresh friend unread + invalidate friends |
| `friend.reject` | invalidate friend requests + refresh friend unread |
| `group.apply` | invalidate group requests + refresh group unread |
| `group.accept` | invalidate group requests + refresh group unread + invalidate groups/conversations |
| `group.reject` | invalidate group requests + refresh group unread |
| `group.invalidated` | invalidate group requests + refresh group unread |
| `group.invite` | invalidate group requests/invitations + refresh group unread，定位“收到的邀请” |

实现要求：
- `chat-store` 只分发失效和 Toast，不维护申请列表副本。
- `FriendRequestList`、`GroupList` 订阅各自版本，使用单一 `refresh()`；正在显示时收到通知立即重拉。
- `GroupList` 的申请区域至少提供“我发起”“待我审批”“收到的邀请”三个 Tab；管理员已处理但个人未查看的记录继续显示终态红点。
- `ContactList` 和 `IMLayout` 不再各自拉全量群申请计算 badge，统一调用 store 的 count action。
- WS connect/reconnect 后统一刷新 notification、好友申请、群申请未读和当前可见列表，补偿离线窗口。
- 通知导航目标携带 `{tab, requestId, groupId}`；刷新成功后定位目标。
- 好友 sent 映射使用 `peer_uid/req_uid`，received 使用 `peer_uid/user_id`。
- 处理成功后仍可本地显示 loading，但最终状态必须由一次 REST refresh 确认。
- 所有用户可见文案进入 `web/src/i18n/locales/zh-CN.json` 与 `en.json`。

### 5. 关系和会话同步

好友：
- 同意事务写 `friend.added` relation outbox。
- relation consumer 更新缓存后向双方发送关系失效事件；前端收到后 `invalidateFriends()`。
- `friend.accept` 通知也触发刷新，relation event 作为独立可靠兜底。

群：
- 入群事务写 `group.member.added` relation outbox。
- consumer 调 IM RPC `EnsureGroupConversation(userId, groupId)`，操作必须幂等。
- Ensure 失败返回可重试错误，不标 relation event 完成；不再由 social API 启动 800ms 后台 goroutine。
- 成功后向加入者发送 `relation.changed(group.member.added)`；向群主/管理员是否推成员刷新按当前页面需要决定，但至少审批操作者本地立即失效群详情。

### 6. 多端 WebSocket

`apps/im/ws/internal/handler/push/notify.go` 改为向 `GetConn(receiver)` 返回的全部连接发送：
- 每个连接独立发送。
- 汇总日志和指标，不因一个连接失败跳过其他连接。
- 接收者无连接时正常返回；通知已持久化。
- 多节点路由仍记录为后续事项。

### 7. 数据迁移

迁移必须是显式、可审计步骤，不直接依赖 AutoMigrate 在脏数据上创建唯一索引。

#### 7.1 预检查
1. 统计 `friend_requests` 同方向重复 pending。
2. 统计 `group_requests` 同群同申请人重复 pending。
3. 识别旧 `group_requests` 中由 `join_source/inviter_user_id` 表达的邀请数据，生成迁移分类报告。
4. 统计 `friends(user_id,friend_uid)` 重复行和单向好友关系。
5. 验证 `group_members(group_id,user_id)` 唯一索引。
6. 输出待归并记录 ID、最终保留规则和受影响用户数，不静默删除。

#### 7.2 清理规则
- 好友重复 pending：保留最新正常状态行；旧 pending 转 ignored，并保留历史，不物理删除。
- 群重复 pending：保留最新行；旧 pending 转 rejected/ignored 的内部终态并记录迁移原因。
- 旧邀请数据：能明确还原 inviter/invitee/status/time 的迁入 `group_invitations`；无法可靠区分的保留在原历史中并输出人工审计，不猜测邀请语义。
- 重复好友关系：每个有向关系保留最早 `created_at` 或最小 ID，合并非空 remark/tags 后删除重复；单向关系只报告，不自动补双向，避免掩盖业务异常。
- 已存在好友/群成员但申请仍 pending：申请规范为 accepted，并创建 result receipt；是否补发历史通知默认否，防止上线噪声。

#### 7.3 回填
1. 新增 nullable `active_key`，初始全部 NULL。
2. 为清理后仍 pending 的申请回填 active key。
3. 按旧 `receiver_read/sender_read` 回填好友 apply/result receipt。
4. 群旧数据无法还原每个管理员是否已读：为当前群主/管理员创建 receipt；旧 `receiver_read=1` 时均标 read，否则均标 unread，并在迁移说明中记录此近似规则。
5. 新建 `group_invitations` 并迁移可可靠识别的旧邀请记录；历史 pending 邀请按原创建时间计算剩余有效期，已超过 7 天迁为 expired。
6. 创建 notification outbox 和 receipt 索引。
7. 清理重复好友关系后创建唯一索引。
8. 执行三库 schema 和数据校验。

#### 7.4 发布顺序
1. 暂停或限流好友/群申请写入口。
2. 备份并执行审计、清理、加列、回填、建索引。
3. 部署支持 receipt、CAS 和 outbox 的 social RPC/API。
4. 部署新 topic 消费者和 IM 通知筛选标读能力。
5. 部署前端版本失效与新 API。
6. 恢复写入口，观察冲突率、outbox backlog、通知延迟和错误率。

#### 7.5 回滚边界
- 新表和新增列不在回滚时删除，旧版本可忽略。
- 创建唯一索引后回滚旧写逻辑会重新产生冲突，因此后端不能单独回滚到先查后插版本；需同时关闭写入口或保持兼容写服务。
- Kafka 新旧 topic 在观察期双消费但不能双生产；以 event ID/通知唯一键兜底。

## 当前实施状态（2026-07-17）

### 分支与提交

- 当前分支：`feat-social-request-reliability`。
- `33bcd65 docs(social): specify request reliability fixes`：创建本 Spec。
- `b1fc5ad feat(deploy): add social migration audit`：步骤 1，只读迁移审计工具。
- `3570bd5 feat(deploy): migrate social request schema`：步骤 2，显式 schema/data/index 分阶段迁移。
- `1b74e21 fix(social): harden friend request state machine`：步骤 3，好友申请授权、并发和 CAS 状态机。
- `36d0c4e feat(social): add group request state machines`：步骤 4 checkpoint。
- `940aadb fix(social): complete group request reliability`：完成步骤 4 复审修复。
- `aa05ad1 feat(social): add personal request receipts`：完成步骤 5 个人回执与 unread/read 切换。
- `7268dcf feat(social): add reliable notification outbox`：完成步骤 6 notification outbox、relay、幂等消费和监控。

### 已完成

- 步骤 1 已完成：审计重复好友申请、重复/单向好友关系、旧群邀请和群成员唯一索引；SQLite 测试与 vet 通过。
- 步骤 2 已完成：增加申请 active key、独立群邀请、个人 receipt、notification outbox 和好友有向唯一索引；迁移支持失败续跑、旧邀请分类和完整清理报告。
- 步骤 2 的 SQLite legacy upgrade、幂等、约束和故障恢复测试通过；MySQL/PostgreSQL 测试入口存在，但当前未配置 DSN，因此尚未实跑。
- 步骤 3 已完成：好友申请只允许目标用户审批，结果仅允许 accept/reject；发起使用 active key 并发裁决；审批使用事务 CAS；双向好友、反向 pending 收口和 relation outbox 原子提交。
- 步骤 3 覆盖 self、用户不存在/停用、重复与并发申请、越权审批、同结果幂等、不同结果冲突、accept/reject 并发、字段方向和 outbox 回滚；Social 全量、race 和目标 vet 通过。

### 步骤 4 Checkpoint

`36d0c4e` 已实现并通过当前测试的内容：

- 主动群申请绑定 JWT actor，RPC 校验 actor；使用 `active_key` 防止并发重复，不再物理删除历史，并实现 60 秒 terminal 冷却。
- 免验证群直接入群时，accepted 申请、群成员和 relation outbox 位于同一事务。
- 群申请审批使用 CAS；同结果幂等、不同结果冲突；accept 收口同群同用户其他 pending，reject 只处理当前申请。
- 新增独立群邀请 create/list/handle RPC 和 HTTP 入口；旧批量邀请入口改为创建独立 invitation，不再无确认直接拉人。
- 只有 invitee 可处理邀请；邀请接受时重新读取邀请人当前角色；管理角色邀请直入群，普通成员邀请生成带 `source_invitation_id` 的独立管理员审批。
- 接受一条邀请会使同群同 invitee 的其他 active 邀请失效；已有主动 pending 不会阻止普通成员邀请生成第二条审批。
- 已新增真实文件 SQLite 测试，覆盖群申请重复/并发、邀请授权与角色变化、多邀请并发、审批 CAS、成员幂等和 relation outbox 故障回滚。
- checkpoint 提交前 `go test ./apps/social/... -count=1` 和 `go test -race ./apps/social/... -count=1` 通过。

### 步骤 4 复审修复

- 邀请确认和群审批锁定授权成员行；设/撤管理员、转让群主、退群和踢人改为事务内按一致顺序锁定群与成员行，SQLite 跳过 `FOR UPDATE`。
- 单条邀请、旧批量邀请和邀请确认均校验当前用户状态，停用用户不能创建或处理邀请。
- accepted invitation 重试按当前成员关系和来源审批终态返回 `joined`、`pending_approval` 或 `approval_rejected`，并保留稳定的来源审批 ID。
- HTTP 已移除普通申请、审批和按用户列表中的危险身份/来源字段；审批 ID 使用 `uint64`，RPC 保留旧 `int32 groupReqId` 并拒绝新旧 ID 冲突。
- 使用 `goctl 1.8.2`、`protoc 5.28.3`、`protoc-gen-go 1.35.2` 核对并重新生成 proto；未生成额外 `social.yaml`，GET form 标签和零值 count 契约保持不变。
- 新增禁用用户、审批终态重试、新旧/大 ID、角色撤销并发和空踢人 outbox 测试；Social 全量、race、目标 vet 和 diff check 通过。MySQL/PostgreSQL 因未配置 DSN 仍未实跑。

### 恢复入口

步骤 10 已完成。下一会话从步骤 11 开始：好友申请列表按新契约分页与 receipt 已读时序。

### 步骤 5 实施结果

- 好友、群申请和群邀请创建/终态事务已原子写入或关闭个人 receipt；并行申请和其他邀请的自动收口同步关闭 actionability。
- 好友和群申请 unread/read 已切换到 receipt 分类计数；新增群申请/邀请 count 与邀请按展示 ID 标读接口，一人标读不影响其他管理员。
- 好友、群申请和邀请列表优先映射个人 receipt；仅无 receipt 的旧数据回退旧 read 字段或状态推导，新代码不再写旧 read 列。
- 好友申请隐藏只关闭当前用户 receipt，不再删除共享历史或复用取消语义。
- 新增邀请 receipt 枚举修复迁移版本，升级已执行旧迁移的数据库时可幂等修正 expired/invalidated。
- 固定工具重新生成 API/RPC；Social 全量、receipt/migration 测试、race、目标 vet 和 diff check 通过。MySQL/PostgreSQL DSN 未配置，仍只实跑 SQLite。

### 步骤 6 实施结果

- 好友、群申请、审批、自动收口和邀请通知均与业务状态、receipt、关系 outbox 在同一 Social 事务写入 `social_notification_outbox`。
- 独立 relay 投递 `social.request.notification.v1`，稳定使用 outbox ID 作为 event ID，支持指数退避、dead、owner-safe Redis 锁及续租丢锁停止。
- task/mq 同时消费旧/新 topic；通知持久化失败触发 Kafka 重试，持久化成功后的 WebSocket 失败只记录错误，不再无效重投。
- payload 固定包含 requestType/requestId/groupId/result/content，并校验 notify type、biz phase 和结果枚举一致性；坏数据直接 dead。
- 增加 pending/dead/delivery latency/failure 指标和显式 ID dead replay 命令；配置支持显式覆盖并兼容旧配置派生。
- Social/task 目标测试、relay 状态机和事务回滚测试、race、目标 vet 与 diff check 通过。Kafka E2E 及 MySQL/PostgreSQL 实跑留待步骤 17。

### WebSocket 在线定位基础实施结果

- `user:online:<uid>` 写入 WS node ID，并用 connection owner token companion key 实现原子续租和清理；跨节点及同节点旧连接均不能删除新 locator。
- claim/refresh/delete 使用按 UID 分片生命周期锁；Redis 故障不会阻塞其他用户，初始 claim 失败会随当前连接重试。
- 客户端入站消息和 pong control frame 更新 liveness，超时统一走幂等连接清理；连接写串行化，`chat.ping` pong 回原连接并保留 method。
- FriendsOnline 使用共享配置 Redis，不再硬编码 localhost；Redis 故障时返回完整离线 map。
- 外部 WS client 测试改为 `HICHAT_WS_TEST_HOST` 显式启用；相关 package、race、目标 vet 和 diff check 通过。
- 本项仅建立 locator 基础，不宣称已完成跨节点消息传输；当前节点多连接广播仍按步骤 9 实施。

### 步骤 8 实施结果

- 群创建者成员关系与 `group.member.added` relation outbox 在同一 Social 事务提交；原有申请、审批、邀请和 token 入群继续复用统一成员/outbox helper。
- 新增内部 `EnsureGroupConversation(userId, groupId, relationVersion)` RPC；全局群会话和用户绑定使用 Mongo 原子 upsert，并由 `conversationId`/`userId` 唯一索引保证并发幂等。
- added/removed 使用 relation outbox version 原子收敛 `isShow/removedAt`；普通会话更新不覆盖 tombstone/version，旧事件不能恢复较新的移除状态。
- relation consumer 强制单 consumer/processor，并在 durable IM/Mongo 状态成功前内部指数退避阻塞后续 offset；进程 shutdown gate 防止预取高 offset 越过失败事件。
- malformed/非法 ID/version/timestamp 作为 poison 记录后丢弃；cache 和 WS push 在 durable 状态成功后保持 best-effort。
- 删除群创建和 token 入群 API 的 800ms IM goroutine；新增群创建 outbox、consumer RPC/poison 测试。
- IM、task consumer、Social 全量目标测试、race、目标 vet 和 diff check 通过；Mongo 实例并发/E2E 覆盖留待步骤 17。

### 步骤 9 实施结果

- WS 本地连接池从单连接改为 `uid -> connection set`；同一用户多个标签页/设备在当前节点共存，不再互踢。
- presence 提升为 UID 级共享 lease：首连接 claim、附加连接复用、末连接 owner-safe delete；失去 locator owner 的旧节点停止续租且不重新抢占。
- `GetConn` 返回去重后的全部目标连接，`GetUsers(nil)` 返回唯一在线 UID；`NewMessageTest` 直接使用不可变 `conn.Uid`。
- `Send` 对全部连接逐一尝试并聚合错误；每次写有 5 秒 deadline，失败连接在完成本轮广播后异步清理，不阻断健康连接。
- notify、relation、trend、call 和 chat push 均广播到当前节点所有目标连接；chat 单聊/群聊局部失败有明确日志。
- 新增双客户端广播、重复 UID 去重、失败连接隔离及首末连接 presence 测试；WS 全量与 race 通过。

### 步骤 10 实施结果

- 前端 store 新增 friend requests、group requests/invitations、groups 和 friends version；通知与 relation 事件按业务类型统一 invalidation。
- 好友申请、群申请、群列表、会话、好友、业务未读和公共通知未读均使用 generation/token guard，旧请求和旧账号响应不能覆盖新状态。
- 好友/群业务未读改为分类 count API 的 centralized actions，不再从分页列表推导或由多个组件直接写入。
- WS 初连/重连执行 quiet REST recovery；notify 不再用本地 `+1` 作为业务未读真相，accept 事件同步失效好友、群和会话资源。
- FriendRequestList 和 GroupList 在 mutation 成功时先失效旧 GET generation，再 bump version 并用服务端状态收敛。
- logout/login 同步清空 chat auth-scoped state并失效旧请求；异步群成员、用户资料和通知中心列表写入均校验当前 token/generation。
- 新增 latest-request 单元测试；`bun test` 与 lint 通过，typecheck 仅保留两个既有 CSS 自定义属性类型错误。

步骤 4 暂不提前实现：

- 个人 receipt 和 read/unread API 属于步骤 5。
- `social_notification_outbox`、relay 和可靠通知属于步骤 6；当前群通知仍是提交后 best-effort `CommonNotify`。
- `group.member.added` 可靠创建 IM 会话属于步骤 8；步骤 4 已移除群申请/审批 API 的 800ms IM goroutine。
- 邀请过期 cron 和管理员角色变更后的 receipt 补发/收口属于步骤 16；确认接口已经阻止过期邀请。

### 工作区注意事项

- 当前数据库未上线且仅含 mock 数据，迁移无需为真实生产历史扩大兼容路径，但仍保留显式、安全和可重跑迁移。
- 不得提交或修改用户已有的 sample 配置、`apps/im/api/im.go`、`docker-compose.dependencies.yaml`、`hichat2.sh`、`docs/specs/auth-pages-redesign.md`、`test.md`。
- `pkg/2fa/totp/totp_qr.png` 是全仓测试生成的工作区副作用，不属于本功能，禁止暂存。
- MySQL/PostgreSQL 群状态机测试入口为 `SOCIAL_GROUP_TEST_MYSQL_DSN` 和 `SOCIAL_GROUP_TEST_POSTGRES_DSN`；当前环境未配置，实际只运行了 SQLite。

## 实现步骤（每步可独立 commit）
1. [x] 增加迁移审计工具或版本化迁移：识别并清理重复申请、重复好友关系，输出报告。
2. [x] 增加 `active_key`、`group_invitations`、`social_request_receipts`、`social_notification_outbox` 和好友唯一索引，完成迁移实现与 SQLite 测试；MySQL/PostgreSQL 实跑留待步骤 17。
3. [x] 修复好友发起/审批授权、self-check、目标校验、结果枚举和 CAS 状态机。
4. [x] 修复群发起身份伪造、列表越权、结果枚举、事务提交和 CAS 状态机；完成 checkpoint 复审修复与 SQLite 并发测试。
5. [x] 好友和群申请事务接入个人 receipt，切换 unread/read API，保留旧字段兼容读取窗口。
6. [x] 实现 notification outbox relay、新 Kafka topic、幂等消费、backoff/dead 状态和监控指标。
7. [x] 完善好友/群申请 RPC 与 HTTP 字段、分页、稳定错误码和通知业务筛选标读。
8. [x] 将 `group.member.added` 会话创建迁移到可靠消费者，删除 API best-effort goroutine；补关系新增事件。
9. [x] 修复 WS 当前节点内多连接广播，并明确 WS 失败为 best-effort。
10. [x] 前端 store 增加好友/群申请/邀请及群列表版本和统一 unread actions，所有相关通知触发重拉。
11. [ ] FriendRequestList 改为分页单请求、顺序标读、字段正确映射和通知定位。
12. [ ] GroupList/GroupProfileCard 接入三个群申请 Tab、分页、个人回执、邀请确认、直接入群结果、群/会话刷新和通知定位。
13. [ ] 联动业务 receipt 与通知中心标读；补中英文文案和可重试错误态。
14. [ ] 统一 JWT 安全取值、修正好友 remark/tags 方向，重新生成 `.api/.proto/model` 代码并增加生成一致性检查。
15. [ ] 公共通知消费者改走幂等 IM RPC，移除 task 对 IM notification model 的直接写入。
16. [ ] 管理员角色授予/撤销接入 pending receipt 补发/收口，增加邀请过期 cron 和确认接口过期 CAS。
17. [ ] 执行后端三库测试、Kafka/WS 集成测试、前端类型/组件测试及双账号多端 E2E。

## 测试计划

### 后端授权与输入
- [ ] 好友申请人不能审批自己发出的申请；被申请人可以审批。
- [ ] 好友自己申请自己、目标不存在、目标停用均不落库。
- [ ] 普通群申请忽略或拒绝伪造的 `req_id/join_source/inviter_uid`。
- [ ] 群申请列表始终绑定 JWT，传其他 UID 无法越权读取。
- [ ] 好友和群 `handle_result` 非 1/2 全部拒绝。
- [ ] 已撤销管理员不能审批原群申请。
- [ ] JWT claim 缺失、空值或类型错误返回 401，不发生 panic。
- [ ] 所有当前群成员均可邀请；非群成员不能邀请。
- [ ] 只有 invitee 可以接受或拒绝自己的邀请。
- [ ] 管理角色邀请确认后直入群，普通成员邀请确认后进入管理员审批。
- [ ] 邀请人降级后按确认时角色转管理员审批；邀请人退群后邀请失效。

### 并发与幂等
- [ ] 20 个并发好友申请只创建一个 active pending。
- [ ] 20 个并发群申请只创建一个 active pending。
- [ ] 同一用户同群 60 秒内被限流；拒绝 60 秒后可创建新申请并保留历史。
- [ ] 已有主动 pending 时接受普通成员邀请，会生成 source_invitation_id 不同的第二条审批。
- [ ] 通过任一并行审批后，其余同群同用户 pending 自动 accepted/resolved；拒绝其中一条不影响其他审批。
- [ ] 两个并发 accept 只产生一套好友关系或一个群成员。
- [ ] accept/reject 并发只有一个终态，关系与终态一致。
- [ ] 相同终态重试幂等成功，不重复 receipt/outbox/notification。
- [ ] 不同终态重试返回 409，并返回当前终态。
- [ ] 好友、群成员和 receipt 唯一约束在 SQLite/MySQL/PostgreSQL 下生效。
- [ ] 同群同 invitee 的多条邀请可独立创建；并发接受只允许一条成功，其余转 invalidated。

### 事务与 Outbox
- [ ] 注入关系插入失败时，申请状态、receipt 和 outbox 全部回滚。
- [ ] 注入 Commit 失败时 API 不返回成功，不产生可投递 outbox。
- [ ] Kafka 停止时业务事务成功且 outbox 保持 pending；恢复后自动投递。
- [ ] relay 重启、重复投递只生成一条 notification。
- [ ] 超过重试阈值进入 dead 并可人工重放。
- [ ] group.apply 为每个当前群主/管理员各生成一条 outbox 和 notification。
- [ ] group.invite 只通知被邀请人；接受和拒绝均不向邀请人生成通知。
- [ ] task 通过 IM RPC 幂等落通知，不直接写 IM model；可重试 RPC 错误触发 Kafka 重投。

### 已读与列表
- [ ] 群管理员 A 标读不改变管理员 B 的 receipt。
- [ ] A 审批后 B 的 `is_actionable=0` 且未读保持，B 查看后才清个人终态红点。
- [ ] 新管理员获得全部现存 pending 的个人未读，撤销管理员不再有可操作记录。
- [ ] 好友和群申请结果为申请人生成 result unread。
- [ ] 进入 received/sent Tab 只标记本次展示的对应 receipt 和 notification。
- [ ] 标读响应 count 与重新查询一致，不发生红点回弹。
- [ ] 分页状态过滤正确，数据库错误不会被转换为空列表。
- [ ] 好友返回 `peer_uid/handle_msg/handled_at`，刷新前后内容一致。
- [ ] 审批人的 remark/tags 只写入审批人视角，申请预设备注只写入申请人视角。
- [ ] 群异常转 invalidated 并向申请者展示原因，不显示具体处理管理员。
- [ ] 其他入口入群会把原 pending 转 accepted 并记录 actual join source。
- [ ] 邀请 7 天后由 cron 收敛为 expired；cron 未运行时确认接口仍拒绝过期邀请。

### 关系与会话
- [ ] 好友 accept 后双方好友列表刷新且数据库无重复关系。
- [ ] 群 accept/免验证入群后群列表和群会话出现。
- [ ] IM 暂时不可用时 `group.member.added` 消费重试，恢复后补建会话。
- [ ] relation 和 notify 任意到达顺序下，前端最终状态一致。

### 前端
- [ ] 页面已打开时收到好友、群申请和邀请通知，列表和 unread 自动刷新。
- [ ] “收到的邀请”Tab 分页展示每条独立邀请，支持接受和带可选原因拒绝。
- [ ] 普通成员邀请被确认后，管理员审批卡片同时显示申请人和邀请人。
- [ ] `friend.accept` 刷新联系人；`group.accept` 刷新群和会话。
- [ ] sent/received 均显示正确对方 UID、昵称、附言和处理时间。
- [ ] 快速连续通知和慢请求不会让旧响应覆盖新状态。
- [ ] 进入页面后业务 badge 和对应通知中心 badge 联动清除，无关通知保留。
- [ ] 免验证群显示“已加入”而不是“申请已发送”。
- [ ] Toast/通知点击后进入正确 Tab 并定位 request ID。
- [ ] API 失败显示重试状态并保留最后成功数据，不伪装为空。
- [ ] 中英文环境的好友、群申请和邀请通知及错误提示均无硬编码中文。

### 多端与离线 E2E
- [ ] 同一用户两个浏览器连接在当前 WS 节点都收到通知。
- [ ] 接收者离线时发起/审批，上线后 unread、列表和通知中心完整恢复。
- [ ] WS 推送失败但通知落库成功时，重连后的 REST 刷新可恢复。
- [ ] 双账号完整走通好友 apply/accept/reject 和群 apply/accept/reject。
- [ ] 三账号走通普通成员邀请 -> 被邀请人确认 -> 管理员审批，以及管理员邀请 -> 确认后直入群。

### 验证命令
- [ ] `go test ./apps/social/... ./apps/im/... ./apps/task/... -count=1`
- [ ] `go test ./... -count=1`
- [ ] `cd web && bunx tsc --noEmit`
- [ ] `cd web && bun run lint`
- [ ] 执行仓库既有前端测试命令（若 `package.json` 已配置）

## 可观测性
- `social_notification_outbox_pending_total`
- `social_notification_outbox_dead_total`
- `social_notification_outbox_delivery_latency_seconds`
- `social_request_cas_conflict_total{type,result}`
- `social_request_duplicate_active_total{type}`
- `social_request_unread_count_error_total{type}`
- `ws_notify_connections_total` 和 `ws_notify_send_fail_total`
- relation consumer 的 `EnsureGroupConversation` 重试和失败数

日志必须带：`request_type`、`request_id`、`actor_uid`、`receiver_uid`、`event_id`、`notify_type`；不得记录 JWT、手机号或完整私密申请附言。

## 参考的现有模式
- `apps/social/rpc/internal/logic/friendputinlogic.go` — 好友申请创建基线。
- `apps/social/rpc/internal/logic/friendputinhandlelogic.go` — 好友审批、关系事务基线。
- `apps/social/rpc/internal/logic/groupputinlogic.go` — 群申请与直接入群基线。
- `apps/social/rpc/internal/logic/groupputinhandlelogic.go` — 群审批与 relation outbox 基线。
- `apps/social/rpc/internal/relay/relay.go` — outbox relay、Redis 单实例锁和批量投递模式。
- `apps/task/mq/internal/handler/msg_transfer/common_notify_transfer.go` — 通知幂等落库和 WS 推送基线。
- `apps/task/mq/internal/handler/msg_transfer/relation_change_transfer.go` — 关系缓存和群会话恢复基线。
- `apps/im/ws/internal/handler/push/notify.go` — 通知 WS 下行基线。
- `web/src/lib/chat-store.ts` — WebSocket 通知分发基线。
- `web/src/lib/im-store.ts` — 请求版本、关系列表失效状态归属。
- `web/src/components/im/FriendRequestList.tsx` — 好友申请列表、审批和已读基线。
- `web/src/components/im/GroupList.tsx` — 群申请列表、审批和群详情基线。
- `web/src/components/im/NotificationCenter.tsx` — 公共通知列表与标读基线。

## 风险与发布检查
- 唯一索引创建前必须完成重复数据审计；禁止直接在生产脏数据上 AutoMigrate。
- receipt 回填会改变群管理员 badge 初始值，发布公告应说明旧群申请已读状态只能近似恢复。
- 新旧前端并存时，旧前端仍可能显示两套红点不一致；后端 API 必须保持新增字段向后兼容。
- outbox backlog 需要容量评估和告警，避免 Kafka 长时间故障后恢复产生突发投递。
- `group.member.added` 消费增加 IM RPC 后，要避免 consumer 中长事务；单事件一次幂等 RPC，失败交给 Kafka 重试。
- 多节点 WS 路由不在本期，若生产已多节点部署，实时在线送达仍可能依赖重连 REST 补偿。

## 待定事项
- 好友申请是否给申请人提供主动取消 pending 的 UI；群申请已确认本期不支持撤回。
- 通知定位目标在列表分页之外时，是自动逐页定位还是直接按 request ID 拉详情；建议新增详情接口，实施时确定。
- 多 WS 节点的在线用户路由方案需另立 Spec。

## MVP 范围
本 Spec 所列 P01-P40、Schema 迁移、好友与群申请及邀请、通知 Outbox、个人回执、实时 REST 重拉、关系/会话刷新、当前节点多端推送、已读联动和全部测试均纳入 MVP。仅“非目标”和“待定事项”中明确列出的扩展能力不纳入。
