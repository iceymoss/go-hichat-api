package websocket

import "time"

type FrameType uint8

const (
	// FrameData 普通消息
	FrameData FrameType = 0x0

	// FramePing 检查消息
	FramePing FrameType = 0x1

	// FrameErr 错误消息
	FrameErr FrameType = 0x2

	// FrameAck 应答
	FrameAck FrameType = 0x3

	// FrameNoAck 不应答
	FrameNoAck FrameType = 0x4

	FrameCAck FrameType = 0x5
)

type Message struct {
	Id        string             `json:"id"`               // 消息id
	AckSeq    int                `json:"ackSeq,omitempty"` // 消息确认序号
	ackTime   time.Time          `json:"ackTime"`          // 消息确认时间
	errCount  int                `json:"errCount"`         // 消息错误次数
	FrameType `json:"frameType"` // 消息类型
	Method    string             `json:"method,omitempty"` // 业务处理方法
	UserId    string             `json:"userId,omitempty"` // 接受者id
	FormId    string             `json:"formId,omitempty"` // 发送者id
	Data      interface{}        `json:"data,omitempty"`   // 消息详细信息
}

func NewMessageTest(srv *Server, conn *Conn, data interface{}) *Message {
	fid := srv.GetUsers([]*Conn{conn})[0]
	return &Message{
		FrameType: FrameData,
		FormId:    fid,
		Data:      data,
	}
}

func NewMessage(formId string, data interface{}) *Message {
	return &Message{
		FrameType: FrameData,
		FormId:    formId,
		Data:      data,
	}
}

func NewErrMessage(err error) *Message {
	return &Message{
		FrameType: FrameErr,
		Data:      err.Error(),
	}
}
