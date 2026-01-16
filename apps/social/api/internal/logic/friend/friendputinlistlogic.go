package friend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

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
	curUid := l.ctx.Value(Identify).(string)

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

	//Type: 0-待处理, 1-已通过, 2-已拒绝, 3-已忽略
	//Class: 0-我发起的申请列表, 1-我收到的申请列表
	res, err := l.svcCtx.Social.FriendPutInList(l.ctx, &social.FriendPutInListReq{
		UserId: curUid,
		Type:   reqType,
		Class:  class,
	})
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, err
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

		handleResult := int(v.HandleResult)
		list = append(list, &types.FriendRequests{
			Id:            int64(v.Id),
			UserId:        v.UserId,
			ReqUid:        v.ReqUid,
			ReqMsg:        reqMsg,
			MessageStatus: int(v.Status), // 消息状态（0:已删除 1:正常显示 2:忽略不显示）
			ReqTime:       v.ReqTime,
			HandleResult:  handleResult, // 处理结果（0:待处理 1:已同意 2:已拒绝 3:已忽略），确保0值也会返回
			Status:        status,
			StatusText:    statusText,
			HandleMsg:     "",
			ReadState:     int(v.ReadState),
		})
	}

	resp = &types.FriendPutInListResp{List: list}
	return
}
