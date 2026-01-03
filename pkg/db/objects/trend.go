package objects

import (
	"database/sql"
	"time"
)

// Trend 动态表
type Trend struct {
	ID            uint64       `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:动态ID"`
	UserID        uint64       `gorm:"column:userid;type:BIGINT UNSIGNED;not null;index:idx_userid_circle,priority:1;comment:发表用户ID"`
	Type          uint64       `gorm:"column:type;type:TINYINT UNSIGNED;not null;comment:动态类型：1文本，2混合(图片)，3长文，4第三方分享(如B站视频)，5视频，6置顶广告"`
	Content       string       `gorm:"column:content;type:TEXT;not null;comment:动态内容"`
	PositionName  string       `gorm:"column:position_name;type:VARCHAR(255);default:'';comment:位置名称"`
	PositionPoint string       `gorm:"column:position_point;type:VARCHAR(30);comment:位置信息（使用JSON数组）"`
	ReplyCount    uint64       `gorm:"column:reply_count;type:BIGINT UNSIGNED;default:0;comment:评论数量"`
	AgreeCount    uint64       `gorm:"column:agree_count;type:BIGINT UNSIGNED;default:0;comment:点赞数量"`
	Createtime    time.Time    `gorm:"column:createtime;type:DATETIME;not null;default:CURRENT_TIMESTAMP;index:idx_circle_time;index:idx_userid_circle,priority:2;comment:原始创建时间"`
	Updatetime    time.Time    `gorm:"column:updatetime;type:DATETIME;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:最后更新时间"`
	CircleState   int64        `gorm:"column:circle_state;type:BIGINT;not null;default:1;comment:朋友圈状态：2-不可见，1-可见，0-朋友圈删除"`
	State         int64        `gorm:"column:state;type:BIGINT;not null;default:0;comment:是否删除 0-删除，1-正常"`
	IsAd          int64        `gorm:"column:is_ad;type:BIGINT;default:0;comment:是否广告：0-普通，1-广告"`
	URL           string       `gorm:"column:url;type:VARCHAR(255);default:'';comment:广告/视频链接（类型5使用）"`
	AdEndTime     sql.NullTime `gorm:"column:ad_end_time;type:DATETIME;comment:广告展示截止时间"`
	OpenReply     int64        `gorm:"column:open_reply;type:BIGINT;default:1;comment:是否开启评论：0-关闭，1-开启"`
	IsTop         int64        `gorm:"column:is_top;type:BIGINT;default:0;comment:是否置顶：0-否，1-是"`
	Title         string       `gorm:"column:title;type:VARCHAR(255);default:'';comment:长文标题（类型3使用）"`
	IDList        string       `gorm:"column:idlist;type:VARCHAR(100);comment:@用户ID列表（使用JSON数组）"`
	PicArr        string       `gorm:"column:pic_arr;type:TEXT;comment:图片url或者视频"`
	ShareURL      string       `gorm:"column:share_url;type:VARCHAR(255);default:'';comment:第三方内容URL（类型4使用）"`
	Cover         string       `gorm:"column:cover;type:VARCHAR(255);default:'';comment:封面图URL（类型3/5使用）"`
	IP            string       `gorm:"column:ip;type:VARCHAR(45);default:'';comment:发布者IP地址"`
	Device        string       `gorm:"column:device;type:VARCHAR(100);default:'';comment:发布设备标识"`
}

func (Trend) TableName() string {
	return "trend"
}

// TrendAgree 动态点赞表
type TrendAgree struct {
	ID         uint64    `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:主键"`
	UserID     uint64    `gorm:"column:userid;type:INT UNSIGNED;not null;uniqueIndex:uniq_user_trend,priority:1;index:idx_author_read_state,priority:1;comment:点赞用户ID"`
	AuthorID   uint64    `gorm:"column:author_id;type:INT UNSIGNED;not null;index:idx_author_read_state,priority:2;comment:被赞动态的作者ID"`
	TrendID    uint64    `gorm:"column:trend_id;type:INT UNSIGNED;not null;uniqueIndex:uniq_user_trend,priority:2;index:idx_trend_state,priority:1;comment:动态ID"`
	OpTime     time.Time `gorm:"column:op_time;type:DATETIME;not null;default:CURRENT_TIMESTAMP;comment:操作时间（点赞/取消时间）"`
	State      uint64    `gorm:"column:state;type:TINYINT UNSIGNED;not null;default:1;index:idx_author_read_state,priority:3;index:idx_trend_state,priority:2;comment:点赞状态 0:取消，1:有效"`
	IsRead     uint64    `gorm:"column:is_read;type:TINYINT UNSIGNED;not null;default:1;comment:已读状态 0:已读，1:未读"`
	CreateTime time.Time `gorm:"column:create_time;type:DATETIME;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time `gorm:"column:update_time;type:DATETIME;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:最后更新时间"`
}

func (TrendAgree) TableName() string {
	return "trend_agree"
}

// TrendDiscuss 动态评论表
type TrendDiscuss struct {
	ID           uint64    `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:主键ID"`
	TrendID      int       `gorm:"column:trendid;type:INT;not null;default:0;index:idx_trendid;index:idx_trendid_createtime,priority:1;index:idx_trendid_updatetime,priority:1;comment:所属动态ID（一级评论关联的源动态）"`
	Father       int       `gorm:"column:father;type:INT;not null;default:0;index:idx_father;index:idx_father_level,priority:1;comment:父评论ID（0表示一级评论）"`
	RootID       uint64    `gorm:"column:rootid;type:INT UNSIGNED;not null;default:0;index:idx_rootid;comment:根评论ID（多级评论的祖先ID）"`
	Replyer      uint64    `gorm:"column:replyer;type:INT UNSIGNED;not null;index:idx_replyer;comment:评论者用户ID"`
	UserID       uint64    `gorm:"column:userid;type:INT UNSIGNED;not null;index:idx_userid;comment:被评论者用户ID"`
	Level        int       `gorm:"column:level;type:TINYINT;not null;default:1;index:idx_level;index:idx_father_level,priority:2;index:idx_path_level,priority:2;comment:评论层级（1=一级评论，2=二级评论...）"`
	Content      string    `gorm:"column:content;type:VARCHAR(2000);not null;comment:评论内容"`
	IDList       string    `gorm:"column:idlist;type:VARCHAR(255);not null;default:'';comment:@用户的ID列表"`
	AgreeCount   uint64    `gorm:"column:agree_count;type:INT UNSIGNED;not null;default:0;comment:点赞数"`
	DiscussCount int       `gorm:"column:discuss_count;type:INT;not null;default:0;comment:子评论数量"`
	State        uint64    `gorm:"column:state;type:TINYINT UNSIGNED;not null;default:1;index:idx_state;index:idx_updatetime_state,priority:2;comment:状态（0=已删除, 1=正常）"`
	Read         int       `gorm:"column:read;type:TINYINT;not null;default:1;comment:已读状态（0=已读, 1=未读）"`
	Path         string    `gorm:"column:path;type:VARCHAR(1000);not null;default:'/';index:idx_path_level,priority:1;index:idx_path;comment:评论路径"`
	Createtime   time.Time `gorm:"column:createtime;type:DATETIME;not null;default:CURRENT_TIMESTAMP;index:idx_createtime;index:idx_trendid_createtime,priority:2;comment:创建时间"`
	Updatetime   time.Time `gorm:"column:updatetime;type:DATETIME;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;index:idx_trendid_updatetime,priority:2;index:idx_updatetime_state,priority:1;comment:最后更新时间"`
}

func (TrendDiscuss) TableName() string {
	return "trend_discuss"
}
