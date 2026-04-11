package emoji

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

type UploadEmojiResp struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"`
}

type UploadEmojiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadEmojiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadEmojiLogic {
	return &UploadEmojiLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UploadEmojiLogic) UploadEmoji(file multipart.File, header *multipart.FileHeader) (*UploadEmojiResp, error) {
	if header == nil {
		return nil, libErr.New(xerr.ErrBadRequest, "请选择要上传的表情")
	}
	if header.Size > 10*1024*1024 {
		return nil, libErr.New(xerr.ErrBadRequest, "表情文件大小不能超过10MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}
	if !allowed[ext] {
		return nil, libErr.New(xerr.ErrBadRequest, "只支持图片格式(jpg/png/gif/webp)")
	}

	fileType := "image"
	if ext == ".gif" {
		fileType = "gif"
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

	fs := storage.NewLocalStorage(basePath, baseURL)
	url, err := fs.UploadFile(l.ctx, file, header.Filename, "emoji")
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "上传失败: "+err.Error())
	}

	return &UploadEmojiResp{
		Url:      url,
		Name:     header.Filename,
		Size:     header.Size,
		FileType: fileType,
	}, nil
}
