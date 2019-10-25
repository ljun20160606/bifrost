package bridge

import (
	"encoding/json"
	"github.com/ljun20160606/bifrost/tunnel"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type NodeListener struct {
	*tunnel.Session
	// 任务读写窗口
	sendCh chan []byte
	// 异常处理
	errFunc func(*NodeListener)
}

// 解析节点信息
func NewNodeListener(session *tunnel.Session, errFunc func(*NodeListener)) *NodeListener {
	return &NodeListener{Session: session, sendCh: make(chan []byte, 64), errFunc: errFunc}
}

func (s *NodeListener) Start() {
	go s.keepAlive()
	go s.send()
}

// Send heart in a loop
func (s *NodeListener) keepAlive() {
	s.Logger.Info("Ready keepAlive")
	s.Session.Conn = tunnel.SetConnectTimeout(s.Session.Conn, 30*time.Second, 10*time.Second)
	heart := make([]byte, 1)
	for {
		select {
		case <-s.Context.Done():
		default:
			_, err := s.Read(heart)
			if err != nil {
				if err == io.EOF {
					s.Logger.Info("Disconnect with client")
				} else {
					s.Logger.Error("heart beats error ", err)
				}
				s.Close()
				// 从注册列表中删除自己
				s.errFunc(s)
				return
			}
		}
	}
}

var heartbeat = []byte{tunnel.Delim}

// Send data in a loop
func (s *NodeListener) send() {
	for {
		var data []byte
		// 每秒更新心跳或接收数据
		select {
		case <-s.Context.Done():
			return
		case data = <-s.sendCh:
		case <-time.After(time.Second):
			data = heartbeat
		}
		// 计数
		if _, err := s.Write(data); err != nil {
			log.Error("send fail", err)
			s.Close()
			return
		}
	}
}

// Send a Task
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
	s.Logger.Infof("Call TaskId: %v", message.TaskId)
	s.sendCh <- append(data, tunnel.Delim)
	return true
}

// Cancel context and close conn
func (s *NodeListener) Close() {
	s.Cancel()
	_ = s.Conn.Close()
}
