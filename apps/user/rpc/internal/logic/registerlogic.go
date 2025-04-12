package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/encrypt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrPhoneIsRegister = errors.New("手机号已经注册过")
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// 1. 验证用户是否注册，根据手机号码验证
	userEntity, err := l.svcCtx.UserModels.FindOneByPhone(l.ctx, in.Phone)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	fmt.Println("in:", in)

	if userEntity != nil {
		return nil, ErrPhoneIsRegister
	}

	if in.Nickname == "" {
		in.Nickname = in.Phone
	}

	// 定义用户数据
	userEntity = &models.Users{
		Avatar:    in.Avatar,
		Nickname:  in.Nickname,
		Phone:     in.Phone,
		Sex:       int(in.Sex),
		LastLogin: time.Now(),
		Status:    1,
		Type:      1,
	}

	if len(in.Password) > 0 {
		genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
		if err != nil {
			return nil, err
		}
		userEntity.Password = string(genPassword)
	}

	_, err = l.svcCtx.UserModels.Insert(l.ctx, userEntity)
	if err != nil {
		return nil, err
	}

	// 生成token
	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(l.svcCtx.Config.Jwt.AccessSecret, now, l.svcCtx.Config.Jwt.AccessExpire,
		strconv.Itoa(int(userEntity.Id)))
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
