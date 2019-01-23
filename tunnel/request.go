package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
)

type Request struct {
	*NodeInfo `json:"nodeInfo"`
	// connect | register
	Method int `json:"method"`
	// attachment
	Attachment json.RawMessage `json:"attachment"`
}

type Session struct {
	// origin conn
	net.Conn
	// context
	Context context.Context
	// cancel
	Cancel context.CancelFunc
	// request
	*Request
	// logger
	Logger *logrus.Entry
}

func NewSession(conn net.Conn) (*Session, error) {
	infoBytes, err := bufio.NewReader(conn).ReadBytes(Delim)
	if err != nil {
		return nil, errors.Wrap(err, "node.info长度有误")
	}
	request := new(Request)
	err = json.Unmarshal(infoBytes, request)
	if err != nil {
		return nil, errors.Wrap(err, "解析node.info失败")
	}
	node := new(Session)
	node.Request = request
	node.Context, node.Cancel = context.WithCancel(context.Background())
	node.Conn = conn
	node.Logger = logrus.WithField("id", request.Id)
	return node, nil
}