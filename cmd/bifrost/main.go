package main

import (
	"fmt"
	tunnelProxy "github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel"
	"github.com/ljun20160606/bifrost/tunnel/bridge"
	"github.com/ljun20160606/bifrost/tunnel/mapping"
	"github.com/ljun20160606/bifrost/tunnel/service"
	"github.com/ljun20160606/di"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/proxy"
	"os"
	"os/signal"
	"strings"
)

var (
	rootCmd   = &cobra.Command{}
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
			addr := a.Config.Service.Addr
			addrs := strings.Split(addr, ",")
			if len(addrs) == 0 {
				fmt.Println("addrs invalid, can receive 0.0.0.0:7000 or a group like 0.0.0.0:7000,0.0.0.0:7001")
				return
			}
			// 连接到网桥地址
			client := service.New(group, name, addrs)
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
				targetAddr := a.Config.Proxy.TargetAddr
				group := a.Config.Proxy.Group
				name := a.Config.Proxy.Name
				// SwitchyOmega调试，SwitchyOmega不支持socks5 auth，所以本地再代理一层
				err := tunnelProxy.NoAuthSock5ProxyToSock5(addr, targetAddr, &proxy.Auth{
					User:     tunnel.BuildRealGroup(group, name),
					Password: tunnel.NewUUID(),
				})
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
			targetAddr := a.Config.Mapping.TargetAddr
			realAddr := a.Config.Mapping.RealAddr
			if realAddr == "" {
				fmt.Println("realAddr不能为空")
				return
			}
			group := a.Config.Mapping.Group
			name := a.Config.Mapping.Name
			//group := cmd.Flags().Lookup("group").Value.String()
			//name := cmd.Flags().Lookup("name").Value.String()
			go func() {
				err := mapping.Rewrite(addr, targetAddr, realAddr, &proxy.Auth{
					User:     tunnel.BuildRealGroup(group, name),
					Password: tunnel.NewUUID(),
				})
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
		fmt.Println(".bifrost Not Found, Use DefaultConfig")
		app.Config = &defaultConfig
	}
	di.Start()
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
