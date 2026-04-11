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

type ListEmojisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListEmojisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEmojisLogic {
	return &ListEmojisLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListEmojisLogic) ListEmojis(req *types.ListEmojisReq) (*types.ListEmojisResp, error) {
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
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	m := models.NewUserEmojiModel()
	list, total, err := m.List(l.ctx, userId, page, pageSize)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "查询失败")
	}

	items := make([]types.EmojiItem, 0, len(list))
	for _, e := range list {
		items = append(items, types.EmojiItem{
			Id:        e.Id,
			Url:       e.Url,
			Name:      e.Name,
			Thumbnail: e.Thumbnail,
			Width:     e.Width,
			Height:    e.Height,
			Size:      e.Size,
			FileType:  e.FileType,
			CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.ListEmojisResp{List: items, Total: total}, nil
}
