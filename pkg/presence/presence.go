package presence

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const keyPrefix = "user:online:"

type Store struct{ redis *redis.Client }

func New(client *redis.Client) *Store { return &Store{redis: client} }
func Key(uid string) string           { return keyPrefix + uid }
func ownerKey(uid string) string      { return keyPrefix + uid + ":owner" }
func (s *Store) Claim(ctx context.Context, uid, nodeID, token string, ttl time.Duration) error {
	_, err := s.redis.Eval(ctx, `redis.call("set",KEYS[1],ARGV[1],"PX",ARGV[3]);redis.call("set",KEYS[2],ARGV[2],"PX",ARGV[3]);return 1`, []string{Key(uid), ownerKey(uid)}, nodeID, token, int64(ttl/time.Millisecond)).Result()
	return err
}
func (s *Store) Refresh(ctx context.Context, uid, token string, ttl time.Duration) (bool, error) {
	n, err := s.redis.Eval(ctx, `if redis.call("get",KEYS[2])==ARGV[1] then redis.call("pexpire",KEYS[1],ARGV[2]);redis.call("pexpire",KEYS[2],ARGV[2]);return 1 else return 0 end`, []string{Key(uid), ownerKey(uid)}, token, int64(ttl/time.Millisecond)).Int()
	return n == 1, err
}
func (s *Store) DeleteIfOwner(ctx context.Context, uid, token string) (bool, error) {
	n, err := s.redis.Eval(ctx, `if redis.call("get",KEYS[2])==ARGV[1] then redis.call("del",KEYS[1]);redis.call("del",KEYS[2]);return 1 else return 0 end`, []string{Key(uid), ownerKey(uid)}, token).Int()
	return n == 1, err
}
func (s *Store) BatchOnline(ctx context.Context, uids []string) (map[string]bool, error) {
	out := make(map[string]bool, len(uids))
	if len(uids) == 0 {
		return out, nil
	}
	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = Key(uid)
	}
	values, err := s.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	for i, value := range values {
		out[uids[i]] = value != nil
	}
	return out, nil
}
