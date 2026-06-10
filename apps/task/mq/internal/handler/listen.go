package handler

import (
	msgTransfer "github.com/iceymoss/go-hichat-api/apps/task/mq/internal/handler/msg_transfer"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

// Listen 服务配置
type Listen struct {
	svc *svc.ServiceContext
}

func NewListen(svc *svc.ServiceContext) *Listen {
	return &Listen{
		svc: svc,
	}
}

// Services 服务监控 service 是go-zero的一个服务模块
func (l *Listen) Services() []service.Service {
	// NewMsgChatTransfer 结构体实现了 kq.ConsumeHandle接口
	msgChatConsumeHandle := msgTransfer.NewMsgChatTransfer(l.svc)
	msgChatMarkReadConsumeHandle := msgTransfer.NewMsgReadTransfer(l.svc)
	msgRecallConsumeHandle := msgTransfer.NewMsgRecallTransfer(l.svc)
	trendNotifyConsumeHandle := msgTransfer.NewTrendNotifyTransfer(l.svc)
	// 注册处理逻辑
	return []service.Service{
		kq.MustNewQueue(l.svc.Config.MsgChatTransfer, msgChatConsumeHandle),
		kq.MustNewQueue(l.svc.Config.MsgReadTransfer, msgChatMarkReadConsumeHandle),
		kq.MustNewQueue(l.svc.Config.MsgRecallTransfer, msgRecallConsumeHandle),
		kq.MustNewQueue(l.svc.Config.TrendNotifyTransfer, trendNotifyConsumeHandle),
	}
}
