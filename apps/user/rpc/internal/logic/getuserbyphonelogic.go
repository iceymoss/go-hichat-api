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

type GetUserByPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByPhoneLogic {
	return &GetUserByPhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByPhoneLogic) GetUserByPhone(in *user.GetUserByPhoneRequest) (*user.GetUserByPhoneResponse, error) {
	if in.Phone == "" {
		return nil, errors.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	userEntiy, err := l.svcCtx.UserModels.FindOneByPhone(l.ctx, in.Phone)
	if err != nil {
		logger.Logger.Error("GetUserByPhone.FindOneByPhone: find user by phone failed", zap.Any("email", in.Phone), zap.Error(err))
		return nil, errors.New(xerr.ErrInternalServer, "获取用户失败")
	}

	if userEntiy == nil {
		return nil, errors.New(xerr.ErrNotFound, "用户不存在")
	}

	resp := user.UserEntity{
		Id:           strconv.Itoa(int(userEntiy.Id)),
		Avatar:       userEntiy.Avatar,
		Nickname:     userEntiy.Nickname,
		Phone:        userEntiy.Phone,
		Email:        userEntiy.Email,
		Status:       int32(userEntiy.Status),
		LastLogin:    userEntiy.LastLogin.Unix(),
		Sex:          int32(userEntiy.Sex),
		Introduction: userEntiy.Introduction,
		Type:         int32(userEntiy.Type),
		State:        int32(userEntiy.Status),
	}
	return &user.GetUserByPhoneResponse{
		User: &resp,
	}, nil
}
