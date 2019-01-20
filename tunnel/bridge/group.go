package bridge

import (
	"sync"
)

type Group interface {
	Register(listener *NodeListener)

	Select(group, name string) (listener *NodeListener, ok bool)

	Delete(nodeInfo *NodeInfo)
}

type GroupCenter struct {
	mutex      sync.Mutex
	nodeCenter map[string]interface{}
}

func NewGroupCenter() Group {
	return &GroupCenter{
		nodeCenter: make(map[string]interface{}),
	}
}

func (g *GroupCenter) Register(listener *NodeListener) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	path := genPath(listener.Group, listener.Name)
	groupParty, ok := g.nodeCenter[path]
	if !ok {
		groupParty = new(GroupParty)
		g.nodeCenter[path] = groupParty
	}
	party := groupParty.(*GroupParty)
	party.Add(listener)
}

func (g *GroupCenter) Select(group, name string) (listener *NodeListener, ok bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	value, has := g.nodeCenter[genPath(group, name)]
	if !has {
		return nil, has
	}
	return value.(*GroupParty).Select(), has
}

func (g *GroupCenter) Delete(nodeInfo *NodeInfo) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	path := genPath(nodeInfo.Group, nodeInfo.Name)
	group, ok := g.nodeCenter[path]
	if !ok {
		return
	}
	party := group.(*GroupParty)
	// delete from group
	party.Delete(nodeInfo)
	if party.Len() == 0 {
		delete(g.nodeCenter, path)
	}
}

func genPath(group, name string) string {
	return group + "/" + name
}

// Robin
type GroupParty struct {
	counter   int
	listeners []*NodeListener
}

func (g *GroupParty) Add(listener *NodeListener) {
	g.listeners = append(g.listeners, listener)
}

func (g *GroupParty) Select() (listener *NodeListener) {
	nodeListener := g.listeners[g.counter%len(g.listeners)]
	g.counter += 1
	return nodeListener
}

func (g *GroupParty) Delete(nodeInfo *NodeInfo) {
	for i := range g.listeners {
		listener := g.listeners[i]
		if listener.Id == nodeInfo.Id {
			g.listeners = append(g.listeners[:i], g.listeners[i+1:]...)
			break
		}
	}
}

func (g *GroupParty) Len() int {
	return len(g.listeners)
}
