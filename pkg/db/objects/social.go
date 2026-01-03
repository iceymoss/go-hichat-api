package objects

import (
	"time"
)

// Friend 好友关系表
type Friend struct {
	ID        uint64     `gorm:"primaryKey;column:id;type:INT UNSIGNED;autoIncrement;comment:自增主键"`
	UserID    uint64     `gorm:"column:user_id;type:INT UNSIGNED;not null;index:idx_user;comment:用户ID"`
	FriendUID uint64     `gorm:"column:friend_uid;type:INT UNSIGNED;not null;comment:好友的用户ID"`
	Remark    string     `gorm:"column:remark;type:VARCHAR(255);comment:好友备注名（用户自定义）"`
	AddSource *int       `gorm:"column:add_source;type:TINYINT;comment:添加来源（0:未知 1:搜索 2:群组 3:二维码...）"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP;comment:好友关系建立时间"`
}

func (Friend) TableName() string {
	return "friends"
}

// FriendRequest 好友请求表
type FriendRequest struct {
	ID           uint64     `gorm:"primaryKey;column:id;type:INT UNSIGNED;autoIncrement;comment:自增主键"`
	UserID       uint64     `gorm:"column:user_id;type:INT UNSIGNED;not null;index:idx_user;comment:申请人用户ID"`
	ReqUID       uint64     `gorm:"column:req_uid;type:INT UNSIGNED;not null;comment:被申请人用户ID"`
	ReqMsg       string     `gorm:"column:req_msg;type:VARCHAR(255);comment:好友申请留言"`
	ReqTime      time.Time  `gorm:"column:req_time;type:TIMESTAMP;not null;comment:申请发起时间"`
	HandleResult *int       `gorm:"column:handle_result;type:TINYINT;comment:处理结果（0:待处理 1:同意 2:拒绝）"`
	HandleMsg    string     `gorm:"column:handle_msg;type:VARCHAR(255);comment:处理结果备注"`
	HandledAt    *time.Time `gorm:"column:handled_at;type:TIMESTAMP;comment:处理操作时间"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}

// Group 群组信息表
type Group struct {
	ID              uint64     `gorm:"primaryKey;column:id;type:INT UNSIGNED;autoIncrement;comment:自增主键"`
	Name            string     `gorm:"column:name;type:VARCHAR(255);not null;comment:群名称"`
	Icon            string     `gorm:"column:icon;type:VARCHAR(255);not null;comment:群头像URL"`
	Status          *int       `gorm:"column:status;type:TINYINT;comment:群状态（0:正常 1:已解散 2:封禁）"`
	CreatorUID      uint64     `gorm:"column:creator_uid;type:INT UNSIGNED;not null;comment:群主用户ID"`
	GroupType       int        `gorm:"column:group_type;type:INT;not null;comment:群类型（1:普通群 2:企业群 3:粉丝群...）"`
	IsVerify        int        `gorm:"column:is_verify;type:TINYINT;not null;comment:入群验证（0:不需要 1:需要）"`
	Notification    string     `gorm:"column:notification;type:VARCHAR(255);comment:群公告内容"`
	NotificationUID *uint64    `gorm:"column:notification_uid;type:INT UNSIGNED;comment:最后更新公告的用户ID"`
	CreatedAt       *time.Time `gorm:"column:created_at;type:TIMESTAMP;comment:创建时间"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;type:TIMESTAMP;comment:最后更新时间"`
}

func (Group) TableName() string {
	return "groups"
}

// GroupMember 群成员表
type GroupMember struct {
	ID          uint64     `gorm:"primaryKey;column:id;type:INT UNSIGNED;autoIncrement;comment:自增主键"`
	GroupID     uint64     `gorm:"column:group_id;type:INT UNSIGNED;not null;uniqueIndex:uk_member,priority:1;comment:关联群ID"`
	UserID      uint64     `gorm:"column:user_id;type:INT UNSIGNED;not null;uniqueIndex:uk_member,priority:2;comment:成员用户ID"`
	RoleLevel   int        `gorm:"column:role_level;type:TINYINT;not null;comment:成员角色（0:普通成员 1:管理员 2:群主）"`
	JoinTime    *time.Time `gorm:"column:join_time;type:TIMESTAMP;comment:加入群聊时间"`
	JoinSource  *int       `gorm:"column:join_source;type:TINYINT;comment:加入来源（1:扫码 2:邀请 3:搜索...）"`
	InviterUID  *uint64    `gorm:"column:inviter_uid;type:INT UNSIGNED;comment:邀请人用户ID"`
	OperatorUID *uint64    `gorm:"column:operator_uid;type:INT UNSIGNED;comment:操作人用户ID"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// GroupRequest 加群请求表
type GroupRequest struct {
	ID            uint64     `gorm:"primaryKey;column:id;type:INT UNSIGNED;autoIncrement;comment:自增主键"`
	ReqID         string     `gorm:"column:req_id;type:VARCHAR(64);not null;comment:业务请求ID（唯一标识）"`
	GroupID       uint64     `gorm:"column:group_id;type:INT UNSIGNED;not null;index:idx_group;comment:目标群ID"`
	ReqMsg        string     `gorm:"column:req_msg;type:VARCHAR(255);comment:入群申请留言"`
	ReqTime       *time.Time `gorm:"column:req_time;type:TIMESTAMP;comment:申请时间"`
	JoinSource    *int       `gorm:"column:join_source;type:TINYINT;comment:申请来源（1:扫码 2:邀请 3:搜索...）"`
	InviterUserID *uint64    `gorm:"column:inviter_user_id;type:INT UNSIGNED;comment:邀请人ID"`
	HandleUserID  *uint64    `gorm:"column:handle_user_id;type:INT UNSIGNED;comment:请求处理人ID"`
	HandleTime    *time.Time `gorm:"column:handle_time;type:TIMESTAMP;comment:处理时间"`
	HandleResult  *int       `gorm:"column:handle_result;type:TINYINT;comment:处理结果（0:待处理 1:同意 2:拒绝）"`
}

func (GroupRequest) TableName() string {
	return "group_requests"
}
