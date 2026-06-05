package trend

import (
	"context"
	"errors"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type CreateTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateTrendLogic 创建新动态
func NewCreateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTrendLogic {
	return &CreateTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTrendLogic) CreateTrend(req *types.CreateTrendRequest) (resp *types.CreateTrendResponse, err error) {
	// 类型校验：0 与 6(广告) 不对普通用户开放
	if isContain(req.TrendType) {
		return nil, errors.New("不支持该类型的动态")
	}

	// 按动态类型校验内容是否为空
	switch req.TrendType {
	case 1: // 纯文本
		if strings.TrimSpace(req.Content) == "" {
			return nil, errors.New("请填写动态内容")
		}
	case 2: // 图文混合
		if strings.TrimSpace(req.Content) == "" && len(req.Resources) == 0 {
			return nil, errors.New("请填写内容或上传图片")
		}
	case 3: // 文章
		if strings.TrimSpace(req.Title) == "" {
			return nil, errors.New("请填写文章标题")
		}
		if strings.TrimSpace(req.Content) == "" && len(req.Resources) == 0 {
			return nil, errors.New("请填写文章内容")
		}
	case 4: // 分享
		if req.ShareUrl == "" {
			return nil, errors.New("请选择你要分享的内容")
		}
	case 5: // 视频
		if len(req.Resources) == 0 {
			return nil, errors.New("请上传视频")
		}
	}

	uid := utils.GetUser(l.ctx)
	res, err := l.svcCtx.Trend.CreateTrend(l.ctx, &trend.CreateTrendRequest{
		UserId:       int64(uid),
		Type:         trend.TrendType(req.TrendType),
		Content:      req.Content,
		Scope:        trend.VisibilityScope(req.Scope),
		Resources:    req.Resources,
		PositionName: req.PositionName,
		PositionPoint: &trend.Point{
			Longitude: req.Longitude,
			Latitude:  req.Latitude,
		},
		Title:     req.Title,
		AtUserIds: req.AtUserIDs,
		OpenReply: req.OpenReply,
		CoverUrl:  req.CoverUrl,
		ShareUrl:  req.ShareUrl,
		Device:    req.Device,
		Ip:        req.IP,
	})
	if err != nil {
		zLog.Error("创建动态失败", zap.Error(err))
		return nil, err
	}

	resp = &types.CreateTrendResponse{
		TrendID: int(res.TrendId),
		Code:    int(res.Code),
	}

	return
}

func isContain(t int) bool {
	checkList := []int{0, 6}
	for _, v := range checkList {
		if t == v {
			return true
		}
	}
	return false
}
