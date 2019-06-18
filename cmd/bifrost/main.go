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

type App struct {
	Config *Config `di:"#.{}"`
}

var (
	rootCmd   = &cobra.Command{}
	app       = new(App)
	bridgeCmd = &cobra.Command{
		Use:   "bridge",
		Short: "net bridge",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := app.Config.Bridge.Addr
				proxyAddr := app.Config.Bridge.ProxyAddr
				// 本地网桥启动，网桥地址:7000，代理地址:8888
				err := bridge.ListenAndServer(addr, proxyAddr)
				done <- err
			}()
			<-done
		},
	}

	serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "node service",
		Run: func(cmd *cobra.Command, args []string) {
			// 连接到网桥地址
			client, err := service.New(&app.Config.Service)
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

	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "local proxy",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := app.Config.Proxy.Addr
				targetAddr := app.Config.Proxy.BridgeProxyAddr
				group := app.Config.Proxy.Group
				name := app.Config.Proxy.Name
				password := app.Config.Proxy.Password
				proxyType := app.Config.Proxy.Type

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

	mappingCmd = &cobra.Command{
		Use:   "mapping",
		Short: "local mapping agent",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			addr := app.Config.Mapping.Addr
			targetAddr := app.Config.Mapping.BridgeProxyAddr
			realAddr := app.Config.Mapping.RealAddr
			password := app.Config.Mapping.Password
			if realAddr == "" {
				fmt.Println("realAddr can not be null")
				return
			}
			group := app.Config.Mapping.Group
			name := app.Config.Mapping.Name

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
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	rootCmd.AddCommand(bridgeCmd, proxyCmd, serviceCmd, mappingCmd)

	// Name of config
	const bifrostYaml = ".bifrost.yaml"
	rootCmd.PersistentFlags().StringP("file", "f", bifrostYaml, "Name of .biforst.yaml(Default is '.biforst.yaml')")

	// Load config
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		di.Put(app)
		configPath := rootCmd.PersistentFlags().Lookup("file").Value.String()
		err := di.ConfigLoadFile(configPath, di.YAML)
		if err != nil {
			fmt.Println(configPath + " not found, use defaultConfig")
			app.Config = &defaultConfig
		}
		di.Start()
	}
}

func main() {
	// Run
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
