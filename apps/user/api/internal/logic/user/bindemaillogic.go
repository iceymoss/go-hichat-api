package user

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type BindEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewBindEmailLogic 绑定/更新邮箱
func NewBindEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindEmailLogic {
	return &BindEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindEmailLogic) BindEmail(req *types.BindEmailReq) (resp *types.BindEmailResp, err error) {
	uid := ctxdata.GetUId(l.ctx)

	// 验证邮箱格式
	if req.Email == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "邮箱不能为空")
	}

	// 验证验证码
	if req.Code == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "验证码不能为空")
	}
	
	// 验证邮箱验证码（直接使用 Redis 检查，不删除验证码，让 RPC 层验证后再删除）
	rdb := db.GetRedisConn()
	key := verification.GetRedisKey(verification.CodeTypeEmail, req.Email)

	// 先检查验证码是否存在且正确（不删除，让 RPC 层验证后再删除）
	storedCode, err := rdb.Get(l.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, libErr.New(xerr.ErrInvalidInput, "邮箱验证码错误或已过期")
		}
		logx.Error("验证邮箱验证码失败", zap.Any("email", req.Email), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
	}

	if storedCode != req.Code {
		return nil, libErr.New(xerr.ErrInvalidInput, "邮箱验证码错误或已过期")
	}

	// 验证通过，更新用户邮箱（RPC 层会再次验证并删除验证码）
	_, err = l.svcCtx.User.UpdateUser(l.ctx, &user.UpdateUserReq{
		Id:        uid,
		Email:     req.Email,
		EmailCode: req.Code, // 传递验证码给 RPC 层（RPC 层会验证并删除）
	})
	if err != nil {
		return nil, err
	}

	return &types.BindEmailResp{}, nil
}
