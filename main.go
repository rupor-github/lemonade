package main

import (
	"fmt"
	"log"
	"os"

	"lemonade/client"
	"lemonade/lemon"
	"lemonade/server"
)

// Status codes to be set by os.Exit()
const (
	exitSuccess = iota
	_
	_
	_
	_
	_
	exitFlagParseError
	exitRPCError
	exitHelp
)

// os.Exit() prevents defers from proper cleanup.
func run() int {

	cli := lemon.New()

	if err := cli.ParseFlags(os.Args, false); err != nil {
		fmt.Fprintf(os.Stderr, "\n\n*** ERROR: %s\n", err.Error())
		return exitFlagParseError
	}

	if cli.Help {
		cli.Flags.Usage()
		return exitHelp
	}

	var err error
	switch cli.Cmd {
	case lemon.CmdOpen:
		err = client.Open(cli)
	case lemon.CmdCopy:
		err = client.Copy(cli)
	case lemon.CmdPaste:
		var text string
		text, err = client.Paste(cli)
		os.Stdout.Write([]byte(text))
	case lemon.CmdServer:
		err = server.Serve(cli)
	default:
		panic("Unreachable code")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\n*** ERROR: %s\n", err.Error())
		return exitRPCError
	}
	return exitSuccess
}

func main() {
	log.SetPrefix("[LMND] ")
	log.SetFlags(0)
	os.Exit(run())
}
