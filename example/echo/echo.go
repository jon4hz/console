package main

import (
	"fmt"
	"log"
	"os/user"
	"strings"

	"github.com/jon4hz/console"
)

func main() {
	u, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	c, err := console.New(
		console.WithWelcomeMsg(fmt.Sprintf("Hello %s!", u.Username)),
		console.WithHandleCtrlC(true),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()
	if err := c.RegisterCommands(echoCmd); err != nil {
		log.Fatalln(err)
	}

	if err := c.Start(); err != nil {
		log.Fatalln(err)
	}
}

var echoCmd = &console.Cmd{
	Name:        "echo",
	Description: "echo",
	Handler: func(c *console.Console, args []string) error {
		fmt.Println(strings.Join(args, " "))
		return nil
	},
}
