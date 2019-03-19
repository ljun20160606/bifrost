package bridge

import (
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Caller interface {
	// 拉取一个可用的代理节点
	Call(user, password string) (*tunnel.Session, error)
	// 注册服务节点
	Register(node *tunnel.Session) error
	// 处理上报的代理节点
	Connect(node *tunnel.Session) error
}

func NewCaller(bridgeAddress string) Caller {
	return &NodeCaller{
		bridgeAddr: bridgeAddress,
		registry:   NewRegistry(),
	}
}

type NodeCaller struct {
	// 通信地址
	bridgeAddr string
	// 任务锁
	mutex sync.RWMutex
	// 所有通信中心
	registry Registry
	// 任务中心
	taskCenter sync.Map
}

// Lookup a node that group and name is same with username and password
func (n *NodeCaller) Call(user, password string) (*tunnel.Session, error) {
	listener, has := n.registry.Select(user, password)
	if !has {
		return nil, errors.Errorf("不存在相同组 %v", user)
	}
	taskId := tunnel.NewUUID()
	channel := NewChannel(10 * time.Second)
	n.taskCenter.Store(taskId, channel)
	defer n.taskCenter.Delete(taskId)
	ok := listener.Notify(&Message{
		TaskId:  taskId,
		Address: n.bridgeAddr,
	})
	if !ok {
		return nil, errors.Errorf("对应的服务节点无法接收任务 %v", user)
	}
	ret := channel.Get()
	if ret == nil {
		return nil, errors.New("等待任务超时")
	}
	session := ret.(*tunnel.Session)
	return session, nil
}

// Event register
func (n *NodeCaller) Register(session *tunnel.Session) error {
	nodeListener := NewNodeListener(session, func(listener *NodeListener) {
		listener.Logger.Info("Unregister service")
		n.registry.Delete(session.Request.NodeInfo)
	})
	n.mutex.Lock()
	defer n.mutex.Unlock()
	session.Logger.Info("Register service")
	n.registry.Register(nodeListener)
	nodeListener.Start()
	return nil
}

// Event connect
func (n *NodeCaller) Connect(session *tunnel.Session) error {
	slave, err := NewNodeSlave(session)
	if err != nil {
		return err
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	value, has := n.taskCenter.Load(slave.Message.TaskId)
	if !has {
		_ = slave.Close()
		return errors.Errorf("找不到任务 %v", slave.Message.TaskId)
	}
	err = value.(*Channel).Set(slave.Session)
	if err != nil {
		_ = slave.Close()
		return err
	}
	return nil
}
