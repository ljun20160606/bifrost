package config

type Service struct {
	Group      string `yaml:"group"`
	Name       string `yaml:"name"`
	BridgeAddr string `yaml:"bridgeAddr"`
	Password   string `yaml:"password"`
	Cipher     string `yaml:"cipher"`
}

type Bridge struct {
	Addr      string `yaml:"addr"`
	ProxyAddr string `yaml:"proxyAddr"`
}

type Proxy struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
	Password        string `yaml:"password"`
	Cipher          string `yaml:"cipher"`
	Type            string `yaml:"type"`
}

type Mapping struct {
	Addr            string `yaml:"addr"`
	BridgeProxyAddr string `yaml:"bridgeProxyAddr"`
	RealAddr        string `yaml:"realAddr"`
	Group           string `yaml:"group"`
	Name            string `yaml:"name"`
	Password        string `yaml:"password"`
	Cipher          string `yaml:"cipher"`
}
