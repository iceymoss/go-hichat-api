package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type SendPhoneCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendPhoneCodeLogic 发送手机验证码
func NewSendPhoneCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendPhoneCodeLogic {
	return &SendPhoneCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendPhoneCodeLogic) SendPhoneCode(req *types.SendPhoneCodeReq) (resp *types.SendPhoneCodeResp, err error) {
	if req.Phone == "" {
		return nil, errors.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	_, err = l.svcCtx.User.SendPhoneCode(l.ctx, &user.SendPhoneCodeRequest{Phone: req.Phone})
	if err != nil {
		logx.Error("发送验证码失败", zap.Any("phone", req.Phone), zap.Error(err))
		return nil, err
	}

	return &types.SendPhoneCodeResp{}, nil
}

