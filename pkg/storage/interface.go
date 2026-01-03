package storage

import (
	"context"
	"io"
)

// FileStorage 文件存储接口
// 通过实现此接口，可以轻松切换不同的存储服务（本地存储、OSS、S3等）
type FileStorage interface {
	// UploadFile 上传文件
	// file: 文件内容
	// filename: 原始文件名
	// folder: 存储文件夹（如 "avatar", "image" 等）
	// 返回: 文件的访问URL
	UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error)
	
	// DeleteFile 删除文件
	// url: 文件的访问URL
	DeleteFile(ctx context.Context, url string) error
	
	// GetFileURL 获取文件的访问URL
	// path: 文件的存储路径
	GetFileURL(path string) string
}

