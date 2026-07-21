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

// ContentRecall 撤回控制帧：复用 MsgChatTransfer 链路下发，携带被撤回的 msgId + 操作者，
// 前端按此类型把对应消息原位置为"已撤回"。编号续 8 之后，不复用。
const ContentRecall MType = 9

// CallMType 音视频通话记录消息：通话结束由主叫端经聊天链路投递，content 为 JSON
// {callType, status, duration}，前端渲染为通话气泡（可点击回拨）。编号续 9 之后，不复用。
const CallMType MType = 10

// 消息状态（ChatLog.Status）
const (
	// MsgStatusNormal 正常
	MsgStatusNormal = 0

	// MsgStatusRecalled 已撤回
	MsgStatusRecalled = 1
)

const (
	// SingleChatType 单聊
	SingleChatType ChatType = iota + 1

	// GroupChatType 群聊
	GroupChatType
)
