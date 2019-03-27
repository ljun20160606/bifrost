package main

import (
	"fmt"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/bifrost/tunnel/bridge"
	"github.com/ljun20160606/bifrost/tunnel/mapping"
	tunnelProxy "github.com/ljun20160606/bifrost/tunnel/proxy"
	"github.com/ljun20160606/bifrost/tunnel/service"
	"github.com/ljun20160606/di"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
)

var (
	rootCmd = &cobra.Command{}
)

type App struct {
	Config *Config `di:"#.{}"`
}

func (a *App) Init() {
	bridgeCmd := &cobra.Command{
		Use:   "bridge",
		Short: "网桥",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := a.Config.Bridge.Addr
				proxyAddr := a.Config.Bridge.ProxyAddr
				// 本地网桥启动，网桥地址:7000，代理地址:8888
				err := bridge.ListenAndServer(addr, proxyAddr)
				done <- err
			}()
			<-done
		},
	}

	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "上报服务",
		Run: func(cmd *cobra.Command, args []string) {
			group := a.Config.Service.Group
			name := a.Config.Service.Name
			addr := a.Config.Service.BridgeAddr
			password := a.Config.Service.Password
			addrs := strings.Split(addr, ",")
			if len(addrs) == 0 {
				fmt.Println("addrs invalid, can receive 0.0.0.0:7000 or a group like 0.0.0.0:7000,0.0.0.0:7001")
				return
			}
			if password == "" || len(password) > 255 {
				fmt.Println("length of password must be > 0 and < 255")
				return
			}

			// 连接到网桥地址
			client := service.New(&tunnel.NodeInfo{Group: group, Name: name, Id: tunnel.NewUUID(), Password: password}, addrs)
			client.Upstream()

			// Block till ctrl+c or kill
			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, os.Kill)
			<-c
		},
	}

	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "本地代理",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := a.Config.Proxy.Addr
				targetAddr := a.Config.Proxy.BridgeProxyAddr
				group := a.Config.Proxy.Group
				name := a.Config.Proxy.Name
				password := a.Config.Proxy.Password

				nodeInfo := &tunnel.NodeInfo{Group: group, Name: name, Id: tunnel.NewUUID(), Password: password}
				// SwitchyOmega调试，SwitchyOmega不支持socks5 auth，所以本地再代理一层
				auth, err := nodeInfo.ProxyAuth()
				if err != nil {
					done <- err
				}
				err = tunnelProxy.NoAuthSock5ProxyToSock5(addr, targetAddr, auth)
				done <- err
			}()
			fmt.Println(<-done)
		},
	}

	mappingCmd := &cobra.Command{
		Use:   "mapping",
		Short: "映射",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			addr := a.Config.Mapping.Addr
			targetAddr := a.Config.Mapping.BridgeProxyAddr
			realAddr := a.Config.Mapping.RealAddr
			password := a.Config.Mapping.Password
			if realAddr == "" {
				fmt.Println("realAddr can not be null")
				return
			}
			group := a.Config.Mapping.Group
			name := a.Config.Mapping.Name

			nodeInfo := &tunnel.NodeInfo{Group: group, Name: name, Id: tunnel.NewUUID(), Password: password}
			go func() {
				auth, err := nodeInfo.ProxyAuth()
				if err != nil {
					done <- err
				}
				err = mapping.Rewrite(addr, targetAddr, realAddr, auth)
				done <- err
			}()
			fmt.Println(<-done)
		},
	}
	rootCmd.AddCommand(bridgeCmd, proxyCmd, serviceCmd, mappingCmd)
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	app := new(App)
	di.Put(app)
	err := di.ConfigLoadFile(".bifrost.yaml", di.YAML)
	if err != nil {
		fmt.Println(".bifrost.yaml not found, use defaultConfig")
		app.Config = &defaultConfig
	}
	di.Start()
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
