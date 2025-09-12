package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyEmailLogic {
	return &VerifyEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyEmailLogic) VerifyEmail(in *user.VerifyEmailRequest) (*user.VerifyEmailResponse, error) {
	// todo: add your logic here and delete this line

	return &user.VerifyEmailResponse{}, nil
}
