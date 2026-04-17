package logic

import (
	"context"
	"fmt"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type SetUpUserConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUpUserConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUpUserConversationLogic {
	return &SetUpUserConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetUpUserConversation 给用户建立会话: 群聊, 私聊
func (l *SetUpUserConversationLogic) SetUpUserConversation(in *im.SetUpUserConversationReq) (*im.SetUpUserConversationResp, error) {
	var res im.SetUpUserConversationResp
	switch constants.ChatType(in.ChatType) {
	case constants.SingleChatType: // 单聊
		// 建立私聊的关系：是在用户点击发起聊天后触发

		// 生成会话id
		conversationId := wuid.CombineId(in.SendId, in.RecvId)

		// 建立两者的会话
		// 查询当前会话id是否存在
		_, err := l.svcCtx.ConversationModel.FindOne(l.ctx, conversationId)
		if err != nil {
			if err == models.ErrNotFound { // 如果会话不存在，插入新会话
				err = l.svcCtx.ConversationModel.Insert(l.ctx, &models.Conversation{
					ConversationId: conversationId,
					Msg:            &models.ChatLog{},
					ChatType:       constants.SingleChatType,
				})
				if err != nil {
					zLog.Error(fmt.Sprintf("create conversation err %v, req %v\", err, in", err, in))
					return nil, errors.Wrapf(xerr.NewDBErr(), "create conversation err %v, req %v", err, in)
				}
			} else {
				zLog.Error(fmt.Sprintf("find conversation err %v, req %v", err, conversationId))
				return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation err %v, req %v", err, conversationId)
			}
		}

		// 给发送者用户建议会话记录
		err = l.setUpUserConversation(conversationId, in.SendId, constants.SingleChatType, true)
		if err != nil {
			zLog.Error("SetUpUserConversation.setUpUserConversationset: up user single conversation err", zap.Any("in", in), zap.Error(err))
			return &res, errors.Wrapf(xerr.NewDBErr(), "set up user single conversation err %v, req %v", err, in)
		}

		// 给接受者用户建议会话记录，但是不做展示，等发送者用户这种发送消息后，才将其设置为展示
		// 接收者是被动与目标用户建立连接，因此理论上是不需要在会话列表里展示，而是在用户发起聊天后展示
		err = l.setUpUserConversation(conversationId, in.RecvId, constants.SingleChatType, false)
		if err != nil {
			zLog.Error("SetUpUserConversation.setUpUserConversation2: up user single conversation err", zap.Any("in", in), zap.Error(err))
			return &res, errors.Wrapf(xerr.NewDBErr(), "set up user single conversation err %v, req %v", err, in)
		}
	case constants.GroupChatType:
		// 建立群会话详情
		conversationId := in.RecvId
		_, err := l.svcCtx.ConversationModel.FindOne(l.ctx, conversationId)
		if err != nil {
			if err == models.ErrNotFound { // 如果会话不存在，插入新会话
				err = l.svcCtx.ConversationModel.Insert(l.ctx, &models.Conversation{
					ConversationId: conversationId,
					Msg:            &models.ChatLog{},
					ChatType:       constants.GroupChatType,
				})
				if err != nil {
					zLog.Error(fmt.Sprintf("create conversation err %v, req %v\", err, in", err, in))
					return nil, errors.Wrapf(xerr.NewDBErr(), "create conversation err %v, req %v", err, in)
				}
			} else {
				zLog.Error(fmt.Sprintf("find conversation err %v, req %v", err, conversationId))
				return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation err %v, req %v", err, conversationId)
			}
		}
		// 建立群聊: 动作的触发时在加群通过后
		// 用户加入群后应在会话里显示群聊
		// 每一个用户和群建立会话，群id就是会话id
		err = l.setUpUserConversation(in.RecvId, in.SendId, constants.GroupChatType, true)
		if err != nil {
			zLog.Error("SetUpUserConversation.setUpUserConversation3: up user single conversation err", zap.Any("in", in), zap.Error(err))
			return &res, errors.Wrapf(xerr.NewDBErr(), "set up user group conversation err %v, req %v", err, in)
		}
	}

	return &res, nil
}

// setUpUserConversation 设置会话
func (l *SetUpUserConversationLogic) setUpUserConversation(conversationId, userId string, chatType constants.ChatType, isShow bool) error {
	// 发送者
	conversations, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, userId)
	if err != nil {
		if err == models.ErrNotFound {
			conversations = &models.Conversations{
				ID:               primitive.NewObjectID(),
				UserId:           userId,
				ConversationList: make(map[string]*models.Conversation),
			}
		} else {
			return err
		}
	}
	// 更新会话记录，如果用户会话存在
	if existing, ok := conversations.ConversationList[conversationId]; ok {
		// 会话已存在，如果当前要求显示且之前被隐藏了，则恢复显示
		if isShow && !existing.IsShow {
			existing.IsShow = true
			_, err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
			return err
		}
		return nil
	}

	// 需要建立
	conversations.ConversationList[conversationId] = &models.Conversation{
		ConversationId: conversationId,
		ChatType:       chatType,
		IsShow:         isShow,
	}

	// 存在即更新、不存在则修改
	_, err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
	return err
}
