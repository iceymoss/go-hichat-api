package group

import (
	"encoding/json"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseGroupIDPreservesExactUint64(t *testing.T) {
	id, err := parseGroupID("9007199254740993", "group request id")
	require.NoError(t, err)
	require.Equal(t, uint64(9007199254740993), id)
}

func TestParseGroupIDRejectsNonCanonicalValues(t *testing.T) {
	for _, value := range []string{"", "0", "01", "-1", "+1", " 1", "1 ", "1.0", "invalid", "18446744073709551616"} {
		t.Run(value, func(t *testing.T) {
			_, err := parseGroupID(value, "group request id")
			require.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestGroupPutInResponseSerializesExplicitIsPassAndCanonicalGroupID(t *testing.T) {
	data, err := json.Marshal(types.GroupPutInResp{GroupId: "9007199254740993"})
	require.NoError(t, err)
	require.JSONEq(t, `{"group_id":"9007199254740993","is_pass":0,"status":0}`, string(data))
}

func TestParseGroupIDsRequiresNonEmptyExactIDs(t *testing.T) {
	_, err := parseGroupIDs(nil, "group request id")
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	ids, err := parseGroupIDs([]string{"9007199254740993", "18446744073709551615"}, "group request id")
	require.NoError(t, err)
	require.Equal(t, []uint64{9007199254740993, 18446744073709551615}, ids)
}
