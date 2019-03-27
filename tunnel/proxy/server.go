package proxy

import (
	"github.com/ljun20160606/go-socks5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
)

func NoAuthSock5ProxyToSock5(listenAdder, targetAddr string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", targetAddr, auth, nil)
	if err != nil {
		return err
	}
	log.Infof("Now listening on: %v, Target address: %v, User: %v", listenAdder, targetAddr, auth.User)
	s, _ := socks5.New(&socks5.Config{
		Dial: func(ctx context.Context, network string, addr *socks5.AddrSpec) (conn net.Conn, e error) {
			address := addr.Address()
			log.Info("Addr: ", addr.String())
			return dialer.Dial("tcp", address)
		},
	})
	return s.ListenAndServe("tcp", listenAdder)
}
