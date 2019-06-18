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
		Short: "net bridge",
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
		Short: "node service",
		Run: func(cmd *cobra.Command, args []string) {
			// 连接到网桥地址
			client, err := service.New(&a.Config.Service)
			if err != nil {
				fmt.Println(err)
				return
			}
			client.Upstream()

			// Block till ctrl+c or kill
			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, os.Kill)
			<-c
		},
	}

	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "local proxy",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := a.Config.Proxy.Addr
				targetAddr := a.Config.Proxy.BridgeProxyAddr
				group := a.Config.Proxy.Group
				name := a.Config.Proxy.Name
				password := a.Config.Proxy.Password
				proxyType := a.Config.Proxy.Type

				nodeInfo := &tunnel.NodeInfo{Group: group, Name: name, Id: tunnel.NewUUID(), Password: password}
				// SwitchyOmega调试，SwitchyOmega不支持socks5 auth，所以本地再代理一层
				auth, err := nodeInfo.ProxyAuth()
				if err != nil {
					done <- err
				}

				lowerProxyType := strings.ToLower(proxyType)
				var localSkipAuthProxy tunnelProxy.LocalSkipAuthProxy
				switch lowerProxyType {
				case "http":
					localSkipAuthProxy = tunnelProxy.HttpProxyToSock5
				case "socks5":
					fallthrough
				default:
					localSkipAuthProxy = tunnelProxy.NoAuthSock5ProxyToSocks5
				}
				err = localSkipAuthProxy(addr, targetAddr, auth)
				done <- err
			}()
			fmt.Println(<-done)
		},
	}

	mappingCmd := &cobra.Command{
		Use:   "mapping",
		Short: "local mapping agent",
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
	const bifrostYaml = ".bifrost.yaml"
	err := di.ConfigLoadFile(bifrostYaml, di.YAML)
	if err != nil {
		fmt.Println(bifrostYaml + " not found, use defaultConfig")
		app.Config = &defaultConfig
	}
	di.Start()
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
