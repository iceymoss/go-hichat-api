package logic

import (
	"testing"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
)

// Test_chatLogResp_RemovedAtCutoff 守护「被移出群成员只能看被移出前消息」的历史截断：
// removedAt>0 时剔除 sendTime>removedAt 的消息；removedAt<=0 不截断。
func Test_chatLogResp_RemovedAtCutoff(t *testing.T) {
	logs := []*model.ChatLog{
		{SendId: "11", SendTime: 100},
		{SendId: "11", SendTime: 200}, // 被移出时刻
		{SendId: "11", SendTime: 300}, // 被移出后
		{SendId: "11", SendTime: 400}, // 被移出后
	}

	tests := []struct {
		name      string
		removedAt int64
		wantCount int
	}{
		{"未截断(0)：全部返回", 0, 4},
		{"未截断(负)：全部返回", -1, 4},
		{"截断于200：含边界保留前两条", 200, 2},
		{"截断于早于全部：返回空", 50, 0},
		{"截断于晚于全部：全部返回", 1000, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := chatLogResp(logs, tt.removedAt)
			if got := len(resp.List); got != tt.wantCount {
				t.Errorf("chatLogResp(removedAt=%d) 返回 %d 条, 期望 %d", tt.removedAt, got, tt.wantCount)
			}
			// 截断后绝不含晚于 removedAt 的消息
			if tt.removedAt > 0 {
				for _, c := range resp.List {
					if c.SendTime > tt.removedAt {
						t.Errorf("泄漏了 sendTime=%d > removedAt=%d 的消息", c.SendTime, tt.removedAt)
					}
				}
			}
		})
	}
}
