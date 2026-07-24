package logic

import (
	"context"
	stdErrors "errors"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, libErr.New(xerr.ErrBadRequest, "参数错误")
	}

	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	userEntiy, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if stdErrors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		logger.Logger.Error("GetUserById.FindOne: find user by id failed", zap.Any("id", in.Id), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "获取用户失败")
	}

	if userEntiy == nil {
		return nil, libErr.New(xerr.ErrNotFound, "用户不存在")
	}

	// 使用统一的转换函数
	resp := ToUserEntity(userEntiy)

	return &user.GetUserByIdResponse{
		User: resp,
	}, nil
}
