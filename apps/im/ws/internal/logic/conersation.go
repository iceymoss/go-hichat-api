package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"
	"go.uber.org/zap"
	"time"
)

// ChatLogSlg 会话
type ChatLogSlg interface {
	SingleChatLog(data *ws.Chat, userId string) error
}

// UserLogic 用户单聊会话
type UserLogic struct {
	ctx    context.Context
	srv    *websocket.Server
	svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, srv *websocket.Server, svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		ctx:    ctx,
		srv:    srv,
		svcCtx: svcCtx,
	}
}

// Chat 聊天，构建聊天内容
func (l *UserLogic) Chat(data *ws.Chat, userId string) error {
	if data.ConversationId == "" {
		// 生成会话
		data.ConversationId = wuid.CombineId(data.RecvId, userId)
	}

	now := time.Now().UnixNano()

	//push msg
	pushMsg := ws.Push{
		ConversationId: data.ConversationId,
		ChatType:       data.ChatType,
		SendId:         userId,
		RecvId:         data.RecvId,
		SendTime:       now,
		MType:          constants.TextMType,
		Content:        data.Content,
	}

	msg := websocket.Message{
		FrameType: websocket.FrameData,
		UserId:    data.RecvId,
		FormId:    userId,
		Data:      pushMsg,
	}

	// 获取发送的对象的连接
	toConn := l.srv.GetConn([]string{data.RecvId})

	//不存在，说明离线了, 发给自己的消息可以入库，但是不能推送
	if len(toConn) != 0 && data.RecvId != userId {
		//在线情况，push到目标conn中
		pushErr := l.srv.SendByUserId(msg, data.RecvId)
		// ToDo：如果push失败，需要做什么？
		if pushErr != nil {
			zLog.Error("Chat.SendByUserId: push msg failed", zap.Any("msg", msg))
		}
	}

	chatLog := model.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         userId,
		RecvId:         data.RecvId,
		MsgType:        data.MType,
		MsgContent:     data.Content,
		SendTime:       time.Now().UnixNano(),
	}

	// 插入聊天记录中
	_, err := l.svcCtx.ChatLogModel.Insert(l.ctx, &chatLog)
	return err
}
