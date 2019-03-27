package service

import (
	"github.com/ljun20160606/bifrost/tunnel"
)

type Tower struct {
	*tunnel.NodeInfo
	Clients map[string]*Client
}

func New(nodeInfo *tunnel.NodeInfo, addrs []string) *Tower {
	tower := &Tower{
		NodeInfo: nodeInfo,
		Clients:  make(map[string]*Client),
	}
	for i := range addrs {
		addr := addrs[i]
		tower.Clients[addr] = NewClient(nodeInfo, addr)
	}
	return tower
}

func (t *Tower) Upstream() {
	for i := range t.Clients {
		client := t.Clients[i]
		go client.Upstream()
	}
}
