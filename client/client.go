package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rupor-github/lemonade/lemon"
	"github.com/rupor-github/lemonade/param"
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

func getAddresses(port int, translateLoopbackIP, debug bool) (string, string) {

	portStr := ":" + strconv.Itoa(port)

	addrListen, addrSend := `localhost`+portStr, `http://localhost:%d/%s`
	if translateLoopbackIP {
		// open to the outside world - direct connection expected
		addrListen, addrSend = portStr, `http://127.0.0.1:%d/%s`
	} else {
		if addr := getSSHSessionAddr(); len(addr) != 0 {
			// if we run in SSH session - expect dynamic port forwarding
			addrListen, addrSend = portStr, `http://`+addr+`:%d/%s`
		} else if addr = getLocalHostIPv4(); len(addr) != 0 {
			// See if could derive extrnal IPv4 for our host
			addrListen, addrSend = portStr, `http://`+addr+`:%d/%s`
		}
	}
	if debug {
		log.Printf("serveFile listen address: '%s' send URL: '%s'", addrListen, addrSend)
	}
	return addrListen, addrSend
}

func getfileHandler(fname string, srv *http.Server, finished chan *http.Server, debug bool) func(w http.ResponseWriter, r *http.Request) {
	// NOTE: There is still a chance that serving actual file will be completed before any additional requests from the browser
	// generating ssh "channel X: open failed: connect failed: Connection refused" messages, especially when everything is slow
	// due to network congestion or excessive debug logging.
	return func(w http.ResponseWriter, r *http.Request) {

		if debug {
			log.Printf("Processing request '%s'", r.URL)
		}

		if filepath.Base(fname) != path.Base(r.URL.String()) {
			if debug {
				log.Print("Not serving...")
			}
			http.Error(w, "not serving...", http.StatusNotFound)
			return
		}

		// Kill caching
		for _, v := range etagHeaders {
			r.Header.Del(v)
		}
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		if debug {
			log.Printf("Transferring file '%s'", fname)
		}

		f, err := os.Open(fname)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if debug {
				log.Printf("Servng '%s' is completed", fname)
			}
			f.Close()
		}()

		http.ServeContent(w, r, fname, time.Unix(0, 0), f)
		finished <- srv
	}
}

// NOTE: we actuall need real server here - browsers like to ask for /favicon.ico etc. especially when ports are selected randomly and
// request url is changing. If not answered properly it will generate channel errors when ssh dynamic port forwarding is used.
func serveFile(fname string, port int, timeout time.Duration, translateLoopbackIP, debug bool) (string, <-chan *http.Server, error) {

	addrListen, addrSend := getAddresses(port, translateLoopbackIP, debug)

	l, err := net.Listen("tcp", addrListen)
	if err != nil {
		return "", nil, err
	}

	finished := make(chan *http.Server)

	srv := &http.Server{
		Addr:         addrListen,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	m := http.NewServeMux()
	m.HandleFunc("/", getfileHandler(fname, srv, finished, debug))
	srv.Handler = m

	go func() {
		if debug {
			log.Printf("Starting http server for '%s'", fname)
		}
		_ = srv.Serve(l)
	}()

	if port == 0 {
		port = l.Addr().(*net.TCPAddr).Port
	}
	return fmt.Sprintf(addrSend, port, fname), finished, nil
}

// Open implements client "open" command.
func Open(c *lemon.CLI) error {

	uri := c.DataSource
	if c.Debug {
		log.Printf("Client URI.Open '%s' to %s:%d", uri, c.Host, c.Port)
	}

	var finished <-chan *http.Server
	if c.TransLocalfile && fileExists(uri) {
		var err error
		uri, finished, err = serveFile(uri, c.TransFilePort, c.TransFileTimeout, c.TransLoopback, c.Debug)
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
		// First we wait for file to be served
		timer := time.NewTimer(c.TransFileTimeout)
		defer timer.Stop()
		select {
		case srv := <-finished:
			if c.Debug {
				log.Printf("Client URI.Open to %s:%d done", c.Host, c.Port)
			}
			// And then we try to end gracefully to avoid ssh channel complaints.
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.TransFileTimeout))
			_ = srv.Shutdown(ctx)
			cancel()
		case <-timer.C:
			log.Printf("Client URI.Open to %s:%d timeout waiting for file request", c.Host, c.Port)
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
