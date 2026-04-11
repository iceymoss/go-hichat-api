package models

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
	"gorm.io/gorm"
)

type Favorite struct {
	Id        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId    uint64    `gorm:"not null;index:idx_user_id;index:idx_user_type" json:"userId"`
	Type      string    `gorm:"type:varchar(20);not null;default:text;index:idx_user_type" json:"type"`
	Title     string    `gorm:"type:varchar(255);not null;default:''" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Source    string    `gorm:"type:varchar(20);not null;default:''" json:"source"`
	SourceId  string    `gorm:"type:varchar(64);not null;default:''" json:"sourceId"`
	Tags      string    `gorm:"type:varchar(500);not null;default:''" json:"tags"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Favorite) TableName() string {
	return "favorites"
}

type FavoriteModel struct{}

func NewFavoriteModel() *FavoriteModel {
	return &FavoriteModel{}
}

func (m *FavoriteModel) conn(ctx context.Context) *gorm.DB {
	return transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
}

// Create 创建收藏
func (m *FavoriteModel) Create(ctx context.Context, data *Favorite) error {
	return m.conn(ctx).Create(data).Error
}

// Delete 删除收藏(仅限本人)
func (m *FavoriteModel) Delete(ctx context.Context, id, userId uint64) error {
	res := m.conn(ctx).Where("id = ? AND user_id = ?", id, userId).Delete(&Favorite{})
	return res.Error
}

// FindOne 查询单条
func (m *FavoriteModel) FindOne(ctx context.Context, id, userId uint64) (*Favorite, error) {
	var fav Favorite
	err := m.conn(ctx).Where("id = ? AND user_id = ?", id, userId).First(&fav).Error
	return &fav, err
}

// List 分页查询用户收藏
func (m *FavoriteModel) List(ctx context.Context, userId uint64, favType string, keyword string, page, pageSize int) ([]*Favorite, int64, error) {
	q := m.conn(ctx).Model(&Favorite{}).Where("user_id = ?", userId)

	if favType != "" {
		q = q.Where("type = ?", favType)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*Favorite
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
