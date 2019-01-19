package tunnel

import "net"

func LocalIP() (internet net.IP, intranet net.IP) {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil {
				ip = ip.To4()
				if len(ip) != 4 {
					continue
				}
				if ip.IsLoopback() {
					continue
				}
				if ip[0] == 192 && ip[1] == 168 {
					intranet = ip
					continue
				}
				if ip[0] == 172 && ip[1] <= 31 && ip[1] >= 16 {
					intranet = ip
					continue
				}
				internet = ip
			} // process IP address
		}
	}
	if internet == nil {
		internet = intranet
	}
	return
}
