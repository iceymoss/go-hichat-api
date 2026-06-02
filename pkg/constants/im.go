package constants

// MType 消息类型
type MType int

// ChatType 聊天类型
type ChatType int

const (
	// TextMType 文中
	TextMType MType = iota + 1

	// FileMType 文件类型
	FileMType

	// VoiceMType 语音类型
	VoiceMType

	// ImageMType 图片类型
	ImageMType

	// MemesMType 表情包类型
	MemesMType
)

// ContentMakeRead 已读回执消息类型（独立显式值，避免 iota 重排造成协议不兼容）
const ContentMakeRead MType = 6

// ContentMsgAck 发送方回响：携带 MongoDB 真实 MsgId，用于把前端本地占位 ID 升级为真实 ID
const ContentMsgAck MType = 7

// VideoMType 视频类型（追加在 6/7 控制类型之后，编号不复用）
const VideoMType MType = 8

const (
	// SingleChatType 单聊
	SingleChatType ChatType = iota + 1

	// GroupChatType 群聊
	GroupChatType
)
