package handler

import (
	"sync"

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
	relationChangeConsumeHandle := msgTransfer.NewRelationChangeTransfer(l.svc)
	commonNotifyConsumeHandle := msgTransfer.NewCommonNotifyTransfer(l.svc)
	socialRequestNotification := socialRequestNotificationConf(l.svc.Config.SocialRequestNotification, l.svc.Config.CommonNotifyTransfer)
	relationChange := l.svc.Config.RelationChangeTransfer
	relationChange.ForceCommit = false
	relationChange.CommitInOrder = false
	relationChange.Consumers = 1
	relationChange.Processors = 1
	commonNotify := reliableNotificationQueue(l.svc.Config.CommonNotifyTransfer)
	socialRequestNotification = reliableNotificationQueue(socialRequestNotification)
	// 注册处理逻辑
	return []service.Service{
		kq.MustNewQueue(l.svc.Config.MsgChatTransfer, msgChatConsumeHandle),
		kq.MustNewQueue(l.svc.Config.MsgReadTransfer, msgChatMarkReadConsumeHandle),
		kq.MustNewQueue(l.svc.Config.MsgRecallTransfer, msgRecallConsumeHandle),
		kq.MustNewQueue(l.svc.Config.TrendNotifyTransfer, trendNotifyConsumeHandle),
		kq.MustNewQueue(relationChange, relationChangeConsumeHandle),
		kq.MustNewQueue(commonNotify, commonNotifyConsumeHandle),
		kq.MustNewQueue(socialRequestNotification, commonNotifyConsumeHandle),
	}
}

func socialRequestNotificationConf(socialRequestNotification, commonNotify kq.KqConf) kq.KqConf {
	if socialRequestNotification.Topic == "" {
		socialRequestNotification = commonNotify
		socialRequestNotification.Topic = "social.request.notification.v1"
		socialRequestNotification.Name = "socialRequestNotificationV1"
		socialRequestNotification.Group = "socialRequestNotificationV1"
	}
	if len(socialRequestNotification.Brokers) == 0 {
		socialRequestNotification.Brokers = commonNotify.Brokers
	}
	if socialRequestNotification.Name == "" {
		socialRequestNotification.Name = "socialRequestNotificationV1"
	}
	if socialRequestNotification.Group == "" {
		socialRequestNotification.Group = "socialRequestNotificationV1"
	}
	if socialRequestNotification.Offset == "" {
		socialRequestNotification.Offset = "first"
	}
	return socialRequestNotification
}

// Service groups task queues behind an ordered shutdown boundary: retries are canceled before queues wait for handlers.
func (l *Listen) Service() service.Service {
	return newTaskMQService(l.svc.Cancel, l.Services()...)
}

type taskMQService struct {
	queues   []service.Service
	cancel   func()
	stopOnce sync.Once
}

func newTaskMQService(cancel func(), queues ...service.Service) *taskMQService {
	return &taskMQService{queues: queues, cancel: cancel}
}

func (s *taskMQService) Start() {
	var group sync.WaitGroup
	group.Add(len(s.queues))
	for _, queue := range s.queues {
		queue := queue
		go func() {
			defer group.Done()
			queue.Start()
		}()
	}
	group.Wait()
}

func (s *taskMQService) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
		var group sync.WaitGroup
		group.Add(len(s.queues))
		for _, queue := range s.queues {
			queue := queue
			go func() {
				defer group.Done()
				queue.Stop()
			}()
		}
		group.Wait()
	})
}

func reliableNotificationQueue(conf kq.KqConf) kq.KqConf {
	conf.Consumers = 1
	conf.Processors = 1
	conf.ForceCommit = false
	conf.CommitInOrder = false
	return conf
}
