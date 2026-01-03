package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/encrypt"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("获取用户失败", zap.Any("phone", in.Phone), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "获取用户失败")
	}

	if userEntity != nil && userEntity.Id != 0 {
		return nil, ErrPhoneIsRegister
	}

	// 2. 验证手机验证码（必填）
	if in.PhoneCode == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "手机验证码不能为空")
	}
	
	// 验证验证码长度（必须是6位）
	if len(in.PhoneCode) != 6 {
		return nil, libErr.New(xerr.ErrBadRequest, "手机验证码必须是6位数")
	}
	
	// 使用工厂模式获取验证码发送器
	codeSender, err := verification.GetCodeSender(verification.CodeTypeSMS)
	if err != nil {
		logger.Error("获取短信验证码发送器失败", zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
	}
	
	rdb := db.GetRedisConn()
	
	// 构建Redis key
	key := verification.GetRedisKey(verification.CodeTypeSMS, in.Phone)
	
	pass, err := codeSender.VerifyCode(l.ctx, rdb, key, in.PhoneCode)
	if err != nil {
		logger.Error("验证手机验证码失败", zap.Any("phone", in.Phone), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
	}
	
	if !pass {
		return nil, libErr.New(xerr.ErrInvalidInput, "手机验证码错误或已过期")
	}

	if in.Nickname == "" {
		in.Nickname = in.Phone
	}

	// 定义用户数据
	// 注意：Email 字段不设置（保持零值），但为了避免唯一索引冲突，
	// 我们需要在插入时使用 NULL 而不是空字符串
	userEntity = &models.Users{
		Avatar:    "", // 注册时不设置头像，用户后续可在个人主页编辑
		Nickname:  in.Nickname,
		Phone:     in.Phone,
		Sex:       0, // 注册时不设置性别，默认为0（未知），用户后续可在个人主页编辑
		LastLogin: time.Now(),
		// Email 字段不设置，使用 GORM 的 Omit 来忽略它，让数据库使用 NULL
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

	err = l.svcCtx.UserModels.Create(l.ctx, userEntity)
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
