package sfu

import "testing"

func Test_SFU_AddPeer_DuplicateReturnsExistingPeer(t *testing.T) {
	s := NewSFU()

	first, err := s.AddPeer("call1", "alice", nil)
	if err != nil {
		t.Fatalf("first AddPeer: %v", err)
	}
	defer first.Close()

	second, err := s.AddPeer("call1", "alice", nil)
	if err != nil {
		t.Fatalf("duplicate AddPeer: %v", err)
	}
	if second != first {
		t.Fatal("duplicate AddPeer created a second peer")
	}
	if got := len(s.rooms["call1"].Participants()); got != 1 {
		t.Fatalf("participants = %d, want 1", got)
	}
}

func Test_SFU_RemovePeer_LastParticipantDeletesRoomState(t *testing.T) {
	s := NewSFU()
	if _, err := s.AddPeer("call1", "alice", nil); err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	s.setPubSSRC("alice", "video", 42)

	s.RemovePeer("alice")

	if s.GetPeer("alice") != nil {
		t.Fatal("peer still registered after RemovePeer")
	}
	if _, ok := s.rooms["call1"]; ok {
		t.Fatal("empty room still registered after RemovePeer")
	}
	if _, ok := s.getPubSSRC("alice", "video"); ok {
		t.Fatal("publisher SSRC still registered after RemovePeer")
	}
}

func Test_SFU_RemoveRoom_ClosesAllParticipants(t *testing.T) {
	s := NewSFU()
	for _, uid := range []string{"alice", "bob"} {
		if _, err := s.AddPeer("call1", uid, nil); err != nil {
			t.Fatalf("AddPeer(%s): %v", uid, err)
		}
	}

	s.RemoveRoom("call1")

	if _, ok := s.rooms["call1"]; ok {
		t.Fatal("room still registered after RemoveRoom")
	}
	for _, uid := range []string{"alice", "bob"} {
		if s.GetPeer(uid) != nil {
			t.Fatalf("peer %s still registered after RemoveRoom", uid)
		}
	}
}
