package logic

import (
	"context"
	"fmt"
	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetConversations 获取指定用户的会话数据，用户的会话列表
func (l *GetConversationsLogic) GetConversations(in *im.GetConversationsReq) (*im.GetConversationsResp, error) {
	// 获取用户会话信息
	data, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	fmt.Printf("data list: %+v\n", data)
	if err != nil {
		if err == models.ErrNotFound {
			return &im.GetConversationsResp{}, nil
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation by userId err %v req %v", err, in.UserId)
	}

	var res im.GetConversationsResp
	//conversationMap := make(map[string]*im.Conversation, len(data.ConversationList))
	//for k, v := range data.ConversationList {
	//	var msg im.ChatLog
	//	if v.Msg != nil {
	//		msg = im.ChatLog{
	//			Id:             v.Msg.ID.String(),
	//			ConversationId: v.Msg.ConversationId,
	//			SendId:         v.Msg.SendId,
	//			RecvId:         v.Msg.RecvId,
	//			MsgType:        int32(v.Msg.MsgType),
	//			MsgContent:     v.Msg.MsgContent,
	//			ChatType:       int32(v.Msg.ChatType),
	//			SendTime:       v.Msg.SendTime,
	//			ReadRecords:    v.Msg.ReadRecords,
	//		}
	//	}
	//	conversationMap[k] = &im.Conversation{
	//		ConversationId: v.ConversationId,
	//		ChatType:       int32(v.ChatType),
	//		TargetId:       v.Msg.RecvId,
	//		IsShow:         v.IsShow,
	//		Seq:            v.Seq,
	//		Total:          int32(v.Total),
	//		Msg:            &msg,
	//	}
	//}
	//res.ConversationList = conversationMap

	copier.Copy(&res, &data)

	// 会话id
	ids := make([]string, 0, len(data.ConversationList))
	for _, conversation := range data.ConversationList {
		ids = append(ids, conversation.ConversationId)
	}

	// 批量获取会话，统计会话的消息情况
	list, err := l.svcCtx.ConversationModel.ListByConversationIds(l.ctx, ids)
	if err != nil {
		return nil, errorx.Wrapf(xerr.NewDBErr(), "list conversation by ids err %v, req %v", err, ids)
	}

	for _, conversation := range list {
		if _, ok := res.ConversationList[conversation.ConversationId]; !ok {
			continue
		}

		// 用户读取的消息量
		total := res.ConversationList[conversation.ConversationId].Total
		if total < int32(conversation.Total) { // 如果读取的消息量 < 会话实际的消息量 => 存在未读消息
			// 有新的消息，记录新的实际消息量给前端
			res.ConversationList[conversation.ConversationId].Total = int32(conversation.Total)
			// 待读消息量
			res.ConversationList[conversation.ConversationId].ToRead = int32(conversation.Total) - total
			// 有新消息一定显示
			res.ConversationList[conversation.ConversationId].IsShow = true
		}

		if conversation.Msg != nil {
			readRecords := make([]byte, 0)
			if conversation.Msg.ReadRecords != nil {
				readRecords = conversation.Msg.ReadRecords
			}
			res.ConversationList[conversation.ConversationId].Msg = &im.ChatLog{
				ConversationId: conversation.Msg.ConversationId,
				SendId:         conversation.Msg.SendId,
				RecvId:         conversation.Msg.RecvId,
				MsgType:        int32(conversation.Msg.MsgType),
				MsgContent:     conversation.Msg.MsgContent,
				ChatType:       int32(conversation.Msg.ChatType),
				SendTime:       conversation.Msg.SendTime,
				ReadRecords:    readRecords,
			}
		}
	}

	return &res, nil
}
