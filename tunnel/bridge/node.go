package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	Delim = '\n'
)

type NodeInfo struct {
	// id
	Id string `json:"id"`
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
	// logger
	Logger *logrus.Entry
}

func NewNode(conn net.Conn) (*Node, error) {
	infoBytes, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "node.info长度有误")
	}
	nodeInfo := new(NodeInfo)
	err = json.Unmarshal(infoBytes, nodeInfo)
	if err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "解析node.info失败")
	}
	node := new(Node)
	node.NodeInfo = nodeInfo
	node.Context, node.Cancel = context.WithCancel(context.Background())
	node.Conn = conn
	node.Logger = logrus.WithField("id", nodeInfo.Id)
	return node, nil
}
