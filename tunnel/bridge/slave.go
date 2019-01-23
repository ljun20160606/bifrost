package bridge

import (
	"encoding/json"
	"github.com/ljun20160606/bifrost/tunnel"
)

// A info of service when service connect bridge
type NodeSlave struct {
	*tunnel.Session
	// 任务信息
	Message *Message
}

func NewNodeSlave(session *tunnel.Session) (*NodeSlave, error) {
	message := new(Message)
	err := json.Unmarshal(session.Attachment, message)
	if err != nil {
		_ = session.Close()
		return nil, err
	}
	return &NodeSlave{Session: session, Message: message}, nil
}
