package console_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jon4hz/console"
	"github.com/stretchr/testify/assert"
)

func TestEmptyConsole(t *testing.T) {
	c, err := console.New()
	assert.NoError(t, err)
	defer c.Close()
}

var echoCmd = &console.Cmd{
	Name:        "echo",
	Description: "echo",
}

func TestAddingEchoCmd(t *testing.T) {
	c, err := console.New()
	assert.NoError(t, err)
	defer c.Close()
	err = c.RegisterCommands(echoCmd)
	assert.NoError(t, err)
}

func TestAddingCmdTwice(t *testing.T) {
	c, err := console.New()
	assert.NoError(t, err)
	defer c.Close()
	err = c.RegisterCommands(echoCmd, echoCmd)
	assert.Error(t, err)
}

func TestEchoCmdWithoutHandler(t *testing.T) {
	c, err := console.New()
	assert.NoError(t, err)
	defer c.Close()

	err = c.RegisterCommands(echoCmd)
	assert.NoError(t, err)

	err = echoCmd.Handle("echo")
	assert.ErrorIs(t, err, console.ErrCmdNoHandler)
}

func TestEchoCmdWithHandler(t *testing.T) {
	c, err := console.New()
	assert.NoError(t, err)
	defer c.Close()

	err = c.RegisterCommands(echoCmd)
	assert.NoError(t, err)

	echoCmd.Handler = func(c *console.Console, args []string) error {
		fmt.Println(strings.Join(args, " "))
		return nil
	}
	err = echoCmd.Handle("echo")
	assert.NoError(t, err)
}

func TestMatchEchoCmd(t *testing.T) {
	assert.True(t, echoCmd.Match("echo"))
	assert.True(t, echoCmd.Match("echo test"))
}

func TestMismatchEchoCmd(t *testing.T) {
	assert.False(t, echoCmd.Match("foo"))
	assert.False(t, echoCmd.Match("foo test"))
}
