package rpcauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	metadataPrincipal = "x-hichat-rpc-principal"
	metadataSubject   = "x-hichat-rpc-subject"
	metadataTimestamp = "x-hichat-rpc-timestamp"
	metadataNonce     = "x-hichat-rpc-nonce"
	metadataDigest    = "x-hichat-rpc-digest"
	metadataSignature = "x-hichat-rpc-signature"
	principalUser     = "user"
	principalTask     = "task"
	taskSubject       = "task"
	maxClockSkew      = 5 * time.Minute
	replayTTL         = 10 * time.Minute
	EnvSecret         = "HICHAT_IM_RPC_AUTH_SECRET"
)

var (
	ErrMissingSecret   = errors.New("rpc auth secret is required")
	ErrUnauthenticated = errors.New("missing or invalid rpc caller principal")
	ErrWrongPrincipal  = errors.New("rpc caller principal type is not allowed")
	ErrReplay          = errors.New("rpc authentication metadata was replayed")
	ErrReplayStore     = errors.New("rpc replay protection is unavailable")
)

type contextKey uint8

const (
	outboundPrincipalKey contextKey = iota
	verifiedPrincipalKey
)

type principal struct {
	kind    string
	subject string
}

type ReplayStore interface {
	Consume(context.Context, string, time.Duration) (bool, error)
}

type Auth struct {
	secret []byte
	now    func() time.Time
	random io.Reader
}

type RedisReplayStore struct {
	client *redis.Client
}

func New(secret string) (*Auth, error) {
	if len(secret) < 32 {
		return nil, ErrMissingSecret
	}
	return &Auth{secret: []byte(secret), now: time.Now, random: rand.Reader}, nil
}

func LoadSecret(configured string) string {
	if secret, ok := os.LookupEnv(EnvSecret); ok && secret != "" {
		return secret
	}
	return configured
}

func NewRedisReplayStore(client *redis.Client) *RedisReplayStore {
	return &RedisReplayStore{client: client}
}

func (s *RedisReplayStore) Consume(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
	if s == nil || s.client == nil {
		return false, ErrReplayStore
	}
	keyDigest := sha256.Sum256([]byte(nonce))
	consumed, err := s.client.SetNX(ctx, "im:rpc-auth:nonce:"+hex.EncodeToString(keyDigest[:]), "1", ttl).Result()
	if err != nil {
		return false, errors.Join(ErrReplayStore, err)
	}
	return consumed, nil
}

func WithUser(ctx context.Context, uid string) (context.Context, error) {
	if !CanonicalUID(uid) {
		return nil, ErrUnauthenticated
	}
	return context.WithValue(ctx, outboundPrincipalKey, principal{kind: principalUser, subject: uid}), nil
}

func WithTask(ctx context.Context) context.Context {
	return context.WithValue(ctx, outboundPrincipalKey, principal{kind: principalTask, subject: taskSubject})
}

func RequireUser(ctx context.Context, expectedUID string) error {
	p, ok := ctx.Value(verifiedPrincipalKey).(principal)
	if !ok {
		return ErrUnauthenticated
	}
	if p.kind != principalUser || p.subject != expectedUID || !CanonicalUID(p.subject) {
		return ErrWrongPrincipal
	}
	return nil
}

func RequireTask(ctx context.Context) error {
	p, ok := ctx.Value(verifiedPrincipalKey).(principal)
	if !ok {
		return ErrUnauthenticated
	}
	if p.kind != principalTask || p.subject != taskSubject {
		return ErrWrongPrincipal
	}
	return nil
}

func (a *Auth) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		p, ok := ctx.Value(outboundPrincipalKey).(principal)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		message, ok := req.(proto.Message)
		if !ok {
			return ErrUnauthenticated
		}
		signedCtx, err := a.sign(ctx, p, method, message)
		if err != nil {
			return err
		}
		return invoker(signedCtx, method, req, reply, cc, opts...)
	}
}

func (a *Auth) UnaryServerInterceptor(store ReplayStore, protectedMethods ...string) grpc.UnaryServerInterceptor {
	protected := make(map[string]struct{}, len(protectedMethods))
	for _, method := range protectedMethods {
		protected[method] = struct{}{}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := protected[info.FullMethod]; !ok {
			return handler(ctx, req)
		}
		message, ok := req.(proto.Message)
		if !ok {
			return nil, ErrUnauthenticated
		}
		p, err := a.verify(ctx, info.FullMethod, message, store)
		if err != nil {
			switch {
			case errors.Is(err, ErrReplayStore):
				return nil, status.Error(codes.Unavailable, "rpc replay protection is unavailable")
			default:
				return nil, status.Error(codes.Unauthenticated, "missing or invalid rpc authentication metadata")
			}
		}
		return handler(context.WithValue(ctx, verifiedPrincipalKey, p), req)
	}
}

func CanonicalUID(uid string) bool {
	parsed, err := strconv.ParseUint(uid, 10, 64)
	return err == nil && parsed > 0 && strconv.FormatUint(parsed, 10) == uid
}

func (a *Auth) sign(ctx context.Context, p principal, method string, message proto.Message) (context.Context, error) {
	nonceBytes := make([]byte, 32)
	if _, err := io.ReadFull(a.random, nonceBytes); err != nil {
		return nil, err
	}
	timestamp := strconv.FormatInt(a.now().Unix(), 10)
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	digest, err := requestDigest(message)
	if err != nil {
		return nil, err
	}
	signature := a.signature(p.kind, p.subject, timestamp, nonce, method, digest)
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set(metadataPrincipal, p.kind)
	md.Set(metadataSubject, p.subject)
	md.Set(metadataTimestamp, timestamp)
	md.Set(metadataNonce, nonce)
	md.Set(metadataDigest, digest)
	md.Set(metadataSignature, signature)
	return metadata.NewOutgoingContext(ctx, md), nil
}

func (a *Auth) verify(ctx context.Context, method string, message proto.Message, store ReplayStore) (principal, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return principal{}, ErrUnauthenticated
	}
	kind, kindOK := singleValue(md, metadataPrincipal)
	subject, subjectOK := singleValue(md, metadataSubject)
	timestamp, timestampOK := singleValue(md, metadataTimestamp)
	nonce, nonceOK := singleValue(md, metadataNonce)
	digest, digestOK := singleValue(md, metadataDigest)
	signature, signatureOK := singleValue(md, metadataSignature)
	if !kindOK || !subjectOK || !timestampOK || !nonceOK || !digestOK || !signatureOK || subject == "" {
		return principal{}, ErrUnauthenticated
	}
	nonceBytes, err := base64.RawURLEncoding.DecodeString(nonce)
	if err != nil || len(nonceBytes) != 32 {
		return principal{}, ErrUnauthenticated
	}
	issuedAt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || a.now().Sub(time.Unix(issuedAt, 0)) > maxClockSkew || time.Unix(issuedAt, 0).Sub(a.now()) > maxClockSkew {
		return principal{}, ErrUnauthenticated
	}
	wantDigest, err := requestDigest(message)
	if err != nil || !hmac.Equal([]byte(digest), []byte(wantDigest)) {
		return principal{}, ErrUnauthenticated
	}
	wantSignature, err := base64.RawURLEncoding.DecodeString(a.signature(kind, subject, timestamp, nonce, method, digest))
	if err != nil {
		return principal{}, ErrUnauthenticated
	}
	gotSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(gotSignature, wantSignature) {
		return principal{}, ErrUnauthenticated
	}
	consumed, err := store.Consume(ctx, nonce, replayTTL)
	if err != nil {
		return principal{}, err
	}
	if !consumed {
		return principal{}, ErrReplay
	}
	return principal{kind: kind, subject: subject}, nil
}

func requestDigest(message proto.Message) (string, error) {
	body, err := (proto.MarshalOptions{Deterministic: true}).Marshal(message)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:]), nil
}

func (a *Auth) signature(kind, subject, timestamp, nonce, method, digest string) string {
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte("hichat-rpc-auth-v1\n" + kind + "\n" + subject + "\n" + timestamp + "\n" + nonce + "\n" + method + "\n" + digest))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func singleValue(md metadata.MD, key string) (string, bool) {
	values := md.Get(key)
	if len(values) != 1 {
		return "", false
	}
	return values[0], true
}
