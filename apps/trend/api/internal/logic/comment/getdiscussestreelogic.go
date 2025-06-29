package comment

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiscussesTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetDiscussesTreeLogic 获取评论树
func NewGetDiscussesTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiscussesTreeLogic {
	return &GetDiscussesTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDiscussesTreeLogic) GetDiscussesTree(req *types.GetDiscussesListReq) (resp *types.DiscussesTreeResp, err error) {
	discuss, err := l.svcCtx.Trend.GetTrendDiscusses(l.ctx, &trend.GetDiscussesListReq{
		TrendId: req.TrendID,
		LastId:  int32(req.LastID),
	})
	if err != nil {
		zLog.Error("获取动态失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 获取用户信息
	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: discuss.UserIds,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Any("userIds", discuss.UserIds), zap.Error(err))
		return nil, nil
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

	res := types.GetDiscussesListResp{
		DiscussesMap: map[uint64]types.TrendDiscusses{},
	}
	for k, v := range discuss.Discusses {
		if _, ok := res.DiscussesMap[uint64(k)]; !ok {
			res.DiscussesMap[uint64(k)] = types.TrendDiscusses{
				Discusses: make([]*types.Discuss, 0),
			}
		}
		// 一级评论
		discussTemp := res.DiscussesMap[uint64(k)]
		for _, dis := range v.List {
			// 处理二级评论
			seDisTemp := make([]*types.Discuss, 0, len(dis.Children))
			for _, secDis := range dis.Children {
				secDisTemp := discuss2Resp(secDis)
				secDisTemp.Replyer = userBind[strconv.Itoa(int(secDis.Replyer))]
				secDisTemp.User = userBind[strconv.Itoa(int(secDis.UserId))]
				seDisTemp = append(seDisTemp, secDisTemp)
			}
			rootTemp := discuss2Resp(dis)
			rootTemp.Children = seDisTemp
			rootTemp.Replyer = userBind[strconv.Itoa(int(dis.Replyer))]
			rootTemp.User = userBind[strconv.Itoa(int(dis.UserId))]
			discussTemp.Discusses = append(discussTemp.Discusses, rootTemp)
		}
		res.DiscussesMap[uint64(k)] = discussTemp
	}

	resp = &types.DiscussesTreeResp{
		Discusses: res,
	}

	return
}
