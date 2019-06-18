package tunnel

import (
	"errors"
	"github.com/satori/go.uuid"
	"golang.org/x/net/proxy"
	"strings"
)

const (
	Delim     = '\n'
	userDelim = ":"
)

const (
	// 注册
	MethodRegister = 1
	// 准备接收任务
	MethodConn = 2
)

type NodeInfo struct {
	Id       string `json:"id"`
	Group    string `json:"group"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Cipher   string `json:"cipher"`
}

func (n *NodeInfo) Account() string {
	return n.Group + userDelim + n.Name
}

func (n *NodeInfo) User() string {
	return n.Account() + userDelim + n.Id
}

func (n *NodeInfo) ProxyAuth() (*proxy.Auth, error) {
	auth := &proxy.Auth{
		User:     n.User(),
		Password: n.Password,
	}
	if len(auth.User) == 0 || len(auth.User) > 255 || len(auth.Password) == 0 || len(auth.Password) > 255 {
		return nil, errors.New("invalid username/password")
	}
	return auth, nil
}

func ParseUser(str string) (*NodeInfo, error) {
	n := strings.SplitN(str, userDelim, 3)
	if len(n) < 3 {
		return nil, errors.New(str + " illegal")
	}
	return &NodeInfo{Id: n[2], Group: n[0], Name: n[1]}, nil
}

func NewUUID() string {
	uuids := uuid.NewV4()
	return uuids.String()
}
