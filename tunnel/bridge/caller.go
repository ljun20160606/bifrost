package bridge

import (
	"github.com/ljun20160606/bifrost/net/socks"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Caller interface {
	// 拉取一个可用的代理节点
	Call(auth *socks.UsernamePassword) (*Node, error)
	// 注册服务节点
	Register(node *Node) error
	// 处理上报的代理节点
	Connect(node *Node) error
}

func NewCaller(bridgeAddress string) Caller {
	return &NodeCaller{
		bridgeAddr: bridgeAddress,
	}
}

type NodeCaller struct {
	// 通信地址
	bridgeAddr string
	// 任务锁
	mutex sync.RWMutex
	// 所有通信中心
	nodeCenter sync.Map
	// 任务中心
	taskCenter sync.Map
}

func (n *NodeCaller) Call(auth *socks.UsernamePassword) (*Node, error) {
	value, has := n.nodeCenter.Load(auth.Username)
	if !has {
		return nil, errors.Errorf("不存在相同组 %v", auth.Username)
	}
	uuids, _ := uuid.NewV4()
	taskId := uuids.String()
	channel := NewChannel(10 * time.Second)
	n.taskCenter.Store(taskId, channel)
	defer n.taskCenter.Delete(taskId)
	ok := value.(*NodeListener).Notify(&Message{
		TaskId:  taskId,
		Address: n.bridgeAddr,
	})
	if !ok {
		return nil, errors.Errorf("对应的服务节点无法接收任务 %v", auth.Username)
	}
	log.Info("等待任务响应", auth.Username)
	ret := channel.Get()
	if ret == nil {
		return nil, errors.New("等待任务超时")
	}
	node := ret.(*Node)
	return node, nil
}

func (n *NodeCaller) Register(node *Node) error {
	nodeListener := NewNodeListener(node, func(listener *NodeListener) {
		n.mutex.Lock()
		n.nodeCenter.Delete(listener.Group)
		n.mutex.Unlock()
	})
	n.mutex.Lock()
	defer n.mutex.Unlock()
	_, has := n.nodeCenter.Load(nodeListener.Group)
	if has {
		nodeListener.Close()
		return errors.Errorf("存在相同组，无法再注册 %v", nodeListener.Group)
	}
	n.nodeCenter.Store(nodeListener.Group, nodeListener)
	nodeListener.Start()
	return nil
}

func (n *NodeCaller) Connect(node *Node) error {
	slave, err := NewNodeSlave(node)
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
	err = value.(*Channel).Set(slave.Node)
	if err != nil {
		_ = slave.Close()
		return err
	}
	return nil
}