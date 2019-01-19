package service

import (
	"bufio"
	"encoding/json"
	"github.com/ljun20160606/bifrost/net/socks"
	"github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/bifrost/tunnel/bridge"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Client struct {
	// account
	Group string
	// password
	Name string
	// bridge address
	Addr string
	// conn
	Conn net.Conn
	// reader
	Reader *bufio.Reader
}

func NewClient(group, name, addr string) *Client {
	return &Client{
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
	c.Reader = bufio.NewReader(conn)
	c.Conn = tunnel.SetConnectTimeout(conn, 30*time.Second, 10*time.Second)
	// 上报
	err = c.Register()
	if err != nil {
		return err
	}
	go c.WriteHeart()
	return c.ConnectLoop()
}

func (c *Client) WriteHeart() {
	for {
		select {
		case <-time.After(2 * time.Second):
			_, err := c.Conn.Write([]byte{'\n'})
			if err != nil {
				log.Error("上报心跳失败", err)
				return
			}
			log.Info("上报心跳")
		}
	}
}

func (c *Client) ConnectLoop() error {
	for {
		message, isHeart, err := c.ReadTask()
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

func (c *Client) Connect(message []byte) {
	log.Info("读取任务成功 ", string(message))
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		log.Error("Connect fail", string(message))
		return
	}
	defer conn.Close()
	bytes, err := json.Marshal(bridge.NodeInfo{
		Group:      c.Group,
		Name:       c.Name,
		Method:     tunnel.MethodConn,
		Attachment: message,
	})
	if err != nil {
		log.Warn("序列化节点信息失败")
		return
	}
	_, err = conn.Write(append(bytes, '\n'))
	if err != nil {
		log.Warn("代理上报失败")
		return
	}
	// cmd
	req, err := proxy.ParseCmdRequest(conn)
	if err != nil {
		log.Error(err)
		return
	}
	if req.Cmd != socks.CmdConnect {
		log.Error(errors.New("cmd not support"))
		return
	}
	// connect
	addr := (&net.TCPAddr{IP: req.Addr.IP, Port: req.Addr.Port}).String()
	log.Info("Addr: ", addr)
	target, err := net.Dial("tcp", addr)
	if err != nil {
		log.Error(err)
		return
	}
	defer target.Close()
	// cmd reply
	cmdReply, err := socks.MarshalCmdReply(socks.Version5, socks.StatusSucceeded, &req.Addr)
	if err != nil {
		log.Error(err)
		return
	}
	_, _ = conn.Write(cmdReply)
	// transport
	err = proxy.Transport(target, conn)
	if err != nil {
		log.Error(err)
	}
}

func (c *Client) ReadTask() (message []byte, isHeart bool, err error) {
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

func (c *Client) Register() error {
	bytes, err := json.Marshal(bridge.NodeInfo{
		Group:  c.Group,
		Name:   c.Name,
		Method: tunnel.MethodRegister,
	})
	if err != nil {
		return errors.Wrap(err, "序列化节点信息失败")
	}
	_, _ = c.Conn.Write(append(bytes, '\n'))
	return nil
}
