package constants

// MType 消息类型
type MType int

// ChatType 聊天类型
type ChatType int

const (
	// ChatMsg 聊天类型
	ChatMsg MType = iota + 1

	// FileMsg 文件类型
	FileMsg

	// VoiceMsg 语音类型
	VoiceMsg

	// ImageMsg 图片类型
	ImageMsg

	// MemesMsg 表情包类型
)

const (
	UserType ChatType = iota + 1
	GroupType
)
