package sfu

import (
	"testing"
)

// Test_ActiveSpeakers_TopByLoudness 活跃发言人 top-N：
// RFC6464 音量 0..127（0=最响，127=静音）。持续响的排前，切换后随平滑更新，Remove 后剔除。
func Test_ActiveSpeakers_TopByLoudness(t *testing.T) {
	as := NewActiveSpeakers()

	// A 一直很响(level 低)，B 很轻，C 中等
	for i := 0; i < 20; i++ {
		as.Observe("A", 10)
		as.Observe("B", 120)
		as.Observe("C", 60)
	}
	if got := as.Top(1); len(got) != 1 || got[0] != "A" {
		t.Fatalf("Top(1) = %v, want [A]", got)
	}
	if got := as.Top(2); len(got) != 2 || got[0] != "A" || got[1] != "C" {
		t.Fatalf("Top(2) = %v, want [A C]", got)
	}

	// 切换：B 变响、A 变轻，持续一段时间后 B 应上榜首
	for i := 0; i < 40; i++ {
		as.Observe("A", 120)
		as.Observe("B", 5)
		as.Observe("C", 60)
	}
	if got := as.Top(1); len(got) != 1 || got[0] != "B" {
		t.Fatalf("after switch Top(1) = %v, want [B]", got)
	}

	// Remove 后不再出现
	as.Remove("B")
	for _, uid := range as.Top(3) {
		if uid == "B" {
			t.Fatalf("Top after Remove(B) still contains B: %v", as.Top(3))
		}
	}

	// n 大于人数时返回全部；n<=0 返回空
	if got := as.Top(10); len(got) != 2 {
		t.Errorf("Top(10) = %v, want 2 speakers (A,C)", got)
	}
	if got := as.Top(0); len(got) != 0 {
		t.Errorf("Top(0) = %v, want empty", got)
	}
}
