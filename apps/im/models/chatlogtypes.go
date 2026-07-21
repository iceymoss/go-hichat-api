package model

import (
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatLog 聊天记录
type ChatLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 会话id
	ConversationId string `bson:"conversationId"`

	// 发送者id
	SendId string `bson:"sendId"`

	// 接受者id
	RecvId string `bson:"recvId"`

	// 消息来源
	MsgFrom int `bson:"msgFrom"`

	// 消息类型
	MsgType constants.MType `bson:"msgType"`

	// 消息内容
	MsgContent string `bson:"msgContent"`

	// 引用/回复消息（JSON 字符串：{"id","name","preview"}）
	Quote string `bson:"quote,omitempty"`

	// 发送时间
	SendTime int64 `bson:"sendTime"`

	// 消息状态：0=正常 1=已撤回（见 constants.MsgStatusNormal/Recalled）
	Status int `bson:"status"`

	// 撤回操作者 uid（本人或群管理员）。前端据此区分"你/对方/管理员撤回了一条消息"
	RecalledBy string `bson:"recalledBy,omitempty" json:"recalledBy,omitempty"`

	// 撤回时间（unix nano）
	RecalledAt int64 `bson:"recalledAt,omitempty" json:"recalledAt,omitempty"`

	// 被 @ 的成员 uid 列表（仅群聊）
	AtUsers []string `bson:"atUsers,omitempty" json:"atUsers,omitempty"`

	// 是否 @所有人（仅群聊，权限已在生产端把关）
	AtAll bool `bson:"atAll,omitempty" json:"atAll,omitempty"`

	// 聊天类型
	ChatType constants.ChatType `bson:"chatType"`

	// 已读记录
	ReadRecords []byte `bson:"readRecords"`

	// 每个成员的已读时间戳（unix nano），key=userId。群聊里用来展示"某成员何时读了"
	// sender 自己由 addChatLog 记为发送即读；其他成员由 msg_read_transfer 填入。
	ReadTimes map[string]int64 `bson:"readTimes,omitempty" json:"readTimes,omitempty"`

	// 更新时间
	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`

	// 创建时间
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
