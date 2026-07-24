package objects

import (
	"time"
)

// Friend 好友关系表
type Friend struct {
	ID                uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	UserID            uint64     `gorm:"column:user_id;not null;index:idx_user;uniqueIndex:uk_friends_user_friend,priority:1;comment:用户ID"`
	FriendUID         uint64     `gorm:"column:friend_uid;not null;uniqueIndex:uk_friends_user_friend,priority:2;comment:好友的用户ID"`
	Remark            string     `gorm:"column:remark;type:VARCHAR(255);comment:好友备注名（用户自定义）"`
	AddSource         *int       `gorm:"column:add_source;size:8;comment:添加来源（0:未知 1:搜索 2:群组 3:二维码...）"`
	Blacklisted       bool       `gorm:"column:blacklisted;default:false;not null;comment:是否拉黑 0否 1是"`
	MomentsPermission int        `gorm:"column:moments_permission;size:8;default:0;not null;comment:朋友圈权限 0允许 1仅聊天 2屏蔽朋友圈"`
	NotifyEnabled     bool       `gorm:"column:notify_enabled;default:true;not null;comment:消息通知 1开 0关"`
	Pinned            bool       `gorm:"column:pinned;default:false;not null;comment:置顶聊天 0否 1是"`
	Muted             bool       `gorm:"column:muted;default:false;not null;comment:静音 0否 1是"`
	FriendTags        string     `gorm:"column:friend_tags;type:TEXT;comment:好友标签(JSON数组)"`
	CreatedAt         *time.Time `gorm:"column:created_at;type:TIMESTAMP;comment:好友关系建立时间"`
}

func (Friend) TableName() string {
	return "friends"
}

// FriendRequest 好友请求表
type FriendRequest struct {
	ID           uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	UserID       uint64     `gorm:"column:user_id;not null;index:idx_user;comment:申请人用户ID"`
	ReqUID       uint64     `gorm:"column:req_uid;not null;comment:被申请人用户ID"`
	ReqMsg       string     `gorm:"column:req_msg;type:VARCHAR(255);comment:好友申请留言"`
	ReqTime      time.Time  `gorm:"column:req_time;type:TIMESTAMP;not null;comment:申请发起时间"`
	HandleResult *int       `gorm:"column:handle_result;size:8;comment:处理结果（0:待处理 1:同意 2:拒绝）"`
	HandleMsg    string     `gorm:"column:handle_msg;type:VARCHAR(255);comment:处理结果备注"`
	HandledAt    *time.Time `gorm:"column:handled_at;type:TIMESTAMP;comment:处理操作时间"`
	Status       *int       `gorm:"column:status;type:INT;comment:消息状态（0:已删除 1:正常显示 2:忽略不显示）"`
	ReadState    int        `gorm:"column:read_state;size:8;default:0;not null;comment:读取状态（0:未读 1:已读）"`
	ReceiverRead int        `gorm:"column:receiver_read;size:8;default:0;not null;comment:接收方已读（0:未读 1:已读）"`
	SenderRead   int        `gorm:"column:sender_read;size:8;default:0;not null;comment:发起方已读处理结果（0:未读 1:已读）"`
	Remark       string     `gorm:"column:remark;type:VARCHAR(64);default:'';not null;comment:申请人为对方预设的备注"`
	ActiveKey    *string    `gorm:"column:active_key;size:160;uniqueIndex:uk_friend_requests_active_key;comment:待处理申请唯一键"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}

// Group 群组信息表
type Group struct {
	ID              uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	Name            string     `gorm:"column:name;type:VARCHAR(255);not null;comment:群名称"`
	Icon            string     `gorm:"column:icon;type:VARCHAR(255);not null;comment:群头像URL"`
	Description     string     `gorm:"column:description;type:VARCHAR(255);not null;default:'';comment:群描述"`
	Status          *int       `gorm:"column:status;size:8;comment:群状态（0:正常 1:已解散 2:封禁）"`
	CreatorUID      uint64     `gorm:"column:creator_uid;not null;comment:群主用户ID"`
	GroupType       int        `gorm:"column:group_type;type:INT;not null;comment:群类型（1:普通群 2:企业群 3:粉丝群...）"`
	IsVerify        int        `gorm:"column:is_verify;size:8;not null;comment:入群验证（0:不需要 1:需要）"`
	Notification    string     `gorm:"column:notification;type:VARCHAR(255);comment:群公告内容"`
	NotificationUID *uint64    `gorm:"column:notification_uid;comment:最后更新公告的用户ID"`
	CreatedAt       *time.Time `gorm:"column:created_at;type:TIMESTAMP;comment:创建时间"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;type:TIMESTAMP;comment:最后更新时间"`
}

func (Group) TableName() string {
	return "groups"
}

// GroupMember 群成员表
type GroupMember struct {
	ID          uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	GroupID     uint64     `gorm:"column:group_id;not null;uniqueIndex:uk_member,priority:1;comment:关联群ID"`
	UserID      uint64     `gorm:"column:user_id;not null;uniqueIndex:uk_member,priority:2;comment:成员用户ID"`
	RoleLevel   int        `gorm:"column:role_level;size:8;not null;comment:成员角色（0:普通成员 1:管理员 2:群主）"`
	JoinTime    *time.Time `gorm:"column:join_time;type:TIMESTAMP;comment:加入群聊时间"`
	JoinSource  *int       `gorm:"column:join_source;size:8;comment:加入来源（1:扫码 2:邀请 3:搜索...）"`
	InviterUID  *uint64    `gorm:"column:inviter_uid;comment:邀请人用户ID"`
	OperatorUID *uint64    `gorm:"column:operator_uid;comment:操作人用户ID"`
	GroupNick   string     `gorm:"column:group_nickname;type:VARCHAR(64);not null;default:'';comment:群内昵称"`
	GroupRemark string     `gorm:"column:group_remark;type:VARCHAR(255);not null;default:'';comment:群备注（仅自己可见）"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// GroupRequest 加群请求表
type GroupRequest struct {
	ID                 uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	ReqID              string     `gorm:"column:req_id;type:VARCHAR(64);not null;index:idx_group_request_lookup,priority:2;comment:业务请求ID（唯一标识）"`
	GroupID            uint64     `gorm:"column:group_id;not null;index:idx_group;index:idx_group_request_lookup,priority:1;comment:目标群ID"`
	ReqMsg             string     `gorm:"column:req_msg;type:VARCHAR(255);comment:入群申请留言"`
	ReqTime            *time.Time `gorm:"column:req_time;type:TIMESTAMP;comment:申请时间"`
	JoinSource         *int       `gorm:"column:join_source;size:8;comment:申请来源（1:扫码 2:邀请 3:搜索...）"`
	InviterUserID      *uint64    `gorm:"column:inviter_user_id;comment:邀请人ID"`
	HandleUserID       *uint64    `gorm:"column:handle_user_id;comment:请求处理人ID"`
	HandleTime         *time.Time `gorm:"column:handle_time;type:TIMESTAMP;comment:处理时间"`
	HandleResult       *int       `gorm:"column:handle_result;size:8;index:idx_group_request_lookup,priority:3;comment:处理结果（0:待处理 1:同意 2:拒绝）"`
	ReceiverRead       int        `gorm:"column:receiver_read;size:8;default:0;not null;comment:接收方(群主/管理员)已读 0未读 1已读"`
	ActiveKey          *string    `gorm:"column:active_key;size:160;uniqueIndex:uk_group_requests_active_key;comment:主动待处理申请唯一键"`
	SourceType         int        `gorm:"column:source_type;not null;default:1;comment:来源类型 1主动申请 2成员邀请"`
	SourceInvitationID *uint64    `gorm:"column:source_invitation_id;uniqueIndex:uk_group_requests_source_invitation;comment:来源邀请ID"`
	ActualJoinSource   *int       `gorm:"column:actual_join_source;comment:最终实际入群来源"`
	InvalidReason      string     `gorm:"column:invalid_reason;size:128;not null;default:'';comment:系统失效原因"`
	HandleMsg          string     `gorm:"column:handle_msg;size:255;not null;default:'';comment:审批附言或拒绝原因"`
}

func (GroupRequest) TableName() string {
	return "group_requests"
}

// GroupInvitation 独立的成员邀请记录。Social 服务的用户和群主键均为 uint64，
// 因此邀请双方继续使用数值 UID，不在迁移层引入字符串身份类型。
type GroupInvitation struct {
	ID                  uint64     `gorm:"primaryKey;column:id;autoIncrement"`
	GroupID             uint64     `gorm:"column:group_id;not null;index:idx_group_invitation_group_invitee,priority:1"`
	InviterUID          uint64     `gorm:"column:inviter_uid;not null"`
	InviteeUID          uint64     `gorm:"column:invitee_uid;not null;index:idx_group_invitation_invitee,priority:2;index:idx_group_invitation_group_invitee,priority:2"`
	InviterRoleSnapshot int        `gorm:"column:inviter_role_snapshot;not null;default:0"`
	Message             string     `gorm:"column:message;size:255;not null;default:''"`
	Status              int        `gorm:"column:status;not null;default:0;index:idx_group_invitation_invitee,priority:3;index:idx_group_invitation_group_invitee,priority:3;index:idx_group_invitation_expiry,priority:1"`
	RejectReason        string     `gorm:"column:reject_reason;size:255;not null;default:''"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;index:idx_group_invitation_invitee,priority:4"`
	HandledAt           *time.Time `gorm:"column:handled_at"`
	ExpiresAt           time.Time  `gorm:"column:expires_at;not null;index:idx_group_invitation_expiry,priority:2"`
}

func (GroupInvitation) TableName() string {
	return "group_invitations"
}

// SocialRequestReceipt stores per-user read and action state for social requests.
type SocialRequestReceipt struct {
	ID           uint64     `gorm:"primaryKey;column:id;autoIncrement"`
	RequestType  string     `gorm:"column:request_type;size:16;not null;uniqueIndex:uk_social_request_receipt,priority:1;index:idx_social_receipt_request,priority:1;index:idx_social_receipt_unread,priority:3;index:idx_social_receipt_actionable,priority:3"`
	RequestID    uint64     `gorm:"column:request_id;not null;uniqueIndex:uk_social_request_receipt,priority:2;index:idx_social_receipt_request,priority:2"`
	ReceiverID   string     `gorm:"column:receiver_id;size:64;not null;uniqueIndex:uk_social_request_receipt,priority:3;index:idx_social_receipt_unread,priority:1;index:idx_social_receipt_actionable,priority:1"`
	ReceiptKind  string     `gorm:"column:receipt_kind;size:16;not null;uniqueIndex:uk_social_request_receipt,priority:4"`
	IsRead       int        `gorm:"column:is_read;not null;default:0;index:idx_social_receipt_unread,priority:2"`
	IsActionable int        `gorm:"column:is_actionable;not null;default:0;index:idx_social_receipt_actionable,priority:2"`
	Result       int        `gorm:"column:result;not null;default:0"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	ReadAt       *time.Time `gorm:"column:read_at"`
	ResolvedAt   *time.Time `gorm:"column:resolved_at"`
}

func (SocialRequestReceipt) TableName() string {
	return "social_request_receipts"
}

// SocialNotificationOutbox stores notifications in the same transaction as request changes.
type SocialNotificationOutbox struct {
	ID          uint64     `gorm:"primaryKey;column:id;autoIncrement"`
	NotifyType  string     `gorm:"column:notify_type;size:64;not null;uniqueIndex:uk_social_notification,priority:1"`
	ReceiverID  string     `gorm:"column:receiver_id;size:64;not null;uniqueIndex:uk_social_notification,priority:2"`
	ActorID     string     `gorm:"column:actor_id;size:64;not null"`
	BizID       string     `gorm:"column:biz_id;size:128;not null;uniqueIndex:uk_social_notification,priority:3"`
	GroupID     string     `gorm:"column:group_id;size:64;not null;default:''"`
	Payload     string     `gorm:"column:payload;type:TEXT;not null"`
	Status      int        `gorm:"column:status;not null;default:0;index:idx_social_notification_retry,priority:1"`
	Attempts    int        `gorm:"column:attempts;not null;default:0"`
	NextRetryAt *time.Time `gorm:"column:next_retry_at;index:idx_social_notification_retry,priority:2"`
	LastError   string     `gorm:"column:last_error;size:512;not null;default:''"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	SentAt      *time.Time `gorm:"column:sent_at"`
}

func (SocialNotificationOutbox) TableName() string {
	return "social_notification_outbox"
}

// GroupInviteLink 群邀请链接表（链接/二维码统一）
type GroupInviteLink struct {
	ID        uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键"`
	GroupID   uint64     `gorm:"column:group_id;not null;index:idx_group;comment:群ID"`
	Token     string     `gorm:"column:token;type:VARCHAR(64);not null;uniqueIndex:uk_token;comment:邀请token（唯一）"`
	CreatedBy uint64     `gorm:"column:created_by;not null;index:idx_creator;comment:创建人"`
	ExpireAt  *time.Time `gorm:"column:expire_at;type:TIMESTAMP;comment:过期时间（NULL=永不过期）"`
	MaxUses   uint       `gorm:"column:max_uses;not null;default:0;comment:最大可用次数（0=无限）"`
	UsedCount uint       `gorm:"column:used_count;not null;default:0;comment:已使用次数"`
	Revoked   bool       `gorm:"column:revoked;not null;default:false;comment:是否撤销 0否 1是"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:TIMESTAMP;comment:撤销时间"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:创建时间"`
}

func (GroupInviteLink) TableName() string {
	return "group_invite_links"
}

// GroupMemberSetting 群成员资料/设置
type GroupMemberSetting struct {
	ID          uint64     `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:自增主键"`
	GroupID     uint64     `gorm:"column:group_id;type:INT UNSIGNED;not null;uniqueIndex:uk_group_user,priority:1;comment:群ID"`
	UserID      uint64     `gorm:"column:user_id;type:INT UNSIGNED;not null;uniqueIndex:uk_group_user,priority:2;index:idx_user;comment:用户ID"`
	GroupNick   string     `gorm:"column:group_nickname;type:VARCHAR(64);comment:我在本群的昵称"`
	GroupRemark string     `gorm:"column:group_remark;type:VARCHAR(255);comment:群备注（仅自己可见）"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (GroupMemberSetting) TableName() string {
	return "group_member_settings"
}

// GroupAnnouncement 群公告历史
type GroupAnnouncement struct {
	ID        uint64     `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:自增主键"`
	GroupID   uint64     `gorm:"column:group_id;type:INT UNSIGNED;not null;index:idx_group;comment:群ID"`
	Content   string     `gorm:"column:content;type:VARCHAR(1024);not null;comment:公告内容"`
	CreatedBy uint64     `gorm:"column:created_by;type:INT UNSIGNED;not null;index:idx_creator;comment:发布人"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:发布时间"`
	Pinned    bool       `gorm:"column:pinned;type:TINYINT(1);not null;default:0;comment:是否置顶"`
	PinnedAt  *time.Time `gorm:"column:pinned_at;type:TIMESTAMP;comment:置顶时间"`
	Deleted   bool       `gorm:"column:deleted;type:TINYINT(1);not null;default:0;comment:是否删除"`
}

func (GroupAnnouncement) TableName() string {
	return "group_announcements"
}

// FriendReport 好友举报表
type FriendReport struct {
	ID          uint64     `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:自增主键"`
	ReporterUID uint64     `gorm:"column:reporter_uid;type:BIGINT UNSIGNED;not null;index:idx_reporter;comment:举报人"`
	TargetUID   uint64     `gorm:"column:target_uid;type:BIGINT UNSIGNED;not null;index:idx_target;comment:被举报人"`
	Reason      string     `gorm:"column:reason;type:VARCHAR(512);comment:举报原因"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:创建时间"`
}

func (FriendReport) TableName() string {
	return "friend_reports"
}

// RelationOutbox 关系变更事务性发件箱。
// 关系变更（踢人/退群/解散/删好友）与本表插入在同一事务提交，relay 再异步投递到 Kafka，
// 保证"DB 变更 ⇔ 事件不丢"。自增 id 兼作单调版本号（消费端版本门用）。
type RelationOutbox struct {
	ID        uint64     `gorm:"primaryKey;column:id;autoIncrement;comment:自增主键，兼作单调版本号"`
	EventType string     `gorm:"column:event_type;type:VARCHAR(64);not null;comment:事件类型 group.member.removed/group.disbanded/friend.deleted"`
	GroupID   string     `gorm:"column:group_id;type:VARCHAR(64);not null;default:'';index:idx_group;comment:群ID（好友事件留空，作分区/索引）"`
	Payload   string     `gorm:"column:payload;type:TEXT;not null;comment:事件JSON载荷"`
	Status    int        `gorm:"column:status;size:8;not null;default:0;index:idx_status;comment:0待投递 1已投递"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:创建时间"`
	SentAt    *time.Time `gorm:"column:sent_at;type:TIMESTAMP;comment:投递完成时间"`
}

func (RelationOutbox) TableName() string {
	return "relation_outbox"
}
