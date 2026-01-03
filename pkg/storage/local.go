package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// LocalStorage 本地文件存储实现
type LocalStorage struct {
	basePath string // 基础存储路径，如 /temp
	baseURL  string // 基础访问URL，如 http://localhost:8887/static
}

// NewLocalStorage 创建本地文件存储实例
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// 确保基础目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		logger.Error("创建存储目录失败", zap.String("path", basePath), zap.Error(err))
	}
	
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// UploadFile 上传文件到本地存储
func (s *LocalStorage) UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error) {
	// 生成唯一文件名：时间戳_原始文件名
	ext := filepath.Ext(filename)
	name := filepath.Base(filename)
	name = name[:len(name)-len(ext)]
	
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%d_%s%s", timestamp, name, ext)
	
	// 构建完整路径
	folderPath := filepath.Join(s.basePath, folder)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("创建文件夹失败: %w", err)
	}
	
	filePath := filepath.Join(folderPath, newFilename)
	
	// 创建文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()
	
	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // 如果复制失败，删除已创建的文件
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	
	// 返回访问URL
	relativePath := filepath.Join(folder, newFilename)
	return s.GetFileURL(relativePath), nil
}

// DeleteFile 删除文件
func (s *LocalStorage) DeleteFile(ctx context.Context, url string) error {
	// 从URL中提取相对路径
	// URL格式: http://localhost:8887/static/avatar/1234567890_filename.jpg
	// 需要提取: avatar/1234567890_filename.jpg
	
	// 简单实现：假设URL包含baseURL，提取后面的路径
	relativePath := url
	if len(s.baseURL) > 0 && len(url) > len(s.baseURL) {
		if url[:len(s.baseURL)] == s.baseURL {
			relativePath = url[len(s.baseURL):]
			// 移除开头的斜杠
			if len(relativePath) > 0 && relativePath[0] == '/' {
				relativePath = relativePath[1:]
			}
		}
	}
	
	filePath := filepath.Join(s.basePath, relativePath)
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，认为删除成功
	}
	
	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	
	return nil
}

// GetFileURL 获取文件的访问URL
func (s *LocalStorage) GetFileURL(path string) string {
	// 确保路径使用正斜杠（URL格式）
	urlPath := filepath.ToSlash(path)
	
	// 确保baseURL以斜杠结尾，path不以斜杠开头
	if len(s.baseURL) > 0 {
		baseURL := s.baseURL
		if baseURL[len(baseURL)-1] != '/' {
			baseURL += "/"
		}
		if len(urlPath) > 0 && urlPath[0] == '/' {
			urlPath = urlPath[1:]
		}
		return baseURL + urlPath
	}
	
	return "/" + urlPath
}

