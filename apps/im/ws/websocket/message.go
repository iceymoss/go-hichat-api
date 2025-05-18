package websocket

type FrameType uint8

const (
	// FrameData 普通消息
	FrameData FrameType = 0x0

	// FramePing 检查消息
	FramePing FrameType = 0x1

	// FrameErr 错误消息
	FrameErr FrameType = 0x2

	// FrameNoAck 不应答
	FrameNoAck FrameType = 0x3
)

type Message struct {
	FrameType `json:"frameType"`
	Method    string      `json:"method,omitempty"`
	UserId    string      `json:"userId,omitempty"`
	FormId    string      `json:"formId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
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
