package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EnsureGroupConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnsureGroupConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnsureGroupConversationLogic {
	return &EnsureGroupConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnsureGroupConversationLogic) EnsureGroupConversation(in *im.EnsureGroupConversationReq) (*im.EnsureGroupConversationResp, error) {
	if uid, err := strconv.ParseUint(in.UserId, 10, 64); err != nil || uid == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if gid, err := strconv.ParseUint(in.GroupId, 10, 64); err != nil || gid == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid group id")
	}
	conversationCreated, err := l.svcCtx.ConversationModel.EnsureGroup(l.ctx, in.GroupId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to ensure group conversation")
	}
	bindingCreated, err := l.svcCtx.ConversationsModel.EnsureGroup(l.ctx, in.UserId, in.GroupId, in.RelationVersion)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to ensure group conversation binding")
	}
	return &im.EnsureGroupConversationResp{ConversationCreated: conversationCreated, UserBindingChanged: bindingCreated}, nil
}
