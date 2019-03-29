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
	ServiceId string `json:"serviceId"`
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
	// serviceInfo
	*NodeInfo
	// logger
	Logger *logrus.Entry
}

func NewSession(conn net.Conn) (*Session, error) {
	infoBytes, err := bufio.NewReader(conn).ReadBytes(Delim)
	if err != nil {
		return nil, errors.Wrap(err, "request.info length error")
	}
	request := new(Request)
	err = json.Unmarshal(infoBytes, request)
	if err != nil {
		return nil, errors.Wrap(err, "parse request.info fail")
	}
	node := new(Session)
	node.Request = request
	node.Context, node.Cancel = context.WithCancel(context.Background())
	node.Conn = conn
	node.Logger = logrus.WithField("serviceId", request.ServiceId)
	return node, nil
}
