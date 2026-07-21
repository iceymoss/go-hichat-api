package socialmodels

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GroupsModel = (*customGroupsModel)(nil)

type (
	// GroupsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGroupsModel.
	GroupsModel interface {
		groupsModel
		// SearchGroups 群号精确 OR 群名模糊分页搜索，返回当前页与匹配总数
		SearchGroups(ctx context.Context, keyword string, offset, limit int64) ([]*Groups, int64, error)
	}

	customGroupsModel struct {
		*defaultGroupsModel
	}
)

// NewGroupsModel returns a model for the database table.
func NewGroupsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GroupsModel {
	return &customGroupsModel{
		defaultGroupsModel: newGroupsModel(conn, c, opts...),
	}
}

// SearchGroups 群号精确 OR 群名模糊分页搜索（GORM，三库兼容）。
// 群号(id)为整型列：仅当 keyword 为纯数字时才参与精确匹配，避免 PostgreSQL 整型列与非数字字符串比较报错。
func (m *customGroupsModel) SearchGroups(ctx context.Context, keyword string, offset, limit int64) ([]*Groups, int64, error) {
	var (
		list  []*Groups
		total int64
	)

	tx := m.mysqlConn.WithContext(ctx).Table(m.table)
	if isAllDigits(keyword) {
		tx = tx.Where("id = ? OR name LIKE ?", keyword, "%"+keyword+"%")
	} else {
		tx = tx.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return list, 0, nil
	}
	if err := tx.Offset(int(offset)).Limit(int(limit)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// isAllDigits 判断字符串是否为非空纯数字
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
