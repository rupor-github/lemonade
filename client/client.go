package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"lemonade/lemon"
	"lemonade/param"
)

var dummy = &struct{}{}

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

func serveFile(fname string) (string, <-chan struct{}, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, err
	}
	finished := make(chan struct{})

	go func() {

		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadFile(fname)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(b)

			w.(http.Flusher).Flush()
			finished <- struct{}{}
		}))
	}()

	return fmt.Sprintf("http://127.0.0.1:%d/%s", l.Addr().(*net.TCPAddr).Port, fname), finished, nil
}

// Open implements client "open" command.
func Open(c *lemon.CLI) error {

	uri := c.DataSource

	var finished <-chan struct{}
	if c.TransLocalfile && fileExists(uri) {
		var err error
		uri, finished, err = serveFile(uri)
		if err != nil {
			return err
		}
	}

	err := c.ProcessRPC(func(rc *rpc.Client) error {
		p := &param.OpenParam{
			URI:           uri,
			TransLoopback: c.TransLoopback || c.TransLocalfile,
		}
		if c.Debug {
			log.Printf("Client URI.Open rpc call to %s:%d with '%+v'", c.Host, c.Port, *p)
		}
		return rc.Call("URI.Open", p, dummy)
	})
	if err != nil {
		return err
	}

	if finished != nil {
		<-finished
	}
	return nil
}

// Paste implements client "paste" command.
func Paste(c *lemon.CLI) (string, error) {

	var resp string

	err := c.ProcessRPC(func(rc *rpc.Client) (rer error) {
		if c.Debug {
			log.Printf("Client Clipboard.Paste to %s:%d", c.Host, c.Port)
		}
		defer func() {
			if c.Debug {
				if rer == nil {
					log.Printf("Client Clipboard.Paste received %d length", len(resp))
				} else {
					log.Printf("Client Clipboard.Paste received error: '%s'", rer.Error())
				}
			}
		}()
		return rc.Call("Clipboard.Paste", dummy, &resp)
	})
	if err != nil {
		return "", err
	}
	return c.ConvertLineEnding(resp), nil
}

// Copy implements client "copy" command.
func Copy(c *lemon.CLI) error {

	text := c.DataSource

	return c.ProcessRPC(func(rc *rpc.Client) (rer error) {
		if c.Debug {
			log.Printf("Client Clipboard.Copy to %s:%d - %d length", c.Host, c.Port, len(text))
		}
		defer func() {
			if c.Debug && rer != nil {
				log.Printf("Client Clipboard.Copy received error: '%s'", rer.Error())
			}
		}()
		return rc.Call("Clipboard.Copy", text, dummy)
	})
}
