package ws

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type Msg struct {
	// 消息类型
	constants.MType `json:"mType" mapstructure:"mType"`

	// 消息内容
	Content string `json:"content" mapstructure:"content"`

	// 引用/回复消息（JSON 字符串：{"id","name","preview"}），空表示非引用消息
	Quote string `json:"quote,omitempty" mapstructure:"quote"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`

	// 被 @ 的成员 uid 列表（仅群聊）
	AtUsers []string `json:"atUsers,omitempty" mapstructure:"atUsers"`

	// 是否 @所有人（仅群聊）
	AtAll bool `json:"atAll,omitempty" mapstructure:"atAll"`
}

// Chat 聊天会话， message 结构中的data字段
type Chat struct {
	// 会话id
	ConversationId string `json:"conversationId" mapstructure:"conversationId"`

	// 聊天类型
	constants.ChatType `json:"chatType" mapstructure:"chatType"`

	// 发送者
	SendId string `json:"sendId" mapstructure:"sendId"`

	// 接收者
	RecvId string `json:"recvId" mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `json:"sendTime" mapstructure:"sendTime"`

	// 附加内容类型（如 ContentMakeRead=6 已读回执, ContentMsgAck=7 发送方回响, ContentRecall=9 撤回），0 表示普通消息
	ContentType constants.MType `json:"contentType,omitempty" mapstructure:"contentType"`

	// 撤回操作者 uid，仅 ContentType=ContentRecall 时有值，前端据此区分"你/对方/管理员撤回了一条消息"
	RecalledBy string `json:"recalledBy,omitempty" mapstructure:"recalledBy"`

	// 发送内容
	Msg `json:"msg" mapstructure:"msg"`
}

type Push struct {
	// 会话id
	ConversationId string `json:"conversationId" mapstructure:"conversationId"`

	// 聊天类型：1. 私聊、2. 群聊
	constants.ChatType `json:"chatType" mapstructure:"chatType"`

	// 发送者
	SendId string `json:"sendId" mapstructure:"sendId"`

	RecvIdList []string `json:"recvIdList" mapstructure:"recvIdList"`

	// 接收者
	RecvId string `json:"recvId" mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `json:"sendTime" mapstructure:"sendTime"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`

	constants.MType `json:"mType" mapstructure:"mType"`
	Content         string `json:"content" mapstructure:"content"`

	// 引用/回复消息（JSON 字符串），随推送转发给接收方
	Quote string `json:"quote,omitempty" mapstructure:"quote"`

	// MongoDB 聊天记录 ObjectID（hex），用于前端把 local_ 占位 ID 替换为真实 ID
	MsgId string `json:"msgId" mapstructure:"msgId"`
	// 附加内容类型（ContentType=7 表示给发送方的回响）
	ContentType constants.MType `json:"contentType" mapstructure:"contentType"`

	// 撤回操作者 uid，仅 ContentType=ContentRecall 时有值
	RecalledBy string `json:"recalledBy,omitempty" mapstructure:"recalledBy"`

	// 被 @ 的成员 uid 列表（仅群聊普通消息）
	AtUsers []string `json:"atUsers,omitempty" mapstructure:"atUsers"`

	// 是否 @所有人（仅群聊普通消息）
	AtAll bool `json:"atAll,omitempty" mapstructure:"atAll"`
}

// MarkRead 已读标记
type MarkRead struct {
	constants.ChatType `json:"chatType" mapstructure:"chatType"`
	RecvId             string `json:"recvId" mapstructure:"recvId"`
	ConversationId     string `json:"conversationId" mapstructure:"conversationId"`
	// 发送者
	SendId      string            `json:"sendId" mapstructure:"sendId"`
	MsgIds      []string          `json:"msgIds" mapstructure:"msgIds"`
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`
}

// TrendNotify 动态消息通知推送帧（与聊天会话解耦，走 push.trend 路由单推给 ReceiverId）。
// mq 消费者构造此结构投给 ws，ws 用 push.trend handler 解码后下发，前端用 method=trend.notify 接收。
type TrendNotify struct {
	ReceiverId      string `json:"receiverId" mapstructure:"receiverId"`                     // 接收者(被通知人)uid
	ActorId         string `json:"actorId" mapstructure:"actorId"`                           // 触发者uid
	MsgType         int    `json:"msgType" mapstructure:"msgType"`                           // 1赞 2评论 3回复 4发动态@ 5评论@
	TrendId         uint64 `json:"trendId" mapstructure:"trendId"`                           // 动态ID
	CommentId       uint64 `json:"commentId,omitempty" mapstructure:"commentId"`             // 评论ID
	ParentCommentId uint64 `json:"parentCommentId,omitempty" mapstructure:"parentCommentId"` // 父评论ID
	MessageId       uint64 `json:"messageId" mapstructure:"messageId"`                       // trend_message.id
	Content         string `json:"content,omitempty" mapstructure:"content"`                 // 内容快照
	CreateTime      int64  `json:"createTime" mapstructure:"createTime"`                     // 创建时间(时间戳)
}

// RelationChanged 关系变更通知帧（被踢/退群/解散/删好友后，实时通知受影响用户禁用会话输入）。
// mq 消费者构造此结构投给 ws，ws 用 push.relation handler 下发，前端用 method=relation.changed 接收。
type RelationChanged struct {
	ReceiverId     string `json:"receiverId" mapstructure:"receiverId"`         // 接收者(被影响人)uid
	EventType      string `json:"eventType" mapstructure:"eventType"`           // constants.RelationEvent*
	ConversationId string `json:"conversationId" mapstructure:"conversationId"` // 受影响会话id（前端据此禁用输入）
	GroupId        string `json:"groupId,omitempty" mapstructure:"groupId"`     // 群类事件的群ID
	PeerId         string `json:"peerId,omitempty" mapstructure:"peerId"`       // 好友类事件的对端 uid
	OperatorId     string `json:"operatorId,omitempty" mapstructure:"operatorId"` // 操作者 uid
	Timestamp      int64  `json:"timestamp" mapstructure:"timestamp"`
}

// Notify 公共通知推送帧（公共通知通道：业务方投递 -> task/mq 消费落库 -> 走 push.notify 路由单推给 ReceiverId）。
// 前端用 method=notify 接收，按 notifyType 分发（红点/toast/通知中心/会话内系统消息）。
type Notify struct {
	NotifyType string `json:"notifyType" mapstructure:"notifyType"` // friend.apply / group.accept ...
	ReceiverId string `json:"receiverId" mapstructure:"receiverId"` // 接收者 uid
	ActorId    string `json:"actorId,omitempty" mapstructure:"actorId"`
	BizId      string `json:"bizId,omitempty" mapstructure:"bizId"`     // 业务关联ID(跳转)
	GroupId    string `json:"groupId,omitempty" mapstructure:"groupId"` // 群相关通知的群 id
	Title      string `json:"title,omitempty" mapstructure:"title"`
	Content    string `json:"content,omitempty" mapstructure:"content"`
	Payload    string `json:"payload,omitempty" mapstructure:"payload"` // 扩展字段 JSON
	CreateTime int64  `json:"createTime" mapstructure:"createTime"`
}
