package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	if in.Id == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "用户id不能为空")
	}

	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	// get user
	userEntity, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, libErr.New(xerr.ErrNotFound, "用户不存在")
		}
		logger.Error("user not found", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	if userEntity.Status != 1 {
		return nil, libErr.New(10005, "用户已禁用")
	}

	now := time.Now()
	userObj := models.Users{
		Id:        uint64(userId),
		UpdatedAt: now,
	}

	if in.Name != "" {
		userObj.Nickname = in.Name
	}

	if in.Phone != "" {
		userObj.Phone = in.Phone
	}

	if in.Email != "" {
		userObj.Email = in.Email
	}

	if in.Sex != 0 {
		userObj.Sex = int(in.Sex)
	}

	if in.Avatar != "" {
		userObj.Avatar = in.Avatar
	}

	if in.Introduction != "" {
		userObj.Introduction = in.Introduction
	}

	if in.Type != "" {
		userType, err := strconv.ParseInt(in.Type, 10, 64)
		if err == nil {
			userObj.Type = uint64(userType)
		}
	}

	err = l.svcCtx.UserModels.UpdateByID(l.ctx, &userObj)
	if err != nil {
		logger.Error("UpdateUser.UpdateByID: update user failed", zap.Any("userId", userId), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "更新用户信息失败")
	}

	return &user.UpdateUserResp{}, nil
}
