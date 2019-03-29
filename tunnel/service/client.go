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
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

var socks5Server, _ = socks5.New(&socks5.Config{})

type Client struct {
	*tunnel.NodeInfo
	// bridge address
	Addr string
	// writer
	Writer io.Writer
	// reader
	Reader *bufio.Reader
	// session for mux
	Session *yamux.Session
	// logger
	logger *log.Entry
}

func NewClient(nodeInfo *tunnel.NodeInfo, addr string) *Client {
	return &Client{
		NodeInfo: nodeInfo,
		Addr:     addr,
		logger:   log.WithField("addr", addr),
	}
}

// first action upstream
func (c *Client) Upstream() {
	c.logger.Infof("Bridge Address: %v", c.Addr)
	for {
		c.logger.Info("Connect bridge")
		err := c.upstream()
		if err != nil {
			c.logger.Error(err)
		}
		c.logger.Info("Disconnect bridge try again later")
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
	nodeInfoBytes, err := json.Marshal(c.NodeInfo)
	if err != nil {
		return errors.Wrap(err, "Register marshal nodeInfo fail")
	}
	bytes, err := json.Marshal(tunnel.Request{
		ServiceId:  c.NodeInfo.Id,
		Method:     tunnel.MethodRegister,
		Attachment: nodeInfoBytes,
	})
	if err != nil {
		return errors.Wrap(err, "Register marshal request fail")
	}
	_, _ = c.Writer.Write(append(bytes, tunnel.Delim))
	return nil
}

// Async write heart
func (c *Client) keepAlive() {
	c.logger.Info("Heart beats start")
	for {
		select {
		case <-time.After(2 * time.Second):
			_, err := c.Writer.Write([]byte{tunnel.Delim})
			if err != nil {
				c.logger.Error("Heart beats fail", err)
				return
			}
		}
	}
}

// Wait a task with loop
func (c *Client) ConnectLoop() error {
	for {
		r, err := c.readTask()
		if err != nil {
			c.logger.Error(err)
			return err
		}
		if r.isHeart {
			continue
		}
		go c.Connect(r)
	}
}

type connectResp struct {
	content []byte
	isHeart bool
	message *bridge.Message
}

// Read a task from connect
func (c *Client) readTask() (r *connectResp, err error) {
	r = new(connectResp)
	r.content, err = c.Reader.ReadBytes(tunnel.Delim)
	if err != nil {
		err = errors.Wrap(err, "Task read fail")
		return
	}
	if len(r.content) == 0 || (len(r.content) == 1 && r.content[0] == 10) {
		r.isHeart = true
		return
	}
	r.message = new(bridge.Message)
	err = json.Unmarshal(r.content, r.message)
	if err != nil {
		err = errors.Wrap(err, "Task content is illegal "+string(r.content))
		return
	}
	return
}

// Connect event
func (c *Client) Connect(r *connectResp) {
	withId := log.WithField("taskId", r.message.TaskId)
	withId.Info("Get call")
	bytes, err := json.Marshal(tunnel.Request{
		ServiceId:  c.NodeInfo.Id,
		Method:     tunnel.MethodConn,
		Attachment: r.content,
	})
	if err != nil {
		withId.Error("marshal request fail")
		return
	}

	conn, err := c.Session.Open()
	if err != nil {
		withId.Error("connect fail")
		return
	}
	defer conn.Close()

	// connect task
	_, err = conn.Write(append(bytes, tunnel.Delim))
	if err != nil {
		withId.Warn("answer fail")
		return
	}

	socks5Server.ServeCmdConn(context.Background(), conn, bufio.NewReader(conn))
}
