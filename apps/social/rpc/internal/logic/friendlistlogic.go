package logic

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendList 好友列表
func (l *FriendListLogic) FriendList(in *social.FriendListReq) (*social.FriendListResp, error) {
	friendsList, err := l.svcCtx.FriendsModel.ListByUserid(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list friend by uid err %v req %v ", err,
			in.UserId)
	}

	respList := make([]*social.Friends, 0, len(friendsList))
	for _, v := range friendsList {
		// 处理 FriendTags，如果是 NULL 则返回空数组
		var friendTags []string
		if v.FriendTags.Valid && v.FriendTags.String != "" {
			// 尝试解析 JSON 数组
			if err := json.Unmarshal([]byte(v.FriendTags.String), &friendTags); err != nil {
				// 如果解析失败，返回空数组
				friendTags = []string{}
			}
		}

		respList = append(respList, &social.Friends{
			Id:                int32(v.Id),
			UserId:            strconv.Itoa(int(v.UserId)),
			Remark:            v.Remark,
			AddSource:         int32(v.AddSource),
			FriendUid:         strconv.Itoa(int(v.FriendUid)),
			Blacklisted:       v.Blacklisted == 1,
			MomentsPermission: int32(v.MomentsPermission),
			NotifyEnabled:     v.NotifyEnabled == 1,
			Pinned:            v.Pinned == 1,
			Muted:             v.Muted == 1,
			FriendTags:        friendTags,
		})
	}

	return &social.FriendListResp{
		List: respList,
	}, nil
}
