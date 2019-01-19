package tunnel

import "net"

// A Handler of Conn
type Handler interface {
	Serve(conn net.Conn)
}

// For quickly defining Handler
type HandlerFunc func(conn net.Conn)

// Implement Handler
func (f HandlerFunc) Serve(conn net.Conn) {
	f(conn)
}

type Server struct {
	Addr    string
	Handler Handler
}

func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.Handler.Serve(conn)
	}
}
