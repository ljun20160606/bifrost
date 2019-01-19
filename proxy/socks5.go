package proxy

import (
	"github.com/ljun20160606/bifrost/net/socks"
	"github.com/pkg/errors"
	"net"
)

func HandleCmdRequest(conn net.Conn) (net.Conn, error) {
	req, err := ParseCmdRequest(conn)
	if err != nil {
		return nil, err
	}
	if req.Cmd != socks.CmdConnect {
		return nil, errors.New("cmd not support")
	}
	addr := (&net.TCPAddr{IP: req.Addr.IP, Port: req.Addr.Port}).String()
	target, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	local := target.LocalAddr().(*net.TCPAddr)
	bind := &socks.Addr{IP: local.IP, Port: local.Port}
	cmdReply, err := socks.MarshalCmdReply(socks.Version5, socks.StatusSucceeded, bind)
	if err != nil {
		_ = target.Close()
		return nil, err
	}
	_, _ = conn.Write(cmdReply)
	return target, nil
}

func ParseCmdRequest(conn net.Conn) (*socks.CmdRequest, error) {
	req, err := socks.ParseCmdRequest(conn)
	if err != nil {
		return nil, errors.Wrap(err, "解析CmdRequest失败")
	}
	err = ResolveDNS(req)
	if err != nil {
		return nil, errors.Wrap(err, "解析DNS失败")
	}
	return req, nil
}

//如果cmd中的name不为空则使用本地dns解析
func ResolveDNS(req *socks.CmdRequest) error {
	name := req.Addr.Name
	if name == "" {
		return nil
	}
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return err
	}
	req.Addr.IP = addr.IP
	return nil
}
