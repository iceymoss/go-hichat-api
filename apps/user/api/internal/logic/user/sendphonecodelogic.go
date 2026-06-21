package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type SendPhoneCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendPhoneCodeLogic 发送手机验证码
func NewSendPhoneCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendPhoneCodeLogic {
	return &SendPhoneCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendPhoneCodeLogic) SendPhoneCode(req *types.SendPhoneCodeReq) (resp *types.SendPhoneCodeResp, err error) {
	if req.Phone == "" {
		return nil, errors.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	rpcResp, err := l.svcCtx.User.SendPhoneCode(l.ctx, &user.SendPhoneCodeRequest{Phone: req.Phone})
	if err != nil {
		logx.Error("发送验证码失败", zap.Any("phone", req.Phone), zap.Error(err))
		return nil, err
	}

	resp = &types.SendPhoneCodeResp{}
	// 测试/演示模式下 rpc 把验证码放在 Message 回传，这里透传给前端自动填入输入框。
	if isDemoSMSMode() {
		resp.Code = rpcResp.Message
	}
	return resp, nil
}

// isDemoSMSMode 未配置真实短信服务商（provider 为空或 console）时视为测试/演示模式。
func isDemoSMSMode() bool {
	c := pkcCfg.ServiceConf
	if c == nil || c.Verification == nil {
		return true
	}
	p := c.Verification.SMS.Provider
	return p == "" || p == "console"
}

