package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"lemonade/lemon"
)

// Serve starts "lemonade" server backend.
func Serve(c *lemon.CLI) error {

	uri := lemon.NewURI(c)
	if err := rpc.Register(uri); err != nil {
		return fmt.Errorf("unable to register URI rpc: %w", err)
	}
	clip := lemon.NewClipboard(c)
	if err := rpc.Register(clip); err != nil {
		return fmt.Errorf("unable to register Clipboard rpc: %w", err)
	}
	ra, err := lemon.NewRange(c.Allow)
	if err != nil {
		return fmt.Errorf("unable to process allowed IP ranges: %w", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return fmt.Errorf("ResolveTCPAddr error: '%w'", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("ListenTCP error: '%w'", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("lemonade server Accept error: '%w'", err)
		}
		if c.Debug {
			log.Printf("lemonade server request from '%s'", conn.RemoteAddr())
		}
		go func(conn net.Conn) {
			defer conn.Close()

			if c.Debug {
				log.Printf("lemonade server request from '%s'", conn.RemoteAddr())
			}
			if ra.IsConnIn(conn) {
				c.ConnCh <- conn
				rpc.ServeConn(conn)
				if c.Debug {
					log.Printf("lemonade server done with '%s'", conn.RemoteAddr())
				}
			}
		}(conn)
	}
}
