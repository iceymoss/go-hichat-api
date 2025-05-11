package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInListLogic {
	return &FriendPutInListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInList 获取未处理的好友申请列表,或者获取我发起的申请好友列表
func (l *FriendPutInListLogic) FriendPutInList(in *social.FriendPutInListReq) (*social.FriendPutInListResp, error) {
	friendReqList, err := l.svcCtx.FriendRequestsModel.ListFilterHandler(l.ctx, in.UserId, in.Type, in.Class)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find list friend req err %v req %v", err, in.UserId)
	}

	resp := make([]*social.FriendRequests, 0, len(friendReqList))
	for _, v := range friendReqList {
		resp = append(resp, &social.FriendRequests{
			Id:           int32(v.Id),
			UserId:       strconv.Itoa(int(v.UserId)),
			ReqUid:       strconv.Itoa(int(v.ReqUid)),
			ReqMsg:       v.ReqMsg,
			ReqTime:      v.ReqTime.Unix(),
			HandleResult: int32(v.HandleResult),
		})
	}

	return &social.FriendPutInListResp{
		List: resp,
	}, nil
}
