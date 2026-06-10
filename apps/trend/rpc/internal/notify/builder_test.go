package notify

import (
	"reflect"
	"testing"

	"github.com/iceymoss/go-hichat-api/pkg/constants"
)

// Test_BuildTrendMessages 覆盖五类互动事件 -> 应生成的消息集合，
// 验证：自我操作过滤、@去重、评论同时@主接收者去重、@类型派生。
func Test_BuildTrendMessages(t *testing.T) {
	cases := []struct {
		name string
		ev   TrendNotifyEvent
		want []TrendMessageSpec
	}{
		{
			name: "like other's trend -> one like message to author",
			ev: TrendNotifyEvent{
				Type:     constants.TrendMsgLike,
				ActorId:  10,
				TrendId:  100,
				AuthorId: 20,
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgLike, TrendId: 100},
			},
		},
		{
			name: "like own trend -> no message",
			ev: TrendNotifyEvent{
				Type:     constants.TrendMsgLike,
				ActorId:  10,
				TrendId:  100,
				AuthorId: 10,
			},
			want: nil,
		},
		{
			name: "comment other's trend, no @ -> one comment message to author",
			ev: TrendNotifyEvent{
				Type:      constants.TrendMsgComment,
				ActorId:   10,
				TrendId:   100,
				AuthorId:  20,
				CommentId: 500,
				Content:   "nice",
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgComment, TrendId: 100, CommentId: 500, Content: "nice"},
			},
		},
		{
			name: "comment + @ third parties -> comment to author + at_comment to others",
			ev: TrendNotifyEvent{
				Type:      constants.TrendMsgComment,
				ActorId:   10,
				TrendId:   100,
				AuthorId:  20,
				CommentId: 500,
				AtUsers:   []uint64{30, 40},
				Content:   "hey",
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgComment, TrendId: 100, CommentId: 500, Content: "hey"},
				{ReceiverId: 30, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 500, Content: "hey"},
				{ReceiverId: 40, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 500, Content: "hey"},
			},
		},
		{
			name: "comment AND @ the author -> author notified once (comment wins, no duplicate at_comment)",
			ev: TrendNotifyEvent{
				Type:      constants.TrendMsgComment,
				ActorId:   10,
				TrendId:   100,
				AuthorId:  20,
				CommentId: 500,
				AtUsers:   []uint64{20, 30},
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgComment, TrendId: 100, CommentId: 500},
				{ReceiverId: 30, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 500},
			},
		},
		{
			name: "@ list with duplicates / actor self / zero -> deduped & filtered",
			ev: TrendNotifyEvent{
				Type:      constants.TrendMsgComment,
				ActorId:   10,
				TrendId:   100,
				AuthorId:  20,
				CommentId: 500,
				AtUsers:   []uint64{30, 30, 10, 0, 40},
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgComment, TrendId: 100, CommentId: 500},
				{ReceiverId: 30, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 500},
				{ReceiverId: 40, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 500},
			},
		},
		{
			name: "reply other's comment -> reply message to parent comment author",
			ev: TrendNotifyEvent{
				Type:            constants.TrendMsgReply,
				ActorId:         10,
				TrendId:         100,
				AuthorId:        20, // parent comment author
				CommentId:       600,
				ParentCommentId: 500,
				Content:         "agreed",
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgReply, TrendId: 100, CommentId: 600, ParentCommentId: 500, Content: "agreed"},
			},
		},
		{
			name: "reply own comment -> no reply message, but @ others still notified",
			ev: TrendNotifyEvent{
				Type:            constants.TrendMsgReply,
				ActorId:         10,
				TrendId:         100,
				AuthorId:        10,
				CommentId:       600,
				ParentCommentId: 500,
				AtUsers:         []uint64{30},
			},
			want: []TrendMessageSpec{
				{ReceiverId: 30, ActorId: 10, Type: constants.TrendMsgAtComment, TrendId: 100, CommentId: 600, ParentCommentId: 500},
			},
		},
		{
			name: "publish trend @ users -> at_trend message to each, self filtered",
			ev: TrendNotifyEvent{
				Type:    constants.TrendMsgAtTrend,
				ActorId: 10,
				TrendId: 100,
				AtUsers: []uint64{20, 10, 30},
			},
			want: []TrendMessageSpec{
				{ReceiverId: 20, ActorId: 10, Type: constants.TrendMsgAtTrend, TrendId: 100},
				{ReceiverId: 30, ActorId: 10, Type: constants.TrendMsgAtTrend, TrendId: 100},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BuildTrendMessages(c.ev)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("BuildTrendMessages() =\n  %+v\nwant\n  %+v", got, c.want)
			}
		})
	}
}

// Test_SumUnread 未读总数 = 各类型之和。
func Test_SumUnread(t *testing.T) {
	cases := []struct {
		name string
		in   map[uint64]int64
		want int64
	}{
		{"empty", map[uint64]int64{}, 0},
		{"single type", map[uint64]int64{uint64(constants.TrendMsgLike): 3}, 3},
		{
			"mixed types",
			map[uint64]int64{
				uint64(constants.TrendMsgLike):    3,
				uint64(constants.TrendMsgComment): 2,
				uint64(constants.TrendMsgReply):   1,
				uint64(constants.TrendMsgAtTrend): 4,
			},
			10,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SumUnread(c.in); got != c.want {
				t.Errorf("SumUnread(%v) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}
