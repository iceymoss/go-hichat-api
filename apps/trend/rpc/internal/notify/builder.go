// Package notify 动态消息通知的纯领域逻辑：把一次互动事件翻译成应生成的消息集合。
// 不依赖 DB / RPC / ctx，便于单元测试；持久化与推送由上层（rpc logic / api logic）负责。
package notify

import "github.com/iceymoss/go-hichat-api/pkg/constants"

// TrendNotifyEvent 一次互动事件的输入。
//   - Like:    AuthorId=动态作者
//   - Comment: AuthorId=动态作者, CommentId=新评论ID, AtUsers=评论里@的人
//   - Reply:   AuthorId=父评论作者, CommentId=新评论ID, ParentCommentId=父评论ID, AtUsers=@的人
//   - AtTrend: AuthorId 不用, AtUsers=发动态时@的人
type TrendNotifyEvent struct {
	Type            constants.TrendMsgType // 主事件类型
	ActorId         uint64                 // 操作人 uid
	TrendId         uint64                 // 动态 ID
	AuthorId        uint64                 // 主接收者：动态作者(赞/评论) 或 父评论作者(回复)
	CommentId       uint64                 // 关联评论 ID（评论/回复）
	ParentCommentId uint64                 // 父评论 ID（回复）
	AtUsers         []uint64               // 被 @的 uid 列表
	Content         string                 // 内容快照
}

// TrendMessageSpec 一条待落库的消息（持久化前的纯数据）。
type TrendMessageSpec struct {
	ReceiverId      uint64
	ActorId         uint64
	Type            constants.TrendMsgType
	TrendId         uint64
	CommentId       uint64
	ParentCommentId uint64
	Content         string
}

// BuildTrendMessages 把一次互动事件翻译成应生成的消息集合。
//
// 规则：
//   - 永不通知自己（ActorId 被过滤）。
//   - 赞/评论/回复先给主接收者（AuthorId）发一条主类型消息。
//   - @列表去重、去零、去自己；若某人已作为主接收者收到消息，则不再重复发 @消息。
//   - @消息类型派生：评论/回复 -> at_comment；发动态 -> at_trend。
func BuildTrendMessages(ev TrendNotifyEvent) []TrendMessageSpec {
	var out []TrendMessageSpec
	notified := map[uint64]bool{ev.ActorId: true} // 自己永不收到（含 actor=0 的兜底）

	// 主接收者：赞 / 评论 / 回复
	switch ev.Type {
	case constants.TrendMsgLike, constants.TrendMsgComment, constants.TrendMsgReply:
		if ev.AuthorId != 0 && !notified[ev.AuthorId] {
			out = append(out, specFrom(ev, ev.AuthorId, ev.Type))
			notified[ev.AuthorId] = true
		}
	}

	// @接收者：派生 @消息类型
	var atType constants.TrendMsgType
	switch ev.Type {
	case constants.TrendMsgComment, constants.TrendMsgReply:
		atType = constants.TrendMsgAtComment
	case constants.TrendMsgAtTrend:
		atType = constants.TrendMsgAtTrend
	}
	if atType == 0 {
		return out
	}

	for _, uid := range ev.AtUsers {
		if uid == 0 || notified[uid] {
			continue
		}
		out = append(out, specFrom(ev, uid, atType))
		notified[uid] = true
	}
	return out
}

func specFrom(ev TrendNotifyEvent, receiver uint64, t constants.TrendMsgType) TrendMessageSpec {
	return TrendMessageSpec{
		ReceiverId:      receiver,
		ActorId:         ev.ActorId,
		Type:            t,
		TrendId:         ev.TrendId,
		CommentId:       ev.CommentId,
		ParentCommentId: ev.ParentCommentId,
		Content:         ev.Content,
	}
}

// SumUnread 未读总数 = 各类型未读之和。
func SumUnread(byType map[uint64]int64) int64 {
	var total int64
	for _, c := range byType {
		total += c
	}
	return total
}
