package logic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type blockingTask struct {
	started  chan struct{}
	finished chan struct{}
}

func (*blockingTask) GetName() string           { return "blocking" }
func (*blockingTask) GetSpec() string           { return "* * * * * *" }
func (*blockingTask) GetDescription() string    { return "blocks until cancellation" }
func (*blockingTask) GetTimeout() time.Duration { return time.Hour }
func (t *blockingTask) Execute(ctx context.Context) error {
	close(t.started)
	<-ctx.Done()
	close(t.finished)
	return ctx.Err()
}

func TestTaskManagerStopCancelsActiveTaskWithoutDeadlock(t *testing.T) {
	manager := NewTaskManager(true, 10)
	task := &blockingTask{started: make(chan struct{}), finished: make(chan struct{})}
	require.NoError(t, manager.RegisterTask(task))
	require.NoError(t, manager.Start())
	select {
	case <-task.started:
	case <-time.After(2 * time.Second):
		t.Fatal("task did not start")
	}

	stopped := make(chan error, 1)
	go func() { stopped <- manager.Stop() }()
	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Stop deadlocked while task saved its result")
	}
	select {
	case <-task.finished:
	case <-time.After(time.Second):
		t.Fatal("active task was not canceled")
	}
	results, err := manager.GetTaskResults(task.GetName(), 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Contains(t, results[0].Error, context.Canceled.Error())
}
