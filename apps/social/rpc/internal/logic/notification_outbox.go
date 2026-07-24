package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type requestNotificationPayload struct {
	RequestType string `json:"requestType"`
	RequestID   uint64 `json:"requestId"`
	Result      int    `json:"result"`
	Content     string `json:"content,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
}

func emitRequestNotificationInTx(tx *gorm.DB, svcCtx *svc.ServiceContext, notifyType, requestType string, requestID uint64, receiverID, actorID string, groupID uint64, result int, content ...string) error {
	text := ""
	if len(content) > 0 {
		text = content[0]
	}
	groupText := ""
	if groupID != 0 {
		groupText = strconv.FormatUint(groupID, 10)
	}
	payload, err := json.Marshal(requestNotificationPayload{RequestType: requestType, RequestID: requestID, Result: result, Content: text, GroupID: groupText})
	if err != nil {
		return err
	}
	phase := "apply"
	if notifyType == NotifyGroupRequestResolved {
		phase = "resolved"
	} else if notifyType == NotifyGroupInviteInvalidated {
		phase = "invalidated"
	} else if requestType == receiptTypeGroupInvite {
		phase = "invite"
	} else if result == receiptAccepted {
		phase = "accept"
	} else if result == receiptRejected {
		phase = "reject"
	} else if result == receiptInvalidated {
		phase = "invalidated"
	}
	row := &objects.SocialNotificationOutbox{NotifyType: notifyType, ReceiverID: receiverID, ActorID: actorID,
		BizID: fmt.Sprintf("%s:%d:%s", requestType, requestID, phase), GroupID: strconv.FormatUint(groupID, 10), Payload: string(payload), CreatedAt: time.Now()}
	if groupID == 0 {
		row.GroupID = ""
	}
	if svcCtx == nil || svcCtx.SocialNotificationOutboxModel == nil {
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error
	}
	return svcCtx.SocialNotificationOutboxModel.InsertTx(tx, row)
}
