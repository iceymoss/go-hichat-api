package logic

import (
	"context"
	"errors"
	"fmt"
	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/zeromicro/go-zero/core/errorx"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupConversationLogic {
	return &CreateGroupConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateGroupConversation 创建群会话,创建群时创建，群的会话是由创建群的时候创建，在会话表中只有一条群的会话，通过用户会话表进行绑定
func (l *CreateGroupConversationLogic) CreateGroupConversation(in *im.CreateGroupConversationReq) (*im.CreateGroupConversationResp, error) {
	res := &im.CreateGroupConversationResp{}

	// 群聊中，群id就是会话id
	_, err := l.svcCtx.ConversationModel.FindOne(l.ctx, in.GroupId)
	if err == nil {
		return res, nil
	}

	// 获取群学校，判断是否为群主
	groupList, err := l.svcCtx.Social.GroupList(l.ctx, &social.GroupListReq{UserId: in.CreateId})
	if err != nil {
		zLog.Error("CreateGroupConversation.GroupList: get group users failed", zap.Any("req", in), zap.Error(err))
		return nil, err
	}

	fmt.Printf("group:%+v\n", groupList)

	isGroupCreater := false
	for _, group := range groupList.List {
		if group.CreatorUid == in.CreateId && group.Id == in.GroupId {
			isGroupCreater = true
			break
		}
	}

	if !isGroupCreater {
		return nil, errors.New("该用户不是群主")
	}

	err = l.svcCtx.ConversationModel.Insert(l.ctx, &models.Conversation{
		ConversationId: in.GroupId,
		ChatType:       constants.GroupChatType,
	})
	if err != nil {
		zLog.Error(fmt.Sprintf("CreateGroupConversation.Insert err %v, req %v", err, in), zap.Error(err))
		return res, errorx.Wrapf(xerr.NewDBErr(), "insert conversation err %v, req %v", err, in)
	}

	_, err = NewSetUpUserConversationLogic(l.ctx, l.svcCtx).SetUpUserConversation(&im.SetUpUserConversationReq{
		SendId:   in.CreateId,
		RecvId:   in.GroupId,
		ChatType: int32(constants.GroupChatType),
	})

	return res, nil
}
