package user

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"

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
	_, err = l.svcCtx.User.UpdateUser(l.ctx, &user.UpdateUserReq{
		Id:           uid,
		Name:         req.Name,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Type:         req.Type,
		Email:        req.Email,
		Sex:          int32(req.Sex),
		Introduction: req.Introduction,
		Password:     req.Password,
	})
	if err != nil {
		return nil, err
	}

	return
}
