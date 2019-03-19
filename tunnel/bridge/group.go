package bridge

import (
	"github.com/ljun20160606/bifrost/tunnel"
	"stathat.com/c/consistent"
	"sync"
)

type Registry interface {
	Register(listener *NodeListener)

	Select(group, tempId string) (listener *NodeListener, ok bool)

	Delete(nodeInfo *tunnel.NodeInfo)
}

type RegistryCenter struct {
	mutex sync.Mutex
	// path groupParty
	nodeCenter map[string]interface{}
}

func NewRegistry() Registry {
	return &RegistryCenter{
		nodeCenter: make(map[string]interface{}),
	}
}

func (g *RegistryCenter) Register(listener *NodeListener) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	path := listener.Path()
	party, ok := g.nodeCenter[path]
	if !ok {
		party = NewParty()
		g.nodeCenter[path] = party
	}
	consistentParty := party.(*ConsistentParty)
	consistentParty.Add(listener)
}

func (g *RegistryCenter) Select(group, tempId string) (listener *NodeListener, ok bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	value, has := g.nodeCenter[group]
	if !has {
		return nil, has
	}
	return value.(*ConsistentParty).Select(tempId), has
}

func (g *RegistryCenter) Delete(nodeInfo *tunnel.NodeInfo) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	path := nodeInfo.Path()
	group, ok := g.nodeCenter[path]
	if !ok {
		return
	}
	party := group.(*ConsistentParty)
	// delete from group
	party.Delete(nodeInfo)
	if party.Len() == 0 {
		delete(g.nodeCenter, path)
	}
}

// Consistent
type ConsistentParty struct {
	// id node
	listeners map[string]*NodeListener
	// id ring
	consistent *consistent.Consistent
}

func NewParty() *ConsistentParty {
	return &ConsistentParty{
		listeners:  make(map[string]*NodeListener),
		consistent: consistent.New(),
	}
}

func (g *ConsistentParty) Add(listener *NodeListener) {
	g.listeners[listener.Id] = listener
	g.consistent.Add(listener.Id)
}

func (g *ConsistentParty) Select(id string) (listener *NodeListener) {
	listenerId, _ := g.consistent.Get(id)
	return g.listeners[listenerId]
}

func (g *ConsistentParty) Delete(nodeInfo *tunnel.NodeInfo) {
	delete(g.listeners, nodeInfo.Id)
	g.consistent.Remove(nodeInfo.Id)
}

func (g *ConsistentParty) Len() int {
	return len(g.listeners)
}
