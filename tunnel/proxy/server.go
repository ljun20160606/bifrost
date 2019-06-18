package proxy

import (
	"github.com/ljun20160606/go-socks5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"gopkg.in/elazarl/goproxy.v1"
	"net"
	"net/http"
)

type LocalSkipAuthProxy func(listenAdder, targetAddr string, auth *proxy.Auth) error

// Convert socks5 to auth socks5
func NoAuthSock5ProxyToSocks5(listenAdder, targetAddr string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", targetAddr, auth, nil)
	if err != nil {
		return err
	}
	log.Infof("Now listening on: %v, Target address: %v, User: %v, Type: socks5", listenAdder, targetAddr, auth.User)
	s, _ := socks5.New(&socks5.Config{
		Dial: func(ctx context.Context, network string, addr *socks5.AddrSpec) (conn net.Conn, e error) {
			address := addr.Address()
			log.Info("Addr: ", addr.String())
			return dialer.Dial("tcp", address)
		},
	})
	return s.ListenAndServe("tcp", listenAdder)
}

// Convert http to auth socks5
func HttpProxyToSock5(listenAdder, targetAddr string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", targetAddr, auth, nil)
	if err != nil {
		return err
	}
	log.Infof("Now listening on: %v, Target address: %v, User: %v, Type: http", listenAdder, targetAddr, auth.User)
	server := goproxy.NewProxyHttpServer()
	server.Verbose = true
	server.ConnectDial = dialer.Dial
	return http.ListenAndServe(listenAdder, server)
}
