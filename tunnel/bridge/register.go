package bridge

import (
	"stathat.com/c/consistent"
	"sync"
)

type Registry interface {
	Register(listener *NodeListener)

	Select(group, tempId string) (listener *NodeListener, ok bool)

	Find(id string) (listener *NodeListener, ok bool)

	Delete(serviceId string)
}

type RegistryCenter struct {
	mutex sync.RWMutex
	// path:GroupParty
	groupMap map[string]interface{}
	// serviceId:NodeListener
	serviceMap map[string]*NodeListener
}

func NewRegistry() Registry {
	return &RegistryCenter{
		groupMap:   make(map[string]interface{}),
		serviceMap: make(map[string]*NodeListener),
	}
}

// Add group
// Add service
func (g *RegistryCenter) Register(listener *NodeListener) {
	g.mutex.Lock()
	group := listener.Account()
	party, ok := g.groupMap[group]
	if !ok {
		party = NewParty()
		g.groupMap[group] = party
	}
	consistentParty := party.(*ConsistentParty)
	consistentParty.Add(listener)
	g.serviceMap[listener.Id] = listener
	g.mutex.Unlock()
}

// Get serviceId from group
// Find listener from serviceId
func (g *RegistryCenter) Select(group, tempId string) (listener *NodeListener, ok bool) {
	g.mutex.Lock()
	value, has := g.groupMap[group]
	g.mutex.Unlock()
	if !has {
		return nil, has
	}
	serviceId := value.(*ConsistentParty).Select(tempId)
	return g.Find(serviceId)
}

func (g *RegistryCenter) Find(id string) (listener *NodeListener, ok bool) {
	g.mutex.RLock()
	nodeListener, has := g.serviceMap[id]
	g.mutex.RUnlock()
	return nodeListener, has
}

func (g *RegistryCenter) Delete(serviceId string) {
	listener, has := g.Find(serviceId)
	if !has {
		return
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	groupName := listener.Account()
	group, ok := g.groupMap[groupName]
	if !ok {
		return
	}
	party := group.(*ConsistentParty)
	// delete from group
	party.Delete(listener.Id)
	if party.Len() == 0 {
		// delete from groupMap
		delete(g.groupMap, groupName)
	}

	// delete from serviceMap
	delete(g.serviceMap, serviceId)
}

// Consistent
type ConsistentParty struct {
	// id ring
	consistent *consistent.Consistent
}

func NewParty() *ConsistentParty {
	return &ConsistentParty{
		consistent: consistent.New(),
	}
}

func (g *ConsistentParty) Add(listener *NodeListener) {
	g.consistent.Add(listener.Id)
}

func (g *ConsistentParty) Select(id string) string {
	listenerId, _ := g.consistent.Get(id)
	return listenerId
}

func (g *ConsistentParty) Delete(id string) {
	g.consistent.Remove(id)
}

func (g *ConsistentParty) Len() int {
	return len(g.consistent.Members())
}
