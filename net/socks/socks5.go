package socks

import (
	"io"
)

// An AuthRequest represents an authentication request.
type AuthRequest struct {
	Version int
	Methods []AuthMethod
}

// 读取方法
func ReadMethods(r io.Reader) ([]AuthMethod, error) {
	nmethods := []byte{0}
	if _, err := io.ReadAtLeast(r, nmethods, 1); err != nil {
		return nil, err
	}
	length := int(nmethods[0])
	// 等于0则长度为0
	if length == 0 {
		return nil, nil
	}
	methods := make([]byte, length)
	if _, err := io.ReadAtLeast(r, methods, length); err != nil {
		return nil, err
	}
	authMethods := make([]AuthMethod, length)
	for i := range methods {
		authMethods[i] = AuthMethod(methods[i])
	}
	return authMethods, nil
}

// A CmdRequest repesents a command request.
type CmdRequest struct {
	Version int
	Cmd     Command
	Addr    Addr
}
