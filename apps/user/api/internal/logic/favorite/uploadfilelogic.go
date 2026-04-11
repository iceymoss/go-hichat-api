package favorite

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/storage"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileResp struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"` // image/video/file
}

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(file multipart.File, header *multipart.FileHeader) (*UploadFileResp, error) {
	if header == nil {
		return nil, libErr.New(xerr.ErrBadRequest, "请选择要上传的文件")
	}

	if header.Size > 50*1024*1024 {
		return nil, libErr.New(xerr.ErrBadRequest, "文件大小不能超过50MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Determine file type and storage folder
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}
	videoExts := map[string]bool{".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true}

	fileType := "file"
	folder := "favorite/file"
	if imageExts[ext] {
		fileType = "image"
		folder = "favorite/image"
	} else if videoExts[ext] {
		fileType = "video"
		folder = "favorite/video"
	}

	cfg := config.ServiceConf
	basePath := cfg.Upload.BasePath
	if basePath == "" {
		basePath = "./temp"
	}
	baseURL := cfg.Upload.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8887/static"
	}

	fileStorage := storage.NewLocalStorage(basePath, baseURL)
	url, err := fileStorage.UploadFile(l.ctx, file, header.Filename, folder)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "上传文件失败: "+err.Error())
	}

	return &UploadFileResp{
		Url:      url,
		Name:     header.Filename,
		Size:     header.Size,
		FileType: fileType,
	}, nil
}
