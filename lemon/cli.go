package lemon

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"lemonade/misc"
)

// LastGitCommit hold git hash from build.
var LastGitCommit string

// Command enum defines what we are executing.
type Command int

// Commands
const (
	CmdOpen Command = iota + 1
	CmdCopy
	CmdPaste
	CmdServer
)

// CLI holds program state.
type CLI struct {
	Cmd        Command
	DataSource string

	// option flags
	Port             int
	Allow            string
	Host             string
	TransLoopback    bool
	TransLocalfile   bool
	TransFileTimeout time.Duration
	TransFilePort    int
	LineEnding       string
	Help             bool
	Debug            bool
	// and our flagset
	Flags *flag.FlagSet

	// used by our server,
	ConnCh chan net.Conn
}

// New initializes environment.
func New() *CLI {

	c := &CLI{
		ConnCh: make(chan net.Conn, 1),
		Flags:  flag.NewFlagSet("lemonade", flag.ContinueOnError),
	}

	c.Flags.BoolVar(&c.Help, "help", false, "Show this message")
	c.Flags.IntVar(&c.Port, "port", 2489, "TCP port number")
	c.Flags.StringVar(&c.Allow, "allow", "0.0.0.0/0,::/0", "Allow IP range [server only]")
	c.Flags.StringVar(&c.Host, "host", "localhost", "Destination host name [client only]")
	c.Flags.StringVar(&c.LineEnding, "line-ending", "", "Convert Line Endings (LF/CRLF)")
	c.Flags.BoolVar(&c.TransLoopback, "trans-loopback", true, "Replace loopback address [open command only]")
	c.Flags.BoolVar(&c.TransLocalfile, "trans-localfile", true, "Transfer local file [open command only]")
	c.Flags.IntVar(&c.TransFilePort, "trans-localfile-port", 2490, "Port to listen on transfer local file [open command only]")
	c.Flags.DurationVar(&c.TransFileTimeout, "trans-localfile-timeout", time.Second, "How long to wait for local file transfer request [open command only]")
	c.Flags.BoolVar(&c.Debug, "debug", false, "Print verbose debugging information")

	c.Flags.Usage = func() {
		var buf strings.Builder
		c.Flags.SetOutput(&buf)
		fmt.Fprintf(&buf, `
Lemonade - copy, paste and open browser over TCP

Version:
	%s (%s) %s
`, misc.GetVersion(), runtime.Version(), LastGitCommit)

		fmt.Fprintf(&buf, `
Usage:
	lemonade [options]... COMMAND [arg]

Commands:

	copy 'text'	 - send text to server clipboard
	paste		 - output server clipboard locally
	open 'url'	 - open url in server's default browser
	server		 - start server

Options:

`)
		c.Flags.PrintDefaults()
		fmt.Fprint(os.Stderr, buf.String())
	}
	return c
}

// ProcessRPC makes RPC call.
func (c *CLI) ProcessRPC(f func(*rpc.Client) error) error {
	rc, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}
	// Do not leak connections
	defer rc.Close()

	if err = f(rc); err != nil {
		return err
	}
	return nil
}

// ConvertLineEnding is used to normaliza line endings when exchanging clipboard content.
func (c *CLI) ConvertLineEnding(text string) string {
	switch {
	case strings.EqualFold("lf", c.LineEnding):
		text = strings.Replace(text, "\r\n", "\n", -1)
		return strings.Replace(text, "\r", "\n", -1)
	case strings.EqualFold("crlf", c.LineEnding):
		text = regexp.MustCompile(`\r(.)|\r$`).ReplaceAllString(text, "\r\n$1")
		text = regexp.MustCompile(`([^\r])\n|^\n`).ReplaceAllString(text, "$1\r\n")
		return text
	default:
		return text
	}
}
