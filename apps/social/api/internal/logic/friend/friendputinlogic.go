package friend

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

const Identify = "hichat2.com"

type FriendPutInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFriendPutInLogic 好友申请
func NewFriendPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInLogic {
	return &FriendPutInLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInLogic) FriendPutIn(req *types.FriendPutInReq) (resp *types.FriendPutInResp, err error) {

	// 查询被申请用户是否存在
	usersResp, err := l.svcCtx.User.GetUserById(l.ctx, &user.GetUserByIdRequest{Id: req.UserId})
	if err != nil {
		zLog.Error("get user by id err", zap.Error(err))
		return nil, err
	}

	if usersResp != nil && usersResp.User != nil && usersResp.User.Status != 1 {
		return nil, libErr.New(xerr.ErrNotFound, "用户不存在")
	}

	uid := l.ctx.Value(Identify).(string)
	_, err = l.svcCtx.Social.FriendPutIn(l.ctx, &social.FriendPutInReq{
		UserId:  uid,
		ReqUid:  req.UserId,
		ReqMsg:  req.ReqMsg,
		ReqTime: time.Now().Unix(),
	})
	if err != nil {
		zLog.Error("req friend err", zap.Error(err))
		return nil, err
	}

	return
}
