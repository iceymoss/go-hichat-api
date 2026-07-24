package logic

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseNotificationIDs(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []uint64
	}{
		{name: "large IDs remain exact", values: []string{"9007199254740993", "18446744073709551615"}, want: []uint64{9007199254740993, math.MaxUint64}},
		{name: "zero", values: []string{"0"}},
		{name: "leading zero", values: []string{"01"}},
		{name: "signed", values: []string{"+1"}},
		{name: "overflow", values: []string{"18446744073709551616"}},
		{name: "empty", values: []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNotificationIDs(tt.values)
			if tt.want == nil {
				require.Equal(t, codes.InvalidArgument, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
