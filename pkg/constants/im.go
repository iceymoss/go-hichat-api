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

	ContentMakeRead
)

const (
	// SingleChatType 单聊
	SingleChatType ChatType = iota + 1

	// GroupChatType 群聊
	GroupChatType
)
