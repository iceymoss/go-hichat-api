package objects

import (
	"time"
)

// User 用户表结构
type User struct {
	ID           uint64     `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:用户ID（自增主键）"`
	Avatar       string     `gorm:"column:avatar;type:VARCHAR(255);not null;default:'';comment:头像URL"`
	Nickname     string     `gorm:"column:nickname;type:VARCHAR(64);not null;comment:昵称"`
	Phone        string     `gorm:"column:phone;type:VARCHAR(20);not null;uniqueIndex:uk_phone;comment:手机号（唯一）"`
	Email        string     `gorm:"column:email;type:VARCHAR(255);uniqueIndex:uk_email;comment:邮箱（唯一）"`
	Type         uint64     `gorm:"column:type;type:TINYINT UNSIGNED;not null;default:0;comment:用户类型（0-普通用户 1-管理员）"`
	LastLogin    *time.Time `gorm:"column:last_login;type:TIMESTAMP;comment:最后登录时间"`
	Password     string     `gorm:"column:password;type:VARCHAR(255);not null;comment:密码哈希"`
	Status       uint64     `gorm:"column:status;type:TINYINT UNSIGNED;not null;default:1;comment:状态（0-禁用 1-正常）"`
	Sex          *int       `gorm:"column:sex;type:TINYINT UNSIGNED;comment:性别（0-未知 1-男 2-女）"`
	Introduction string     `gorm:"column:introduction;type:VARCHAR(100);comment:个性签名"`
	Region       string     `gorm:"column:region;type:VARCHAR(100);comment:地区"`
	Occupation   string     `gorm:"column:occupation;type:VARCHAR(100);comment:职业"`
	Tags         string     `gorm:"column:tags;type:TEXT;comment:个人标签（JSON数组）"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
