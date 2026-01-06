package friend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 好友列表
func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendListLogic) FriendList(req *types.FriendListReq) (resp *types.FriendListResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	res, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{
		UserId: uid,
	})
	if err != nil || res == nil {
		return nil, err
	}

	ids := make([]string, 0, len(res.List))
	for _, v := range res.List {
		ids = append(ids, v.FriendUid)
	}
	//获取用户信息
	userList, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Name:  "",
		Phone: "",
		Ids:   ids,
	})
	if err != nil || userList == nil {
		return nil, err
	}

	uidBindInfo := make(map[string]*user.UserEntity)
	for _, v := range userList.User {
		uidBindInfo[v.Id] = v
	}

	// 去重用户id
	dup := make(map[string]struct{}, len(res.List))
	list := make([]*types.Friends, 0, len(userList.User))
	for _, item := range res.List {
		v := item
		if user, ok := uidBindInfo[v.FriendUid]; ok {
			// 删除原来自动填充备注的逻辑，保持数据的真实性
			// if v.Remark == "" {
			// 	v.Remark = user.Nickname
			// }

			key := fmt.Sprintf("%s_%s", v.UserId, v.FriendUid)
			if _, ok := dup[key]; ok {
				continue
			}

			// 将用户ID作为id返回（而不是关系表的id）
			// 同时返回完整的用户信息
			item := &types.Friends{
				Id:        strconv.Itoa(int(v.Id)),
				FriendUid: user.Id, // 好友用户ID
				Nickname:  user.Nickname,
				Avatar:    user.Avatar,
				Remark:    v.Remark,
				// TODO: 待 social RPC 返回扩展字段后填充
				Sex:          user.Sex,
				Email:        user.Email,
				Phone:        user.Phone,
				Introduction: user.Introduction,
				Region:       user.Region,
				Occupation:   user.Occupation,
				Tags:         user.Tags,
				Status:       user.Status,
				Type:         user.Type,
				LastLogin:    user.LastLogin,
			}
			list = append(list, item)
			dup[key] = struct{}{}
		}
	}

	// 不在这里排序，由前端按拼音排序和分组（类似微信）
	// 前端会使用 sortFriendsByPinyin 和 groupFriendsByPinyin 进行正确的拼音排序

	resp = &types.FriendListResp{List: list}

	return
}
