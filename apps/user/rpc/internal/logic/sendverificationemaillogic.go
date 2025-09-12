package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendVerificationEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationEmailLogic {
	return &SendVerificationEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendVerificationEmailLogic) SendVerificationEmail(in *user.SendVerificationEmailRequest) (*user.SendVerificationEmailResponse, error) {
	// todo: add your logic here and delete this line

	return &user.SendVerificationEmailResponse{}, nil
}
