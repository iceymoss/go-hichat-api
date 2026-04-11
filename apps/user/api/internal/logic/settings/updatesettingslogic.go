package settings

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSettingsLogic {
	return &UpdateSettingsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateSettingsLogic) UpdateSettings(req *types.UpdateSettingsReq) (*types.UpdateSettingsResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	userId, _ := strconv.ParseUint(uid, 10, 64)
	if userId == 0 {
		return nil, libErr.New(xerr.ErrBadRequest, "用户未登录")
	}

	// Validate it's valid JSON
	if !json.Valid([]byte(req.Settings)) {
		return nil, libErr.New(xerr.ErrBadRequest, "配置格式无效")
	}

	m := models.NewUserSettingsModel()
	if err := m.Upsert(l.ctx, userId, req.Settings); err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "保存配置失败")
	}

	return &types.UpdateSettingsResp{}, nil
}
