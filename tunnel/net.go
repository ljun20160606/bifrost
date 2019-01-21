package tunnel

import (
	"net"
	"time"
)

func SetConnectTimeout(conn net.Conn, readTimeout, writeTimeout time.Duration) net.Conn {
	return &timeoutedNetConn{
		Conn:         conn,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

type timeoutedNetConn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (t *timeoutedNetConn) Read(b []byte) (n int, err error) {
	_ = t.SetDeadline(time.Now().Add(time.Duration(t.ReadTimeout)))
	i, err := t.Conn.Read(b)
	_ = t.SetDeadline(time.Time{})
	return i, err
}

func (t *timeoutedNetConn) Write(b []byte) (n int, err error) {
	_ = t.SetDeadline(time.Now().Add(time.Duration(t.WriteTimeout)))
	i, err := t.Conn.Write(b)
	_ = t.SetDeadline(time.Time{})
	return i, err
}
