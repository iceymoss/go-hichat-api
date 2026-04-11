package favorite

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFavoriteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFavoriteLogic {
	return &AddFavoriteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFavoriteLogic) AddFavorite(req *types.AddFavoriteReq) (*types.AddFavoriteResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	if req.Content == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "收藏内容不能为空")
	}

	validTypes := map[string]bool{"text": true, "image": true, "link": true, "file": true, "video": true, "location": true, "note": true}
	if !validTypes[req.Type] {
		req.Type = "text"
	}

	fav := &models.Favorite{
		UserId:   userId,
		Type:     req.Type,
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		SourceId: req.SourceId,
		Tags:     req.Tags,
	}

	favModel := models.NewFavoriteModel()
	if err := favModel.Create(l.ctx, fav); err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "收藏失败")
	}

	return &types.AddFavoriteResp{Id: fav.Id}, nil
}
