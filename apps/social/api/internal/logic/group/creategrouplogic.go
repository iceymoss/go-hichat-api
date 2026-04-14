package group

import (
	"context"
	"errors"
	"time"

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
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		Status:      0,
		CreatorUid:  uid,
	})
	if err != nil {
		zLog.Error("CreateGroup.GroupCreate: create group failed", zap.Error(err))
		return nil, err
	}

	// Best-effort: do NOT block or fail group creation if im-rpc is down.
	go func(groupId, createId string) {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel()
		if _, err := l.svcCtx.Im.CreateGroupConversation(ctx, &im.CreateGroupConversationReq{
			GroupId:  groupId,
			CreateId: createId,
		}); err != nil {
			zLog.Error("CreateGroup.CreateGroupConversation: best-effort failed", zap.Error(err))
		}
	}(res.GroupId, uid)

	return &types.GroupCreateResp{GroupId: res.GroupId}, nil
}
