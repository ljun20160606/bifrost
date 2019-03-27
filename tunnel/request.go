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
	Method     int             `json:"method"`
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
		return nil, errors.Wrap(err, "node.info length error")
	}
	request := new(Request)
	err = json.Unmarshal(infoBytes, request)
	if err != nil {
		return nil, errors.Wrap(err, "parse node.info fail")
	}
	node := new(Session)
	node.Request = request
	node.Context, node.Cancel = context.WithCancel(context.Background())
	node.Conn = conn
	node.Logger = logrus.WithField("id", request.Id)
	return node, nil
}
