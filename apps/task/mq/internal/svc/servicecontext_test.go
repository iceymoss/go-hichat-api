package svc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-queue/kq"
)

type closeableDLQ struct {
	closed bool
}

func (*closeableDLQ) Publish(context.Context, []byte) error { return nil }

func (d *closeableDLQ) Close() error {
	d.closed = true
	return nil
}

type closeableWebsocket struct {
	closed int
}

func (w *closeableWebsocket) Close() error {
	w.closed++
	return nil
}

func (*closeableWebsocket) Send(any) error { return nil }
func (*closeableWebsocket) Read(any) error { return nil }

type blockingKafkaWriter struct {
	called  chan kafka.Message
	release chan error
}

func (w *blockingKafkaWriter) WriteMessages(_ context.Context, messages ...kafka.Message) error {
	w.called <- messages[0]
	return <-w.release
}

func TestNotificationDLQPublishWaitsForWriterAcknowledgement(t *testing.T) {
	writer := &blockingKafkaWriter{called: make(chan kafka.Message, 1), release: make(chan error, 1)}
	publisher := &notificationDLQPublisher{writer: writer}
	done := make(chan error, 1)
	go func() { done <- publisher.Publish(context.Background(), []byte("event")) }()

	message := <-writer.called
	if string(message.Value) != "event" {
		t.Fatalf("message = %q, want event", message.Value)
	}
	select {
	case err := <-done:
		t.Fatalf("Publish returned before writer acknowledgement: %v", err)
	default:
	}
	wantErr := errors.New("broker rejected write")
	writer.release <- wantErr
	if err := <-done; !errors.Is(err, wantErr) {
		t.Fatalf("Publish error = %v, want %v", err, wantErr)
	}
}

func TestNotificationDLQWriterConfig(t *testing.T) {
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(caFile, testCertificate(t), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := notificationDLQWriterConfig(kq.KqConf{
		Brokers:  []string{"broker-a:9092", "broker-b:9092"},
		Username: "task",
		Password: "secret",
		CaFile:   caFile,
	}, "custom.notification.dead.v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Brokers) != 2 || got.Topic != "custom.notification.dead.v1" || got.RequiredAcks != int(kafka.RequireAll) || got.Async || got.BatchSize != 1 {
		t.Fatalf("writer config = %+v", got)
	}
	if got.Dialer == nil || got.Dialer.SASLMechanism == nil || got.Dialer.TLS == nil || got.Dialer.TLS.RootCAs == nil {
		t.Fatalf("writer security config not derived from KqConf: %+v", got.Dialer)
	}
}

func TestNotificationDLQWriterConfigDefaultsTopic(t *testing.T) {
	got, err := notificationDLQWriterConfig(kq.KqConf{Brokers: []string{"broker:9092"}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Topic != defaultNotificationDLQTopic {
		t.Fatalf("topic = %q, want %q", got.Topic, defaultNotificationDLQTopic)
	}
}

func testCertificate(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestNewTaskRPCAuthRequiresSecret(t *testing.T) {
	t.Setenv(rpcauth.EnvSecret, "")
	_, err := newTaskRPCAuth("")
	if !errors.Is(err, rpcauth.ErrMissingSecret) {
		t.Fatalf("error = %v, want ErrMissingSecret", err)
	}
}

func TestNewTaskRPCAuthLoadsConfigOrEnvironment(t *testing.T) {
	const configured = "configured-secret-at-least-32-bytes-long"
	t.Setenv(rpcauth.EnvSecret, "")
	if _, err := newTaskRPCAuth(configured); err != nil {
		t.Fatalf("configured secret: %v", err)
	}

	t.Setenv(rpcauth.EnvSecret, "environment-secret-at-least-32-bytes-long")
	if _, err := newTaskRPCAuth(""); err != nil {
		t.Fatalf("environment secret: %v", err)
	}
}

func TestServiceContextClosesDLQPublisher(t *testing.T) {
	dlq := &closeableDLQ{}
	ws := &closeableWebsocket{}
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	svcCtx := &ServiceContext{NotificationDLQ: dlq, WsClient: ws, shutdownCtx: shutdownCtx, shutdownCancel: shutdownCancel}
	if err := svcCtx.Close(); err != nil {
		t.Fatal(err)
	}
	if !dlq.closed {
		t.Fatal("DLQ publisher was not closed")
	}
	if !errors.Is(shutdownCtx.Err(), context.Canceled) {
		t.Fatalf("shutdown context error = %v, want context.Canceled", shutdownCtx.Err())
	}
	if ws.closed != 1 {
		t.Fatalf("websocket closes = %d, want 1", ws.closed)
	}
}

func TestServiceContextCancelClosesNotificationGate(t *testing.T) {
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	svcCtx := &ServiceContext{shutdownCtx: shutdownCtx, shutdownCancel: shutdownCancel}
	done, ok := svcCtx.BeginNotification()
	if !ok {
		t.Fatal("notification gate unexpectedly closed")
	}
	canceled := make(chan struct{})
	go func() {
		svcCtx.Cancel()
		close(canceled)
	}()
	<-shutdownCtx.Done()
	select {
	case <-canceled:
		t.Fatal("Cancel returned before the active notification completed")
	default:
	}
	done()
	<-canceled
	if _, ok := svcCtx.BeginNotification(); ok {
		t.Fatal("notification gate accepted work after shutdown")
	}
}
