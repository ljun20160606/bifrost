package bridge

import (
	"context"
	"github.com/hashicorp/yamux"
	"github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/go-socks5"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

const (
	sessionKey = "session"
	logKey     = "log"
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
	// socks5
	SocksServer *socks5.Server
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
	server.SocksServer, _ = socks5.New(&socks5.Config{
		Credentials: server,
	})
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
	defer conn.Close()
	// Setup server side of yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Error("multiplexing fail", err)
		return
	}

	for {
		stream, err := session.Accept()
		if err != nil {
			if err != io.EOF {
				log.Error("session accept fail", err)
			}
			return
		}
		go func() {
			tunnelSession, err := tunnel.NewSession(stream)
			if err != nil {
				log.Error(err)
				return
			}
			switch tunnelSession.Method {
			case tunnel.MethodRegister:
				log.WithField("service", tunnelSession.Request).Info("")
				err = s.Caller.Register(tunnelSession)
				if err != nil {
					tunnelSession.Logger.Error(err)
					return
				}
			case tunnel.MethodConn:
				err = s.Caller.Connect(tunnelSession)
				if err != nil {
					tunnelSession.Logger.Error(err)
					return
				}
			default:
				panic(errors.Errorf("method %v not support", tunnelSession.Method))
			}
		}()
	}
}

func (s *Server) HandleProxy(conn net.Conn) {
	defer conn.Close()
	withIp := log.WithField("remoteAddr", conn.RemoteAddr().String())
	ctx := context.Background()
	ctx = context.WithValue(ctx, logKey, withIp)
	ctx, _, err := s.SocksServer.Authenticate(ctx, conn, conn)
	if err != nil {
		withIp.Error("auth fail ", err)
		return
	}

	err = proxy.Transport(ctx.Value(sessionKey).(*tunnel.Session).Conn, conn)
	if err != nil {
		return
	}
}

// valid username password and call service
func (s *Server) Valid(ctx context.Context, user, password string) (context.Context, bool) {
	withField := ctx.Value(logKey).(*log.Entry).WithField("user", user)
	nodeInfo, err := tunnel.ParseUser(user)
	if err != nil {
		withField.Error("auth fail", err)
		return ctx, false
	}
	node, err := s.Caller.Call(nodeInfo)
	if err != nil {
		withField.Error("auth fail", err)
		return ctx, false
	}
	if password != node.Password {
		withField.Error("auth fail password error")
		return ctx, false
	}
	withField.Info("Auth success")
	ctx = context.WithValue(ctx, sessionKey, node)
	return ctx, true
}
