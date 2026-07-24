package friend

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRequestIDPreservesUint64(t *testing.T) {
	id, err := parseRequestID("9007199254740993", 0)
	require.NoError(t, err)
	require.Equal(t, uint64(9007199254740993), id)
}

func TestParseRequestIDRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "invalid", "18446744073709551616"} {
		t.Run(value, func(t *testing.T) {
			_, err := parseRequestID(value, 0)
			require.Error(t, err)
		})
	}
}
