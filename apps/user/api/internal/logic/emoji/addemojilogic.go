package emoji

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddEmojiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddEmojiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddEmojiLogic {
	return &AddEmojiLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AddEmojiLogic) AddEmoji(req *types.AddEmojiReq) (*types.AddEmojiResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}
	if req.Url == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "表情URL不能为空")
	}

	m := models.NewUserEmojiModel()

	// Check duplicate
	exists, _ := m.ExistsByUrl(l.ctx, userId, req.Url)
	if exists {
		return nil, libErr.New(xerr.ErrBadRequest, "该表情已收藏")
	}

	fileType := req.FileType
	if fileType == "" {
		fileType = "image"
	}

	e := &models.UserEmoji{
		UserId:    userId,
		Url:       req.Url,
		Name:      req.Name,
		Thumbnail: req.Thumbnail,
		Width:     req.Width,
		Height:    req.Height,
		Size:      req.Size,
		FileType:  fileType,
	}

	if err := m.Create(l.ctx, e); err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "添加失败")
	}

	return &types.AddEmojiResp{Id: e.Id}, nil
}
