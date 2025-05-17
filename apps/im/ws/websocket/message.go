package websocket

type FrameType uint8

const (
	// FrameData 普通消息
	FrameData FrameType = 0x0

	// FramePing 检查消息
	FramePing FrameType = 0x1
)

type Message struct {
	FrameType `json:"frameType"`
	Method    string      `json:"method,omitempty"`
	UserId    string      `json:"userId,omitempty"`
	FormId    string      `json:"formId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func NewMessage(srv *Server, conn *Conn, data interface{}) *Message {
	fid := srv.GetUsers([]*Conn{conn})[0]
	return &Message{
		FrameType: FrameData,
		FormId:    fid,
		Data:      data,
	}
}
