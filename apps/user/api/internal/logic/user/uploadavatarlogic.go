package user

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/storage"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(r *types.UploadAvatarReq, fileHeader *multipart.FileHeader) (resp *types.UploadAvatarResp, err error) {
	// 验证文件
	if fileHeader == nil {
		return nil, libErr.New(xerr.ErrBadRequest, "请选择要上传的文件")
	}

	// 验证文件类型（只允许图片）
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return nil, libErr.New(xerr.ErrBadRequest, "只支持上传图片文件（jpg, jpeg, png, gif, webp）")
	}

	// 验证文件大小（限制5MB）
	if fileHeader.Size > 5*1024*1024 {
		return nil, libErr.New(xerr.ErrBadRequest, "图片大小不能超过5MB")
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "打开文件失败")
	}
	defer file.Close()

	// 使用存储接口上传文件
	var fileStorage storage.FileStorage
	cfg := config.ServiceConf

	// 设置默认值
	basePath := cfg.Upload.BasePath
	if basePath == "" {
		basePath = "./temp"
	}
	baseURL := cfg.Upload.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8887/static"
	}

	// 暂时使用本地存储，后续可以切换到OSS
	fileStorage = storage.NewLocalStorage(basePath, baseURL)

	// 上传文件
	url, err := fileStorage.UploadFile(l.ctx, file, fileHeader.Filename, "avatar")
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "上传文件失败: "+err.Error())
	}

	return &types.UploadAvatarResp{
		Url: url,
	}, nil
}
