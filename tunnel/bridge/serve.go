package bridge

import (
	"github.com/ljun20160606/bifrost/net/socks"
	"github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
)

var (
	internet, _ = tunnel.LocalIP()
)

type Server struct {
	// 通信中心
	BridgeServer *tunnel.Server
	// 代理中心
	ProxyServer *tunnel.Server
	// 消息中心
	Caller Caller
}

func NewServer(bridgeAddr, proxyAddr string) *Server {
	server := &Server{}
	server.BridgeServer = &tunnel.Server{
		Addr:    bridgeAddr,
		Handler: tunnel.HandlerFunc(server.HandleCommunication),
	}
	server.ProxyServer = &tunnel.Server{
		Addr:    proxyAddr,
		Handler: tunnel.HandlerFunc(server.HandleProxy),
	}
	_, port, _ := net.SplitHostPort(bridgeAddr)
	server.Caller = NewCaller(internet.String() + ":" + port)
	return server
}

func (s *Server) ListenAndServer() error {
	errChan := make(chan error)
	go func() {
		log.Infof("Bridge listening on: %v", s.BridgeServer.Addr)
		err := s.BridgeServer.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		log.Infof("Proxy listening on: %v", s.ProxyServer.Addr)
		err := s.ProxyServer.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
	return <-errChan
}

// 处理通信
func (s *Server) HandleCommunication(conn net.Conn) {
	node, err := NewNode(conn)
	if err != nil {
		log.Error(err)
		return
	}
	switch node.Method {
	case tunnel.MethodRegister:
		err = s.Caller.Register(node)
		if err != nil {
			log.Error(err)
			return
		}
	case tunnel.MethodConn:
		err = s.Caller.Connect(node)
		if err != nil {
			log.Error(err)
			return
		}
	default:
		panic(errors.Errorf("method %v not support", node.Method))
	}
}

func (s *Server) HandleProxy(conn net.Conn) {
	auth, err := socks.AuthRequired(conn, conn)
	if err != nil {
		_ = conn.Close()
		log.Error(err)
		return
	}
	node, err := s.Caller.Call(auth)
	if err != nil {
		_ = conn.Close()
		log.Error(err)
		return
	}
	log.Info("Auth success username:", auth.Username)
	log.Info("Transport start:", auth.Username)
	err = proxy.Transport(node.Conn, conn)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("代理结束:", auth.Username)
}
