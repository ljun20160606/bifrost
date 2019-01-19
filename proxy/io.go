package proxy

import (
	"io"
)

func Copy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	select {
	case errCh <- err:
	default:
	}
}

func Transport(conn1 io.ReadWriter, conn2 io.ReadWriter) error {
	err := make(chan error)
	go Copy(conn1, conn2, err)
	go Copy(conn2, conn1, err)
	return <-err
}
