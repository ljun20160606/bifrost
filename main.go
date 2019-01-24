package main

import (
	"fmt"
	tunnelProxy "github.com/ljun20160606/bifrost/proxy"
	"github.com/ljun20160606/bifrost/tunnel/bridge"
	"github.com/ljun20160606/bifrost/tunnel/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/proxy"
	"os"
	"os/signal"
	"strings"
)

var (
	rootCmd   = &cobra.Command{}
	bridgeCmd = &cobra.Command{
		Use:   "bridge",
		Short: "网桥",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := cmd.Flags().Lookup("addr").Value.String()
				proxyAddr := cmd.Flags().Lookup("proxyAddr").Value.String()
				// 本地网桥启动，网桥地址:7000，代理地址:8888
				err := bridge.ListenAndServer(addr, proxyAddr)
				done <- err
			}()
			<-done
		},
	}

	serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "上报服务",
		Run: func(cmd *cobra.Command, args []string) {
			group := cmd.Flags().Lookup("group").Value.String()
			name := cmd.Flags().Lookup("name").Value.String()
			addr := cmd.Flags().Lookup("addr").Value.String()
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

	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "本地代理",
		Run: func(cmd *cobra.Command, args []string) {
			done := make(chan error)
			go func() {
				addr := cmd.Flags().Lookup("addr").Value.String()
				targetAddr := cmd.Flags().Lookup("targetAddr").Value.String()
				group := cmd.Flags().Lookup("group").Value.String()
				name := cmd.Flags().Lookup("name").Value.String()
				// SwitchyOmega调试，SwitchyOmega不支持socks5 auth，所以本地再代理一层
				err := tunnelProxy.NoAuthSock5ProxyToSock5(addr, targetAddr, &proxy.Auth{
					User:     group,
					Password: name,
				})
				done <- err
			}()
			<-done
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	bridgeCmd.Flags().StringP("addr", "a", ":7000", "网桥地址")
	bridgeCmd.Flags().StringP("proxyAddr", "p", ":8888", "网桥代理地址")

	serviceCmd.Flags().StringP("group", "g", "tangtangtang", "分组")
	serviceCmd.Flags().StringP("name", "n", "ljun", "名称")
	serviceCmd.Flags().StringP("addr", "a", ":7000", "网桥地址，接受多个网桥地址使用`,`分割，如 :7000,:7001")

	proxyCmd.Flags().StringP("addr", "p", ":8080", "本地代理地址")
	proxyCmd.Flags().StringP("targetAddr", "t", ":8888", "网桥代理地址")
	proxyCmd.Flags().StringP("group", "g", "tangtangtang", "分组")
	proxyCmd.Flags().StringP("name", "n", "ljun", "名称")
}

func main() {
	rootCmd.AddCommand(bridgeCmd, proxyCmd, serviceCmd)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
