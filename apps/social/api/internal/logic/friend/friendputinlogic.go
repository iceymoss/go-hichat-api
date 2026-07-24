package friend

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	uid := ctxdata.GetUId(l.ctx)
	actor, err := strconv.ParseUint(uid, 10, 64)
	if err != nil || actor == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	target, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil || target == 0 {
		return nil, status.Error(codes.InvalidArgument, "target user id must be a positive integer")
	}
	if actor == target {
		return nil, status.Error(codes.InvalidArgument, "cannot send a friend request to yourself")
	}

	// 查询被申请用户是否存在
	usersResp, err := l.svcCtx.User.GetUserById(l.ctx, &user.GetUserByIdRequest{Id: req.UserId})
	if err != nil {
		zLog.Error("get user by id err", zap.Error(err))
		switch status.Code(err) {
		case codes.NotFound:
			return nil, status.Error(codes.NotFound, "target user does not exist or is unavailable")
		case codes.DeadlineExceeded:
			return nil, status.Error(codes.DeadlineExceeded, "user service deadline exceeded")
		case codes.Unavailable:
			return nil, status.Error(codes.Unavailable, "user service unavailable")
		default:
			return nil, status.Error(codes.Internal, "failed to validate target user")
		}
	}

	if usersResp == nil || usersResp.User == nil || usersResp.User.Status != 1 {
		return nil, status.Error(codes.NotFound, "target user does not exist or is unavailable")
	}

	// 使用中国时区的当前时间
	chinaNow := utils.NowInChina()
	rpcResp, err := l.svcCtx.Social.FriendPutIn(l.ctx, &social.FriendPutInReq{
		UserId:   uid,
		ActorUid: uid,
		ReqUid:   req.UserId,
		ReqMsg:   req.ReqMsg,
		Remark:   req.Remark,
		ReqTime:  utils.TimeToChinaUnix(chinaNow),
	})
	if err != nil {
		zLog.Error("req friend err", zap.Error(err))
		return nil, err
	}

	return &types.FriendPutInResp{
		RequestId: strconv.FormatInt(rpcResp.RequestId, 10), Status: rpcResp.Status,
		AlreadyPending: rpcResp.AlreadyPending, AlreadyFriend: rpcResp.AlreadyFriend,
	}, nil
}
