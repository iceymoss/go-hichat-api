package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/bitmap"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatLogReadRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatLogReadRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatLogReadRecordsLogic {
	return &GetChatLogReadRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetChatLogReadRecords 根据消息id获取用户消息阅读情况
func (l *GetChatLogReadRecordsLogic) GetChatLogReadRecords(req *types.GetChatLogReadRecordsReq) (resp *types.GetChatLogReadRecordsResp, err error) {
	chatlogs, err := l.svcCtx.IM.GetChatLog(l.ctx, &im.GetChatLogReq{
		MsgId: req.MsgId,
	})

	if err != nil {
		return nil, err
	}
	if chatlogs == nil || len(chatlogs.List) == 0 {
		// 消息不存在时返回空列表而非 null，避免前端当成错误
		return &types.GetChatLogReadRecordsResp{
			Reads:   []types.ReadRecordUser{},
			UnReads: []types.ReadRecordUser{},
		}, nil
	}

	var (
		chatlog   = chatlogs.List[0]
		readIds   []string // 已读 userId
		unIds     []string // 未读 userId
		allIds    []string
		readTimes = chatlog.ReadTimes // userId → unix nano
	)
	if readTimes == nil {
		readTimes = map[string]int64{}
	}

	switch constants.ChatType(chatlog.ChatType) {
	case constants.SingleChatType:
		// 私聊：发送者本人不在展示范围，只看接收方状态
		if len(chatlog.ReadRecords) == 1 && chatlog.ReadRecords[0] == 0x01 {
			readIds = []string{chatlog.RecvId}
		} else {
			unIds = []string{chatlog.RecvId}
		}
		allIds = []string{chatlog.RecvId}
	case constants.GroupChatType:
		groupUsers, err := l.svcCtx.Social.GroupUsers(l.ctx, &socialclient.GroupUsersReq{
			GroupId: chatlog.RecvId,
		})
		if err != nil {
			return nil, err
		}
		bitmaps := bitmap.Load(chatlog.ReadRecords)
		for _, member := range groupUsers.List {
			// 发送者自己不在已读/未读列表中
			if member.UserId == chatlog.SendId {
				continue
			}
			allIds = append(allIds, member.UserId)
			if bitmaps.IsSet(member.UserId) {
				readIds = append(readIds, member.UserId)
			} else {
				unIds = append(unIds, member.UserId)
			}
		}
	}

	// 批量查询用户资料
	userEntitys, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{Ids: allIds})
	if err != nil {
		return nil, err
	}
	userEntitySet := make(map[string]*user.UserEntity, len(userEntitys.User))
	for i := range userEntitys.User {
		userEntitySet[userEntitys.User[i].Id] = userEntitys.User[i]
	}

	toEntries := func(ids []string, includeTime bool) []types.ReadRecordUser {
		out := make([]types.ReadRecordUser, 0, len(ids))
		for _, id := range ids {
			u := userEntitySet[id]
			nick, avatar := id, ""
			if u != nil {
				if u.Nickname != "" {
					nick = u.Nickname
				}
				avatar = u.Avatar
			}
			var readAt int64
			if includeTime {
				readAt = readTimes[id]
			}
			out = append(out, types.ReadRecordUser{Id: id, Nickname: nick, Avatar: avatar, ReadAt: readAt})
		}
		return out
	}

	return &types.GetChatLogReadRecordsResp{
		Reads:   toEntries(readIds, true),
		UnReads: toEntries(unIds, false),
	}, nil
}
