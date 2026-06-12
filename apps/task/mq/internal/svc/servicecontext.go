package svc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/config"
	userModels "github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	// Config 服务配置
	Config config.Config

	// websocket客户端
	WsClient websocket.Client

	// imChatLogModel 聊天记录集合数据结构
	ChatLogModel model.ChatLogModel

	// ConversationModel 会话详情相关
	ConversationModel model.ConversationModel

	// 用户会话相关
	ConversationsModel model.ConversationsModel

	// 导入social微服务模块
	Social socialclient.Social

	// 关系缓存（群成员集/好友集）：消费关系变更事件维护，群扇出/鉴权读取
	RelationCache *relationcache.Cache

	// 用户个人设置（MySQL）
	UserSettingsModel *userModels.UserSettingsModel

	// 系统级配置（MySQL）
	SystemConfigModel *userModels.SystemConfigModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config:             c,
		ConversationModel:  model.NewConversationModel(),
		ChatLogModel:       model.NewChatLogModel(),
		ConversationsModel: model.NewConversationsModel(),
		Social:             socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		RelationCache:      relationcache.New(db.GetRedisConn()),
		UserSettingsModel:  userModels.NewUserSettingsModel(),
		SystemConfigModel:  userModels.NewSystemConfigModel(),
	}

	token, err := svcCtx.GetToken()
	if err != nil {
		fmt.Errorf("getToekn err %v", err)
	}

	fmt.Println("token:", token)

	header := http.Header{}
	header.Set("Authorization", token)

	// mq消费端连接websocket服务
	svcCtx.WsClient = websocket.NewClient(c.Ws.Host, websocket.WithClientHeader(header))

	return svcCtx
}

func (svcCtx *ServiceContext) GetToken() (string, error) {
	redisConn := db.GetRedisConn()
	res := redisConn.Get(context.Background(), constants.REDIS_SYSTEM_ROOT_TOEKN)
	return res.Val(), res.Err()
}
