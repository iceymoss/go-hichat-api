package models

import (
	"context"
	"encoding/json"
	"strconv"
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

// GetBoolField 从 settings JSON 中读取布尔字段；缺省或解析失败返回 defaultVal。
func (m *UserSettingsModel) GetBoolField(ctx context.Context, userId uint64, field string, defaultVal bool) bool {
	raw, err := m.Get(ctx, userId)
	if err != nil || raw == "" {
		return defaultVal
	}
	return parseBoolField(raw, field, defaultVal)
}

// parseBoolField 从 JSON 字符串中解析布尔字段（pure，便于单测）。
// 兼容 true/false、"true"/"false"、1/0、"1"/"0"。
func parseBoolField(raw, field string, defaultVal bool) bool {
	var kv map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &kv); err != nil {
		return defaultVal
	}
	v, ok := kv[field]
	if !ok {
		return defaultVal
	}
	var b bool
	if err := json.Unmarshal(v, &b); err == nil {
		return b
	}
	var n float64
	if err := json.Unmarshal(v, &n); err == nil {
		return n != 0
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		if parsed, e := strconv.ParseBool(s); e == nil {
			return parsed
		}
	}
	return defaultVal
}
