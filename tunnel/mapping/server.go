package mapping

import (
	biproxy "github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"net"
)

func Rewrite(listenAdder, targetAddr, realAddr string, auth *proxy.Auth) error {
	dialer, err := proxy.SOCKS5("tcp", targetAddr, auth, nil)
	if err != nil {
		return err
	}
	log.Infof("Now listening on: %v, Target address: %v, Real address %v", listenAdder, targetAddr, realAddr)
	server := &tunnel.Server{
		Addr: listenAdder,
		Handler: tunnel.HandlerFunc(func(conn net.Conn) {
			defer conn.Close()
			destConn, err := dialer.Dial("tcp", realAddr)
			if err != nil {
				log.Error("connect realAddr fail ", err)
				return
			}
			err = biproxy.Transport(conn, destConn)
			if err != nil {
				return
			}
		}),
	}
	return server.ListenAndServe()
}
