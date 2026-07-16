package diagnostics

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Logger_Write_SanitizesAndPersistsJSONLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sfu.jsonl")
	logger, err := New(path, 1024*1024)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer logger.Close()

	logger.Write(Event{
		Source:    "client",
		UID:       "user-1",
		CallID:    "call-1",
		SessionID: "session-1",
		Name:      "ice_candidate_error",
		Fields: map[string]any{
			"state":     "failed",
			"error":     strings.Repeat("x", maxFieldLength+20),
			"candidate": "candidate:full-secret-address",
			"sdp":       "v=0 secret",
			"unknown":   "must not persist",
		},
	})

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("diagnostic file has no event")
	}
	var got Event
	if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != "ice_candidate_error" || got.UID != "user-1" || got.Timestamp.IsZero() {
		t.Fatalf("event = %+v", got)
	}
	if _, ok := got.Fields["candidate"]; ok {
		t.Fatal("full ICE candidate was persisted")
	}
	if _, ok := got.Fields["sdp"]; ok {
		t.Fatal("SDP was persisted")
	}
	if _, ok := got.Fields["unknown"]; ok {
		t.Fatal("unknown diagnostic field was persisted")
	}
	if len(got.Fields["error"].(string)) != maxFieldLength {
		t.Fatalf("error length = %d, want %d", len(got.Fields["error"].(string)), maxFieldLength)
	}
}

func Test_Logger_Write_RotatesAtConfiguredSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sfu.jsonl")
	logger, err := New(path, 1)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer logger.Close()

	logger.Write(Event{Source: "server", Name: "first"})
	logger.Write(Event{Source: "server", Name: "second"})

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("rotated file: %v", err)
	}
}

func Test_Logger_Write_PersistsVideoDiagnosticFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sfu.jsonl")
	logger, err := New(path, 1024*1024)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer logger.Close()

	logger.Write(Event{Source: "client", Name: "video_inbound_stats", Fields: map[string]any{
		"peer_uid": "user-2", "codec": "video/VP8", "frame_width": 1280,
		"frame_height": 720, "frames_decoded": 120, "frames_dropped": 3,
		"frames_per_second": 30, "key_frames_decoded": 2, "jitter_ms": 4,
		"packets_lost": 1, "freeze_count": 0, "stalled_ms": 0,
	}})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, field := range []string{
		"codec", "frame_width", "frame_height", "frames_decoded", "frames_dropped",
		"frames_per_second", "key_frames_decoded", "jitter_ms", "packets_lost", "freeze_count", "stalled_ms",
	} {
		if _, ok := got.Fields[field]; !ok {
			t.Errorf("video diagnostic field %q was not persisted", field)
		}
	}
}
