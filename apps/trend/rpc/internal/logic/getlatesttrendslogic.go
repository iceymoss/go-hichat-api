package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetLatestTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTrendsLogic {
	return &GetLatestTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLatestTrends 获取最新动态
func (l *GetLatestTrendsLogic) GetLatestTrends(in *trend.GetLatestTrendsRequest) (*trend.GetLatestTrendsResponse, error) {
	list, err := l.svcCtx.Trend.List(l.ctx, int(in.LastTrendId), 0, in.UserIds, []string{"*"}, "id", -1)
	if err != nil {
		zLog.Error("GetLatestTrends.list: get trend list failed", zap.Any("trendId", in.LastTrendId), zap.Any("userId", in.UserIds), zap.Error(err))
		return nil, errors.New(fmt.Sprintf("GetLatestTrends.List: %v", err))
	}

	var lastId uint64
	if len(*list) > 0 {
		lastId = (*list)[len(*list)-1].Id
	}

	trendList := make([]*trend.Trend, 0, len(*list))
	for _, v := range *list {
		trendList = append(trendList, trendToResp(v))
	}

	// TODO: 获取评论和点赞数据

	hasMore := false
	if len(trendList) >= models.Limit {
		hasMore = true
	}

	return &trend.GetLatestTrendsResponse{
		Trends:      trendList,
		LastTrendId: int64(lastId),
		HasMore:     hasMore,
	}, nil
}

func trendToResp(v models.Trend) *trend.Trend {
	// 可见范围以 scope 列为准(1-仅自己,2-仅好友,3-所有人)
	// 历史数据 scope 未回填(0)时，用 circle_state 兜底派生
	readScope := trend.VisibilityScope(v.Scope)
	if readScope == trend.VisibilityScope_SCOPE_UNSPECIFIED {
		if v.CircleState == 1 {
			readScope = trend.VisibilityScope_FRIENDS // 仅好友
		} else {
			readScope = trend.VisibilityScope_PRIVATE // 仅自己
		}
	}

	var positionPoint []float32
	_ = json.Unmarshal([]byte(v.PositionPoint), &positionPoint)

	var atIdList []int32
	_ = json.Unmarshal([]byte(v.Idlist), &atIdList)

	var picArr []string
	_ = json.Unmarshal([]byte(v.PicArr), &picArr)

	return &trend.Trend{
		Id:            int64(v.Id),
		UserId:        int64(v.Userid),
		Type:          trend.TrendType(v.Type),
		Content:       v.Content,
		Scope:         readScope,
		CreateTime:    v.Createtime.Unix(),
		ReplyCount:    int32(v.ReplyCount),
		AgreeCount:    int32(v.AgreeCount),
		PositionName:  v.PositionName,
		PositionPoint: positionPoint,
		Title:         v.Title,
		AtUserIds:     atIdList,
		Resources:     picArr,
		CoverUrl:      v.Cover,
		ShareUrl:      v.ShareUrl,
		OpenReply:     int32(v.OpenReply),
		DeviceId:      v.Device,
		Ip:            v.Ip,
		IsTop:         int32(v.IsTop),
	}
}
