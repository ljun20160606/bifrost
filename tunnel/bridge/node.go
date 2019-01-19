package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"net"
)

const (
	Delim = '\n'
)

type NodeInfo struct {
	// account
	Group string `json:"group"`
	// password
	Name string `json:"name"`
	// connect | register
	Method int `json:"method"`
	// attachment
	Attachment json.RawMessage `json:"attachment"`
}

type Node struct {
	// origin conn
	net.Conn
	// context
	Context context.Context
	// cancel
	Cancel context.CancelFunc
	// node info
	*NodeInfo
}

func NewNode(conn net.Conn) (*Node, error) {
	infoBytes, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "node.info长度有误")
	}
	nodeInfo := new(NodeInfo)
	err = json.Unmarshal(infoBytes, nodeInfo)
	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "解析node.info失败")
	}
	node := new(Node)
	node.NodeInfo = nodeInfo
	node.Context, node.Cancel = context.WithCancel(context.Background())
	node.Conn = conn
	return node, nil
}
