package main

var defaultConfig = Config{
	Bridge: Bridge{
		Addr:      ":7000",
		ProxyAddr: ":8888",
	},
	Service: Service{
		Group:      "tangtangtang",
		BridgeAddr: ":7000",
		Password:   "20160606",
	},
	Proxy: Proxy{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		Group:           "tangtangtang",
		Password:        "20160606",
	},
	Mapping: Mapping{
		Addr:            ":8080",
		BridgeProxyAddr: ":8888",
		RealAddr:        "",
		Group:           "tangtangtang",
		Password:        "20160606",
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
	Password   string `yaml:"password"`
}

type Proxy struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
	Password        string `yaml:"password"`
}

type Mapping struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	RealAddr        string `yaml:"realAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
	Password        string `yaml:"password"`
}
