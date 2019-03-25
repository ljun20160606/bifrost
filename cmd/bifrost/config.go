package main

var defaultConfig = Config{
	Bridge: Bridge{
		Addr:      ":7000",
		ProxyAddr: ":8888",
	},
	Service: Service{
		Group:      "tangtangtang",
		Name:       "ljun",
		BridgeAddr: ":7000",
	},
	Proxy: Proxy{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		Group:           "tangtangtang",
		Name:            "ljun",
	},
	Mapping: Mapping{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		RealAddr:        "",
		Group:           "tangtangtang",
		Name:            "ljun",
	},
}

type Config struct {
	Bridge  Bridge  `yaml:"bridge"`
	Service Service `yaml:"service"`
	Proxy   Proxy   `yaml:"proxy"`
	Mapping Mapping `yaml:"mapping"`
}

type Bridge struct {
	Addr      string `yaml:"addr"`
	ProxyAddr string `yaml:"proxyAddr"`
}

type Service struct {
	Group      string `yaml:"group"`
	Name       string `yaml:"name"`
	BridgeAddr string `yaml:"bridgeAddr"`
}

type Proxy struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
}

type Mapping struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	RealAddr        string `yaml:"realAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
}
