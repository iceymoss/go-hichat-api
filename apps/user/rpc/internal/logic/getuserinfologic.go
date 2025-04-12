package logic

import (
	"context"
	"errors"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

var ErrUserNotFound = errors.New("这个用户没有")

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	// todo: add your logic here and delete this line

	uidInt, err := strconv.Atoi(in.Id)
	if err != nil {
		return nil, err
	}
	userEntiy, err := l.svcCtx.UserModels.FindOne(l.ctx, uint64(uidInt))
	if err != nil {
		if err == models.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
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
		Introduction: "",
		Type:         int32(userEntiy.Type),
		State:        1,
	}

	return &user.GetUserInfoResp{
		User: &resp,
	}, nil
}

func Copier(userEntiyList []*models.Users) []*user.UserEntity {
	list := make([]*user.UserEntity, 0, len(userEntiyList))
	for _, v := range userEntiyList {
		userEntiy := v
		list = append(list, &user.UserEntity{
			Id:           strconv.Itoa(int(userEntiy.Id)),
			Avatar:       userEntiy.Avatar,
			Nickname:     userEntiy.Nickname,
			Phone:        userEntiy.Phone,
			Email:        userEntiy.Email,
			Status:       int32(userEntiy.Status),
			LastLogin:    userEntiy.LastLogin.Unix(),
			Sex:          int32(userEntiy.Sex),
			Introduction: "",
			Type:         int32(userEntiy.Type),
			State:        1,
		})
	}
	return list
}
