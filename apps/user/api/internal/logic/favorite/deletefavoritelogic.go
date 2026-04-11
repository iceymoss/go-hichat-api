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

type DeleteFavoriteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFavoriteLogic {
	return &DeleteFavoriteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFavoriteLogic) DeleteFavorite(req *types.DeleteFavoriteReq) (*types.DeleteFavoriteResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	favModel := models.NewFavoriteModel()
	if err := favModel.Delete(l.ctx, req.Id, userId); err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "删除失败")
	}

	return &types.DeleteFavoriteResp{}, nil
}
