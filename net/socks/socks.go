// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package socks provides a SOCKS version 5 client implementation.
//
// SOCKS protocol version 5 is defined in RFC 1928.
// Username/Password authentication for SOCKS version 5 is defined in
// RFC 1929.
package socks

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
)

// A Command represents a SOCKS command.
type Command int

func (cmd Command) String() string {
	switch cmd {
	case CmdConnect:
		return "socks connect"
	case cmdBind:
		return "socks bind"
	default:
		return "socks " + strconv.Itoa(int(cmd))
	}
}

// An AuthMethod represents a SOCKS authentication method.
type AuthMethod int

// A Reply represents a SOCKS command reply code.
type Reply int

func (code Reply) String() string {
	switch code {
	case StatusSucceeded:
		return "succeeded"
	case 0x01:
		return "general SOCKS server failure"
	case 0x02:
		return "connection not allowed by ruleset"
	case 0x03:
		return "network unreachable"
	case 0x04:
		return "host unreachable"
	case 0x05:
		return "connection refused"
	case 0x06:
		return "TTL expired"
	case 0x07:
		return "command not supported"
	case 0x08:
		return "address type not supported"
	default:
		return "unknown code: " + strconv.Itoa(int(code))
	}
}

// Wire protocol constants.
const (
	Version5 = 0x05

	AddrTypeIPv4 = 0x01
	AddrTypeFQDN = 0x03
	AddrTypeIPv6 = 0x04

	CmdConnect Command = 0x01 // establishes an active-open forward proxy connection
	cmdBind    Command = 0x02 // establishes a passive-open forward proxy connection

	AuthMethodNotRequired         AuthMethod = 0x00 // no authentication required
	AuthMethodUsernamePassword    AuthMethod = 0x02 // use username/password
	AuthMethodNoAcceptableMethods AuthMethod = 0xff // no acceptable authentication methods

	AuthUsernamePasswordVersion       = 0x01
	StatusSucceeded             Reply = 0x00
)

// An Addr represents a SOCKS-specific address.
// Either Name or IP is used exclusively.
type Addr struct {
	Name string // fully-qualified domain name
	IP   net.IP
	Port int
}

func (a *Addr) Network() string { return "socks" }

func (a *Addr) String() string {
	if a == nil {
		return "<nil>"
	}
	port := strconv.Itoa(a.Port)
	if a.IP == nil {
		return net.JoinHostPort(a.Name, port)
	}
	return net.JoinHostPort(a.IP.String(), port)
}

const (
	authUsernamePasswordVersion = 0x01
	authStatusSucceeded         = 0x00
)

// UsernamePassword are the credentials for the username/password
// authentication method.
type UsernamePassword struct {
	Username string
	Password string
}

// Authenticate authenticates a pair of username and password with the
// proxy server.
func (up *UsernamePassword) Authenticate(ctx context.Context, rw io.ReadWriter, auth AuthMethod) error {
	switch auth {
	case AuthMethodNotRequired:
		return nil
	case AuthMethodUsernamePassword:
		if len(up.Username) == 0 || len(up.Username) > 255 || len(up.Password) == 0 || len(up.Password) > 255 {
			return errors.New("invalid username/password")
		}
		b := []byte{authUsernamePasswordVersion}
		b = append(b, byte(len(up.Username)))
		b = append(b, up.Username...)
		b = append(b, byte(len(up.Password)))
		b = append(b, up.Password...)
		// TODO(mikio): handle IO deadlines and cancelation if
		// necessary
		if _, err := rw.Write(b); err != nil {
			return err
		}
		if _, err := io.ReadFull(rw, b[:2]); err != nil {
			return err
		}
		if b[0] != authUsernamePasswordVersion {
			return errors.New("invalid username/password version")
		}
		if b[1] != authStatusSucceeded {
			return errors.New("username/password authentication failed")
		}
		return nil
	}
	return errors.New("unsupported authentication method " + strconv.Itoa(int(auth)))
}
