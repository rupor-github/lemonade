package lemon

import (
	"log"

	"github.com/atotto/clipboard"
)

// Clipboard is used by "lemonade" to rpc clipboard content.
type Clipboard struct {
	cli *CLI
}

// NewClipboard initializes Clipboard structure.
func NewClipboard(c *CLI) *Clipboard {
	return &Clipboard{
		cli: c,
	}
}

// Copy is implementation of "lemonade" rpc "copy" command.
func (c *Clipboard) Copy(text string, _ *struct{}) error {
	<-c.cli.ConnCh
	if c.cli.Debug {
		log.Printf("lemonade Copy request received len: %d", len(text))
	}
	// Logger instance needs to be passed here somehow?
	return clipboard.WriteAll(c.cli.ConvertLineEnding(text))
}

// Paste is implementation of "lemonade" rpc "paste" command.
func (c *Clipboard) Paste(_ struct{}, resp *string) error {
	<-c.cli.ConnCh
	t, err := clipboard.ReadAll()
	if c.cli.Debug {
		log.Printf("lemonade Paste request received len: %d, error: '%+v'", len(t), err)
	}
	*resp = t
	return err
}
