package websocket

type Message struct {
	Method string      `json:"method,omitempty"` // 方法
	UserId string      `json:"userId,omitempty"` // 接受人
	FormId string      `json:"formId,omitempty"` // 发起人
	Data   interface{} `json:"data,omitempty"`   // 消息内容
}

func NewMessage(fid string, data interface{}) *Message {
	return &Message{
		FormId: fid,
		Data:   data,
	}
}
