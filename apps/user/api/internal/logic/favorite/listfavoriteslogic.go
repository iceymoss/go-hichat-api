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

type ListFavoritesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFavoritesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFavoritesLogic {
	return &ListFavoritesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFavoritesLogic) ListFavorites(req *types.ListFavoritesReq) (*types.ListFavoritesResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	favModel := models.NewFavoriteModel()
	list, total, err := favModel.List(l.ctx, userId, req.Type, req.Keyword, page, pageSize)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "查询失败")
	}

	items := make([]types.FavoriteItem, 0, len(list))
	for _, f := range list {
		items = append(items, types.FavoriteItem{
			Id:        f.Id,
			Type:      f.Type,
			Title:     f.Title,
			Content:   f.Content,
			Source:    f.Source,
			SourceId:  f.SourceId,
			Tags:      f.Tags,
			CreatedAt: f.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.ListFavoritesResp{List: items, Total: total}, nil
}
