package logic

import (
	"testing"
	"time"
)

func TestCallService_Create(t *testing.T) {
	tests := []struct {
		name     string
		caller   string
		callee   string
		callType CallType
		wantErr  error
	}{
		{name: "ok video", caller: "a", callee: "b", callType: CallVideo, wantErr: nil},
		{name: "ok voice", caller: "a", callee: "b", callType: CallVoice, wantErr: nil},
		{name: "self call", caller: "a", callee: "a", callType: CallVoice, wantErr: ErrSelfCall},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewCallService(time.Minute)
			s, err := cs.Create(tt.caller, tt.callee, tt.callType)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if s.State != CallStateInviting {
				t.Errorf("state = %v, want inviting", s.State)
			}
			if !cs.IsBusy(tt.caller) || !cs.IsBusy(tt.callee) {
				t.Errorf("both parties should be busy after create")
			}
		})
	}
}

func TestCallService_Busy(t *testing.T) {
	cs := NewCallService(time.Minute)
	if _, err := cs.Create("a", "b", CallVoice); err != nil {
		t.Fatal(err)
	}
	// b 忙线时，c 再呼 b 应失败
	if _, err := cs.Create("c", "b", CallVoice); err != ErrBusy {
		t.Errorf("err = %v, want ErrBusy", err)
	}
	// a 忙线时，a 再呼 d 应失败
	if _, err := cs.Create("a", "d", CallVoice); err != ErrBusy {
		t.Errorf("err = %v, want ErrBusy", err)
	}
}

func TestCallService_Lifecycle(t *testing.T) {
	tests := []struct {
		name       string
		action     func(cs *CallService, callID string) (*CallSession, error)
		wantState  CallState
		wantReason EndReason
		wantErr    error
	}{
		{
			name:      "callee accept -> connected",
			action:    func(cs *CallService, id string) (*CallSession, error) { return cs.Accept(id, "b") },
			wantState: CallStateConnected,
		},
		{
			name:       "callee reject -> ended/rejected",
			action:     func(cs *CallService, id string) (*CallSession, error) { return cs.Reject(id, "b") },
			wantState:  CallStateEnded,
			wantReason: EndRejected,
		},
		{
			name:       "caller cancel -> ended/canceled",
			action:     func(cs *CallService, id string) (*CallSession, error) { return cs.Cancel(id, "a") },
			wantState:  CallStateEnded,
			wantReason: EndCanceled,
		},
		{
			name:    "caller cannot accept",
			action:  func(cs *CallService, id string) (*CallSession, error) { return cs.Accept(id, "a") },
			wantErr: ErrNotParticipant,
		},
		{
			name:    "callee cannot cancel",
			action:  func(cs *CallService, id string) (*CallSession, error) { return cs.Cancel(id, "b") },
			wantErr: ErrNotParticipant,
		},
		{
			name:    "stranger cannot act",
			action:  func(cs *CallService, id string) (*CallSession, error) { return cs.End(id, "z") },
			wantErr: ErrNotParticipant,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewCallService(time.Minute)
			s, err := cs.Create("a", "b", CallVideo)
			if err != nil {
				t.Fatal(err)
			}
			got, err := tt.action(cs, s.ID)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got.State != tt.wantState {
				t.Errorf("state = %v, want %v", got.State, tt.wantState)
			}
			if tt.wantReason != "" && got.EndReason != tt.wantReason {
				t.Errorf("reason = %v, want %v", got.EndReason, tt.wantReason)
			}
			// 结束后双方应不再忙线
			if tt.wantState == CallStateEnded {
				if cs.IsBusy("a") || cs.IsBusy("b") {
					t.Errorf("parties should be free after end")
				}
			}
		})
	}
}

func TestCallService_EndAfterConnected(t *testing.T) {
	cs := NewCallService(time.Minute)
	s, _ := cs.Create("a", "b", CallVoice)
	if _, err := cs.Accept(s.ID, "b"); err != nil {
		t.Fatal(err)
	}
	ended, err := cs.End(s.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if ended.State != CallStateEnded || ended.EndReason != EndCompleted {
		t.Errorf("got state=%v reason=%v, want ended/completed", ended.State, ended.EndReason)
	}
	if _, ok := cs.Get(s.ID); ok {
		t.Errorf("session should be removed after end")
	}
}

func TestCallService_Timeout(t *testing.T) {
	cs := NewCallService(20 * time.Millisecond)
	fired := make(chan *CallSession, 1)
	cs.SetTimeoutHandler(func(s *CallSession) { fired <- s })

	s, _ := cs.Create("a", "b", CallVideo)
	_ = s
	select {
	case got := <-fired:
		if got.EndReason != EndNoAnswer {
			t.Errorf("reason = %v, want no_answer", got.EndReason)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout handler not fired")
	}
	if cs.IsBusy("a") || cs.IsBusy("b") {
		t.Errorf("parties should be free after timeout")
	}

	// 已接通的通话不应再触发超时
	cs2 := NewCallService(20 * time.Millisecond)
	cs2.SetTimeoutHandler(func(s *CallSession) { t.Errorf("should not fire after accept") })
	s2, _ := cs2.Create("a", "b", CallVideo)
	if _, err := cs2.Accept(s2.ID, "b"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(60 * time.Millisecond)
}
