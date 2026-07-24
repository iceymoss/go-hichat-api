package group

import (
	"context"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateGroupRejectsInvalidJWTActor(t *testing.T) {
	for _, value := range []any{nil, "", "invalid", "0", "01", 1} {
		t.Run("actor", func(t *testing.T) {
			ctx := context.Background()
			if value != nil {
				ctx = context.WithValue(ctx, ctxdata.Identify, value)
			}
			_, err := NewCreateGroupLogic(ctx, &svc.ServiceContext{}).CreateGroup(&types.GroupCreateReq{Name: "group"})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	}
}
