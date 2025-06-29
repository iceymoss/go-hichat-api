package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDiscussesLogic {
	return &GetTrendDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTrendDiscusses 获取动态的多级评论树（树形结构）
func (l *GetTrendDiscussesLogic) GetTrendDiscusses(in *trend.GetDiscussesListReq) (*trend.GetDiscussesListResp, error) {
	// 获取动态一级评论
	// 获取一级评论（使用 key set 分页）
	discusses, err := l.svcCtx.TrendDiscuss.FindFirstByTrendIds(l.ctx, in.TrendId, uint64(in.LastId))
	if err != nil {
		zLog.Error("GetRootDiscusses.FindByTrendWithKeyset: 获取一级评论失败", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, errors.New("获取评论失败")
	}

	uids := make([]string, 0, len(discusses))

	fatherIds := make([]uint64, 0, len(discusses))
	for _, v := range discusses {
		fatherIds = append(fatherIds, v.Id)
	}

	// 获取子评论
	child, err := l.svcCtx.TrendDiscuss.FindChildrenByFathers(l.ctx, fatherIds, 0, 30)
	childBind := make(map[int64][]*trend.Discuss, len(child))
	for _, v := range child {
		if _, ok := childBind[v.Father]; !ok {
			childBind[v.Father] = make([]*trend.Discuss, 0)
		}
		uids = append(uids, []string{strconv.Itoa(int(v.Replyer)), strconv.Itoa(int(v.Userid))}...)
		childBind[v.Father] = append(childBind[v.Father], convertToReply(v))
	}

	// 一级评论
	// 动态id => {[
	//   一级评论 => {[二级评论]}
	//]}
	//
	listBind := make(map[uint32]*trend.TrendDiscusses, len(in.TrendId))
	for _, v := range discusses {
		if _, ok := listBind[uint32(v.Trendid)]; !ok {
			listBind[uint32(v.Trendid)] = &trend.TrendDiscusses{
				List: make([]*trend.Discuss, 0),
			}
		}
		uids = append(uids, []string{strconv.Itoa(int(v.Replyer)), strconv.Itoa(int(v.Userid))}...)
		// 加入二级评论
		rootDiscuss := convertToReply(v)
		rootDiscuss.Children = childBind[int64(v.Id)]
		listBind[uint32(v.Trendid)].List = append(listBind[uint32(v.Trendid)].List, rootDiscuss)
	}

	str, err := json.Marshal(listBind)
	fmt.Println("data:", string(str))

	return &trend.GetDiscussesListResp{
		Discusses: listBind,
		UserIds:   uids,
	}, nil
}
