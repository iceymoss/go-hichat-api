package mq

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type MsgChatTransfer struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType"`
	// 会话id
	ConversationId string `json:"conversationId"`
	// 发送者
	SendId string `json:"sendId"`
	// 接收者
	RecvId string `json:"recvId"`

	RecvIdList []string `json:"recvIdList"`

	// 消息类型
	MsgType constants.MType `json:"mType,omitempty"`

	// 已读记录
	ReadRecords map[string]string `json:"readRecords"`

	// 消息内容
	MsgContent string `json:"content,omitempty"`

	// 引用/回复消息（JSON 字符串：{"id","name","preview"}）
	Quote string `json:"quote,omitempty"`

	// 发送时间
	SendTime int64 `json:"sendTime"`

	ContentType constants.MType

	// MongoDB 持久化后的聊天记录 ID（hex），用于前端把 local_ 占位 ID 替换为真实 ID
	MsgId string `json:"msgId,omitempty"`

	// RecalledBy 撤回操作者 uid，仅 MsgType=ContentRecall 的撤回控制帧填充
	RecalledBy string `json:"recalledBy,omitempty"`

	// AtUsers 被 @ 的成员 uid 列表（仅群聊普通消息）
	AtUsers []string `json:"atUsers,omitempty"`

	// AtAll 是否 @所有人（仅群聊普通消息，权限已在生产端把关）
	AtAll bool `json:"atAll,omitempty"`
}

// MsgRecallTransfer 撤回事件消息结构体。
// 撤回与普通聊天消息是两类语义不同的领域事件，走独立 topic（msgRecallTransfer），
// 消费端据此扇出撤回控制帧，避免与聊天消息共享通道造成的领域耦合与吞吐相互影响。
// 注意：仅承载撤回所需字段；推送给前端的帧仍由消费端翻译成 ContentRecall 控制帧，对外协议不变。
type MsgRecallTransfer struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType"`
	// 会话id
	ConversationId string `json:"conversationId"`
	// 原消息发送者
	SendId string `json:"sendId"`
	// 原消息接收者（私聊为对端 uid，群聊为群 id）
	RecvId string `json:"recvId"`
	// 被撤回的消息 ID
	MsgId string `json:"msgId"`
	// RecalledBy 撤回操作者 uid
	RecalledBy string `json:"recalledBy"`
}

// MsgMarkRead 标记已读消息结构体
type MsgMarkRead struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType,omitempty"`
	// 会话id
	ConversationId string `json:"conversationId,omitempty"`
	// 发送者
	SendId string `json:"sendId,omitempty"`
	// 接收者
	RecvId string `json:"recvId,omitempty"`

	// 已读消息集合
	MsgIds []string `json:"msgIds,omitempty"`

	ReadRecords map[string]string `mapstructure:"readRecords"`
}

// TrendNotifyTransfer 动态消息通知事件（trend-api 生产 -> task/mq 消费 -> ws 推送）。
// 与聊天/撤回链路解耦，走独立 topic（trendNotifyTransfer）。
type TrendNotifyTransfer struct {
	// 接收者(被通知人)uid
	ReceiverId string `json:"receiverId"`
	// 触发者uid
	ActorId string `json:"actorId"`
	// 1赞 2评论 3回复 4发动态@ 5评论@
	MsgType int `json:"msgType"`
	// 动态ID
	TrendId uint64 `json:"trendId"`
	// 评论ID
	CommentId uint64 `json:"commentId,omitempty"`
	// 父评论ID
	ParentCommentId uint64 `json:"parentCommentId,omitempty"`
	// trend_message.id
	MessageId uint64 `json:"messageId"`
	// 内容快照
	Content string `json:"content,omitempty"`
	// 创建时间(时间戳)
	CreateTime int64 `json:"createTime"`
}

// RelationChangeTransfer 关系变更事件（social outbox -> relay 生产 -> task/mq 消费）。
// 消费端据此维护关系缓存（版本门）并推送会话失效帧给受影响用户。走独立 topic relationChangeTransfer。
type RelationChangeTransfer struct {
	// 事件类型，见 constants.RelationEvent*
	EventType string `json:"eventType"`
	// 群ID（群类事件）
	GroupId string `json:"groupId,omitempty"`
	// 受影响成员 uid（被踢/退群者）
	UserId string `json:"userId,omitempty"`
	// 操作者 uid
	OperatorId string `json:"operatorId,omitempty"`
	// 好友对（好友类事件，双向）
	FriendA string `json:"friendA,omitempty"`
	FriendB string `json:"friendB,omitempty"`
	// 单调版本号（= outbox.id），消费端版本门用
	Version int64 `json:"version"`
	// 事件时间戳
	Timestamp int64 `json:"timestamp"`
}

// CommonNotify 公共通知事件（业务方直接 Push -> task/mq 消费 -> 落库 + ws 在线推送）。
// 单一公共 topic（commonNotifyTransfer），notifyType 字符串区分业务，后续新增业务只加 type，不动通道。
// 多接收者（如 group.apply 通知群主+管理员）在生产端展开为多条单接收者消息。
type CommonNotify struct {
	EventId     uint64 `json:"eventId,omitempty"`
	RequestType string `json:"requestType,omitempty"`
	RequestId   uint64 `json:"requestId,omitempty"`
	Result      int32  `json:"result,omitempty"`
	// 通知类型：friend.apply / friend.accept / friend.reject / group.apply / group.accept / group.reject ...
	NotifyType string `json:"notifyType"`
	// 接收者 uid（单接收者）
	ReceiverId string `json:"receiverId"`
	// 触发者 uid
	ActorId string `json:"actorId,omitempty"`
	// 业务关联ID（幂等键 + 前端跳转）。同一事件重投须用相同 bizId，不同事件用不同 bizId。
	BizId string `json:"bizId,omitempty"`
	// 群ID（群相关通知）
	GroupId string `json:"groupId,omitempty"`
	// 冗余标题（可空，前端也可用 notifyType + i18n 渲染）
	Title string `json:"title,omitempty"`
	// 冗余正文/摘要
	Content string `json:"content,omitempty"`
	// 扩展字段 JSON（向前兼容，宽容反序列化）
	Payload string `json:"payload,omitempty"`
	// 创建时间(unix 秒)
	CreateTime int64 `json:"createTime"`
}
