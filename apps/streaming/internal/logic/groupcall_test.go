package logic

import (
	"sort"
	"testing"
	"time"
)

func TestGroupCallService_CreateJoinLeave(t *testing.T) {
	s := NewGroupCallService(4)

	callID, err := s.Create("a", "g1", CallVideo, []string{"b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if !s.HasParticipant(callID, "a") {
		t.Fatal("initiator should be a participant")
	}
	if !s.IsBusy("a") {
		t.Fatal("initiator should be busy")
	}

	// b 加入：返回的 others 应含 a
	others, typ, err := s.Join(callID, "b")
	if err != nil {
		t.Fatal(err)
	}
	if typ != CallVideo {
		t.Errorf("type = %v, want video", typ)
	}
	if len(others) != 1 || others[0] != "a" {
		t.Errorf("others = %v, want [a]", others)
	}

	// c 加入：others 应含 a,b
	others, _, err = s.Join(callID, "c")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(others)
	if len(others) != 2 || others[0] != "a" || others[1] != "b" {
		t.Errorf("others = %v, want [a b]", others)
	}

	ps, ok := s.Participants(callID)
	if !ok || len(ps) != 3 {
		t.Errorf("participants = %v, want 3", ps)
	}

	// b 离开：剩 a,c，未结束
	_, ended, err := s.Leave(callID, "b")
	if err != nil {
		t.Fatal(err)
	}
	if ended {
		t.Error("call should not end with 2 left")
	}
	if s.IsBusy("b") {
		t.Error("b should not be busy after leave")
	}

	// c 离开：剩 a (<2) → 结束
	_, ended, err = s.Leave(callID, "c")
	if err != nil {
		t.Fatal(err)
	}
	if !ended {
		t.Error("call should end with <2 participants")
	}
	if _, ok := s.Participants(callID); ok {
		t.Error("call should be removed after end")
	}
	if s.IsBusy("a") {
		t.Error("a should be freed after call ended")
	}
}

func TestGroupCallService_ActiveByGroup(t *testing.T) {
	s := NewGroupCallService(4)
	// 无通话
	if _, _, _, ok := s.ActiveByGroup("g1"); ok {
		t.Fatal("no active call expected")
	}
	callID, _ := s.Create("a", "g1", CallVideo, []string{"b"})
	if _, _, err := s.Join(callID, "b"); err != nil {
		t.Fatal(err)
	}
	id, typ, ps, ok := s.ActiveByGroup("g1")
	if !ok || id != callID || typ != CallVideo {
		t.Errorf("active = %v %v %v, want %v video true", id, typ, ok, callID)
	}
	sort.Strings(ps)
	if len(ps) != 2 || ps[0] != "a" || ps[1] != "b" {
		t.Errorf("participants = %v, want [a b]", ps)
	}
	// 其他群无通话
	if _, _, _, ok := s.ActiveByGroup("g2"); ok {
		t.Error("g2 should have no active call")
	}
	// 结束后不再活跃
	s.Leave(callID, "b")
	s.Leave(callID, "a")
	if _, _, _, ok := s.ActiveByGroup("g1"); ok {
		t.Error("call ended, should be inactive")
	}
}

func TestGroupCallService_Full(t *testing.T) {
	s := NewGroupCallService(2) // 上限 2 便于测试
	callID, _ := s.Create("a", "g1", CallVoice, []string{"b", "c"})
	if _, _, err := s.Join(callID, "b"); err != nil {
		t.Fatal(err)
	}
	// 满员
	if _, _, err := s.Join(callID, "c"); err != ErrGroupCallFull {
		t.Errorf("err = %v, want ErrGroupCallFull", err)
	}
}

func Test_GroupCallService_DefaultLimit_IsFifty(t *testing.T) {
	s := NewGroupCallService(0)
	if s.maxPart != 50 {
		t.Fatalf("default max participants = %d, want 50", s.maxPart)
	}
}

func TestGroupCallService_Errors(t *testing.T) {
	s := NewGroupCallService(4)
	if _, _, err := s.Join("nope", "x"); err != ErrGroupCallNotFound {
		t.Errorf("err = %v, want ErrGroupCallNotFound", err)
	}
	callID, _ := s.Create("a", "g1", CallVoice, nil)
	// a 已忙，不能再发起
	if _, err := s.Create("a", "g2", CallVoice, nil); err != ErrBusy {
		t.Errorf("err = %v, want ErrBusy", err)
	}
	_ = callID
	_ = time.Now
}
