package group

import (
	"context"
	"errors"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const Identify = "hichat2.com"

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创群
func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGroupLogic) CreateGroup(req *types.GroupCreateReq) (resp *types.GroupCreateResp, err error) {
	if req.Name == "" {
		return nil, errors.New("group name is empty")
	}
	uid := l.ctx.Value(Identify).(string)
	_, err = l.svcCtx.Social.GroupCreate(l.ctx, &social.GroupCreateReq{
		Name:       req.Name,
		Icon:       req.Icon,
		Status:     0,
		CreatorUid: uid,
	})
	if err != nil {
		return nil, err
	}

	return
}
