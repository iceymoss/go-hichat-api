package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/encrypt"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type ResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetPasswordLogic) ResetPassword(in *user.ResetPassWordReq) (*user.ResetPassWordResp, error) {
	if in.Id == "" {
		return nil, xerr.New(xerr.ErrBadRequest, "用户id不能为空")
	}

	if in.Password == "" {
		return nil, xerr.New(xerr.ErrBadRequest, "密码不能为空")
	}

	userId, err := strconv.Atoi(in.Id)
	if err != nil {
		logger.Error("atoi error", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}
	// get user
	userEntity, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(userId))
	if err != nil {
		return nil, libErr.New(10004, "用户不存在")
	}

	if userEntity.Status != 1 {
		return nil, libErr.New(10005, "用户已禁用")
	}

	genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
	if err != nil {
		return nil, err
	}
	userEntity.Password = string(genPassword)

	now := time.Now()
	userObj := models.Users{
		Id:        uint64(userId),
		Password:  string(genPassword),
		UpdatedAt: now,
	}

	err = l.svcCtx.UserModels.UpdateByID(l.ctx, &userObj)
	if err != nil {
		logger.Error("update user error", zap.Any("user", userObj), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "更新用户信息失败")
	}

	return &user.ResetPassWordResp{}, nil
}
