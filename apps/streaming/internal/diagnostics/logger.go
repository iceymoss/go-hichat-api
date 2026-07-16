package diagnostics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxFieldLength = 512

var sensitiveFields = map[string]struct{}{
	"candidate":  {},
	"credential": {},
	"jwt":        {},
	"sdp":        {},
	"token":      {},
}

var allowedFields = map[string]struct{}{
	"attempt": {}, "attempts": {}, "audio_tracks": {}, "candidate_type": {}, "ended": {},
	"error": {}, "kind": {}, "message_type": {}, "participant_count": {}, "peer_uid": {},
	"reconnect_attempt": {},
	"phase":             {}, "protocol": {}, "reason": {}, "remaining_count": {}, "replaced": {}, "sdp_bytes": {}, "signaling_state": {},
	"stage": {}, "state": {}, "tcp_type": {}, "track_state": {}, "video_tracks": {},
}

// Event 是一条可按通话、用户和浏览器会话关联的 SFU 诊断事件。
type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	UID       string         `json:"uid,omitempty"`
	CallID    string         `json:"call_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Name      string         `json:"event"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Logger 并发安全地写 JSONL；达到 maxBytes 后保留一份 .1 旧日志。
type Logger struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	file     *os.File
	size     int64
}

func New(path string, maxBytes int64) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Logger{path: path, maxBytes: maxBytes, file: file, size: info.Size()}, nil
}

func (l *Logger) Write(event Event) {
	event.Timestamp = time.Now().UTC()
	event.Fields = sanitizeFields(event.Fields)
	line, err := json.Marshal(event)
	if err != nil {
		return
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return
	}
	if l.maxBytes > 0 && l.size > 0 && l.size+int64(len(line)) > l.maxBytes {
		if err := l.rotate(); err != nil {
			return
		}
	}
	n, err := l.file.Write(line)
	if err == nil {
		l.size += int64(n)
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

func (l *Logger) rotate() error {
	if err := l.file.Close(); err != nil {
		return err
	}
	_ = os.Remove(l.path + ".1")
	if err := os.Rename(l.path, l.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	l.file = file
	l.size = 0
	return nil
}

func sanitizeFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return nil
	}
	clean := make(map[string]any, len(fields))
	for key, value := range fields {
		if _, allowed := allowedFields[key]; !allowed {
			continue
		}
		if _, sensitive := sensitiveFields[key]; sensitive {
			continue
		}
		switch typed := value.(type) {
		case string:
			if len(typed) > maxFieldLength {
				value = typed[:maxFieldLength]
			}
		case bool, float64, int, int64:
		default:
			continue
		}
		clean[key] = value
	}
	return clean
}
