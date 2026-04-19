package objects

import "time"

// SystemSetting 系统级配置表（管理员后台可维护），key/value 形式
type SystemSetting struct {
	ID        uint64    `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:主键"`
	KeyName   string    `gorm:"column:key_name;type:VARCHAR(100);not null;uniqueIndex:uk_key_name;comment:配置键"`
	Value     string    `gorm:"column:value;type:TEXT;not null;comment:配置值（字符串，业务自解析）"`
	Remark    string    `gorm:"column:remark;type:VARCHAR(255);not null;default:'';comment:备注"`
	CreatedAt time.Time `gorm:"column:created_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}
