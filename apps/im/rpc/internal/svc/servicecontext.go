package svc

import (
	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// 聊天记录相关
	models.ChatLogModel

	// 会话相关
	models.ConversationsModel

	// 会话下聊天相关
	models.ConversationModel

	// 公共通知
	models.NotificationModel

	// 社交模块
	Social socialclient.Social
}

func NewServiceContext(c config.Config) *ServiceContext {

	//// 使用正确的类型声明和检查逻辑
	//var socialRpcClient *zrpc.Client
	//var lastErr error
	//const maxRetry = 5
	//
	//// 尝试连接并验证
	//for i := 0; i <= maxRetry; i++ {
	//	// 正确的创建方式 - 使用 zrpc.NewClient 而不是 MustNewClient
	//	client, err := zrpc.NewClient(c.SocialRpc)
	//	if err != nil {
	//		lastErr = err
	//	} else {
	//		// 通过调用健康检查或直接获取连接对象
	//		conn := client.Conn()
	//
	//		// 检查连接状态
	//		if state := conn.GetState(); state == connectivity.Ready {
	//			socialRpcClient = &client
	//			break
	//		} else {
	//			lastErr = fmt.Errorf("连接状态: %s", state.String())
	//		}
	//	}
	//
	//	if i < maxRetry {
	//		waitTime := time.Duration(i+1) * 500 * time.Millisecond
	//		zLog.Error(fmt.Sprintf("social.rpc 连接失败(尝试 %d/%d)，%v后重试: %v",
	//			i+1, maxRetry, waitTime, lastErr))
	//		time.Sleep(waitTime)
	//	}
	//}
	//
	//// 最终检查
	//if socialRpcClient == nil {
	//	logx.Must(fmt.Errorf("social.rpc 连接失败: %v", lastErr))
	//}
	//
	//// 创建服务客户端
	//socialClient := socialclient.NewSocial(*socialRpcClient)

	return &ServiceContext{
		Config:             c,
		ChatLogModel:       models.NewChatLogModel(),
		ConversationModel:  models.NewConversationModel(),
		ConversationsModel: models.NewConversationsModel(),
		NotificationModel:  models.NewNotificationModel(),
		Social:             socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
	}
}
