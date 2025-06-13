package logic

import (
	"context"
	"fmt"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/sensitive"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTrendLogic {
	return &CreateTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateTrend 发布动态
func (l *CreateTrendLogic) CreateTrend(in *trend.CreateTrendRequest) (*trend.CreateTrendResponse, error) {
	// 获取分布式锁，放置被刷，默认锁30s
	rdb := db.GetRedisConn()
	key := "trend:user:push:time"
	lastPushTimeStr, err := rdb.HGet(l.ctx, key, strconv.Itoa(int(in.UserId))).Result()
	if err != nil {
		return nil, err
	}

	lastPushTime := 0
	if lastPushTimeStr != "" {
		lastPushTime, err = strconv.Atoi(lastPushTimeStr)
	}

	// 30秒内只能发布一条
	if time.Now().Unix()-int64(lastPushTime) < 30 {
		return &trend.CreateTrendResponse{
			Code: 5000,
		}, status.Error(codes.InvalidArgument, "您动态发布得太频繁")
	}

	if len(in.Content) > 2000 {
		return &trend.CreateTrendResponse{
			Code: 4000,
		}, status.Error(codes.InvalidArgument, "内容超过2000字限制")
	}

	// 敏感词过滤
	pass, sensitiveWord, err := checkTrendContent(in.Content + in.Title)
	if err != nil {
		return nil, err
	}
	if !pass {
		return &trend.CreateTrendResponse{
			Code: 2000,
		}, errors.New(fmt.Sprintf("动态标题或内容包含敏感词:%s", sensitiveWord))
	}

	// 图片检查
	if !checkTrendPic(in.Resources) {
		return &trend.CreateTrendResponse{
			Code: 3000,
		}, errors.New("The image content is illegal")
	}

	// 发版 => 朋友圈：直接发
	trendTypeStr := in.Type.String()
	trendTypeInt, err := strconv.Atoi(trendTypeStr)
	if err != nil {
		return nil, err
	}

	circleState := 2                               //默认不可见
	if in.Scope == trend.VisibilityScope_FRIENDS { //朋友圈
		circleState = 1
	}

	var openReply int64
	if in.OpenReply {
		openReply = 1
	}

	// 处理图片或者视频
	positionPoint := make([]float64, 0, 2)
	positionPoint = append(positionPoint, in.PositionPoint.Latitude)
	positionPoint = append(positionPoint, in.PositionPoint.Longitude)

	id, err := l.svcCtx.Trend.Insert(l.ctx, &models.Trend{
		Userid:        uint64(in.UserId),
		Type:          uint64(trendTypeInt),
		Content:       in.Content,
		PositionName:  in.PositionName,
		PositionPoint: positionPoint,
		Createtime:    time.Now(),
		Updatetime:    time.Now(),
		CircleState:   int64(circleState),
		State:         1,
		OpenReply:     openReply,
		Title:         in.Title,
		PicArr:        in.Resources,
		ShareUrl:      in.ShareUrl,
		Cover:         in.CoverUrl,
		Ip:            in.Ip,
		Device:        in.Device,
	})
	if err != nil {
		return nil, err
	}

	rdb.HSet(l.ctx, key, time.Now().Unix())

	return &trend.CreateTrendResponse{TrendId: int64(id)}, nil
}

func checkTrendContent(content string) (bool, string, error) {
	word, err := sensitive.NewWord(sensitive.ALL_FILE)
	if err != nil {
		return false, "", err
	}

	pass, sensitiveWord := word.Validate(content)

	return pass, sensitiveWord, nil
}

func checkTrendPic(urls []string) bool {
	//TODO:
	return true
}
