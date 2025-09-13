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

type SendEmailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendEmailCodeLogic 发送邮件验证码
func NewSendEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailCodeLogic {
	return &SendEmailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendEmailCodeLogic) SendEmailCode(req *types.SendVerificationEmailReq) (resp *types.SendVerificationEmailResp, err error) {
	if req.Email == "" {
		return nil, errors.New(xerr.ErrBadRequest, "邮箱不能为空")
	}

	_, err = l.svcCtx.User.SendVerificationEmail(l.ctx, &user.SendVerificationEmailRequest{Email: req.Email})
	if err != nil {
		logx.Error("发送验证码失败", zap.Any("email", req.Email), zap.Error(err))
		return nil, err
	}

	return &types.SendVerificationEmailResp{}, nil
}
