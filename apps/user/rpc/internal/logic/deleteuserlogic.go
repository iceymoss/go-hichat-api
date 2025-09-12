package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *user.DeleteUserReq) (*user.DeleteUserResp, error) {
	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	// 获取用户
	userEntity, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		return nil, errors.New(10004, "用户不存在")
	}

	if userEntity.Status != 1 {
		return nil, errors.New(10005, "用户已禁用")
	}

	// 删除用户
	err = l.svcCtx.UserModels.Delete(l.ctx, uint64(userId))
	if err != nil {
		logger.Error("DeleteUser.Delete: delete user failed", zap.Any("userId", userId), zap.Error(err))
		return nil, err
	}

	return &user.DeleteUserResp{}, nil
}
