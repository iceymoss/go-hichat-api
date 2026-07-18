package svc

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	mq_client "github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type UserLookup interface {
	GetUserById(ctx context.Context, in *user.GetUserByIdRequest, opts ...grpc.CallOption) (*user.GetUserByIdResponse, error)
}

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB

	socialmodels.FriendsModel        //好友关系表
	socialmodels.FriendRequestsModel //好友申请表
	socialmodels.GroupsModel         //群信息表
	socialmodels.GroupRequestsModel  //群申请表
	socialmodels.GroupMembersModel   //群成员表

	// RelationOutboxModel 关系变更事务性发件箱
	RelationOutboxModel           socialmodels.RelationOutboxModel
	SocialNotificationOutboxModel socialmodels.SocialNotificationOutboxModel
	// RelationChangeTransferClient 关系变更事件生产者（relay 投递用）
	RelationChangeTransferClient mq_client.RelationChangeTransferClient
	// CommonNotifyClient 公共通知事件生产者（好友/群申请等实时通知，直接 Push）
	CommonNotifyClient              mq_client.CommonNotifyClient
	SocialRequestNotificationClient mq_client.CommonNotifyClient
	// RelationCache 关系缓存：变更后 best-effort 同步，让闸门即时生效
	RelationCache *relationcache.Cache

	User UserLookup
}

func NewServiceContext(c config.Config) *ServiceContext {

	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)
	notificationAddrs := c.SocialRequestNotification.Addrs
	if len(notificationAddrs) == 0 {
		notificationAddrs = c.CommonNotifyTransfer.Addrs
	}
	notificationTopic := c.SocialRequestNotification.Topic
	if notificationTopic == "" {
		notificationTopic = "social.request.notification.v1"
	}

	return &ServiceContext{
		Config:              c,
		DB:                  db.GetMysqlConn(db.MYSQL_DB_HICHAT2),
		FriendsModel:        socialmodels.NewFriendsModel(sqlConn, c.Cache),
		FriendRequestsModel: socialmodels.NewFriendRequestsModel(sqlConn, c.Cache),
		GroupsModel:         socialmodels.NewGroupsModel(sqlConn, c.Cache),
		GroupRequestsModel:  socialmodels.NewGroupRequestsModel(sqlConn, c.Cache),
		GroupMembersModel:   socialmodels.NewGroupMembersModel(sqlConn, c.Cache),

		RelationOutboxModel:             socialmodels.NewRelationOutboxModel(),
		SocialNotificationOutboxModel:   socialmodels.NewSocialNotificationOutboxModel(),
		RelationChangeTransferClient:    mq_client.NewRelationChangeTransferClient(c.RelationChangeTransfer.Addrs, c.RelationChangeTransfer.Topic),
		CommonNotifyClient:              mq_client.NewCommonNotifyClient(c.CommonNotifyTransfer.Addrs, c.CommonNotifyTransfer.Topic),
		SocialRequestNotificationClient: mq_client.NewCommonNotifyClient(notificationAddrs, notificationTopic),
		RelationCache:                   relationcache.New(db.GetRedisConn()),

		User: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
