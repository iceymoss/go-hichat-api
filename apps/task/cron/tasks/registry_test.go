package tasks

import (
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/types"

	"github.com/stretchr/testify/require"
)

type recordingTaskManager struct {
	names []string
}

func (m *recordingTaskManager) RegisterTask(task types.Task) error {
	m.names = append(m.names, task.GetName())
	return nil
}
func (*recordingTaskManager) UnregisterTask(string) error        { return nil }
func (*recordingTaskManager) Start() error                       { return nil }
func (*recordingTaskManager) Stop() error                        { return nil }
func (*recordingTaskManager) GetTaskStatus(string) (bool, error) { return false, nil }
func (*recordingTaskManager) GetTaskResults(string, int) ([]types.TaskResult, error) {
	return nil, nil
}

func TestRegisterAllTasksInvitationExpirationIsOptional(t *testing.T) {
	for _, tt := range []struct {
		name string
		spec string
		want bool
	}{
		{name: "disabled"},
		{name: "enabled", spec: "0 * * * * *", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manager := &recordingTaskManager{}
			svcCtx := &svc.ServiceContext{}
			svcCtx.Config.Cron.InvitationExpirationSpec = tt.spec
			RegisterAllTasks(manager, svcCtx)
			require.Equal(t, tt.want, containsTask(manager.names, "group_invitation_expiration"))
			if tt.want {
				require.Equal(t, []string{"group_invitation_expiration"}, manager.names)
			} else {
				require.Empty(t, manager.names)
			}
		})
	}
}

func containsTask(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}
