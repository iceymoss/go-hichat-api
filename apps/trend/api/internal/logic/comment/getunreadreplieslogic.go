package comment

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUnreadRepliesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUnreadRepliesLogic 获取未读回复通知
func NewGetUnreadRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadRepliesLogic {
	return &GetUnreadRepliesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetUnreadReplies 获取未读的评论和点赞
func (l *GetUnreadRepliesLogic) GetUnreadReplies(req *types.GetUnreadRepliesReq) (resp *types.RepliesListResp, err error) {
	uid := utils.GetUser(l.ctx)
	unRead, err := l.svcCtx.Trend.GetUnreadReplies(l.ctx, &trend.GetUnreadRepliesReq{
		UserId: uint64(uid),
		Pagination: &trend.Pagination{
			LastId:   int32(req.DiscussLasID),
			LastTime: int64(req.LastTime),
		},
	})
	if err != nil {
		zLog.Error("获取未读评论失败", zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	// 获取未读点赞
	unReadLike, err := l.svcCtx.Trend.GetUnreadLikes(l.ctx, &trend.GetUnreadLikesRequest{
		UserId: uint32(uid),
		LastId: uint32(req.LikeLastID),
	})
	if err != nil {
		zLog.Error("获取未读点赞失败", zap.Any("req", req), zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	userIds := make([]string, 0, len(unRead.Replies)+len(unReadLike.Likes))
	discussList := make([]*types.Discuss, 0, len(unRead.Replies))
	for _, v := range unRead.Replies {
		discussList = append(discussList, discuss2Resp(v))
		userIds = append(userIds, []string{strconv.Itoa(int(v.UserId)), strconv.Itoa(int(v.Replyer))}...)
	}

	likeList := make([]*types.Like, 0, len(unReadLike.Likes))
	for _, v := range unReadLike.Likes {
		likeList = append(likeList, like2Resp(v))
		userIds = append(userIds, strconv.Itoa(int(v.UserId)))
	}

	// 获取用户信息
	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userIds,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Any("userIds", userIds), zap.Error(err))
		return nil, err
	}

	userBind := make(map[string]*types.User, len(userInfo.User))
	for _, v := range userInfo.User {
		userBind[v.Id] = &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		}
	}

	for i, v := range unRead.Replies {
		discussList[i].User = userBind[strconv.Itoa(int(v.UserId))]
		discussList[i].Replyer = userBind[strconv.Itoa(int(v.Replyer))]
	}

	for i, v := range unReadLike.Likes {
		likeList[i].User = userBind[strconv.Itoa(int(v.UserId))]
	}

	resp = &types.RepliesListResp{
		Replies:      discussList,
		Likes:        likeList,
		LikeLastID:   int(unReadLike.LastId),
		DiscussLasID: int(unRead.Pagination.LastId),
	}
	return
}

func like2Resp(like *trend.LikeInfo) *types.Like {
	return &types.Like{
		Id:        int(like.Id),
		Trend_id:  uint64(like.TrendId),
		Like_time: uint64(like.CreateTime),
	}
}
