package models

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 系统配置键名常量，集中管理避免散落。命名与现网 system_settings.key_name 保持一致（camelCase）
const (
	SysCfgReadReceiptEnabled = "readReceiptEnabled"
)

type SystemConfigModel struct{}

func NewSystemConfigModel() *SystemConfigModel {
	return &SystemConfigModel{}
}

func (m *SystemConfigModel) conn(ctx context.Context) *gorm.DB {
	return transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
}

// GetBool 读取布尔型配置，记录不存在或解析失败都返回 defaultVal，保证调用链稳定
func (m *SystemConfigModel) GetBool(ctx context.Context, key string, defaultVal bool) bool {
	var row objects.SystemSetting
	err := m.conn(ctx).Where("key_name = ?", key).First(&row).Error
	if err != nil {
		return defaultVal
	}
	switch row.Value {
	case "true", "1", "on", "yes":
		return true
	case "false", "0", "off", "no", "":
		return false
	}
	return defaultVal
}

// Upsert 管理员写入；后台接入时复用
func (m *SystemConfigModel) Upsert(ctx context.Context, key, value, remark string) error {
	row := objects.SystemSetting{
		KeyName: key,
		Value:   value,
		Remark:  remark,
	}
	return m.conn(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "remark"}),
		}).
		Create(&row).Error
}

// SetBool 辅助：写入布尔
func (m *SystemConfigModel) SetBool(ctx context.Context, key string, value bool, remark string) error {
	return m.Upsert(ctx, key, strconv.FormatBool(value), remark)
}
