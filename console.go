package console

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

var defaultHistoryFile = filepath.Join(os.TempDir(), ".console_history")

type Opts func(*Console)

func WithExitCmd(e *Cmd) Opts {
	return func(c *Console) {
		c.exitCmd = e
	}
}

func WithContext(ctx context.Context) Opts {
	return func(c *Console) {
		c.parentCtx = ctx
	}
}

func WithHistoryFile(file string) Opts {
	return func(c *Console) {
		c.historyFile = file
	}
}

func WithWelcomeMsg(msg string) Opts {
	return func(c *Console) {
		c.welcomeMsg = msg
	}
}

func WithHandleCtrlC(handle bool) Opts {
	return func(c *Console) {
		c.liner.SetCtrlCAborts(handle)
	}
}

type Console struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	isOsPipe  bool

	liner       *liner.State
	historyFile string
	welcomeMsg  string

	cmds    []*Cmd
	exitCmd *Cmd
}

func New(opts ...Opts) (*Console, error) {
	c := &Console{
		parentCtx:   context.Background(),
		liner:       liner.NewLiner(),
		historyFile: defaultHistoryFile,
		exitCmd:     quitCmd,
		cmds:        defaultCmds,
	}
	c.liner.SetCtrlCAborts(true)

	// check if stdin is a pipe
	if isPipe, err := fileIsPipe(os.Stdin); err != nil {
		return nil, fmt.Errorf("error checking if stdin is a pipe: %s", err)
	} else if isPipe {
		c.isOsPipe = true
	}

	for _, opt := range opts {
		opt(c)
	}

	ctx, cancel := context.WithCancel(c.parentCtx)
	c.ctx = ctx
	c.cancel = cancel

	c.setCompleter()

	return c, nil
}

func fileIsPipe(in *os.File) (bool, error) {
	if fi, _ := in.Stat(); (fi.Mode() & os.ModeNamedPipe) != 0 {
		return true, nil
	}
	return false, nil
}

func (c *Console) RegisterCommands(cmds ...*Cmd) error {
	for _, cmd := range cmds {
		if c.checkCmdRegistered(cmd) {
			return errors.New("command matches an existing command")
		}
		cmd.Console = c
		c.cmds = append(c.cmds, cmd)
	}
	return nil
}

func (c *Console) checkCmdRegistered(cmd *Cmd) bool {
	for _, n := range c.cmds {
		for _, v := range append(n.Aliases, n.Name) {
			if v == cmd.Name {
				return true
			}
			for _, a := range cmd.Aliases {
				if a == v {
					return true
				}
			}
		}
	}
	return false
}

func (c *Console) Start() error {
	if !c.isOsPipe {
		c.printWelcomeMsg()
	}
	c.readHistory()
	return c.read()
}

func (c *Console) Close() error {
	c.cancel()
	c.writeHistory()
	c.liner.Close()
	return nil
}

func (c *Console) ExitCmd() (*Cmd, bool) {
	return c.exitCmd, c.exitCmd != nil
}

func (c *Console) setCompleter() {
	c.liner.SetCompleter(func(line string) (s []string) {
		for _, n := range c.cmds {
			if strings.HasPrefix(n.Name, strings.ToLower(line)) {
				s = append(s, n.Name)
				continue
			}
			for _, a := range n.Aliases {
				if strings.HasPrefix(a, strings.ToLower(line)) {
					s = append(s, a)
				}
			}
		}
		return
	})
}

func (c *Console) printWelcomeMsg() {
	fmt.Println(c.welcomeMsg)
}

func (c *Console) read() error {
	doneC := make(chan struct{})
	go func() {
		defer close(doneC)
		for {
			if in, err := c.liner.Prompt("> "); err == nil {
				in = strings.TrimSpace(in)
				if in == "" {
					continue
				}
				c.liner.AppendHistory(in)
				if exit, err := c.handleInput(in); err != nil {
					fmt.Println(styleError.Render(err.Error()))
				} else if exit { // prevent an unnecessary newline
					break
				}
			} else if err == liner.ErrPromptAborted {
				fmt.Println("Aborted")
				break
			} else if err == io.EOF {
				break
			} else {
				fmt.Println(styleError.Render(fmt.Sprintf("Error reading line: %s", err)))
				break
			}
		}
	}()

	select {
	case <-doneC:
		return nil
	case <-c.ctx.Done():
		return nil
	}
}

func (c *Console) readHistory() {
	f, err := os.Open(c.historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Println(styleError.Render(fmt.Sprintf("Error opening history file: %s", err)))
	}
	defer f.Close()
	if _, err := c.liner.ReadHistory(f); err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error reading history file: %s", err)))
	}
}

func (c *Console) writeHistory() {
	f, err := os.Create(c.historyFile)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error creating history file: %s", err)))
	}
	defer f.Close()
	if _, err := c.liner.WriteHistory(f); err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error writing history file: %s", err)))
	}
}

func (c *Console) handleInput(input string) (exit bool, err error) {
	if e, ok := c.ExitCmd(); ok {
		if e.Match(input) {
			return true, e.Handle(input)
		}
	}
	for _, cmd := range c.cmds {
		if cmd.Match(input) {
			if err := cmd.Handle(input); err != nil {
				fmt.Println(styleError.Render(fmt.Sprintf("error running command %s: %s\n", cmd.Name, err)))
			}
			return false, nil
		}
	}
	return false, nil
}
