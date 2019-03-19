package service

import (
	"github.com/ljun20160606/bifrost/tunnel"
)

type Tower struct {
	*tunnel.NodeInfo
	Clients map[string]*Client
}

func New(group, name string, addrs []string) *Tower {
	nodeInfo := &tunnel.NodeInfo{
		Id:    tunnel.NewUUID(),
		Group: group,
		Name:  name,
	}
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
