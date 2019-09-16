package client

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"

	"lemonade/lemon"
	"lemonade/param"
)

var (
	dummy          = &struct{}{}
	noCacheHeaders = map[string]string{
		"Expires":         time.Unix(0, 0).Format(time.RFC1123),
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}
	etagHeaders = []string{
		"ETag",
		"If-Modified-Since",
		"If-Match",
		"If-None-Match",
		"If-Range",
		"If-Unmodified-Since",
	}
)

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

func getSSHSessionAddr() string {

	ssh := os.Getenv("SSH_CONNECTION")

	if len(ssh) == 0 {
		return ""
	}

	const (
		rHost = iota //nolint:unused
		rPort        //nolint:unused
		lHost        //
		lPort        //nolint:unused
		maxPart
	)

	parts := strings.Split(ssh, " ")
	if len(parts) < maxPart {
		return ""
	}
	return parts[lHost]
}

func getLocalHostIPv4() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func serveFile(fname string, translateLoopbackIP, debug bool) (string, <-chan struct{}, error) {

	const anywhere = `:0`

	addrListen, addrSend := `localhost:0`, `http://localhost:%d/%s`
	if translateLoopbackIP {
		// open to the outside world - direct connection expected
		addrListen, addrSend = anywhere, `http://127.0.0.1:%d/%s`
	} else {
		if addr := getSSHSessionAddr(); len(addr) != 0 {
			// if we run in SSH session - expect dynamic port forwarding
			addrListen, addrSend = anywhere, `http://`+addr+`:%d/%s`
		} else if addr = getLocalHostIPv4(); len(addr) != 0 {
			// See if could derive extrnal IPv4 for our host
			addrListen, addrSend = anywhere, `http://`+addr+`:%d/%s`
		}
	}
	if debug {
		log.Printf("Serving addresses - listen: '%s' send: '%s'", addrListen, addrSend)
	}

	l, err := net.Listen("tcp", addrListen)
	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})
	go func() {

		if debug {
			log.Printf("Serving '%s' on %s", fname, l.Addr())
		}

		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Kill caching
			for _, v := range etagHeaders {
				r.Header.Del(v)
			}
			for k, v := range noCacheHeaders {
				w.Header().Set(k, v)
			}

			f, err := os.Open(fname)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()

			http.ServeContent(w, r, fname, time.Unix(0, 0), f)
			if wf, ok := w.(http.Flusher); ok {
				wf.Flush()
			}

			finished <- struct{}{}
		}))
	}()

	return fmt.Sprintf(addrSend, l.Addr().(*net.TCPAddr).Port, fname), finished, nil
}

// Open implements client "open" command.
func Open(c *lemon.CLI) error {

	uri := c.DataSource
	if c.Debug {
		log.Printf("Client URI.Open '%s' to %s:%d", uri, c.Host, c.Port)
	}

	var finished <-chan struct{}
	if c.TransLocalfile && fileExists(uri) {
		var err error
		uri, finished, err = serveFile(uri, c.TransLoopback, c.Debug)
		if err != nil {
			return err
		}
	}

	err := c.ProcessRPC(func(rc *rpc.Client) error {
		p := &param.OpenParam{
			URI:           uri,
			TransLoopback: c.TransLoopback,
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

		timer := time.NewTimer(c.TransFileTimeout)
		defer timer.Stop()

		select {
		case <-finished:
			if c.Debug {
				log.Printf("Client URI.Open to %s:%d done", c.Host, c.Port)
			}
		case <-timer.C:
			log.Printf("Client URI.Open to %s:%d temeouted waiting for file request", c.Host, c.Port)
		}
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
