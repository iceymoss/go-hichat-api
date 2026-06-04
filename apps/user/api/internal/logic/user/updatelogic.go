package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改用户信息
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogic) Update(req *types.UpdateUserReq) (resp *types.UpdateUserResp, err error) {
	uid := ctxdata.GetUId(l.ctx)

	// 调用 RPC 更新用户信息
	// 注意：Phone 和 Email 不能通过此 API 更新
	// Phone 不可修改，Email 使用独立的绑定 API
	_, err = l.svcCtx.User.UpdateUser(l.ctx, &user.UpdateUserReq{
		Id:           uid,
		Name:         req.Name,
		Avatar:       req.Avatar,
		Type:         req.Type,
		Sex:          int32(req.Sex),
		Introduction: req.Introduction,
		Password:     req.Password,
		Region:       req.Region,
		Occupation:   req.Occupation,
		Tags:         req.Tags,
		MomentsCover: req.MomentsCover,
		// Phone 和 Email 不传递，不允许通过此 API 更新
	})
	if err != nil {
		return nil, err
	}

	return
}
