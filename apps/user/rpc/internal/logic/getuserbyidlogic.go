package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByIdLogic {
	return &GetUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByIdLogic) GetUserById(in *user.GetUserByIdRequest) (*user.GetUserByIdResponse, error) {
	if in.Id == "" {
		return nil, errors.New(xerr.ErrBadRequest, "参数错误")
	}

	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	userEntiy, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		logger.Logger.Error("GetUserById.FindOne: find user by id failed", zap.Any("id", in.Id), zap.Error(err))
		return nil, errors.New(xerr.ErrInternalServer, "获取用户失败")
	}

	if userEntiy == nil {
		return nil, errors.New(xerr.ErrNotFound, "用户不存在")
	}

	// 使用统一的转换函数
	resp := ToUserEntity(userEntiy)

	return &user.GetUserByIdResponse{
		User: resp,
	}, nil
}
