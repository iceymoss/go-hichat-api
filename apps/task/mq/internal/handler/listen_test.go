package handler

import (
	"context"
	"testing"

	"github.com/zeromicro/go-queue/kq"
)

type shutdownAwareService struct {
	shutdown context.Context
	stopped  bool
}

func (*shutdownAwareService) Start() {}

func (s *shutdownAwareService) Stop() {
	if s.shutdown.Err() != nil {
		s.stopped = true
	}
}

func TestReliableNotificationQueue(t *testing.T) {
	got := reliableNotificationQueue(kq.KqConf{Consumers: 9, Processors: 8, ForceCommit: true, CommitInOrder: true})
	if got.Consumers != 1 || got.Processors != 1 || got.ForceCommit || got.CommitInOrder {
		t.Fatalf("reliable notification queue = %+v", got)
	}
}

func TestTaskMQServiceCancelsBeforeStoppingQueues(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := &shutdownAwareService{shutdown: ctx}
	service := newTaskMQService(cancel, queue)

	service.Stop()
	if !queue.stopped {
		t.Fatal("queue stopped before task shutdown context was canceled")
	}
}
