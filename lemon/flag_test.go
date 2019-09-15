package lemon

import (
	"reflect"
	"testing"
)

func TestCLIParse(t *testing.T) {
	assert := func(args []string, expected CLI) {

		c := New()

		if err := c.ParseFlags(args, true); err != nil {
			t.Fatal(err)
		}

		// Do not check those fields
		c.Flags = expected.Flags
		c.ConnCh = expected.ConnCh

		if !reflect.DeepEqual(expected, *c) {
			t.Errorf("Expected:\n %+v, but got\n %+v", expected, c)
		}
	}

	defaultPort := 2489
	defaultHost := "localhost"
	defaultAllow := "0.0.0.0/0,::/0"

	assert([]string{"xdg-open", "http://example.com"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		DataSource:     "http://example.com",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"/usr/bin/xdg-open", "http://example.com"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		DataSource:     "http://example.com",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"xdg-open"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"pbpaste", "--port", "1124"}, CLI{
		Cmd:            CmdPaste,
		Host:           defaultHost,
		Port:           1124,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"/usr/bin/pbpaste", "--port", "1124"}, CLI{
		Cmd:            CmdPaste,
		Host:           defaultHost,
		Port:           1124,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"pbcopy", "hogefuga"}, CLI{
		Cmd:            CmdCopy,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		DataSource:     "hogefuga",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"/usr/bin/pbcopy", "hogefuga"}, CLI{
		Cmd:            CmdCopy,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		DataSource:     "hogefuga",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "--host", "192.168.0.1", "--port", "1124", "open", "http://example.com"}, CLI{
		Cmd:            CmdOpen,
		Host:           "192.168.0.1",
		Port:           1124,
		Allow:          defaultAllow,
		DataSource:     "http://example.com",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "copy", "hogefuga"}, CLI{
		Cmd:            CmdCopy,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		DataSource:     "hogefuga",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "paste"}, CLI{
		Cmd:            CmdPaste,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "--allow", "192.168.0.0/24", "server", "--port", "1124"}, CLI{
		Cmd:            CmdServer,
		Host:           defaultHost,
		Port:           1124,
		Allow:          "192.168.0.0/24",
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "open", "--trans-loopback=false"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  false,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "open", "--trans-loopback=true"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})

	assert([]string{"lemonade", "open", "--trans-localfile=false"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: false,
	})

	assert([]string{"lemonade", "open", "--trans-localfile=true"}, CLI{
		Cmd:            CmdOpen,
		Host:           defaultHost,
		Port:           defaultPort,
		Allow:          defaultAllow,
		TransLoopback:  true,
		TransLocalfile: true,
	})
}
