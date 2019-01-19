package bridge

import (
	"bytes"
	"encoding/json"
	"github.com/ljun20160606/bifrost/tunnel"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type NodeListener struct {
	*Node
	// 任务读写窗口
	rw chan []byte

	errFunc func(*NodeListener)
}

// 解析节点信息
func NewNodeListener(node *Node, errFunc func(*NodeListener)) *NodeListener {
	return &NodeListener{Node: node, rw: make(chan []byte, 64), errFunc: errFunc}
}

func (s *NodeListener) Start() {
	go s.ReadHeart()
	go s.WriteLoop()
}

func (s *NodeListener) Notify(message *Message) bool {
	select {
	case <-s.Context.Done():
		return false
	default:
	}
	data, err := json.Marshal(message)
	if err != nil {
		return false
	}
	log.Info("通知任务", string(data))
	s.rw <- append(data, '\n')
	return true
}

func (s *NodeListener) WriteLoop() {
	buf := bytes.NewBuffer(nil)
	for {
		buf.Reset()
		var data []byte
		var wrote bool
		// 每秒更新心跳或接收数据
		select {
		case <-s.Context.Done():
			return
		case data = <-s.rw:
			// 标记为已写
			wrote = true
		case <-time.After(time.Second):
			data = []byte{Delim}
		}
		buf.Write(data)
		// 计数
		var count int
	COMPOSITE:
		for {
			select {
			case data = <-s.rw:
				if !wrote {
					buf.Reset()
				}
				buf.Write(data)
				count++
				if count > 60 {
					break COMPOSITE
				}
			default:
				break COMPOSITE
			}
		}
		if _, err := s.Write(buf.Bytes()); err != nil {
			log.Error("写失败", err)
			s.Close()
			return
		}
	}
}

func (s *NodeListener) ReadHeart() {
	log.Infof("%v 开始心跳", s.Group)
	s.Node.Conn = tunnel.SetConnectTimeout(s.Node.Conn, 30*time.Second, 10*time.Second)
	heart := make([]byte, 1)
	for {
		select {
		case <-s.Context.Done():
		default:
			_, err := s.Read(heart)
			if err != nil {
				if err != io.EOF {
					log.Errorf("%v 心跳错误 %v", s.Group, err)
				} else {
					log.Errorf("%v 断开", s.Group)
				}
				s.Close()
				// 从注册列表中删除自己
				s.errFunc(s)
				return
			}
		}
	}
}
func (s *NodeListener) Close() {
	s.Cancel()
	_ = s.Conn.Close()
}
