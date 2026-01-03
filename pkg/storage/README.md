# 文件存储接口

## 概述

文件存储接口设计用于支持多种存储后端，目前实现了本地存储，后续可以轻松切换到OSS、S3等云存储服务。

## 接口定义

```go
type FileStorage interface {
    // UploadFile 上传文件
    UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error)
    
    // DeleteFile 删除文件
    DeleteFile(ctx context.Context, url string) error
    
    // GetFileURL 获取文件的访问URL
    GetFileURL(path string) string
}
```

## 当前实现

### 本地存储 (LocalStorage)

- **存储路径**: `./temp/{folder}/`
- **文件命名**: `{timestamp}_{原文件名}`
- **访问URL**: `{baseURL}/{folder}/{filename}`

### 配置

在 `config/config-local.yaml` 中配置：

```yaml
upload:
    basePath: './temp'  # 本地存储路径
    baseURL: 'http://localhost:8887/static'  # 访问URL
```

## 使用示例

```go
// 创建存储实例
fileStorage := storage.NewLocalStorage("./temp", "http://localhost:8887/static")

// 上传文件
url, err := fileStorage.UploadFile(ctx, file, "avatar.jpg", "avatar")
if err != nil {
    return err
}
// url: http://localhost:8887/static/avatar/1234567890_avatar.jpg

// 删除文件
err = fileStorage.DeleteFile(ctx, url)
```

## 切换到OSS存储

1. 创建 `pkg/storage/oss.go` 实现 `FileStorage` 接口
2. 在需要的地方替换存储实例：

```go
// 从
fileStorage := storage.NewLocalStorage(basePath, baseURL)

// 改为
fileStorage := storage.NewOSSStorage(ossConfig)
```

## 注意事项

1. **目录权限**: 确保 `temp` 目录有写入权限
2. **静态文件服务**: 需要配置Web服务器提供静态文件访问（如Nginx或Go的http.FileServer）
3. **文件清理**: 建议定期清理临时文件或实现文件生命周期管理

