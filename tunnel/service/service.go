package service

import (
	"errors"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/bifrost/tunnel/config"
	"strings"
)

type Tower struct {
	*tunnel.NodeInfo
	Clients map[string]*Client
}

func New(config *config.Service) (*Tower, error) {
	addr := config.BridgeAddr
	addrs := strings.Split(addr, ",")
	if len(addrs) == 0 {
		return nil, errors.New("addrs invalid, can receive 0.0.0.0:7000 or a group like 0.0.0.0:7000,0.0.0.0:7001")
	}
	password := config.Password
	if password == "" || len(password) > 255 {
		return nil, errors.New("length of password must be > 0 and < 255")
	}

	nodeInfo := &tunnel.NodeInfo{Group: config.Group, Name: config.Name, Id: tunnel.NewUUID(), Password: config.Password, Cipher: config.Cipher}
	tower := &Tower{
		NodeInfo: nodeInfo,
		Clients:  make(map[string]*Client),
	}
	for i := range addrs {
		addr := addrs[i]
		tower.Clients[addr] = NewClient(nodeInfo, addr)
	}
	return tower, nil
}

func (t *Tower) Upstream() {
	for i := range t.Clients {
		client := t.Clients[i]
		go client.Upstream()
	}
}
