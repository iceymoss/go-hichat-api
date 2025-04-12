package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/encrypt"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"time"
)

var (
	ErrPhoneNotRegister = xerr.New(xerr.SERVER_COMMON_ERROR, "手机号没有注册")
	ErrUserPwdError     = xerr.New(xerr.SERVER_COMMON_ERROR, "密码不正确")
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 验证用户是否注册，根据手机号码验证
	userEntity, err := l.svcCtx.UserModels.FindOneByPhone(l.ctx, in.Phone)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, errors.WithStack(ErrPhoneNotRegister)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find user by phone err %v , req %v", err, in.Phone)
	}

	// 密码验证
	if !encrypt.ValidatePasswordHash(in.Password, userEntity.Password) {
		return nil, errors.WithStack(ErrUserPwdError)
	}

	// 生成token
	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(l.svcCtx.Config.Jwt.AccessSecret, now, l.svcCtx.Config.Jwt.AccessExpire,
		strconv.Itoa(int(userEntity.Id)))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "ctxdata get jwt token err %v", err)
	}

	userEntity.LastLogin = time.Now()

	//更新登录时间
	l.svcCtx.UserModels.Update(l.ctx, userEntity)

	return &user.LoginResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
