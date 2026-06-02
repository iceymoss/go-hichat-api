package logic

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/iceymoss/go-hichat-api/pkg/storage"
)

// fakeStorage 记录最后一次上传的 folder，便于断言归档目录
type fakeStorage struct {
	gotFolder string
	retURL    string
}

func (f *fakeStorage) UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error) {
	f.gotFolder = folder
	return f.retURL, nil
}
func (f *fakeStorage) DeleteFile(ctx context.Context, url string) error { return nil }
func (f *fakeStorage) GetFileURL(path string) string                   { return path }

func Test_uploadMedia_Behaviors(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		size       int64
		wantErr    bool
		wantKind   string
		wantFolder string
	}{
		{"empty filename rejected", "", 100, true, "", ""},
		{"over 100MB rejected", "big.mp4", 100*1024*1024 + 1, true, "", ""},
		{"image goes to im/image", "p.png", 1024, false, storage.MediaImage, "im/image"},
		{"video goes to im/video", "c.mp4", 1024, false, storage.MediaVideo, "im/video"},
		{"voice goes to im/voice", "v.m4a", 1024, false, storage.MediaVoice, "im/voice"},
		{"file goes to im/file", "d.pdf", 1024, false, storage.MediaFile, "im/file"},
		{"exactly 100MB allowed", "ok.png", 100 * 1024 * 1024, false, storage.MediaImage, "im/image"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &fakeStorage{retURL: "http://x/y"}
			url, kind, err := uploadMedia(context.Background(), fs, tt.filename, tt.size, strings.NewReader("data"))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if kind != tt.wantKind {
				t.Errorf("kind = %q, want %q", kind, tt.wantKind)
			}
			if fs.gotFolder != tt.wantFolder {
				t.Errorf("folder = %q, want %q", fs.gotFolder, tt.wantFolder)
			}
			if url != "http://x/y" {
				t.Errorf("url = %q, want passthrough from storage", url)
			}
		})
	}
}
