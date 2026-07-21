package sfu

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

const (
	defaultLoadParticipants = 50
	loadTopAudio            = 4
	loadVisibleVideo        = 8
)

type countingSink struct{ writes atomic.Uint64 }

func (s *countingSink) WriteRTP(*rtp.Packet) error {
	s.writes.Add(1)
	return nil
}

func newLoadRoom(participants, topAudio, visibleVideo int) (*Room, []*countingSink, uint64) {
	room := NewRoom("load")
	sinks := make([]*countingSink, 0, participants*(participants+visibleVideo))
	for i := 0; i < participants; i++ {
		uid := fmt.Sprintf("u%02d", i)
		room.Join(uid)
		room.ManageAudio(uid)
		room.AddPublished(uid, "audio", "audio", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
		room.AddPublishedLayer(uid, "video", "q", "video", webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	}
	active := make([]string, topAudio)
	for i := range active {
		active[i] = fmt.Sprintf("u%02d", i)
	}
	room.SetActiveSpeakers(active)
	for publisher := 0; publisher < participants; publisher++ {
		pubUID := fmt.Sprintf("u%02d", publisher)
		for subscriber := 0; subscriber < participants; subscriber++ {
			if subscriber == publisher {
				continue
			}
			sink := &countingSink{}
			sinks = append(sinks, sink)
			room.Subscribe(fmt.Sprintf("u%02d", subscriber), pubUID, "audio", sink)
		}
		for offset := 1; offset <= visibleVideo; offset++ {
			subUID := fmt.Sprintf("u%02d", (publisher+offset)%participants)
			sink := &countingSink{}
			sinks = append(sinks, sink)
			room.SubscribeLayer(subUID, pubUID, "video", "q", sink)
		}
	}
	expected := uint64(topAudio*(participants-1) + participants*visibleVideo)
	return room, sinks, expected
}

func routeLoadCycle(room *Room, participants int, packet *rtp.Packet) {
	for i := 0; i < participants; i++ {
		uid := fmt.Sprintf("u%02d", i)
		room.RouteRTP(uid, "audio", packet)
		room.RouteRTPLayer(uid, "video", "q", packet)
	}
}

func Test_LoadProfile_BoundsFanout(t *testing.T) {
	participants := loadParticipantCount()
	topAudio := min(loadTopAudio, participants)
	visibleVideo := min(loadVisibleVideo, participants-1)
	room, sinks, expected := newLoadRoom(participants, topAudio, visibleVideo)
	routeLoadCycle(room, participants, &rtp.Packet{})
	var got uint64
	for _, sink := range sinks {
		got += sink.writes.Load()
	}
	if got != expected {
		t.Fatalf("fanout writes = %d, want %d", got, expected)
	}
}

func BenchmarkRoomFanout(b *testing.B) {
	participants := loadParticipantCount()
	room, _, _ := newLoadRoom(participants, min(loadTopAudio, participants), min(loadVisibleVideo, participants-1))
	packet := &rtp.Packet{Header: rtp.Header{PayloadType: 96}, Payload: make([]byte, 1200)}
	b.ReportAllocs()
	b.SetBytes(int64(len(packet.Payload) * (participants * 2)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		routeLoadCycle(room, participants, packet)
	}
}

func loadParticipantCount() int {
	participants, err := strconv.Atoi(os.Getenv("SFU_LOAD_PARTICIPANTS"))
	if err != nil || participants < 2 || participants > 50 {
		return defaultLoadParticipants
	}
	return participants
}
