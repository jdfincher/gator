package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jdfincher/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f := c.commands[cmd.name]
	err := f(s, cmd)
	if err != nil {
		return fmt.Errorf("error: running %v command -> %w", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.commands[name] = f
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("error: login requires Username")
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("Username successfully set to %v\n", cmd.args[0])
	return nil
}

func newCommands() *commands {
	c := new(commands)
	c.commands = make(map[string]func(*state, command) error)
	return c
}

func main() {
	var err error
	state := new(state)
	state.cfg, err = config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	coms := newCommands()
	if err = coms.register("login", handlerLogin); err != nil {
		fmt.Printf("%v\n", err)
	}
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("error: missing command arguments, use gator 'command_name'\n")
		os.Exit(1)
	}
	usercom := command{
		name: strings.ToLower(args[1]),
		args: args[2:],
	}
	if err := coms.run(state, usercom); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
