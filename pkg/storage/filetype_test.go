package storage

import "testing"

func Test_ClassifyMedia_ByExtension_ReturnsKind(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"jpg is image", "photo.jpg", MediaImage},
		{"JPEG uppercase is image", "PHOTO.JPEG", MediaImage},
		{"png is image", "a.png", MediaImage},
		{"gif is image", "sticker.gif", MediaImage},
		{"webp is image", "x.webp", MediaImage},
		{"mp4 is video", "clip.mp4", MediaVideo},
		{"mov is video", "clip.MOV", MediaVideo},
		{"mkv is video", "clip.mkv", MediaVideo},
		{"mp3 is voice", "v.mp3", MediaVoice},
		{"m4a is voice", "v.m4a", MediaVoice},
		{"aac is voice", "v.aac", MediaVoice},
		{"ogg is voice", "v.ogg", MediaVoice},
		{"wav is voice", "v.wav", MediaVoice},
		{"pdf is file", "doc.pdf", MediaFile},
		{"zip is file", "a.zip", MediaFile},
		{"no extension is file", "README", MediaFile},
		{"unknown extension is file", "a.xyz", MediaFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyMedia(tt.filename); got != tt.want {
				t.Errorf("ClassifyMedia(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
