package proxy

import (
	"github.com/ljun20160606/bifrost/net/socks"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"net"
)

func NoAuthSock5ProxyToSock5(listenAdder, targetAddr string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", targetAddr, auth, nil)
	if err != nil {
		return err
	}
	log.Infof("Now listening on: %v, Target address: %v", listenAdder, targetAddr)
	server := tunnel.Server{
		Addr: listenAdder,
		Handler: tunnel.HandlerFunc(func(conn net.Conn) {
			defer conn.Close()
			// auth
			err := socks.NoAuthRequired(conn, conn)
			if err != nil {
				log.Error(err)
				return
			}
			// cmd
			req, err := socks.ParseCmdRequest(conn)
			if err != nil {
				log.Error(err)
				return
			}
			if req.Cmd != socks.CmdConnect {
				log.Error(errors.New("cmd not support"))
				return
			}
			// connect
			log.Info("Addr: ", req.Addr.String())
			target, err := dialer.Dial("tcp", req.Addr.String())
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
			_ = Transport(target, conn)
		}),
	}
	return server.ListenAndServe()
}

func ListenAndServerNoAuthSock5(address string) error {
	log.Infof("Listening on: %v", address)
	return (&tunnel.Server{
		Addr: address,
		Handler: tunnel.HandlerFunc(func(conn net.Conn) {
			defer conn.Close()
			// 免鉴权交互
			if err := socks.NoAuthRequired(conn, conn); err != nil {
				log.Error(err)
				return
			}
			target, err := HandleCmdRequest(conn)
			if err != nil {
				log.Error(err)
				return
			}
			defer target.Close()
			err = Transport(target, conn)
			if err != nil {
				log.Error(err)
			}
		}),
	}).ListenAndServe()
}
