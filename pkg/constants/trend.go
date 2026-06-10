package constants

// TrendMsgType 动态消息通知类型（trend_message.type）
type TrendMsgType int

const (
	// TrendMsgLike 点赞动态
	TrendMsgLike TrendMsgType = iota + 1

	// TrendMsgComment 评论动态
	TrendMsgComment

	// TrendMsgReply 回复评论
	TrendMsgReply

	// TrendMsgAtTrend 发动态 @
	TrendMsgAtTrend

	// TrendMsgAtComment 评论中 @
	TrendMsgAtComment
)

// 动态消息已读状态（trend_message.is_read）
const (
	// TrendMsgUnread 未读
	TrendMsgUnread = 0

	// TrendMsgRead 已读
	TrendMsgRead = 1
)

// 动态消息状态（trend_message.state）
const (
	// TrendMsgStateDeleted 已删除（级联软删）
	TrendMsgStateDeleted = 0

	// TrendMsgStateNormal 正常
	TrendMsgStateNormal = 1
)
