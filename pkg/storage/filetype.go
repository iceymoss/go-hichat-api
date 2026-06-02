package storage

import (
	"path/filepath"
	"strings"
)

// 媒体类型分类，用于按类型归档存储与前端渲染分支
const (
	MediaImage = "image"
	MediaVideo = "video"
	MediaVoice = "voice"
	MediaFile  = "file"
)

var (
	imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}
	videoExts = map[string]bool{".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true}
	voiceExts = map[string]bool{".mp3": true, ".m4a": true, ".aac": true, ".ogg": true, ".wav": true, ".opus": true, ".amr": true}
)

// ClassifyMedia 根据文件名扩展名判断媒体类型，未知类型归为 MediaFile 兜底。
func ClassifyMedia(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch {
	case imageExts[ext]:
		return MediaImage
	case videoExts[ext]:
		return MediaVideo
	case voiceExts[ext]:
		return MediaVoice
	default:
		return MediaFile
	}
}
