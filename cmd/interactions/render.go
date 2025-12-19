package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/interactions"
)

var _ Cmd = (*renderCmd)(nil)

type renderCmd struct {
	*RootCmd
	Flags *flag.FlagSet

	output string

	columns int

	SubCommands map[string]Cmd
}

func (c *renderCmd) Usage() {
	err := executeUsage(os.Stderr, "render_usage.txt", c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *renderCmd) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	err := c.Flags.Parse(args)
	if err != nil {
		return NewUserError(err, fmt.Sprintf("flag parse error %s", err.Error()))
	}
	interactions.Render(c.output, c.columns)
	return nil
}

func (c *RootCmd) NewrenderCmd() *renderCmd {
	set := flag.NewFlagSet("render", flag.ContinueOnError)
	v := &renderCmd{
		RootCmd:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}

	set.StringVar(&v.output, "output", "", "TODO: Add usage text")

	set.IntVar(&v.columns, "columns", 0, "TODO: Add usage text")

	set.Usage = v.Usage

	return v
}
