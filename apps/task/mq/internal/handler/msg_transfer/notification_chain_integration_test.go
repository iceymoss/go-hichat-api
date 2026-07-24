package msg_transfer

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	immodel "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const notificationIntTestTopic = "social.request.notification.v1.inttest"

type notificationModelRPCAdapter struct {
	model immodel.NotificationModel
}

func (a *notificationModelRPCAdapter) CreateNotification(ctx context.Context, in *im.CreateNotificationReq, _ ...grpc.CallOption) (*im.CreateNotificationResp, error) {
	row := &immodel.Notification{
		ReceiverId: in.ReceiverId, NotifyType: in.NotifyType, BizId: in.BizId,
		ActorId: in.ActorId, GroupId: in.GroupId, Title: in.Title, Content: in.Content,
		Payload: in.Payload, CreatedAt: time.Unix(in.CreateTime, 0),
	}
	inserted, err := a.model.Insert(ctx, row)
	return &im.CreateNotificationResp{Id: row.Id, Inserted: inserted}, err
}

func TestNotificationChainIntegration(t *testing.T) {
	broker := os.Getenv("HICHAT_NOTIFICATION_CHAIN_KAFKA_BROKER")
	if broker == "" {
		broker = "127.0.0.1:9092"
	}
	dsn := os.Getenv("HICHAT_NOTIFICATION_CHAIN_MYSQL_DSN")
	if dsn == "" || os.Getenv("HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS") != "1" {
		t.Skip("set HICHAT_NOTIFICATION_CHAIN_MYSQL_DSN and HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS=1 to run")
	}
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open notification integration MySQL: %v", err)
	}
	var databaseName string
	if err := database.Raw("SELECT DATABASE()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("read integration database: %v", err)
	}
	matched, err := regexp.MatchString(`^hichat_[a-z0-9_]*_test$`, strings.ToLower(databaseName))
	if err != nil || !matched {
		t.Fatalf("notification integration requires a dedicated hichat_*_test database, got %q", databaseName)
	}
	if err := database.AutoMigrate(&objects.Notification{}, &objects.NotificationReadIntent{}); err != nil {
		t.Fatalf("migrate notification integration tables: %v", err)
	}

	uniq := strconv.FormatInt(time.Now().UnixNano(), 10)
	receiver := uniq
	event := mq.CommonNotify{
		EventId: 1, RequestType: "group_invite", RequestId: 9, Result: 3,
		NotifyType: "group.invite.invalidated", ReceiverId: receiver, ActorId: "1",
		BizId: "group_invite:9:invalidated", GroupId: "7", CreateTime: time.Now().Unix(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		database.Where("receiver_id = ?", receiver).Delete(&objects.Notification{})
		database.Where("receiver_id = ?", receiver).Delete(&objects.NotificationReadIntent{})
	})

	writer := &kafka.Writer{Addr: kafka.TCP(broker), Topic: notificationIntTestTopic, RequiredAcks: kafka.RequireAll}
	defer writer.Close()
	if err := writer.WriteMessages(context.Background(), kafka.Message{Key: []byte(uniq), Value: body}); err != nil {
		t.Skipf("kafka unavailable: %v", err)
	}

	read := readNotificationByReceiver(t, broker, receiver)
	if read == nil {
		t.Skip("notification Kafka round-trip did not receive its own message")
	}
	readBody, err := json.Marshal(read)
	if err != nil {
		t.Fatal(err)
	}
	sender := &fakeNotificationSender{}
	transfer := &CommonNotifyTransfer{
		im: &notificationModelRPCAdapter{model: immodel.NewNotificationModelWithDB(database)},
		ws: sender, dlq: &fakeDeadLetterPublisher{}, shutdown: context.Background(),
		retryDelay: func(int) time.Duration { return 0 },
	}
	if err := transfer.Consume(context.Background(), "", string(readBody)); err != nil {
		t.Fatalf("consume notification: %v", err)
	}
	if err := transfer.Consume(context.Background(), "", string(readBody)); err != nil {
		t.Fatalf("consume duplicate notification: %v", err)
	}

	var count int64
	if err := database.Model(&objects.Notification{}).Where("receiver_id = ? AND notify_type = ? AND biz_id = ?", receiver, event.NotifyType, event.BizId).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("persisted notifications = %d, want 1", count)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("websocket pushes = %d, want 1", len(sender.messages))
	}
	message := sender.messages[0].(websocket.Message)
	if message.Method != "push.notify" {
		t.Fatalf("websocket method = %q, want push.notify", message.Method)
	}
}

func readNotificationByReceiver(t *testing.T, broker, receiver string) *mq.CommonNotify {
	t.Helper()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker}, Topic: notificationIntTestTopic,
		GroupID: "notification_inttest_" + receiver, StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			return nil
		}
		var event mq.CommonNotify
		if json.Unmarshal(message.Value, &event) == nil && event.ReceiverId == receiver {
			return &event
		}
	}
}
