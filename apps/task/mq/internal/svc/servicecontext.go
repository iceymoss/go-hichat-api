package svc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"sync"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/imclient"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/config"
	userModels "github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/zrpc"
)

const defaultNotificationDLQTopic = "im.notification.dead.v1"

type DLQPublisher interface {
	Publish(context.Context, []byte) error
}

type kafkaMessageWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

type notificationDLQPublisher struct {
	writer kafkaMessageWriter
	close  func() error
}

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
	Social           socialclient.Social
	Im               imclient.Im
	NotificationDLQ  DLQPublisher
	shutdownCtx      context.Context
	shutdownCancel   context.CancelFunc
	notificationGate notificationGate
	wsCloseOnce      sync.Once
	wsCloseErr       error

	// 关系缓存（群成员集/好友集）：消费关系变更事件维护，群扇出/鉴权读取
	RelationCache *relationcache.Cache

	// 用户个人设置（MySQL）
	UserSettingsModel *userModels.UserSettingsModel

	// 系统级配置（MySQL）
	SystemConfigModel *userModels.SystemConfigModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpcAuth, err := newTaskRPCAuth(c.RpcAuthSecret)
	if err != nil {
		panic(err)
	}
	dlq, err := newNotificationDLQ(c.CommonNotifyTransfer, c.NotificationDLQTopic)
	if err != nil {
		panic(fmt.Errorf("task mq startup: create notification DLQ publisher: %w", err))
	}
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	svcCtx := &ServiceContext{
		Config:             c,
		ConversationModel:  model.NewConversationModel(),
		ChatLogModel:       model.NewChatLogModel(),
		ConversationsModel: model.NewConversationsModel(),
		Social:             socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		Im: imclient.NewIm(zrpc.MustNewClient(c.ImRpc,
			zrpc.WithUnaryClientInterceptor(rpcAuth.UnaryClientInterceptor()))),
		NotificationDLQ:   dlq,
		shutdownCtx:       shutdownCtx,
		shutdownCancel:    shutdownCancel,
		RelationCache:     relationcache.New(db.GetRedisConn()),
		UserSettingsModel: userModels.NewUserSettingsModel(),
		SystemConfigModel: userModels.NewSystemConfigModel(),
	}

	token, err := svcCtx.GetToken()
	if err != nil {
		fmt.Printf("getToken err: %v\n", err)
	}

	fmt.Println("token:", token)

	header := http.Header{}
	header.Set("Authorization", token)

	// mq消费端连接websocket服务
	svcCtx.WsClient = websocket.NewClient(c.Ws.Host, websocket.WithClientHeader(header))

	return svcCtx
}

func newNotificationDLQ(conf kq.KqConf, topic string) (*notificationDLQPublisher, error) {
	writerConfig, err := notificationDLQWriterConfig(conf, topic)
	if err != nil {
		return nil, err
	}
	writer := kafka.NewWriter(writerConfig)
	return &notificationDLQPublisher{writer: writer, close: writer.Close}, nil
}

func notificationDLQWriterConfig(conf kq.KqConf, topic string) (kafka.WriterConfig, error) {
	if topic == "" {
		topic = defaultNotificationDLQTopic
	}
	dialer := &kafka.Dialer{}
	if conf.Username != "" && conf.Password != "" {
		dialer.SASLMechanism = plain.Mechanism{Username: conf.Username, Password: conf.Password}
	}
	if conf.CaFile != "" {
		caCert, err := os.ReadFile(conf.CaFile)
		if err != nil {
			return kafka.WriterConfig{}, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return kafka.WriterConfig{}, fmt.Errorf("parse Kafka CA certificate %q", conf.CaFile)
		}
		dialer.TLS = &tls.Config{RootCAs: caCertPool, MinVersion: tls.VersionTLS12}
	}
	return kafka.WriterConfig{
		Brokers:      conf.Brokers,
		Topic:        topic,
		Dialer:       dialer,
		RequiredAcks: int(kafka.RequireAll),
		Async:        false,
		BatchSize:    1,
	}, nil
}

func (p *notificationDLQPublisher) Publish(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Value: value})
}

func (p *notificationDLQPublisher) Close() error {
	if p.close == nil {
		return nil
	}
	return p.close()
}

func newTaskRPCAuth(configured string) (*rpcauth.Auth, error) {
	auth, err := rpcauth.New(rpcauth.LoadSecret(configured))
	if err != nil {
		return nil, fmt.Errorf("task mq startup: HICHAT_IM_RPC_AUTH_SECRET is required: %w", err)
	}
	return auth, nil
}

func (svcCtx *ServiceContext) Close() error {
	svcCtx.Cancel()
	var firstErr error
	if closer, ok := svcCtx.NotificationDLQ.(interface{ Close() error }); ok {
		firstErr = closer.Close()
	}
	if err := svcCtx.closeWebsocket(); firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (svcCtx *ServiceContext) ShutdownContext() context.Context {
	if svcCtx.shutdownCtx == nil {
		return context.Background()
	}
	return svcCtx.shutdownCtx
}

func (svcCtx *ServiceContext) Cancel() {
	if svcCtx.shutdownCancel != nil {
		svcCtx.shutdownCancel()
	}
	// Closing the socket unblocks an in-flight best-effort push before the gate
	// waits for active notification handlers to leave.
	_ = svcCtx.closeWebsocket()
	svcCtx.notificationGate.close()
}

func (svcCtx *ServiceContext) closeWebsocket() error {
	svcCtx.wsCloseOnce.Do(func() {
		if svcCtx.WsClient != nil {
			svcCtx.wsCloseErr = svcCtx.WsClient.Close()
		}
	})
	return svcCtx.wsCloseErr
}

func (svcCtx *ServiceContext) BeginNotification() (func(), bool) {
	return svcCtx.notificationGate.begin()
}

type notificationGate struct {
	mu     sync.RWMutex
	closed bool
}

func (g *notificationGate) begin() (func(), bool) {
	g.mu.RLock()
	if g.closed {
		g.mu.RUnlock()
		return nil, false
	}
	return g.mu.RUnlock, true
}

func (g *notificationGate) close() {
	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()
}

func (svcCtx *ServiceContext) GetToken() (string, error) {
	redisConn := db.GetRedisConn()
	res := redisConn.Get(context.Background(), constants.REDIS_SYSTEM_ROOT_TOEKN)
	return res.Val(), res.Err()
}
