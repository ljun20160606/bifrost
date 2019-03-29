package main

import "github.com/ljun20160606/bifrost/tunnel/config"

var defaultConfig = Config{
	Bridge: config.Bridge{
		Addr:      ":7000",
		ProxyAddr: ":8888",
	},
	Service: config.Service{
		Group:      "tangtangtang",
		BridgeAddr: ":7000",
		Password:   "20160606",
	},
	Proxy: config.Proxy{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		Group:           "tangtangtang",
		Password:        "20160606",
	},
	Mapping: config.Mapping{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		RealAddr:        "",
		Group:           "tangtangtang",
		Password:        "20160606",
	},
}

type Config struct {
	Bridge  config.Bridge  `yaml:"bridge"`
	Service config.Service `yaml:"service"`
	Proxy   config.Proxy   `yaml:"proxy"`
	Mapping config.Mapping `yaml:"mapping"`
}
