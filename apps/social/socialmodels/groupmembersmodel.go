package socialmodels

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GroupMembersModel = (*customGroupMembersModel)(nil)

type (
	// GroupMembersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGroupMembersModel.
	GroupMembersModel interface {
		groupMembersModel

		// DeleteByGroupId 删除整个群的成员（解散群用）
		DeleteByGroupId(ctx context.Context, groupId string) error

		// UpdateRoleLevelByUserIds 批量更新成员角色（设管理员/取消管理员/转让群主用）
		UpdateRoleLevelByUserIds(ctx context.Context, groupId string, userIds []string, roleLevel int) error

		// UpdateRoleLevel 更新单个成员角色
		UpdateRoleLevel(ctx context.Context, groupId string, userId string, roleLevel int) error
	}

	customGroupMembersModel struct {
		*defaultGroupMembersModel
	}
)

// NewGroupMembersModel returns a model for the database table.
func NewGroupMembersModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GroupMembersModel {
	return &customGroupMembersModel{
		defaultGroupMembersModel: newGroupMembersModel(conn, c, opts...),
	}
}

func (m *customGroupMembersModel) DeleteByGroupId(ctx context.Context, groupId string) error {
	return m.mysqlConn.WithContext(ctx).Table(m.table).Where("group_id = ?", groupId).Delete(&GroupMembers{}).Error
}

func (m *customGroupMembersModel) UpdateRoleLevelByUserIds(ctx context.Context, groupId string, userIds []string, roleLevel int) error {
	if len(userIds) == 0 {
		return nil
	}
	return m.mysqlConn.WithContext(ctx).Table(m.table).
		Where("group_id = ?", groupId).
		Where("user_id in (?)", userIds).
		Updates(map[string]any{"role_level": roleLevel}).Error
}

func (m *customGroupMembersModel) UpdateRoleLevel(ctx context.Context, groupId string, userId string, roleLevel int) error {
	return m.mysqlConn.WithContext(ctx).Table(m.table).
		Where("group_id = ?", groupId).
		Where("user_id = ?", userId).
		Updates(map[string]any{"role_level": roleLevel}).Error
}
