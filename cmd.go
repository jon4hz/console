package console

import (
	"errors"
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

var (
	ErrCmdNoHandler = errors.New("command has no handler")
)
var defaultCmds = []*Cmd{
	helpCmd,
	clearCmd,
}

type Cmd struct {
	Name                 string
	Aliases              []string
	Description          string
	IgnorePipe           bool
	Matcher              func(cmd string) bool
	IgnoreDefaultMatcher bool
	Handler              func(c *Console, cmd string) error
	Console              *Console
}

func (c *Cmd) defaultMatcher(cmd string) bool {
	cmd, _ = splitCmdArgs(cmd)
	if cmd == c.Name {
		return true
	}
	for _, alias := range c.Aliases {
		if cmd == alias {
			return true
		}
	}
	return false
}

func splitCmdArgs(cmd string) (string, []string) {
	args := strings.Split(cmd, " ")
	return args[0], args[1:]
}

func (c *Cmd) Match(cmd string) bool {
	if c.defaultMatcher(cmd) && !c.IgnoreDefaultMatcher {
		return true
	}
	if c.Matcher != nil {
		return c.Matcher(cmd)
	}
	return false
}

func (c *Cmd) Handle(cmd string) error {
	if c.Console.isOsPipe && c.IgnorePipe {
		return nil
	}
	if c.Handler != nil {
		return c.Handler(c.Console, cmd)
	}
	return ErrCmdNoHandler
}

var helpCmd = &Cmd{
	Name:        "help",
	Description: "Show the help",
	Handler: func(c *Console, cmd string) error {
		fmt.Println(helpView(c))
		return nil
	},
}

func helpView(c *Console) string {
	s := "Available commands:"
	for _, cmd := range c.cmds {
		if cmd.Name != "" && cmd.Description != "" {
			s += fmt.Sprintf("\n  %s - %s", cmd.Name, cmd.Description)
		}
	}
	if c.exitCmd != nil {
		s += fmt.Sprintf("\n  %s - Exit the console", c.exitCmd.Name)
	}
	return s
}

var quitCmd = &Cmd{
	Name:        "quit",
	Aliases:     []string{"exit"},
	Description: "Quit the console",
	IgnorePipe:  true,
	Handler: func(c *Console, cmd string) error {
		c.Close()
		return nil
	},
}

var clearCmd = &Cmd{
	Name:        "clear",
	Description: "Clear the screen",
	IgnorePipe:  true,
	Handler: func(c *Console, cmd string) error {
		termenv.ClearScreen()
		return nil
	},
}
