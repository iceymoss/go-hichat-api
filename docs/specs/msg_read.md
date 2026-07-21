# 消息已读回执（Read Receipt）

> 微信式已读回执：发送方能看到自己发出的消息是否已被接收方阅读。
> 相关改动跨前后端多个模块，本文是功能规格 + 实现参考，任何改动都应保持现有消息收发链路不受影响。

---

## 一、需求概述

实现类似微信的消息已读回执：
- 私聊：发送方看到「已读」状态。
- 群聊：发送方看到「X/N 人已读」状态。

## 二、三层开关

已读回执是否生效，取决于三层开关联合判定：

| 层级   | 开关归属 | 作用范围     | 控制者   |
| ------ | -------- | ------------ | -------- |
| 系统层 | 系统设置 | 全局总开关   | 管理员   |
| 用户层 | 个人设置 | 当前用户自己 | 用户本人 |
| 显示层 | UI 渲染  | 会话详情页   | 前端联合判定 |

**生效条件**：系统开关 ON 且 发送方开启 且 接收方开启 → 显示已读差异。任一层关闭，UI 均显示「已发送」状态（灰色 ✓✓），不泄漏已读信息。

## 三、功能流程

### 1. 发送方（User A）
- 发出消息后 UI 显示灰色 ✓✓（未读状态）。
- 收到已读回执 → 变为蓝色 ✓✓ + 「已读」（私聊）或「X/N 人已读」（群聊）。

### 2. 接收方（User B）
- 进入聊天窗口 → 自动对未读消息发送 `chat.markChat` WS 方法，携带 `msgIds` 列表。

### 3. 后端
- `im/ws` 收到 `chat.markChat` → 投递到 Kafka `MsgMarkRead` 主题。
- `task/mq` 的 `msg_read_transfer` 消费：
  1. 无条件更新 `chat_logs.readRecords`（用于未读数、服务端状态正确性）。
  2. 读取系统开关 + 发送方 + 接收方设置，三者均开启 → 下游推送回执消息。
  3. 下游消息 `MsgType = ContentMakeRead (6)`，通过 `FrameNoAck` 推给原发送方。

### 4. 前端
- WS 推送消息 `msg.mType === 6` → 识别为已读回执。
- 解析 `msg.readRecords`（key=MongoDB msgId，value=base64(bitmap字节)）。
- 将对应消息标记 `isRead = true`、填充 `readCount`（bitmap 计数）。
- **乱序缓存**：回执可能比消息先到，需要按 msgId 暂存，待消息落本地时合并。

## 四、UI 展示规则

| 消息状态     | 图标颜色        | 文案                       |
| ------------ | --------------- | -------------------------- |
| 发送中       | —               | `...`                      |
| 发送失败     | 红色 `!`        | 「发送失败」（可点击重试） |
| 已发送未读   | 灰色 ✓✓ #C8CCD0 | 时间戳                     |
| 已读（私聊） | 蓝色 ✓✓ #3390EC | 「已读 · 时间戳」          |
| 已读（群聊） | 蓝色 ✓✓ #3390EC | 「X/N 人已读」             |

X = bitmap 中已读人数（不含发送者自己）；N = 群成员总数 - 1。

## 五、技术要点

- 私聊、群聊都支持；群聊用 `pkg/bitmap` 记录每个成员。
- 已读回执走 `FrameNoAck` 避免占用 RigorAck 队列。
- DB 写不受三层开关影响（保证未读计数正确）。
- 回执乱序需前端缓存。
- 回执本身不产生新 `ChatLog`，不占用聊天记录页面。

---

## 六、实现清单

### 批次 A — 常量、表、用户设置字段

| 文件                                                             | 改动                                                                |
| ---------------------------------------------------------------- | ------------------------------------------------------------------- |
| `pkg/constants/im.go`                                            | 显式 `ContentMakeRead MType = 6`                                    |
| `pkg/db/objects/system.go` *(新)*                                | 定义 `SystemSetting` 结构体                                         |
| `deploy/sql_init.go`                                             | 在 `tables` 列表加 `&objects.SystemSetting{}`                       |
| `apps/user/models/systemconfigmodel.go` *(新)*                   | 提供 `GetBool(key) (bool, error)` 等读取接口                        |
| `web/src/lib/settings-store.ts`                                  | 加 `readReceiptEnabled: boolean`（默认 `true`）并进入 saveToBackend |

后端用户设置 JSON 字段：由 `/api/user/settings` 透明转发，无需改 Go API 层。

### 批次 B — 三层开关过滤

| 文件                                                                  | 改动                                                                                                      |
| --------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `apps/task/mq/internal/svc/servicecontext.go`                         | 注入 `UserSettingsModel`、`SystemConfigModel`                                                             |
| `apps/task/mq/internal/handler/msg_transfer/msg_read_transfer.go`     | 调用 `shouldSendReceipt(systemOn, readerOn, senderOn)`，任一关则跳过 `MsgChatTransfer`                   |
| `apps/task/mq/internal/handler/msg_transfer/msg_read_transfer.go`     | 把 `MsgType: ContentMakeRead` 同时填入 `MsgChatTransfer`（保证 pusher 把 mType 带到前端），去掉未用的 `ContentType` |

### 批次 C — 前端 store、WS 分发、乱序缓存

| 文件                                | 改动                                                                                     |
| ----------------------------------- | ---------------------------------------------------------------------------------------- |
| `web/src/lib/ws-client.ts`          | `MsgType.ContentMakeRead = 6`，`WsChatData.msg.mType` 保持 number                        |
| `web/src/lib/mock-data.ts`          | `Message` 加 `isRead?`、`readCount?`、`readTotal?`                                       |
| `web/src/lib/chat-store.ts`         | `handlePush` 识别 `mType===6` → 更新 `readRecords/isRead/readCount`；失配则进入 pending 缓存；消息落本地时消费缓存 |

### 批次 D — UI、设置开关、自动上报

| 文件                                                | 改动                                                                                                       |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `web/src/components/im/SettingsPage.tsx`            | Notification 区新增 `ToggleRow`「消息已读回执」，绑定 `readReceiptEnabled`                                  |
| `web/src/components/im/ChatDetail.tsx`              | 状态渲染：私聊显示「已读 · 时间」，群聊显示「X/N 人已读」；进入会话对未读消息调用已存在的 `markRead`；仅当 `readReceiptEnabled` 为 true 时区分未读/已读 |

## 七、数据结构

### `system_settings`

沿用现网已有 schema（prefixed camelCase key_name）：

```sql
CREATE TABLE `system_settings` (
  `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `key_name`   VARCHAR(100)     NOT NULL,
  `value`      TEXT             NOT NULL,
  `remark`     VARCHAR(255)     NOT NULL DEFAULT '',
  `created_at` TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key_name` (`key_name`)
);
```

初始化：`INSERT ... ON DUPLICATE KEY UPDATE` 写入 `readReceiptEnabled='true'`（系统默认开）。

### 用户设置 JSON

```json
{
  "themeMode": "system",
  "notifyEnabled": true,
  "...": "...",
  "readReceiptEnabled": true
}
```

默认 `true`（若字段缺失按开启处理，保持现网友好默认）。

### WS 推送（回执）

```json
{
  "conversationId": "10_20",
  "chatType": 1,
  "sendId": "20",
  "recvId": "10",
  "sendTime": 1718534400000000000,
  "msg": {
    "mType": 6,
    "content": "",
    "readRecords": { "65a...": "AQ==" }
  }
}
```

- `mType=6` 标记这是已读回执。
- `readRecords.key` = MongoDB msgId；`value` = bitmap 字节的 base64。
- `sendId` 是"阅读者"，`recvId` 是"原消息发送者"。

## 八、不影响现有功能的约束

1. 保留 `ContentMakeRead` 值仍为 6；外层 `MsgType` 枚举不新增位置。
2. DB 侧读写路径（`UpdateChatLogRead`、未读计数）无条件执行，仅 WS 下游推送受开关影响。
3. 默认值均为 ON，老用户无感。
4. 前端 UI：关闭时退化为现有「灰色 ✓✓ + 时间」渲染，与老版本视觉一致。
5. 乱序缓存只处理 `mType=6` 分支，不干预普通消息流。
6. 所有新增数据库字段、列均允许缺省，避免老数据迁移阻塞。
