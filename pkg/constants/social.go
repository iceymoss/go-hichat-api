package constants

// HandlerResult 处理结果
type HandlerResult int

const (
	NoHandlerResult     HandlerResult = iota // 0-待处理
	PassHandlerResult                        // 1-已同意
	RefuseHandlerResult                      // 2-已拒绝
	IgnoreHandlerResult                      // 3-已忽略
	CancelHandlerResult                      // 4-已取消（保留，暂未使用）
)

// GroupRoleLevel 群等级 2. 创建者，1. 管理者，0. 普通
type GroupRoleLevel int

const (

	// AtLargeGroupRoleLevel 普通成员
	AtLargeGroupRoleLevel GroupRoleLevel = iota

	// ManagerGroupRoleLevel 管理员
	ManagerGroupRoleLevel

	// CreatorGroupRoleLevel 群主
	CreatorGroupRoleLevel
)

// GroupJoinSource 进群申请的方式： 1. 邀请， 2. 申请
type GroupJoinSource int

const (

	// PutInGroupJoinSource 申请入群
	PutInGroupJoinSource GroupJoinSource = iota + 1

	// InviteGroupJoinSource 邀请入群
	InviteGroupJoinSource
)

const GroupRequests = "group_requests"
