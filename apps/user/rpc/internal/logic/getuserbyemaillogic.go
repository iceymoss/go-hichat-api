package logic

import (
	"context"
	"database/sql"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUserByEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByEmailLogic {
	return &GetUserByEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByEmailLogic) GetUserByEmail(in *user.GetUserByEmailRequest) (*user.GetUserByEmailResponse, error) {
	userEntiy, err := l.svcCtx.UserModels.FindOneByEmail(l.ctx, sql.NullString{
		String: in.Email,
		Valid:  true,
	})
	if err != nil {
		logger.Logger.Error("GetUserByEmail.FindOneByEmail: find user by email failed", zap.Any("email", in.Email), zap.Error(err))
		return nil, errors.New(xerr.ErrInternalServer, "获取用户失败")
	}

	if userEntiy == nil {
		return nil, errors.New(xerr.ErrNotFound, "用户不存在")
	}

	// 使用统一的转换函数
	resp := ToUserEntity(userEntiy)

	return &user.GetUserByEmailResponse{
		User: resp,
	}, nil
}
