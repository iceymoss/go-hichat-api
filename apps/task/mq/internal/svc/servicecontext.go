package svc

import (
	"context"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/config"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"net/http"
)

type ServiceContext struct {
	// Config 服务配置
	Config config.Config

	// websocket客户端
	WsClient websocket.Client

	// imChatLogModel 聊天记录集合数据结构
	ChatLogModel model.ChatLogModel

	// ConversationModel 会话相关
	ConversationModel model.ConversationModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config:            c,
		ConversationModel: model.NewConversationModel(),
		ChatLogModel:      model.NewChatLogModel(),
	}

	token, err := svcCtx.GetToken()
	if err != nil {
		fmt.Errorf("getToekn err %v", err)
	}

	fmt.Println("token:", token)

	header := http.Header{}
	header.Set("Authorization", token)
	svcCtx.WsClient = websocket.NewClient(c.Ws.Host, websocket.WithClientHeader(header))

	return svcCtx
}

func (svcCtx *ServiceContext) GetToken() (string, error) {
	redisConn := db.GetRedisConn()
	res := redisConn.Get(context.Background(), constants.REDIS_SYSTEM_ROOT_TOEKN)
	return res.Val(), res.Err()
}
