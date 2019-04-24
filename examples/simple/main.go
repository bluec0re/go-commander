package main

import (
	"fmt"

	commander "github.com/bluec0re/go-commander"
)

func printCmd(cr commander.Commander, c commander.Command, args []string) error {
	fmt.Printf("Called %s with %s\n", c.Cmd, args)
	return nil
}

func main() {
	c := commander.New(">>> ")
	c.AddCommand(commander.Command{
		Cmd:         "foo",
		RunCallback: printCmd,
	})
	c.AddCommand(commander.Command{
		Cmd:         "bar",
		RunCallback: printCmd,
	})
	c.AddCommand(commander.Command{
		Cmd:         "baz",
		RunCallback: printCmd,
	})
	c.Run()
}
