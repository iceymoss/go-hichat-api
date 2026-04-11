package models

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
	"gorm.io/gorm"
)

type UserEmoji struct {
	Id        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId    uint64    `gorm:"not null;index:idx_user;index:idx_user_sort" json:"userId"`
	Url       string    `gorm:"type:varchar(500);not null" json:"url"`
	Name      string    `gorm:"type:varchar(100);not null;default:''" json:"name"`
	Thumbnail string    `gorm:"type:varchar(500);not null;default:''" json:"thumbnail"`
	Width     uint      `gorm:"not null;default:0" json:"width"`
	Height    uint      `gorm:"not null;default:0" json:"height"`
	Size      uint      `gorm:"not null;default:0" json:"size"`
	FileType  string    `gorm:"type:varchar(20);not null;default:image" json:"fileType"`
	SortOrder int       `gorm:"not null;default:0;index:idx_user_sort" json:"sortOrder"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserEmoji) TableName() string {
	return "user_emojis"
}

type UserEmojiModel struct{}

func NewUserEmojiModel() *UserEmojiModel {
	return &UserEmojiModel{}
}

func (m *UserEmojiModel) conn(ctx context.Context) *gorm.DB {
	return transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
}

// Create 添加表情
func (m *UserEmojiModel) Create(ctx context.Context, data *UserEmoji) error {
	return m.conn(ctx).Create(data).Error
}

// Delete 删除表情(仅限本人)
func (m *UserEmojiModel) Delete(ctx context.Context, id, userId uint64) error {
	return m.conn(ctx).Where("id = ? AND user_id = ?", id, userId).Delete(&UserEmoji{}).Error
}

// List 获取用户表情包列表(按 sort_order DESC, created_at DESC)
func (m *UserEmojiModel) List(ctx context.Context, userId uint64, page, pageSize int) ([]*UserEmoji, int64, error) {
	q := m.conn(ctx).Model(&UserEmoji{}).Where("user_id = ?", userId)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*UserEmoji
	err := q.Order("sort_order DESC, created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ExistsByUrl 检查用户是否已收藏该表情
func (m *UserEmojiModel) ExistsByUrl(ctx context.Context, userId uint64, url string) (bool, error) {
	var count int64
	err := m.conn(ctx).Model(&UserEmoji{}).Where("user_id = ? AND url = ?", userId, url).Count(&count).Error
	return count > 0, err
}
