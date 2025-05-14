package friend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"net/http"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

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
	reqTypeStr := l.r.URL.Query().Get("type")
	var reqTypeInt int
	reqTypeInt, err = strconv.Atoi(reqTypeStr)
	if err != nil {
		return
	}
	//Type 1：表示已经通过，2表示已决绝
	//Class 申请列表类型：0我发起的申请列表；1我接受到的好友申请列表
	class := l.r.URL.Query().Get("class")
	res, err := l.svcCtx.Social.FriendPutInList(l.ctx, &social.FriendPutInListReq{
		UserId: curUid,
		Type:   int32(reqTypeInt),
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
		list = append(list, &types.FriendRequests{
			Id:           int64(v.Id),
			UserId:       v.UserId,
			ReqUid:       v.ReqUid,
			ReqMsg:       v.ReqMsg,
			ReqTime:      v.ReqTime,
			HandleResult: int(v.HandleResult),
			HandleMsg:    "",
		})
	}

	resp = &types.FriendPutInListResp{List: list}
	return
}
