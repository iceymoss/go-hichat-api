package constants

// 关系变更事件类型（social 生产 outbox -> relay 投 relationChangeTransfer topic -> task/mq 消费）。
// 字符串值进 Kafka 与 outbox.payload，演进时只增不改，保证跨版本兼容。
const (
	// RelationEventGroupMemberRemoved 群成员被移除（踢人/退群）
	RelationEventGroupMemberRemoved = "group.member.removed"
	// RelationEventGroupDisbanded 群解散（全员失效）
	RelationEventGroupDisbanded = "group.disbanded"
	// RelationEventFriendDeleted 好友删除（双向失效）
	RelationEventFriendDeleted = "friend.deleted"
	// RelationEventFriendAdded 好友新增（双向加入好友集，无版本门）
	RelationEventFriendAdded = "friend.added"
	// RelationEventGroupMemberAdded 群成员新增（入群/邀请/邀请链接，版本门）
	RelationEventGroupMemberAdded = "group.member.added"
)
