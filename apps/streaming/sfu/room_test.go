package sfu

import (
	"sort"
	"testing"
)

// Test_Room_JoinLeave_TracksMembership 房间成员进出簿记：加入幂等、离开移除、离开未知成员无副作用。
func Test_Room_JoinLeave_TracksMembership(t *testing.T) {
	tests := []struct {
		name string
		ops  func(r *Room)
		want []string
	}{
		{
			name: "join adds participants",
			ops:  func(r *Room) { r.Join("a"); r.Join("b") },
			want: []string{"a", "b"},
		},
		{
			name: "join is idempotent",
			ops:  func(r *Room) { r.Join("a"); r.Join("a") },
			want: []string{"a"},
		},
		{
			name: "leave removes participant",
			ops:  func(r *Room) { r.Join("a"); r.Join("b"); r.Leave("a") },
			want: []string{"b"},
		},
		{
			name: "leave unknown is no-op",
			ops:  func(r *Room) { r.Join("a"); r.Leave("x") },
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoom("call1")
			tt.ops(r)
			got := r.Participants()
			sort.Strings(got)
			if !equalStrs(got, tt.want) {
				t.Errorf("Participants() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
