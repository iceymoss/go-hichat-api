package group

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

const Identify = "hichat2.com"

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateGroupLogic 创群
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
	res, err := l.svcCtx.Social.GroupCreate(l.ctx, &social.GroupCreateReq{
		Name:       req.Name,
		Icon:       req.Icon,
		Status:     0,
		CreatorUid: uid,
	})
	if err != nil {
		zLog.Error("CreateGroup.GroupCreate: create group failed", zap.Error(err))
		return nil, err
	}

	_, err = l.svcCtx.Im.CreateGroupConversation(l.ctx, &im.CreateGroupConversationReq{
		GroupId:  res.GroupId,
		CreateId: uid,
	})
	if err != nil {
		zLog.Error("CreateGroup.CreateGroupConversation: create conversation failed", zap.Error(err))
		return nil, err
	}

	return
}
