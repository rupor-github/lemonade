package lemon

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/mitchellh/go-homedir"
	"github.com/monochromegane/conflag"
)

// ParseFlags prepares and parses carguments.
// NOTE: function parameters are used in tests.
func (c *CLI) ParseFlags(args []string, skip bool) error {
	if aliased, err := c.getCommand(args); err != nil {
		return err
	} else if !aliased {
		args = args[:len(args)-1]
	}
	return c.parse(args, skip)
}

func (c *CLI) getCommand(args []string) (bool, error) {

	aliased := true
	switch {
	case regexp.MustCompile(`/?xdg-open$`).MatchString(args[0]):
		c.Cmd = CmdOpen
		return aliased, nil
	case regexp.MustCompile(`/?pbpaste$`).MatchString(args[0]):
		c.Cmd = CmdPaste
		return aliased, nil
	case regexp.MustCompile(`/?pbcopy$`).MatchString(args[0]):
		c.Cmd = CmdCopy
		return aliased, nil
	default:
		aliased = false
	}

	del := func(i int) {
		copy(args[i+1:], args[i+2:])
		args[len(args)-1] = ""
	}

	for i, v := range args[1:] {
		switch v {
		case "open":
			c.Cmd = CmdOpen
			del(i)
			return aliased, nil
		case "paste":
			c.Cmd = CmdPaste
			del(i)
			return aliased, nil
		case "copy":
			c.Cmd = CmdCopy
			del(i)
			return aliased, nil
		case "server":
			c.Cmd = CmdServer
			del(i)
			return aliased, nil
		}
	}

	c.Flags.Usage()

	return aliased, errors.New("unknown subcommand")
}

func (c *CLI) parse(args []string, skip bool) error {

	confPath, err := homedir.Expand("~/.config/lemonade.toml")
	if err == nil && !skip {
		if confArgs, err := conflag.ArgsFrom(confPath); err == nil {
			_ = c.Flags.Parse(confArgs)
		}
	}

	var arg string
	err = c.Flags.Parse(args[1:])
	if err != nil {
		return err
	}
	if c.Cmd == CmdPaste || c.Cmd == CmdServer {
		return nil
	}

	for 0 < c.Flags.NArg() {
		arg = c.Flags.Arg(0)
		err := c.Flags.Parse(c.Flags.Args()[1:])
		if err != nil {
			return err
		}

	}

	if c.Help {
		return nil
	}

	if arg != "" {
		c.DataSource = arg
	} else {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		c.DataSource = string(b)
	}

	return nil
}
