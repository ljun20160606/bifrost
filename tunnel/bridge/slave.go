package bridge

import "encoding/json"

type NodeSlave struct {
	*Node
	// 任务信息
	Message *Message
}

func NewNodeSlave(node *Node) (*NodeSlave, error) {
	message := new(Message)
	err := json.Unmarshal(node.Attachment, message)
	if err != nil {
		node.Close()
		return nil, err
	}
	return &NodeSlave{Node: node, Message: message}, nil
}
