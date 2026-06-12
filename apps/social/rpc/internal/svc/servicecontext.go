package svc

import (
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	mq_client "github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	socialmodels.FriendsModel        //好友关系表
	socialmodels.FriendRequestsModel //好友申请表
	socialmodels.GroupsModel         //群信息表
	socialmodels.GroupRequestsModel  //群申请表
	socialmodels.GroupMembersModel   //群成员表

	// RelationOutboxModel 关系变更事务性发件箱
	RelationOutboxModel socialmodels.RelationOutboxModel
	// RelationChangeTransferClient 关系变更事件生产者（relay 投递用）
	RelationChangeTransferClient mq_client.RelationChangeTransferClient
	// RelationCache 关系缓存：变更后 best-effort 同步，让闸门即时生效
	RelationCache *relationcache.Cache

	User userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {

	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:              c,
		FriendsModel:        socialmodels.NewFriendsModel(sqlConn, c.Cache),
		FriendRequestsModel: socialmodels.NewFriendRequestsModel(sqlConn, c.Cache),
		GroupsModel:         socialmodels.NewGroupsModel(sqlConn, c.Cache),
		GroupRequestsModel:  socialmodels.NewGroupRequestsModel(sqlConn, c.Cache),
		GroupMembersModel:   socialmodels.NewGroupMembersModel(sqlConn, c.Cache),

		RelationOutboxModel:          socialmodels.NewRelationOutboxModel(),
		RelationChangeTransferClient: mq_client.NewRelationChangeTransferClient(c.RelationChangeTransfer.Addrs, c.RelationChangeTransfer.Topic),
		RelationCache:                relationcache.New(db.GetRedisConn()),

		User: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
