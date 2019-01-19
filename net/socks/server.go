package socks

import (
	"errors"
	"io"
	"net"
)

// 协商版本及认证方式
// 只支持socks5
// VER NMETHODS METHOD
// 1   1        1-255
func ParseAuthRequest(r io.Reader) (*AuthRequest, error) {
	lengthByte := []byte{0}
	if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
		return nil, errors.New("short auth request")
	}
	ver := int(lengthByte[0])
	if ver != Version5 {
		return nil, errors.New("unexpected protocol version")
	}
	authMethods, err := ReadMethods(r)
	if err != nil {
		return nil, errors.New("short auth request")
	}
	return &AuthRequest{Version: ver, Methods: authMethods}, nil
}

// MarshalAuthReply returns an authentication reply in wire format.
func MarshalAuthReply(ver int, m AuthMethod) ([]byte, error) {
	return []byte{byte(ver), byte(m)}, nil
}

// AuthRequired handles authentication-required signaling
func AuthRequired(w io.Writer, r io.Reader) (*UsernamePassword, error) {
	req, err := ParseAuthRequest(r)
	if err != nil {
		return nil, err
	}
	b, _ := MarshalAuthReply(req.Version, AuthMethodUsernamePassword)
	n, err := w.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, errors.New("short write")
	}
	// version, username length
	lengthByte := []byte{0, 0}
	if _, err := io.ReadAtLeast(r, lengthByte, 2); err != nil {
		return nil, shortAuth
	}
	if lengthByte[0] != AuthUsernamePasswordVersion {
		return nil, errors.New("unexpected auth version")
	}
	// username
	usernameBytes := make([]byte, lengthByte[1])
	if _, err := io.ReadAtLeast(r, usernameBytes, int(lengthByte[1])); err != nil {
		return nil, shortAuth
	}
	// password length
	if _, err := io.ReadAtLeast(r, lengthByte[:1], 1); err != nil {
		return nil, shortAuth
	}
	// password
	passwordBytes := make([]byte, lengthByte[0])
	if _, err := io.ReadAtLeast(r, passwordBytes, int(lengthByte[0])); err != nil {
		return nil, shortAuth
	}
	// replay
	n, err = w.Write([]byte{byte(AuthUsernamePasswordVersion), authStatusSucceeded})
	if err != nil {
		return nil, err
	}
	return &UsernamePassword{Username: string(usernameBytes), Password: string(passwordBytes)}, nil
}

// NoAuthRequired handles a no-authentication-required signaling.
func NoAuthRequired(w io.Writer, r io.Reader) error {
	req, err := ParseAuthRequest(r)
	if err != nil {
		return err
	}
	b, err := MarshalAuthReply(req.Version, AuthMethodNotRequired)
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errors.New("short write")
	}
	return nil
}

var shortAuth = errors.New("short auth request")

// Only socks5
// ParseCmdRequest parses a command request.
func ParseCmdRequest(r io.Reader) (*CmdRequest, error) {
	// VER
	lengthByte := []byte{0}
	if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
		return nil, errors.New("short auth request")
	}
	ver := int(lengthByte[0])
	if ver != Version5 {
		return nil, errors.New("unexpected protocol version")
	}
	// CMD
	if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
		return nil, shortAuth
	}
	command := Command(lengthByte[0])
	// RSV
	if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
		return nil, shortAuth
	}
	if lengthByte[0] != 0 {
		return nil, errors.New("non-zero reserved field")
	}
	req := &CmdRequest{Version: ver, Cmd: command}
	// ATYP
	if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
		return nil, shortAuth
	}
	var l int
	switch lengthByte[0] {
	case AddrTypeIPv4:
		l = net.IPv4len
		req.Addr.IP = make(net.IP, net.IPv4len)
	case AddrTypeIPv6:
		l = net.IPv6len
		req.Addr.IP = make(net.IP, net.IPv6len)
	case AddrTypeFQDN:
		if _, err := io.ReadAtLeast(r, lengthByte, 1); err != nil {
			return nil, shortAuth
		}
		l = int(lengthByte[0])
	default:
		return nil, errors.New("unknown address type")
	}
	if req.Addr.IP != nil {
		if _, err := io.ReadAtLeast(r, req.Addr.IP, l); err != nil {
			return nil, shortAuth
		}
	} else {
		fqdnBytes := make([]byte, l)
		if _, err := io.ReadAtLeast(r, fqdnBytes, l); err != nil {
			return nil, shortAuth
		}
		req.Addr.Name = string(fqdnBytes)
	}
	// PORT
	portBytes := make([]byte, 2)
	if _, err := io.ReadAtLeast(r, portBytes, 2); err != nil {
		return nil, shortAuth
	}
	req.Addr.Port = (int(portBytes[0]) << 8) | int(portBytes[1])
	return req, nil
}

// MarshalCmdReply returns a command reply in wire format.
func MarshalCmdReply(ver int, reply Reply, a *Addr) ([]byte, error) {
	b := make([]byte, 4)
	b[0] = byte(ver)
	b[1] = byte(reply)
	if a.Name != "" {
		if len(a.Name) > 255 {
			return nil, errors.New("fqdn too long")
		}
		b[3] = AddrTypeFQDN
		b = append(b, byte(len(a.Name)))
		b = append(b, a.Name...)
	} else if ip4 := a.IP.To4(); ip4 != nil {
		b[3] = AddrTypeIPv4
		b = append(b, ip4...)
	} else if ip6 := a.IP.To16(); ip6 != nil {
		b[3] = AddrTypeIPv6
		b = append(b, ip6...)
	} else {
		return nil, errors.New("unknown address type")
	}
	b = append(b, byte(a.Port>>8), byte(a.Port))
	return b, nil
}
