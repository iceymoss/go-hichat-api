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
		entry, ok := res.ConversationList[conversation.ConversationId]
		if !ok {
			continue
		}

		// 该会话上请求者的「被移出群时刻」（0=正常会话/未被移出）。
		var removedAt int64
		if mEntry, ok := data.ConversationList[conversation.ConversationId]; ok && mEntry != nil {
			removedAt = mEntry.RemovedAt
		}

		// 未读量：正常会话按实际消息量算增量；被移出群成员冻结未读（被移出后的新消息不计未读）。
		if removedAt == 0 {
			total := entry.Total
			if total < int32(conversation.Total) { // 如果读取的消息量 < 会话实际的消息量 => 存在未读消息
				entry.Total = int32(conversation.Total)
				entry.ToRead = int32(conversation.Total) - total
				// 有新消息一定要显示（即使用户之前删除过会话）
				entry.IsShow = true
			}
		}

		// 列表预览消息：默认用会话最新一条；被移出群成员只展示「被移出前的最后一条」，
		// 既不泄漏被移出后的内容，也不至于空预览（与 GetChatLog 历史截断一致）。
		previewMsg := conversation.Msg
		if removedAt > 0 && (previewMsg == nil || previewMsg.SendTime > removedAt) {
			if frozen, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, conversation.ConversationId, 0, removedAt, 1); err == nil && len(frozen) > 0 {
				previewMsg = frozen[0]
			} else {
				previewMsg = nil
			}
		}
		if previewMsg != nil {
			readRecords := make([]byte, 0)
			if previewMsg.ReadRecords != nil {
				readRecords = previewMsg.ReadRecords
			}
			entry.Msg = &im.ChatLog{
				ConversationId: previewMsg.ConversationId,
				SendId:         previewMsg.SendId,
				RecvId:         previewMsg.RecvId,
				MsgType:        int32(previewMsg.MsgType),
				MsgContent:     previewMsg.MsgContent,
				ChatType:       int32(previewMsg.ChatType),
				SendTime:       previewMsg.SendTime,
				ReadRecords:    readRecords,
			}
		}
	}

	return &res, nil
}
