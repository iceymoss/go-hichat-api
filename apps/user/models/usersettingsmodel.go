package models

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSettings struct {
	Id        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId    uint64    `gorm:"not null;uniqueIndex:uk_user" json:"userId"`
	Settings  string    `gorm:"type:json;not null" json:"settings"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

type UserSettingsModel struct{}

func NewUserSettingsModel() *UserSettingsModel {
	return &UserSettingsModel{}
}

func (m *UserSettingsModel) conn(ctx context.Context) *gorm.DB {
	return transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
}

// Get 获取用户配置，不存在返回空字符串
func (m *UserSettingsModel) Get(ctx context.Context, userId uint64) (string, error) {
	var row UserSettings
	err := m.conn(ctx).Where("user_id = ?", userId).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return row.Settings, nil
}

// Upsert 创建或更新用户配置
func (m *UserSettingsModel) Upsert(ctx context.Context, userId uint64, settings string) error {
	row := UserSettings{UserId: userId, Settings: settings}
	return m.conn(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"settings"}),
		}).
		Create(&row).Error
}
