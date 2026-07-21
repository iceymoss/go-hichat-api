package trend

import (
	"context"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadTrendMediaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传动态媒体
func NewUploadTrendMediaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadTrendMediaLogic {
	return &UploadTrendMediaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadTrendMediaLogic) UploadTrendMedia(file multipart.File, header *multipart.FileHeader) (resp *types.UploadTrendMediaResponse, err error) {
	if header == nil {
		return nil, errors.New("请选择要上传的文件")
	}
	if header.Size == 0 {
		return nil, errors.New("上传文件不能为空")
	}

	cfg := defaultTrendPublishConfig()
	fileType := storage.ClassifyMedia(header.Filename)
	if fileType != storage.MediaImage && fileType != storage.MediaVideo {
		return nil, errors.New("仅支持上传图片或视频")
	}

	if fileType == storage.MediaImage && header.Size > int64(cfg.MaxImageSizeMB)*1024*1024 {
		return nil, errors.New("图片大小超过限制")
	}
	if fileType == storage.MediaVideo && header.Size > int64(cfg.MaxVideoSizeMB)*1024*1024 {
		return nil, errors.New("视频大小超过限制")
	}

	contentType, width, height, err := inspectMedia(file, header.Filename, fileType)
	if err != nil {
		return nil, err
	}

	folder := "trend/image"
	if fileType == storage.MediaVideo {
		folder = "trend/video"
	}

	basePath := "./temp"
	baseURL := "http://localhost:8887/static"
	if config.ServiceConf != nil {
		if config.ServiceConf.Upload.BasePath != "" {
			basePath = config.ServiceConf.Upload.BasePath
		}
		if config.ServiceConf.Upload.BaseURL != "" {
			baseURL = config.ServiceConf.Upload.BaseURL
		}
	}

	url, err := storage.NewLocalStorage(basePath, baseURL).UploadFile(l.ctx, file, header.Filename, folder)
	if err != nil {
		return nil, err
	}

	return &types.UploadTrendMediaResponse{
		Url:         url,
		Name:        filepath.Base(header.Filename),
		Size:        header.Size,
		FileType:    fileType,
		ContentType: contentType,
		Width:       width,
		Height:      height,
	}, nil
}

func inspectMedia(file multipart.File, filename string, fileType string) (string, int, int, error) {
	seeker, ok := file.(io.Seeker)
	if !ok {
		return "", 0, 0, errors.New("上传文件不支持安全检查")
	}
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", 0, 0, err
	}
	contentType := http.DetectContentType(buf[:n])
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return "", 0, 0, err
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	if fileType == storage.MediaImage {
		if !isAllowedImage(ext, contentType) {
			return "", 0, 0, errors.New("图片类型不支持")
		}
		if ext == "webp" {
			return contentType, 0, 0, nil
		}
		conf, _, err := image.DecodeConfig(file)
		if err != nil {
			return "", 0, 0, errors.New("图片文件无效")
		}
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", 0, 0, err
		}
		if conf.Width <= 0 || conf.Height <= 0 || conf.Width*conf.Height > 10000*10000 {
			return "", 0, 0, errors.New("图片尺寸异常")
		}
		return contentType, conf.Width, conf.Height, nil
	}

	if !isAllowedVideo(ext, contentType) {
		return "", 0, 0, errors.New("视频类型不支持")
	}
	return contentType, 0, 0, nil
}

func isAllowedImage(ext string, contentType string) bool {
	switch ext {
	case "jpg", "jpeg":
		return contentType == "image/jpeg"
	case "png":
		return contentType == "image/png"
	case "gif":
		return contentType == "image/gif"
	case "webp":
		return contentType == "image/webp" || contentType == "application/octet-stream"
	default:
		return false
	}
}

func isAllowedVideo(ext string, contentType string) bool {
	if ext != "mp4" && ext != "mov" && ext != "webm" {
		return false
	}
	return strings.HasPrefix(contentType, "video/") || contentType == "application/octet-stream"
}
