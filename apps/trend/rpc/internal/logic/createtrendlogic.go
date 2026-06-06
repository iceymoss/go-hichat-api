package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/sensitive"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		// 特殊处理键不存在的情况（redis.Nil 错误）
		if errors.Is(err, redis.Nil) {
			// 键不存在时不报错，使用默认空值继续执行
			lastPushTimeStr = ""
			zLog.Info("CreateTrend.HGet: key not found, using default",
				zap.Any("user", in.UserId),
				zap.String("key", key))
		} else {
			// 其他错误需要记录并返回
			zLog.Error("CreateTrend.HGet: get trend:user:push:time failed",
				zap.Any("user", in.UserId),
				zap.Error(err))
			return nil, err
		}
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

	// 媒体资源检查。上传接口会做文件级安全校验，这里兜底校验数量和类型。
	if !checkTrendResources(in.Type, in.Resources) {
		return &trend.CreateTrendResponse{
			Code: 3000,
		}, errors.New("动态媒体资源不合法")
	}

	// 可见范围(前端语义)：1-仅自己，2-仅好友，3-所有人
	// circle_state(朋友圈流可见状态)：1-可见，2-不可见
	// 仅自己(1) 不进好友/朋友圈流；仅好友(2)、所有人(3) 进朋友圈流
	scope := int(in.Scope)
	if scope < 1 || scope > 3 {
		scope = 2 // 容错：非法值默认仅好友
	}
	circleState := 1 // 默认进朋友圈
	if scope == 1 {  // 仅自己
		circleState = 2
	}

	var openReply int64
	if in.OpenReply {
		openReply = 1
	}

	// 处理图片或者视频

	// 处理位置
	positionPoint := make([]float32, 0, 2)
	positionPoint = append(positionPoint, float32(in.PositionPoint.Latitude))
	positionPoint = append(positionPoint, float32(in.PositionPoint.Longitude))

	var positionPointStr string
	if len(positionPoint) > 0 {
		positionPointByte, err := json.Marshal(positionPoint)
		if err != nil {
			zLog.Error("CreateTrend.Marshal: marshal fail", zap.Any("positionPoint", positionPoint), zap.Error(err))
			return nil, err
		}
		positionPointStr = string(positionPointByte)
	}

	var atUserIdsStr string
	if len(in.AtUserIds) > 0 {
		AtUserIds, err := json.Marshal(in.AtUserIds)
		if err != nil {
			zLog.Error("CreateTrend.Marshal: marshal fail", zap.Any("atUserIds", in.AtUserIds), zap.Error(err))
			return nil, err
		}
		atUserIdsStr = string(AtUserIds)
		fmt.Println("data:", atUserIdsStr, AtUserIds, in.AtUserIds)
	}

	var resourcesStr string
	if len(in.Resources) > 0 {
		resources, err := json.Marshal(in.Resources)
		if err != nil {
			zLog.Error("CreateTrend.Marshal: marshal fail", zap.Any("Resources", in.Resources), zap.Error(err))
			return nil, err
		}
		resourcesStr = string(resources)
	}

	id, err := l.svcCtx.Trend.Insert(l.ctx, &models.Trend{
		Userid:        uint64(in.UserId),
		Type:          uint64(in.Type),
		Content:       in.Content,
		PositionName:  in.PositionName,
		PositionPoint: positionPointStr,
		Createtime:    time.Now(),
		Updatetime:    time.Now(),
		CircleState:   int64(circleState),
		Scope:         int64(scope),
		State:         1,
		Idlist:        atUserIdsStr,
		OpenReply:     openReply,
		Title:         in.Title,
		PicArr:        resourcesStr,
		ShareUrl:      in.ShareUrl,
		Cover:         in.CoverUrl,
		Ip:            in.Ip,
		Device:        in.Device,
	})
	if err != nil {
		return nil, err
	}

	rdb.HSet(l.ctx, key, in.UserId, time.Now().Unix())

	return &trend.CreateTrendResponse{TrendId: int64(id), Code: 1000}, nil
}

func checkTrendContent(content string) (bool, string, error) {
	word, err := sensitive.NewWord(sensitive.OTHER_FILE)
	if err != nil {
		return false, "", err
	}

	pass, sensitiveWord := word.Validate(content)

	return pass, sensitiveWord, nil
}

func checkTrendResources(trendType trend.TrendType, urls []string) bool {
	if len(urls) == 0 {
		return true
	}
	switch trendType {
	case trend.TrendType_MIXED:
		if len(urls) > 9 {
			return false
		}
		for _, url := range urls {
			if !hasAllowedExt(url, []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}) {
				return false
			}
		}
	case trend.TrendType_VIDEO:
		if len(urls) != 1 {
			return false
		}
		return hasAllowedExt(urls[0], []string{".mp4", ".mov", ".webm"})
	}
	return true
}

func hasAllowedExt(rawURL string, allowed []string) bool {
	path := rawURL
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, item := range allowed {
		if ext == item {
			return true
		}
	}
	return false
}
