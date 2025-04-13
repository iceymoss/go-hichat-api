package svc

import (
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	socialmodels.FriendsModel        //好友关系表
	socialmodels.FriendRequestsModel //好友申请表
	socialmodels.GroupsModel         //群信息表
	socialmodels.GroupRequestsModel  //群申请表
	socialmodels.GroupMembersModel   //群成员表
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
	}
}
