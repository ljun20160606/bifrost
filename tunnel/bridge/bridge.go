package bridge

func ListenAndServer(bridgeAddr, proxyAddr string) error {
	return NewServer(bridgeAddr, proxyAddr).ListenAndServer()
}
