package settings

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettingsLogic {
	return &GetSettingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetSettingsLogic) GetSettings(req *types.GetSettingsReq) (*types.GetSettingsResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	m := models.NewUserSettingsModel()
	s, err := m.Get(l.ctx, userId)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "获取配置失败")
	}

	// Return default if not set
	if s == "" {
		s = "{}"
	}

	return &types.GetSettingsResp{Settings: s}, nil
}
