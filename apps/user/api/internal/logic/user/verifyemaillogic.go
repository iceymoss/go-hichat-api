package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type VerifyEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewVerifyEmailLogic 验证邮箱
func NewVerifyEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyEmailLogic {
	return &VerifyEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyEmailLogic) VerifyEmail(req *types.VerifyEmailReq) (resp *types.VerifyEmailResp, err error) {
	if req.Email == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "邮箱不能为空")
	}

	if req.Code == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "验证码不能为空")
	}

	_, err = l.svcCtx.User.VerifyEmail(l.ctx, &user.VerifyEmailRequest{Email: req.Email, Code: req.Code})
	if err != nil {
		logx.Error("验证邮箱失败", zap.Any("email", req.Email), zap.Error(err))
		return nil, err
	}

	return
}
