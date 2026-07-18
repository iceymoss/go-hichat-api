package friend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

// 好友申请列表
func NewFriendPutInListLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *FriendPutInListLogic {
	return &FriendPutInListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *FriendPutInListLogic) FriendPutInList(req *types.FriendPutInListReq) (resp *types.FriendPutInListResp, err error) {
	curUid := ctxdata.GetUId(l.ctx)

	// 从请求参数获取type和class，如果没有则使用默认值
	reqType := req.Type
	if reqType == 0 {
		// 如果请求中没有type，尝试从URL query获取（向后兼容）
		reqTypeStr := l.r.URL.Query().Get("type")
		if reqTypeStr != "" {
			var reqTypeInt int
			reqTypeInt, err = strconv.Atoi(reqTypeStr)
			if err == nil {
				reqType = int32(reqTypeInt)
			}
		}
	}

	class := req.Class
	if class == "" {
		// 如果请求中没有class，尝试从URL query获取（向后兼容）
		class = l.r.URL.Query().Get("class")
		if class == "" {
			class = "1" // 默认值：我收到的申请
		}
	}
	statusFilter := req.Status
	if statusFilter == nil && l.r.URL.Query().Has("type") {
		statusFilter = &reqType
	}
	legacyType := reqType
	if statusFilter == nil {
		legacyType = -1
	}

	//Type: 0-待处理, 1-已通过, 2-已拒绝, 3-已忽略
	//Class: 0-我发起的申请列表, 1-我收到的申请列表
	res, err := l.svcCtx.Social.FriendPutInList(l.ctx, &social.FriendPutInListReq{
		UserId: curUid,
		Type:   legacyType,
		Class:  class,
		Status: statusFilter,
		Page:   req.Page,
		Size:   req.Size,
	})
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, err
	}

	// 收集需要获取用户信息的用户ID列表
	// 根据数据库schema：user_id是申请人，req_uid是被申请人
	// class=1（我收到的申请）：req_uid是我，user_id是对方（申请人），应该获取user_id的用户信息
	// class=0（我发起的申请）：user_id是我，req_uid是对方（被申请人），应该获取req_uid的用户信息
	userIds := make([]string, 0, len(res.List))
	for _, v := range res.List {
		if class == "1" {
			// 我收到的申请，获取申请人的信息（user_id，对方的信息）
			userIds = append(userIds, v.UserId)
		} else {
			// 我发起的申请，获取被申请人的信息（req_uid，对方的信息）
			userIds = append(userIds, v.ReqUid)
		}
	}

	// 批量获取用户信息
	uidBindInfo := make(map[string]*user.UserEntity)
	if len(userIds) > 0 {
		userList, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
			Name:  "",
			Phone: "",
			Ids:   userIds,
		})
		if err != nil {
			return nil, err
		}
		if userList != nil {
			for _, u := range userList.User {
				uidBindInfo[u.Id] = u
			}
		}
	}

	list := make([]*types.FriendRequests, 0, len(res.List))
	for _, v := range res.List {
		// 根据 HandleResult 设置状态文本
		var status, statusText string
		switch v.HandleResult {
		case 0:
			status = "pending"
			statusText = "待处理"
		case 1:
			status = "accepted"
			statusText = "已同意"
		case 2:
			status = "rejected"
			statusText = "已拒绝"
		case 3:
			status = "ignored"
			statusText = "已忽略"
		default:
			status = "pending"
			statusText = "待处理"
		}

		// 根据 status 字段控制消息显示：status=2（忽略）时不返回消息
		reqMsg := v.ReqMsg
		if v.Status == 2 {
			reqMsg = "" // status=2 时不显示消息
		}

		// 确定需要获取的用户ID
		// 根据数据库schema：user_id是申请人，req_uid是被申请人
		// class=1（我收到的申请）：应该获取user_id的用户信息（申请人的信息，对方）
		// class=0（我发起的申请）：应该获取req_uid的用户信息（被申请人的信息，对方）
		var targetUserId string
		if class == "1" {
			targetUserId = v.UserId // 我收到的申请，获取申请人的信息（user_id，对方）
		} else {
			targetUserId = v.ReqUid // 我发起的申请，获取被申请人的信息（req_uid，对方）
		}

		// 获取用户信息
		var userInfo *user.UserEntity
		if u, ok := uidBindInfo[targetUserId]; ok {
			userInfo = u
		}

		handleResult := int(v.HandleResult)
		item := &types.FriendRequests{
			Id:            int64(v.Id),
			UserId:        v.UserId,
			ReqUid:        v.ReqUid,
			ReqMsg:        reqMsg,
			MessageStatus: int(v.Status), // 消息状态（0:已删除 1:正常显示 2:忽略不显示）
			ReqTime:       v.ReqTime,
			HandleResult:  handleResult, // 处理结果（0:待处理 1:已同意 2:已拒绝 3:已忽略），确保0值也会返回
			Status:        status,
			StatusText:    statusText,
			HandleMsg:     v.HandleMsg,
			ReadState:     int(v.ReadState),
			RequestId:     v.RequestId,
			PeerUid:       v.PeerUid,
			HandledAt:     v.HandledAt,
		}

		// 填充用户信息
		if userInfo != nil {
			item.Nickname = userInfo.Nickname
			item.Avatar = userInfo.Avatar
			item.Sex = userInfo.Sex
			item.Email = userInfo.Email
			item.Phone = userInfo.Phone
			item.Introduction = userInfo.Introduction
			item.Region = userInfo.Region
			item.Occupation = userInfo.Occupation
			item.Tags = userInfo.Tags
			item.UserStatus = userInfo.Status
			item.UserType = userInfo.Type
			item.LastLogin = userInfo.LastLogin
		}

		list = append(list, item)
	}

	resp = &types.FriendPutInListResp{List: list, Total: res.Total}
	return
}
