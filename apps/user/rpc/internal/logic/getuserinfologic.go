package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
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
	resp := ToUserEntity(userEntiy)

	return &user.GetUserInfoResp{
		User: resp,
	}, nil
}

func Copier(userEntiyList []*models.Users) []*user.UserEntity {
	list := make([]*user.UserEntity, 0, len(userEntiyList))
	for _, v := range userEntiyList {
		list = append(list, ToUserEntity(v))
	}
	return list
}

// getStringValue 从 sql.NullString 获取字符串值
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// ToUserEntity 将 models.Users 转换为 user.UserEntity（统一转换函数，导出供其他包使用）
func ToUserEntity(userEntiy *models.Users) *user.UserEntity {
	return &user.UserEntity{
		Id:           strconv.Itoa(int(userEntiy.Id)),
		Avatar:       userEntiy.Avatar,
		Nickname:     userEntiy.Nickname,
		Phone:        userEntiy.Phone,
		Email:        getStringValue(userEntiy.Email), // 使用 getStringValue 处理 sql.NullString
		Status:       int32(userEntiy.Status),
		LastLogin:    userEntiy.LastLogin.Unix(),
		Sex:          int32(userEntiy.Sex),
		Introduction: userEntiy.Introduction,
		Type:         int32(userEntiy.Type),
		State:        1,
		Region:       getStringValue(userEntiy.Region),
		Occupation:   getStringValue(userEntiy.Occupation),
		Tags:         getStringValue(userEntiy.Tags),
		MomentsCover: userEntiy.MomentsCover,
	}
}
