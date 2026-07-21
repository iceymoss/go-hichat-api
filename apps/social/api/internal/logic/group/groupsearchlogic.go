package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 搜索群（群号精确+群名模糊，分页）
func NewGroupSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSearchLogic {
	return &GroupSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupSearchLogic) GroupSearch(req *types.GroupSearchReq) (resp *types.GroupSearchResp, err error) {
	rpcResp, err := l.svcCtx.Social.GroupSearch(l.ctx, &social.GroupSearchReq{
		Keyword: req.Keyword,
		Page:    req.Page,
		Size:    req.Size,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.GroupSearchItem, 0, len(rpcResp.List))
	for _, g := range rpcResp.List {
		list = append(list, types.GroupSearchItem{
			Id:          g.Id,
			Name:        g.Name,
			Icon:        g.Icon,
			Description: g.Description,
			CreatorUid:  g.CreatorUid,
			MemberCount: g.MemberCount,
		})
	}

	return &types.GroupSearchResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}
