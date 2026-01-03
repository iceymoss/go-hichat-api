package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyPhoneCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewVerifyPhoneCodeLogic 验证手机验证码
func NewVerifyPhoneCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyPhoneCodeLogic {
	return &VerifyPhoneCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyPhoneCodeLogic) VerifyPhoneCode(req *types.VerifyPhoneCodeReq) (resp *types.VerifyPhoneCodeResp, err error) {
	if req.Phone == "" {
		return nil, errors.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	if req.Code == "" {
		return nil, errors.New(xerr.ErrBadRequest, "验证码不能为空")
	}

	_, err = l.svcCtx.User.VerifyPhoneCode(l.ctx, &user.VerifyPhoneCodeRequest{
		Phone: req.Phone,
		Code:  req.Code,
	})
	if err != nil {
		return nil, err
	}

	return &types.VerifyPhoneCodeResp{}, nil
}

