package logic

import (
	"context"
	"mime/multipart"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传富媒体文件(图片/视频/文件/语音)
func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadLogic) Upload(file multipart.File, header *multipart.FileHeader) (resp *types.UploadResp, err error) {
	if header == nil {
		return nil, libErr.New(xerr.ErrBadRequest, "请选择要上传的文件")
	}

	url, kind, err := uploadMedia(l.ctx, l.svcCtx.FileStorage, header.Filename, header.Size, file)
	if err != nil {
		return nil, err
	}

	return &types.UploadResp{
		Url:      url,
		Name:     header.Filename,
		Size:     header.Size,
		FileType: kind,
	}, nil
}
