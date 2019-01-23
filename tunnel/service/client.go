package service

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/hashicorp/yamux"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/bifrost/tunnel/bridge"
	"github.com/ljun20160606/go-socks5"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

var socks5Server, _ = socks5.New(&socks5.Config{})

type Client struct {
	// ClientId
	Id string
	// account
	Group string
	// password
	Name string
	// bridge address
	Addr string
	// writer
	Writer io.Writer
	// reader
	Reader *bufio.Reader
	// session for mux
	Session *yamux.Session
}

func NewClient(group, name, addr string) *Client {
	uuids, _ := uuid.NewV4()
	return &Client{
		Id:    uuids.String(),
		Group: group,
		Name:  name,
		Addr:  addr,
	}
}

// first action upstream
func (c *Client) Upstream() {
	log.Infof("Bridge Address: %v", c.Addr)
	for {
		log.Info("开始连接网桥")
		err := c.upstream()
		if err != nil {
			log.Error(err)
		}
		log.Info("无法连接网桥稍后尝试重连")
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) upstream() error {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	// mux
	c.Session, _ = yamux.Client(conn, nil)
	stream, err := c.Session.OpenStream()
	if err != nil {
		return err
	}
	defer stream.Close()
	// read
	c.Reader = bufio.NewReader(stream)
	// write
	c.Writer = tunnel.SetConnectTimeout(stream, 30*time.Second, 10*time.Second)
	err = c.Register()
	if err != nil {
		return err
	}
	go c.keepAlive()
	return c.ConnectLoop()
}

// Register service
func (c *Client) Register() error {
	bytes, err := json.Marshal(bridge.NodeInfo{
		Id:     c.Id,
		Group:  c.Group,
		Name:   c.Name,
		Method: tunnel.MethodRegister,
	})
	if err != nil {
		return errors.Wrap(err, "序列化节点信息失败")
	}
	_, _ = c.Writer.Write(append(bytes, '\n'))
	return nil
}

// Async write heart
func (c *Client) keepAlive() {
	log.Info("上报心跳")
	for {
		select {
		case <-time.After(2 * time.Second):
			_, err := c.Writer.Write([]byte{'\n'})
			if err != nil {
				log.Error("上报心跳失败", err)
				return
			}
		}
	}
}

// Wait a task with loop
func (c *Client) ConnectLoop() error {
	for {
		message, isHeart, err := c.readTask()
		if err != nil {
			log.Error(err)
			return err
		}
		if isHeart {
			continue
		}
		go c.Connect(message)
	}
}

// Read a task from connect
func (c *Client) readTask() (message []byte, isHeart bool, err error) {
	bytes, err := c.Reader.ReadBytes('\n')
	if err != nil {
		err = errors.Wrap(err, "读取任务失败")
		return
	}
	if len(bytes) == 0 || (len(bytes) == 1 && bytes[0] == 10) {
		isHeart = true
		return
	}
	err = json.Unmarshal(bytes, new(bridge.Message))
	if err != nil {
		err = errors.Wrap(err, "读取数据格式有误"+string(bytes))
		return
	}
	message = bytes[:len(bytes)-1]
	return
}

// Connect event
func (c *Client) Connect(messageBytes []byte) {
	log.Info("读取任务成功", string(messageBytes))
	bytes, err := json.Marshal(bridge.NodeInfo{
		Id:         c.Id,
		Group:      c.Group,
		Name:       c.Name,
		Method:     tunnel.MethodConn,
		Attachment: messageBytes,
	})
	if err != nil {
		log.Error("序列化节点信息失败")
		return
	}

	message := new(bridge.Message)
	err = json.Unmarshal(messageBytes, message)
	if err != nil {
		log.Error("反序列化节点信息失败")
		return
	}

	conn, err := c.Session.Open()
	//conn, err := net.Dial("tcp", message.Address)
	if err != nil {
		log.Error("Connect fail", string(messageBytes))
		return
	}
	defer conn.Close()

	// connect task
	_, err = conn.Write(append(bytes, '\n'))
	if err != nil {
		log.Warn("代理上报失败")
		return
	}

	socks5Server.ServeCmdConn(context.Background(), conn, bufio.NewReader(conn))
}
