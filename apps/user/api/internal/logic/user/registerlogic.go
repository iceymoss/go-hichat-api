package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	rpcRegisterResp, err := l.svcCtx.User.Register(l.ctx, &user.RegisterReq{
		Phone:     req.Phone,
		Nickname:  req.Nickname,
		Password:  req.Password,
		PhoneCode: req.PhoneCode,
		// Avatar、Sex 和 Email 不再从注册请求中获取，使用默认值
		// Email 需要在个人资料中绑定，并需要验证码验证
		Avatar:    "",
		Sex:       0,
		Email:     "",
	})
	if err != nil {
		return nil, err
	}

	res := types.RegisterResp{
		Token:  rpcRegisterResp.Token,
		Expire: rpcRegisterResp.Expire,
	}

	return &res, nil
}
