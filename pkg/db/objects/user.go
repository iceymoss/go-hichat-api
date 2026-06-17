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
	MomentsCover string     `gorm:"column:moments_cover;type:VARCHAR(255);not null;default:'';comment:朋友圈封面图URL"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:TIMESTAMP;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserSettings 用户配置表（每个用户一行，settings 整体存 JSON）
type UserSettings struct {
	ID        uint64    `gorm:"primaryKey;column:id;autoIncrement;comment:主键"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user;comment:用户ID"`
	Settings  string    `gorm:"column:settings;type:json;not null;comment:用户配置JSON"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// Favorite 收藏表
type Favorite struct {
	ID        uint64    `gorm:"primaryKey;column:id;autoIncrement;comment:主键"`
	UserID    uint64    `gorm:"column:user_id;not null;index:idx_user_id;index:idx_user_type,priority:1;comment:用户ID"`
	Type      string    `gorm:"column:type;type:varchar(20);not null;default:text;index:idx_user_type,priority:2;comment:收藏类型"`
	Title     string    `gorm:"column:title;type:varchar(255);not null;default:'';comment:标题"`
	Content   string    `gorm:"column:content;type:text;not null;comment:内容"`
	Source    string    `gorm:"column:source;type:varchar(20);not null;default:'';comment:来源"`
	SourceID  string    `gorm:"column:source_id;type:varchar(64);not null;default:'';comment:来源ID"`
	Tags      string    `gorm:"column:tags;type:varchar(500);not null;default:'';comment:标签"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}

func (Favorite) TableName() string {
	return "favorites"
}

// UserEmoji 用户自定义表情表
type UserEmoji struct {
	ID        uint64    `gorm:"primaryKey;column:id;autoIncrement;comment:主键"`
	UserID    uint64    `gorm:"column:user_id;not null;index:idx_user;index:idx_user_sort,priority:1;comment:用户ID"`
	URL       string    `gorm:"column:url;type:varchar(500);not null;comment:表情URL"`
	Name      string    `gorm:"column:name;type:varchar(100);not null;default:'';comment:名称"`
	Thumbnail string    `gorm:"column:thumbnail;type:varchar(500);not null;default:'';comment:缩略图URL"`
	Width     uint      `gorm:"column:width;not null;default:0;comment:宽"`
	Height    uint      `gorm:"column:height;not null;default:0;comment:高"`
	Size      uint      `gorm:"column:size;not null;default:0;comment:文件大小"`
	FileType  string    `gorm:"column:file_type;type:varchar(20);not null;default:image;comment:文件类型"`
	SortOrder int       `gorm:"column:sort_order;not null;default:0;index:idx_user_sort,priority:2;comment:排序"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}

func (UserEmoji) TableName() string {
	return "user_emojis"
}
